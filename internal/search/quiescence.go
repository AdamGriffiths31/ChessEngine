package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	evalpkg "github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

func (m *MinimaxEngine) enterQuiescenceNode(ctx context.Context) bool {
	m.searchState.searchStats.NodesSearched++
	m.searchState.searchStats.QNodes++

	return m.pollCancel(ctx)
}

func (m *MinimaxEngine) quiescence(ctx context.Context, player movegen.Player, alpha, beta evalpkg.EvaluationScore, ply int) evalpkg.EvaluationScore {
	b := m.searchState.board

	if m.enterQuiescenceNode(ctx) {
		return alpha
	}

	hash := b.GetHash()
	score, alpha, beta, done := m.probeQuiescenceTT(hash, ply, alpha, beta)
	if done {
		return score
	}

	return m.quiescenceBody(ctx, player, alpha, beta, ply, hash, m.generator.IsKingInCheck(b, player))
}

func (m *MinimaxEngine) quiescenceBody(ctx context.Context, player movegen.Player, alpha, beta evalpkg.EvaluationScore, ply int, hash uint64, inCheck bool) evalpkg.EvaluationScore {
	b := m.searchState.board
	if ply >= MaxKillerDepth-1 {
		return sideRelative(m.evaluator.Evaluate(b), player)
	}

	originalAlpha := alpha

	var standPat evalpkg.EvaluationScore
	var movesToSearch *movegen.MoveList
	bestScore := -scoreInfinity
	evasions := allSquares

	if inCheck {
		movesToSearch = m.generator.GeneratePseudoLegalMoves(b, player)
		m.scoreMoves(b, movesToSearch, ply, board.Move{})
		evasions = m.generator.CheckEvasionMask(b, player)
	} else {
		standPat = sideRelative(m.evaluator.Evaluate(b), player)
		if standPat >= beta {
			return beta
		}
		alpha = max(alpha, standPat)
		bestScore = standPat

		movesToSearch = m.generator.GenerateCaptures(b, player)
		m.scoreCaptures(movesToSearch, ply)
	}
	defer movegen.ReleaseMoveList(movesToSearch)

	legalMoveCount := 0

	for i := range movesToSearch.Count {
		m.pickNextMove(movesToSearch, i, ply)
		move := movesToSearch.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		if cannotEvade(&move, evasions) {
			continue
		}

		if legalMoveCount > 0 && !inCheck && move.IsCapture && m.isDeltaPruned(move, standPat, alpha) {
			continue
		}

		undo, ok := m.tryMove(move, player)
		if !ok {
			continue
		}

		legalMoveCount++

		if !inCheck && move.IsCapture && m.shouldPruneCapture(b, move, standPat, alpha) {
			b.UnmakeMove(undo)
			continue
		}

		score := -m.quiescence(ctx, oppositePlayer(player), -beta, -alpha, ply+1)
		b.UnmakeMove(undo)

		if score > bestScore {
			bestScore = score
		}

		if score > alpha {
			alpha = score
			if alpha >= beta {
				break
			}
		}
	}

	if legalMoveCount == 0 {
		if inCheck {
			return -evalpkg.MateScore + evalpkg.EvaluationScore(ply)
		}
		return standPat
	}

	m.storeQuiescenceResult(hash, ply, originalAlpha, beta, bestScore)
	return bestScore
}

func (m *MinimaxEngine) probeQuiescenceTT(hash uint64, ply int, alpha, beta evalpkg.EvaluationScore) (score, alphaOut, betaOut evalpkg.EvaluationScore, done bool) {
	if m.transpositionTable == nil {
		return 0, alpha, beta, false
	}
	m.searchState.searchStats.TTProbes++
	entry, found := m.transpositionTable.Probe(hash)
	if !found {
		return 0, alpha, beta, false
	}
	m.searchState.searchStats.TTHits++

	ttScore := scoreFromTT(entry.Score, ply)
	switch entry.GetType() {
	case EntryExact:
		return ttScore, alpha, beta, true
	case EntryLowerBound:
		if ttScore >= beta {
			return ttScore, alpha, beta, true
		}
		alpha = max(alpha, ttScore)
	case EntryUpperBound:
		if ttScore <= alpha {
			return ttScore, alpha, beta, true
		}
		beta = min(beta, ttScore)
	}
	return 0, alpha, beta, false
}

func (m *MinimaxEngine) isDeltaPruned(move board.Move, standPat, alpha evalpkg.EvaluationScore) bool {
	gain := deltaCaptureValue(move.Captured)
	if move.HasPromotion() {
		gain += QDeltaPromoBonus
	}
	if standPat+gain+QDeltaMargin < alpha {
		m.searchState.searchStats.DeltaPruned++
		return true
	}
	return false
}

func (m *MinimaxEngine) shouldPruneCapture(b *board.Board, move board.Move, standPat, alpha evalpkg.EvaluationScore) bool {
	return m.isDeltaPruned(move, standPat, alpha) || m.seeCalculator.SEE(b, move) < QSEEPruneThreshold
}

func deltaCaptureValue(victim board.Piece) evalpkg.EvaluationScore {
	switch victim {
	case board.WhitePawn, board.BlackPawn:
		return 100
	case board.WhiteKnight, board.BlackKnight, board.WhiteBishop, board.BlackBishop:
		return 300
	case board.WhiteRook, board.BlackRook:
		return 500
	case board.WhiteQueen, board.BlackQueen:
		return 900
	}
	return 0
}

func (m *MinimaxEngine) storeQuiescenceResult(hash uint64, ply int, originalAlpha, beta, bestScore evalpkg.EvaluationScore) {
	if m.transpositionTable == nil || m.searchState.searchCancelled {
		return
	}
	var entryType EntryType
	switch {
	case bestScore <= originalAlpha:
		entryType = EntryUpperBound
	case bestScore >= beta:
		entryType = EntryLowerBound
	default:
		entryType = EntryExact
	}
	m.transpositionTable.Store(hash, 0, scoreToTT(bestScore, ply), entryType, board.Move{})
}

func oppositePlayer(player movegen.Player) movegen.Player {
	if player == movegen.White {
		return movegen.Black
	}
	return movegen.White
}
