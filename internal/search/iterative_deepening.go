// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"
	"errors"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// runIterativeDeepening runs the core iterative deepening search
//
//nolint:gocyclo // refactored in Phase 4
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player movegen.Player, config Config, startTime time.Time) Result {
	// Per-search invariants shared with negamax/quiescence via State.
	m.searchState.board = b
	m.searchState.player = player
	m.searchState.config = config

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer movegen.ReleaseMoveList(pseudoMoves)

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

	// The first legal move in root order is the answer if no iteration
	// completes. Pseudo-legal moves can all be illegal (checkmate, stalemate),
	// so having generated moves says nothing about having legal ones.
	firstLegal, hasLegal := m.firstLegalRootMove(pseudoMoves, player)
	if !hasLegal {
		score := eval.DrawScore
		if m.generator.IsKingInCheck(b, player) {
			score = -eval.MateScore
		}
		return Result{
			BestMove: board.Move{},
			Score:    score,
			Stats:    Stats{StopReason: StopReasonNoLegalMoves},
		}
	}

	lastCompletedBestMove := firstLegal
	lastCompletedScore := eval.EvaluationScore(0)
	lastCompletedDepth := 0
	var finalStats Stats

	maxPlyDepth := config.MaxDepth + PVArrayMargin
	m.searchState.pv = newPVTable(maxPlyDepth)
	pv := m.searchState.pv

	// Per-move scores from the current iteration, used to re-order root moves
	// between depths (indexed in step with pseudoMoves.Moves)
	rootScores := make([]eval.EvaluationScore, pseudoMoves.Count)
	// rootExact marks the entries of rootScores that are exact scores. Only a
	// move searched with a full window is exact; the rest failed low on a
	// null-window scout, so their score is an upper bound.
	rootExact := make([]bool, pseudoMoves.Count)

	startingDepth := 1

	for currentDepth := startingDepth; currentDepth <= config.MaxDepth; currentDepth++ {
		m.searchState.searchCancelled = false

		pv.reset()

		select {
		case <-ctx.Done():
			finalStats.Time = time.Since(startTime)
			finalStats.Depth = lastCompletedDepth
			finalStats.StopReason = StopReasonCancelled
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				finalStats.StopReason = StopReasonTime
			}
			return Result{
				BestMove: lastCompletedBestMove,
				Score:    lastCompletedScore,
				Stats:    finalStats,
			}
		default:
		}

		if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
			finalStats.StopReason = StopReasonTime
			break
		}

		bestScore := -eval.MateScore - 1
		var bestMove board.Move

		useAspirationWindow := currentDepth > 1
		window := eval.EvaluationScore(AspirationWindow)
		var failLows, failHighs int

		// Keep trying with wider windows until we get a score within bounds
		for {
			// Reset both bounds on every attempt: a failed pass leaves alpha
			// raised above the window, and re-searching with one-sided bounds
			// can accept an upper-bound score as exact
			alpha := -eval.MateScore - 1
			beta := eval.MateScore + 1
			if useAspirationWindow {
				alpha = lastCompletedScore - window
				beta = lastCompletedScore + window
			}

			tempBestScore := -eval.MateScore - 1
			tempBestMove := board.Move{}
			moveIndex := 0

			for i := range rootScores {
				rootScores[i] = -eval.MateScore - 1
				rootExact[i] = false
			}

			for moveIdx := range pseudoMoves.Count {
				move := pseudoMoves.Moves[moveIdx]
				if m.searchState.searchCancelled {
					break
				}

				undo, ok := m.tryMove(move, player)
				if !ok {
					continue
				}

				m.addHistory(b.GetHash())

				// Clear the row each root move's search will build into before
				// evaluating it, mirroring negamax's own per-move
				// `pv.clearPly(pvPly + 1)`. Without this, a root move whose
				// own negamax call never improves its (possibly very narrow,
				// for a null-window scout) alpha never calls pv.store itself,
				// leaving row 1 holding a PREVIOUS root move's stale line -
				// which a later move's fail-high (score >= beta, skipping the
				// full re-search below) can then read and promote into pv[0].
				pv.clearPly(1)

				var score eval.EvaluationScore
				windowAlpha := alpha
				fullWindow := moveIndex == 0

				// The root move has been made on the board, so these negamax
				// nodes sit one ply below the root: ply == 1. (Formerly this was
				// originalMaxDepth - depth == currentDepth - (currentDepth-1) == 1.)
				if moveIndex == 0 {
					m.searchState.player = oppositePlayer(player)
					m.searchState.collectPV = true
					score = -m.negamax(ctx, currentDepth-1, -beta, -alpha, 1)
				} else {
					m.searchState.player = oppositePlayer(player)
					m.searchState.collectPV = true
					score = -m.negamax(ctx, currentDepth-1, -alpha-1, -alpha, 1)

					if score > alpha && score < beta {
						m.searchState.player = oppositePlayer(player)
						m.searchState.collectPV = true
						score = -m.negamax(ctx, currentDepth-1, -beta, -alpha, 1)
						fullWindow = true
					}
				}

				b.UnmakeMove(undo)

				m.removeHistory()

				rootScores[moveIdx] = score
				rootExact[moveIdx] = fullWindow && score > windowAlpha && score < beta

				if score > tempBestScore {
					tempBestScore = score
					tempBestMove = move

					if score > alpha {
						alpha = score
					}

					pv.store(0, move)
				}

				moveIndex++

				if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
					break
				}
			}

			if !useAspirationWindow || (tempBestScore > lastCompletedScore-window && tempBestScore < lastCompletedScore+window) {
				// Score is within aspiration window or we're not using aspiration
				bestScore = tempBestScore
				bestMove = tempBestMove
				break
			}

			// Aspiration window failed - widen and retry with fresh bounds
			if tempBestScore <= lastCompletedScore-window {
				failLows++
			} else {
				failHighs++
			}
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
		finalStats.NullMoves = m.searchState.searchStats.NullMoves
		finalStats.NullCutoffs = m.searchState.searchStats.NullCutoffs
		finalStats.NullMoveZugzwangSkipped = m.searchState.searchStats.NullMoveZugzwangSkipped
		finalStats.NullMoveSparseSkipped = m.searchState.searchStats.NullMoveSparseSkipped
		finalStats.QNodes = m.searchState.searchStats.QNodes
		finalStats.TTCutoffs = m.searchState.searchStats.TTCutoffs
		finalStats.FirstMoveCutoffs = m.searchState.searchStats.FirstMoveCutoffs
		finalStats.TotalCutoffs = m.searchState.searchStats.TotalCutoffs
		finalStats.DeltaPruned = m.searchState.searchStats.DeltaPruned
		finalStats.RFPCutoffs = m.searchState.searchStats.RFPCutoffs
		finalStats.RazoringAttempts = m.searchState.searchStats.RazoringAttempts
		finalStats.RazoringCutoffs = m.searchState.searchStats.RazoringCutoffs
		finalStats.RazoringFailed = m.searchState.searchStats.RazoringFailed
		finalStats.FutilityPrunes = m.searchState.searchStats.FutilityPrunes
		finalStats.LMPPrunes = m.searchState.searchStats.LMPPrunes
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
			finalStats.Iterations = append(finalStats.Iterations, Iteration{
				Depth:    currentDepth,
				Score:    bestScore,
				BestMove: bestMove,
				Nodes:    m.searchState.searchStats.NodesSearched,
				Elapsed:  time.Since(startTime),

				AspirationFailLows:  failLows,
				AspirationFailHighs: failHighs,
			})
			finalStats.RootMoves = topRootMoves(pseudoMoves, rootScores, rootExact, MaxLoggedRootMoves)

			// Re-order root moves by this iteration's scores so the best move
			// is searched first at the next depth - PVS gives only the first
			// move a full window and aspiration assumes the score carries over
			reorderRootMoves(pseudoMoves, rootScores)

			finalStats.PrincipalVariation = pv.lineAt(0)

			// Only break on mate if we found the shortest possible mate (mate in 1)
			// This allows us to search deeper for accurate mate distance calculation.
			// Checked only for a completed iteration: a cancelled one leaves
			// bestScore at its -MateScore-1 starting value, which is not a mate.
			if bestScore >= eval.MateScore-1 || bestScore <= -eval.MateScore+1 {
				finalStats.StopReason = StopReasonMate
				break
			}

			if shouldStopOnSoftTime(config.SoftTime, time.Since(startTime), finalStats.Iterations) {
				finalStats.StopReason = StopReasonTime
				break
			}
		}
	}

	if finalStats.StopReason == "" {
		// The loop ran out of depths. If the last one did not complete, the
		// clock or the context cut it short.
		switch {
		case lastCompletedDepth >= config.MaxDepth:
			finalStats.StopReason = StopReasonDepth
		case errors.Is(ctx.Err(), context.Canceled):
			finalStats.StopReason = StopReasonCancelled
		default:
			finalStats.StopReason = StopReasonTime
		}
	}
	finalStats.Time = time.Since(startTime)
	finalStats.Depth = lastCompletedDepth

	return Result{
		BestMove: lastCompletedBestMove,
		Score:    lastCompletedScore,
		Stats:    finalStats,
	}
}

