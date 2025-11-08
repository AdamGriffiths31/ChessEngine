// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// quiescence performs quiescence search to avoid the horizon effect.
// Only searches "quiet" positions by examining tactical sequences (captures and checks)
// until a stable position is reached. This prevents the engine from making decisions
// based on incomplete tactical analysis at the end of main search depth.
func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, depthFromRoot int, stats *ai.SearchStats) ai.EvaluationScore {
	m.searchState.searchStats.NodesSearched++
	m.searchState.searchStats.QNodes++

	select {
	case <-ctx.Done():
		m.searchState.searchCancelled = true
		return alpha
	default:
	}

	hash := b.GetHash()
	if m.transpositionTable != nil {
		m.searchState.searchStats.TTProbes++
		if entry, found := m.transpositionTable.Probe(hash); found {
			m.searchState.searchStats.TTHits++
			if entry.GetDepth() >= 0 {
				switch entry.GetType() {
				case EntryExact:
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						return entry.Score
					}
					if entry.Score > alpha {
						alpha = entry.Score
					}
				case EntryUpperBound:
					if entry.Score <= alpha {
						return entry.Score
					}
					if entry.Score < beta {
						beta = entry.Score
					}
				}
			}
		}
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	originalAlpha := alpha

	eval := m.evaluator.Evaluate(b)
	if player == moves.Black {
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

	// Generate moves based on whether we're in check
	var movesToSearch *moves.MoveList

	if inCheck {
		// When in check, we must consider ALL legal moves (including quiet escapes)
		// This is the idiomatic approach used by strong chess engines
		allMoves := m.generator.GeneratePseudoLegalMoves(b, player)
		movesToSearch = allMoves

	} else {
		// Normal quiescence - only captures and promotions
		allMoves := m.generator.GeneratePseudoLegalMoves(b, player)
		defer moves.ReleaseMoveList(allMoves)

		captureList := moves.GetMoveList()
		for i := 0; i < allMoves.Count; i++ {
			move := allMoves.Moves[i]
			if move.IsCapture || move.Promotion != board.Empty {
				captureList.AddMove(move)
			}
		}
		movesToSearch = captureList
	}

	defer moves.ReleaseMoveList(movesToSearch)

	// Order moves appropriately
	if inCheck {
		m.orderMoves(b, movesToSearch, 0, board.Move{}) // Order all moves when in check
	} else {
		m.orderCaptures(movesToSearch) // Order captures in normal quiescence
	}

	legalMoveCount := 0
	bestScore := eval

	for i := 0; i < movesToSearch.Count; i++ {
		move := movesToSearch.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		// Try the move to see if it's legal
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue // Skip invalid move
		}

		// Check if king is in check after move (illegal)
		if m.generator.IsKingInCheck(b, player) {
			b.UnmakeMove(undo)
			continue // Skip illegal move
		}

		legalMoveCount++

		// Apply pruning only when not in check and only for captures
		if !inCheck && move.IsCapture {
			// Delta pruning - skip captures that can't improve alpha significantly
			captureValue := ai.EvaluationScore(0)
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

			margin := ai.EvaluationScore(200)
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
		score := -m.quiescence(ctx, b, oppositePlayer(player), -beta, -alpha, depthFromRoot+1, stats)
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

	// Handle the case where we have no legal moves
	if inCheck && legalMoveCount == 0 {
		// Checkmate - no legal moves when in check
		// The mate distance should be how many plies from the original search root
		// In quiescence, depthFromRoot represents the distance from the search root
		return -ai.MateScore + ai.EvaluationScore(depthFromRoot)
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
		m.transpositionTable.Store(hash, 0, bestScore, entryType, board.Move{})
	}

	return bestScore
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}
