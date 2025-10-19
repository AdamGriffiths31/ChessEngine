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

	m.orderMoves(b, pseudoMoves, 0, rootTTMove)

	lastCompletedBestMove := pseudoMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)
	lastCompletedDepth := 0
	var finalStats ai.SearchStats

	maxPlyDepth := config.MaxDepth + 20
	pv := make([][]board.Move, maxPlyDepth)
	for i := range pv {
		pv[i] = make([]board.Move, maxPlyDepth)
	}

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

		alpha := -ai.MateScore - 1
		beta := ai.MateScore + 1

		// Use aspiration windows for depths > 1
		useAspirationWindow := currentDepth > 1
		window := ai.EvaluationScore(50)

		if useAspirationWindow {
			alpha = lastCompletedScore - window
			beta = lastCompletedScore + window
		}

		// Keep trying with wider windows until we get a score within bounds
		for {
			tempBestScore := -ai.MateScore - 1
			tempBestMove := board.Move{}
			moveIndex := 0

			for _, move := range pseudoMoves.Moves[:pseudoMoves.Count] {
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

			// Aspiration window failed - widen and retry
			if tempBestScore <= lastCompletedScore-window {
				// Fail low - score is worse than expected
				window *= 2
				alpha = lastCompletedScore - window
				// Keep beta the same to avoid re-searching moves that already failed high
			} else if tempBestScore >= lastCompletedScore+window {
				// Fail high - score is better than expected
				window *= 2
				beta = lastCompletedScore + window
				// Keep alpha at the current best score found
			}

			// Safety: if window gets too large, disable aspiration
			if window > 1000 {
				useAspirationWindow = false
				alpha = -ai.MateScore - 1
				beta = ai.MateScore + 1
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
