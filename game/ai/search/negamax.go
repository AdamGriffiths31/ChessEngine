// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// negamax performs negamax search with alpha-beta pruning and optimizations
func (m *MinimaxEngine) negamax(ctx context.Context, b *board.Board, player moves.Player, depth int, alpha, beta ai.EvaluationScore, originalMaxDepth int, config ai.SearchConfig, stats *ai.SearchStats) ai.EvaluationScore {
	m.searchState.searchStats.NodesSearched++

	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	select {
	case <-ctx.Done():
		m.searchState.searchCancelled = true
		return alpha
	default:
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	if inCheck && depth < originalMaxDepth {
		depth++
	}

	var ttMove board.Move
	hash := b.GetHash()

	// Check for draw by repetition
	if m.isDrawByRepetition(hash) {
		return ai.DrawScore
	}

	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			ttMove = entry.GetMove()

			if ttMove.From.File >= 0 && ttMove.From.File <= 7 &&
				ttMove.To.File >= 0 && ttMove.To.File <= 7 {
				if ttMove.From == ttMove.To {
					ttMove = board.Move{}
				}
			} else {
				ttMove = board.Move{}
			}

			if entry.GetDepth() >= depth {
				switch entry.GetType() {
				case EntryExact:
					m.searchState.searchStats.TTCutoffs++
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						m.searchState.searchStats.TTCutoffs++
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

	staticEval := m.evaluator.Evaluate(b)
	// Null move pruning: If our position is so good we can give opponent a free move
	// and still achieve a beta cutoff, we can prune this branch
	if depth >= 3 &&
		staticEval >= beta &&
		beta < ai.MateScore-MateDistanceThreshold &&
		beta > -ai.MateScore+MateDistanceThreshold {
		if !inCheck {
			m.searchState.searchStats.NullMoves++

			nullReduction := m.searchState.searchParams.NullMoveReduction
			if depth >= 6 && nullReduction < 3 {
				nullReduction++
			}

			nullUndo := b.MakeNullMove()

			nullScore := -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-nullReduction, -beta, -beta+1, originalMaxDepth, config, stats)

			b.UnmakeNullMove(nullUndo)

			if nullScore >= beta {
				if nullScore < ai.MateScore-MateDistanceThreshold {
					m.searchState.searchStats.NullCutoffs++
					return beta
				}
			}
		}
	}

	// Razoring: If static eval + margin is below alpha at low depths,
	// verify with quiescence search and potentially prune
	if m.searchState.searchParams.RazoringEnabled &&
		!inCheck &&
		depth <= m.searchState.searchParams.RazoringMaxDepth &&
		depth > 0 {

		razoringMargin := m.searchState.searchParams.RazoringMargins[depth]

		if staticEval+razoringMargin < alpha {
			// Don't check for TT move at all
			m.searchState.searchStats.RazoringAttempts++

			qScore := m.quiescence(ctx, b, player, alpha, beta,
				originalMaxDepth-depth, stats)

			if qScore <= alpha {
				m.searchState.searchStats.RazoringCutoffs++
				return qScore
			}
			m.searchState.searchStats.RazoringFailed++
		}
	}

	if depth <= 0 {
		return m.quiescence(ctx, b, player, alpha, beta, originalMaxDepth-depth, stats)
	}

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer moves.ReleaseMoveList(pseudoMoves)

	if pseudoMoves.Count == 0 {
		return m.handleNoLegalMoves(b, player, depth, originalMaxDepth, hash)
	}

	m.orderMoves(b, pseudoMoves, currentDepth, ttMove)

	bestScore := -ai.MateScore - 1
	bestMove := board.Move{}
	legalMoveCount := 0

	// Track if we improved alpha to determine correct entry type
	alphaImproved := false

	for i := 0; i < pseudoMoves.Count; i++ {
		move := pseudoMoves.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue // Skip invalid move
		}

		// Check if king is in check after move (illegal)
		if m.generator.IsKingInCheck(b, player) {
			b.UnmakeMove(undo)
			continue // Skip illegal move
		}

		// Add position to repetition history
		m.addHistory(b.GetHash())

		legalMoveCount++
		var score ai.EvaluationScore

		// Late Move Reductions: Search later moves at reduced depth since
		// move ordering should place best moves first
		reduction := m.calculateLMRReduction(b, depth, legalMoveCount, inCheck, move, currentDepth, config)

		if reduction > 0 {
			m.searchState.searchStats.LMRReductions++

			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-reduction, -alpha-1, -alpha, originalMaxDepth, config, stats)

			if score > alpha {
				m.searchState.searchStats.LMRReSearches++

				score = -m.negamax(ctx, b, oppositePlayer(player),
					depth-1, -beta, -alpha, originalMaxDepth, config, stats)
			}
		} else {
			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1, -beta, -alpha, originalMaxDepth, config, stats)
		}

		b.UnmakeMove(undo)

		// Remove position from repetition history
		m.removeHistory()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if alpha >= beta {
				// Track move ordering statistics
				m.searchState.searchStats.TotalCutoffs++
				if legalMoveCount == 1 {
					m.searchState.searchStats.FirstMoveCutoffs++
				}

				if !move.IsCapture {
					m.storeKiller(move, currentDepth)
				}

				if !move.IsCapture {
					if m.historyTable != nil {
						m.historyTable.UpdateHistory(move, depth)
					}
				}

				if m.transpositionTable != nil && !m.searchState.searchCancelled {
					m.transpositionTable.Store(hash, depth, bestScore, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	// Check if we had no legal moves
	if legalMoveCount == 0 {
		return m.handleNoLegalMoves(b, player, depth, originalMaxDepth, hash)
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		// Determine correct entry type based on bounds
		var entryType EntryType
		if alphaImproved {
			// We found a move that improved alpha
			if bestScore >= beta {
				// This shouldn't happen as we return early on beta cutoff
				entryType = EntryLowerBound
			} else {
				// Score is between original alpha and beta
				entryType = EntryExact
			}
		} else {
			// No move improved alpha - this is an upper bound
			entryType = EntryUpperBound
		}

		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
}

// handleNoLegalMoves returns the appropriate score when no legal moves are available
// Returns checkmate score if in check, stalemate score otherwise
func (m *MinimaxEngine) handleNoLegalMoves(b *board.Board, player moves.Player, depth, originalMaxDepth int, hash uint64) ai.EvaluationScore {
	if m.generator.IsKingInCheck(b, player) {
		// Checkmate - the mate distance is how many plies from the root we are
		// Negative score because it's mate AGAINST the current player
		pliesFromRoot := originalMaxDepth - depth
		score := -ai.MateScore + ai.EvaluationScore(pliesFromRoot)
		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, depth, score, EntryExact, board.Move{})
		}
		return score
	}
	// Stalemate
	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, ai.DrawScore, EntryExact, board.Move{})
	}
	return ai.DrawScore
}