// firstLegalRootMove returns the first move in moves that leaves the mover's
// king safe, and whether there is one. The board is restored before returning.
func (m *MinimaxEngine) firstLegalRootMove(moves *movegen.MoveList, player movegen.Player) (board.Move, bool) {
	for i := range moves.Count {
		undo, ok := m.tryMove(moves.Moves[i], player)
		if !ok {
			continue
		}
		m.searchState.board.UnmakeMove(undo)
		return moves.Moves[i], true
	}
	return board.Move{}, false
}

// topRootMoves returns up to n of the moves with an exact score this
// iteration, best score first (ties keep root order, which puts the chosen
// move first). Moves that failed low, or that the iteration never reached,
// are left out. The slices are index-aligned and are read before
// reorderRootMoves permutes them.
func topRootMoves(moves *movegen.MoveList, scores []eval.EvaluationScore, exact []bool, n int) []RootMove {
	count := min(moves.Count, len(scores))
	out := make([]RootMove, 0, n)
	for i := range count {
		if !exact[i] {
			continue
		}
		rm := RootMove{Move: moves.Moves[i], Score: scores[i]}
		pos := len(out)
		for pos > 0 && out[pos-1].Score < rm.Score {
			pos--
		}
		if pos >= n {
			continue
		}
		if len(out) < n {
			out = append(out, RootMove{})
		}
		copy(out[pos+1:], out[pos:])
		out[pos] = rm
	}
	return out
}

// reorderRootMoves stably sorts the first moveList.Count moves in descending
// order of scores, keeping the two slices index-aligned. Insertion sort is
// used for stability: moves with equal scores keep their relative order, and
// unscored moves (illegal or unsearched, recorded as -MateScore-1) sink to
// the back without being shuffled among themselves.
func reorderRootMoves(moveList *movegen.MoveList, scores []eval.EvaluationScore) {
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
