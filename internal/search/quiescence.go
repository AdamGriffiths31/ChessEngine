// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	evalpkg "github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// quiescence performs quiescence search to avoid the horizon effect.
// Only searches "quiet" positions by examining tactical sequences (captures and checks)
// until a stable position is reached. This prevents the engine from making decisions
// based on incomplete tactical analysis at the end of main search depth.
//
//nolint:gocyclo // refactored in Phase 4
func (m *MinimaxEngine) quiescence(ctx context.Context, alpha, beta evalpkg.EvaluationScore, ply int) evalpkg.EvaluationScore {
	b := m.searchState.board
	player := m.searchState.player

	m.searchState.searchStats.NodesSearched++
	m.searchState.searchStats.QNodes++

	// Cancellation cadence: poll ctx once per ~1024 node entries. NodesSearched
	// (shared with negamax and just incremented above) firing the &1023==0 mask
	// samples the combined node stream roughly every 1024 nodes. See negamax.go
	// for the full rationale; behavior on Done is identical to the former
	// per-node select.
	if m.searchState.searchStats.NodesSearched&1023 == 0 {
		select {
		case <-ctx.Done():
			m.searchState.searchCancelled = true
			return alpha
		default:
		}
	}

	hash := b.GetHash()
	if m.transpositionTable != nil {
		m.searchState.searchStats.TTProbes++
		if entry, found := m.transpositionTable.Probe(hash); found {
			m.searchState.searchStats.TTHits++
			if entry.GetDepth() >= 0 {
				ttScore := scoreFromTT(entry.Score, ply)
				switch entry.GetType() {
				case EntryExact:
					return ttScore
				case EntryLowerBound:
					if ttScore >= beta {
						return ttScore
					}
					if ttScore > alpha {
						alpha = ttScore
					}
				case EntryUpperBound:
					if ttScore <= alpha {
						return ttScore
					}
					if ttScore < beta {
						beta = ttScore
					}
				}
			}
		}
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	originalAlpha := alpha

	eval := m.evaluator.Evaluate(b)
	if player == movegen.Black {
		eval = -eval
	}

	if !inCheck {
		if eval >= beta {
			return beta
		}
		if eval > alpha {
			alpha = eval
		}
	}

	var movesToSearch *movegen.MoveList

	if inCheck {
		// When in check, we must consider ALL legal moves (including quiet escapes)
		// This is the idiomatic approach used by strong chess engines
		movesToSearch = m.generator.GeneratePseudoLegalMoves(b, player)
	} else {
		// Normal quiescence - only captures and promotions
		movesToSearch = m.generator.GenerateCaptures(b, player)
	}

	defer movegen.ReleaseMoveList(movesToSearch)

	// Use ply as buffer index for buffer selection
	bufferPly := ply
	if bufferPly >= MaxKillerDepth {
		bufferPly = MaxKillerDepth - 1
	}

	if inCheck {
		m.scoreMoves(b, movesToSearch, bufferPly, board.Move{})
	} else {
		m.scoreCaptures(movesToSearch, bufferPly)
	}

	legalMoveCount := 0
	bestScore := eval

	for i := 0; i < movesToSearch.Count; i++ {
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

		// Apply pruning only when not in check and only for captures
		if !inCheck && move.IsCapture {
			// Delta pruning - skip captures that can't improve alpha significantly
			captureValue := evalpkg.EvaluationScore(0)
			switch move.Captured {
			case board.WhitePawn, board.BlackPawn:
				captureValue = 100
			case board.WhiteKnight, board.BlackKnight, board.WhiteBishop, board.BlackBishop:
				captureValue = 300
			case board.WhiteRook, board.BlackRook:
				captureValue = 500
			case board.WhiteQueen, board.BlackQueen:
				captureValue = 900
			}

			if move.Promotion != board.Empty {
				captureValue += 800
			}

			margin := evalpkg.EvaluationScore(200)
			if eval+captureValue+margin < alpha {
				m.searchState.searchStats.DeltaPruned++
				b.UnmakeMove(undo)
				continue
			}

			// SEE pruning - skip bad captures
			if seeScore := m.seeCalculator.SEE(b, move); seeScore < -100 {
				b.UnmakeMove(undo)
				continue
			}
		}

		// Move is already made and verified as legal above
		m.searchState.player = oppositePlayer(player)
		score := -m.quiescence(ctx, -beta, -alpha, ply+1)
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

	if inCheck && legalMoveCount == 0 {
		// Checkmate - no legal moves when in check
		// The mate distance should be how many plies from the original search root
		// In quiescence, ply represents the distance from the search root
		return -evalpkg.MateScore + evalpkg.EvaluationScore(ply)
	}

	// If not in check and we had no moves to search (no captures), return static eval
	if !inCheck && legalMoveCount == 0 {
		return eval
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		var entryType EntryType
		if bestScore <= originalAlpha {
			entryType = EntryUpperBound
		} else if bestScore >= beta {
			entryType = EntryLowerBound
		} else {
			entryType = EntryExact
		}
		m.transpositionTable.Store(hash, 0, scoreToTT(bestScore, ply), entryType, board.Move{})
	}

	return bestScore
}

// oppositePlayer returns the opposite player
func oppositePlayer(player movegen.Player) movegen.Player {
	if player == movegen.White {
		return movegen.Black
	}
	return movegen.White
}