// calculateLMRReduction calculates the Late Move Reduction amount for a move.
// Returns the number of plies to reduce search depth by, based on depth, move count,
// and history heuristic. Returns 0 if no reduction should be applied.
func (m *MinimaxEngine) calculateLMRReduction(b *board.Board, depth, legalMoveCount int, inCheck bool, move board.Move, currentDepth int, config ai.SearchConfig) int {
	// Don't reduce if conditions aren't met
	if depth < config.LMRMinDepth ||
		legalMoveCount <= config.LMRMinMoves ||
		inCheck ||
		move.IsCapture ||
		move.Promotion != board.Empty ||
		m.isKillerMove(move, currentDepth) {
		return 0
	}

	// Don't reduce moves that give check
	if board.MoveGivesCheck(b, move) {
		return 0
	}

	// Use pre-calculated LMR table for base reduction (performance optimization)
	tableDepth := min(depth, 15)
	tableMoveCount := min(legalMoveCount, 63)
	if tableDepth < 1 || tableMoveCount < 1 {
		return 0
	}
	reduction := LMRTable[tableDepth][tableMoveCount]

	// Adjust reduction based on history heuristic
	historyScore := m.getHistoryScore(move)
	if historyScore > m.searchState.searchParams.HistoryHighThreshold {
		// Very good history - don't reduce
		return 0
	} else if historyScore > m.searchState.searchParams.HistoryMedThreshold && reduction > 0 {
		// Good history - reduce less
		reduction = reduction * 2 / 3
	} else if historyScore < m.searchState.searchParams.HistoryLowThreshold {
		// Bad history - reduce more
		reduction = reduction * 4 / 3
	}

	// Clamp reduction to valid range
	if reduction >= depth {
		reduction = depth - 1
	}
	if reduction < 0 {
		reduction = 0
	}

	return reduction
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
