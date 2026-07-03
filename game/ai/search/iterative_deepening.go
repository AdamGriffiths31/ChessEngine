// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// runIterativeDeepening runs the core iterative deepening search
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig, startTime time.Time) ai.SearchResult {

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer moves.ReleaseMoveList(pseudoMoves)

	if pseudoMoves.Count == 0 {
		isCheck := m.generator.IsKingInCheck(b, player)
		if isCheck {
			return ai.SearchResult{
				BestMove: board.Move{},
				Score:    -ai.MateScore,
				Stats:    ai.SearchStats{},
			}
		}
		return ai.SearchResult{
			BestMove: board.Move{},
			Score:    ai.DrawScore,
			Stats:    ai.SearchStats{},
		}
	}

	var rootTTMove board.Move
	if m.transpositionTable != nil {
		hash := b.GetHash()
		m.searchState.searchStats.TTProbes++
		if entry, found := m.transpositionTable.Probe(hash); found {
			m.searchState.searchStats.TTHits++
			rootTTMove = entry.GetMove()
		}
	}

	m.orderMovesAtRoot(b, pseudoMoves, rootTTMove)

	lastCompletedBestMove := pseudoMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)
	lastCompletedDepth := 0
	var finalStats ai.SearchStats

	maxPlyDepth := config.MaxDepth + 20
	pv := make([][]board.Move, maxPlyDepth)
	for i := range pv {
		pv[i] = make([]board.Move, maxPlyDepth)
	}

	// Per-move scores from the current iteration, used to re-order root moves
	// between depths (indexed in step with pseudoMoves.Moves)
	rootScores := make([]ai.EvaluationScore, pseudoMoves.Count)

	startingDepth := 1

	for currentDepth := startingDepth; currentDepth <= config.MaxDepth; currentDepth++ {
		m.searchState.searchCancelled = false

		for i := range pv {
			pv[i][0] = board.Move{}
		}

		select {
		case <-ctx.Done():
			finalStats.Time = time.Since(startTime)
			finalStats.Depth = lastCompletedDepth
			return ai.SearchResult{
				BestMove: lastCompletedBestMove,
				Score:    lastCompletedScore,
				Stats:    finalStats,
			}
		default:
		}

		if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
			break
		}

		bestScore := -ai.MateScore - 1
		var bestMove board.Move

		// Use aspiration windows for depths > 1
		useAspirationWindow := currentDepth > 1
		window := ai.EvaluationScore(50)

		// Keep trying with wider windows until we get a score within bounds
		for {
			// Reset both bounds on every attempt: a failed pass leaves alpha
			// raised above the window, and re-searching with one-sided bounds
			// can accept an upper-bound score as exact
			alpha := -ai.MateScore - 1
			beta := ai.MateScore + 1
			if useAspirationWindow {
				alpha = lastCompletedScore - window
				beta = lastCompletedScore + window
			}

			tempBestScore := -ai.MateScore - 1
			tempBestMove := board.Move{}
			moveIndex := 0

			for i := range rootScores {
				rootScores[i] = -ai.MateScore - 1
			}

			for moveIdx := 0; moveIdx < pseudoMoves.Count; moveIdx++ {
				move := pseudoMoves.Moves[moveIdx]
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

				var score ai.EvaluationScore

				if moveIndex == 0 {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, &finalStats, pv)
				} else {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -alpha-1, -alpha, currentDepth, config, &finalStats, pv)

					if score > alpha && score < beta {
						score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, &finalStats, pv)
					}
				}

				b.UnmakeMove(undo)

				// Remove position from repetition history
				m.removeHistory()

				rootScores[moveIdx] = score

				if score > tempBestScore {
					tempBestScore = score
					tempBestMove = move

					if score > alpha {
						alpha = score
					}

					pv[0][0] = move
					pvLen := 1
					for i := 0; i < len(pv[1]) && pv[1][i] != (board.Move{}); i++ {
						pv[0][pvLen] = pv[1][i]
						pvLen++
						if pvLen >= len(pv[0]) {
							break
						}
					}
					if pvLen < len(pv[0]) {
						pv[0][pvLen] = board.Move{}
					}
				}

				moveIndex++

				if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
					break
				}
			}

			// Check if we need to re-search with wider window
			if !useAspirationWindow || (tempBestScore > lastCompletedScore-window && tempBestScore < lastCompletedScore+window) {
				// Score is within aspiration window or we're not using aspiration
				bestScore = tempBestScore
				bestMove = tempBestMove
				break
			}

			// Aspiration window failed - widen and retry with fresh bounds
			window *= 2

			// Safety: if window gets too large, drop to a full-width search
			if window > 1000 {
				useAspirationWindow = false
			}
		}

		finalStats.NodesSearched = m.searchState.searchStats.NodesSearched
		finalStats.Depth = currentDepth
		finalStats.LMRReductions = m.searchState.searchStats.LMRReductions
		finalStats.LMRReSearches = m.searchState.searchStats.LMRReSearches
		finalStats.LMRNodesSkipped = m.searchState.searchStats.LMRNodesSkipped
		finalStats.NullMoves = m.searchState.searchStats.NullMoves
		finalStats.NullCutoffs = m.searchState.searchStats.NullCutoffs
		finalStats.QNodes = m.searchState.searchStats.QNodes
		finalStats.TTCutoffs = m.searchState.searchStats.TTCutoffs
		finalStats.FirstMoveCutoffs = m.searchState.searchStats.FirstMoveCutoffs
		finalStats.TotalCutoffs = m.searchState.searchStats.TotalCutoffs
		finalStats.DeltaPruned = m.searchState.searchStats.DeltaPruned
		finalStats.RazoringAttempts = m.searchState.searchStats.RazoringAttempts
		finalStats.RazoringCutoffs = m.searchState.searchStats.RazoringCutoffs
		finalStats.RazoringFailed = m.searchState.searchStats.RazoringFailed
		finalStats.TTProbes = m.searchState.searchStats.TTProbes
		finalStats.TTHits = m.searchState.searchStats.TTHits
		finalStats.PVNodes = m.searchState.searchStats.PVNodes
		finalStats.CutNodes = m.searchState.searchStats.CutNodes
		finalStats.AllNodes = m.searchState.searchStats.AllNodes
		finalStats.CutoffsByMoveIndex = m.searchState.searchStats.CutoffsByMoveIndex
		finalStats.NodesByDepth = m.searchState.searchStats.NodesByDepth

		if !m.searchState.searchCancelled && (config.MaxTime == 0 || time.Since(startTime) < config.MaxTime) {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			lastCompletedDepth = currentDepth

			// Re-order root moves by this iteration's scores so the best move
			// is searched first at the next depth - PVS gives only the first
			// move a full window and aspiration assumes the score carries over
			reorderRootMoves(pseudoMoves, rootScores)

			finalStats.PrincipalVariation = make([]board.Move, 0, currentDepth)
			for i := 0; i < len(pv[0]) && pv[0][i] != (board.Move{}); i++ {
				finalStats.PrincipalVariation = append(finalStats.PrincipalVariation, pv[0][i])
			}
		}

		// Only break on mate if we found the shortest possible mate (mate in 1)
		// This allows us to search deeper for accurate mate distance calculation
		if bestScore >= ai.MateScore-1 || bestScore <= -ai.MateScore+1 {
			break
		}
	}

	finalStats.Time = time.Since(startTime)
	finalStats.Depth = lastCompletedDepth

	return ai.SearchResult{
		BestMove: lastCompletedBestMove,
		Score:    lastCompletedScore,
		Stats:    finalStats,
	}
}

// reorderRootMoves stably sorts the first moveList.Count moves in descending
// order of scores, keeping the two slices index-aligned. Insertion sort is
// used for stability: moves with equal scores keep their relative order, and
// unscored moves (illegal or unsearched, recorded as -MateScore-1) sink to
// the back without being shuffled among themselves.
func reorderRootMoves(moveList *moves.MoveList, scores []ai.EvaluationScore) {
	count := moveList.Count
	if count > len(scores) {
		count = len(scores)
	}
	for i := 1; i < count; i++ {
		mv, sc := moveList.Moves[i], scores[i]
		j := i - 1
		for j >= 0 && scores[j] < sc {
			moveList.Moves[j+1] = moveList.Moves[j]
			scores[j+1] = scores[j]
			j--
		}
		moveList.Moves[j+1] = mv
		scores[j+1] = sc
	}
}
