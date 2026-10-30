package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	evalpkg "github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

func (m *MinimaxEngine) quiescence(ctx context.Context, player movegen.Player, alpha, beta evalpkg.EvaluationScore, ply int) evalpkg.EvaluationScore {
	b := m.searchState.board

	m.searchState.searchStats.NodesSearched++
	m.searchState.searchStats.QNodes++

	if m.pollCancel(ctx) {
		return alpha
	}

	hash := b.GetHash()
	score, alpha, beta, done := m.probeQuiescenceTT(hash, ply, alpha, beta)
	if done {
		return score
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	originalAlpha := alpha

	standPat := sideRelative(m.evaluator.Evaluate(b), player)

	if !inCheck {
		if standPat >= beta {
			return beta
		}
		if standPat > alpha {
			alpha = standPat
		}
	}

	var movesToSearch *movegen.MoveList
	if inCheck {
		movesToSearch = m.generator.GeneratePseudoLegalMoves(b, player)
	} else {
		movesToSearch = m.generator.GenerateCaptures(b, player)
	}
	defer movegen.ReleaseMoveList(movesToSearch)

	bufferPly := min(ply, MaxKillerDepth-1)

	if inCheck {
		m.scoreMoves(b, movesToSearch, bufferPly, board.Move{})
	} else {
		m.scoreCaptures(movesToSearch, bufferPly)
	}

	legalMoveCount := 0
	bestScore := standPat
	if inCheck {
		bestScore = -evalpkg.MateScore - 1
	}

	for i := range movesToSearch.Count {
		m.pickNextMove(movesToSearch, i, bufferPly)
		move := movesToSearch.Moves[i]

		if m.searchState.searchCancelled {
			break
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
	if entry.GetDepth() < 0 {
		return 0, alpha, beta, false
	}

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

func (m *MinimaxEngine) shouldPruneCapture(b *board.Board, move board.Move, standPat, alpha evalpkg.EvaluationScore) bool {
	gain := deltaCaptureValue(move.Captured)
	if move.HasPromotion() {
		gain += QDeltaPromoBonus
	}
	if standPat+gain+QDeltaMargin < alpha {
		m.searchState.searchStats.DeltaPruned++
		return true
	}
	return m.seeCalculator.SEE(b, move) < QSEEPruneThreshold
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
