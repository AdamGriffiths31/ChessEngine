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

// rootSearch holds the per-search state shared by the root-move helpers.
type rootSearch struct {
	player, opponent movegen.Player
	moves            *movegen.MoveList // pseudo-legal root moves, reordered between depths

	// Per-move scores from the current pass, indexed in step with moves.Moves.
	// exact marks the entries that are exact scores: only a move searched with
	// a full window is exact; the rest failed low on a null-window scout, so
	// their score is an upper bound.
	scores []eval.EvaluationScore
	exact  []bool

	startTime time.Time
}

// runIterativeDeepening runs the core iterative deepening search
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player movegen.Player, config Config, startTime time.Time) Result {
	// Per-search invariants shared with negamax/quiescence via State.
	m.searchState.board = b
	m.searchState.config = config

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer movegen.ReleaseMoveList(pseudoMoves)

	m.orderMovesAtRoot(b, pseudoMoves, m.probeRootTTMove(b))

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

	m.searchState.pv = newPVTable(config.MaxDepth + PVArrayMargin)
	pv := m.searchState.pv

	rs := &rootSearch{
		player:    player,
		opponent:  oppositePlayer(player),
		moves:     pseudoMoves,
		scores:    make([]eval.EvaluationScore, pseudoMoves.Count),
		exact:     make([]bool, pseudoMoves.Count),
		startTime: startTime,
	}

	for currentDepth := 1; currentDepth <= config.MaxDepth; currentDepth++ {
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

		bestScore, bestMove, failLows, failHighs := m.searchWithAspiration(ctx, rs, currentDepth, lastCompletedScore)

		finalStats.adoptCounters(m.searchState.searchStats)
		finalStats.Depth = currentDepth

		if !m.searchState.searchCancelled && (config.MaxTime == 0 || time.Since(startTime) < config.MaxTime) {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			lastCompletedDepth = currentDepth
			m.recordIteration(&finalStats, rs, currentDepth, bestScore, bestMove, failLows, failHighs)

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

// probeRootTTMove returns the TT move stored for the root position, if any,
// to seed the root move ordering.
func (m *MinimaxEngine) probeRootTTMove(b *board.Board) board.Move {
	if m.transpositionTable == nil {
		return board.Move{}
	}
	m.searchState.searchStats.TTProbes++
	entry, found := m.transpositionTable.Probe(b.GetHash())
	if !found {
		return board.Move{}
	}
	m.searchState.searchStats.TTHits++
	return entry.GetMove()
}

// searchWithAspiration searches all root moves at depth, starting from a
// window around prevScore and widening it until the best score lands inside.
// It returns the best score and move plus how many times the window failed low
// and high on the way.
func (m *MinimaxEngine) searchWithAspiration(ctx context.Context, rs *rootSearch, depth int, prevScore eval.EvaluationScore) (bestScore eval.EvaluationScore, bestMove board.Move, failLows, failHighs int) {
	useAspirationWindow := depth > 1
	window := eval.EvaluationScore(AspirationWindow)

	// Keep trying with wider windows until we get a score within bounds
	for {
		// Reset both bounds on every attempt: a failed pass leaves alpha
		// raised above the window, and re-searching with one-sided bounds
		// can accept an upper-bound score as exact
		alpha := -eval.MateScore - 1
		beta := eval.MateScore + 1
		if useAspirationWindow {
			alpha = prevScore - window
			beta = prevScore + window
		}

		bestScore, bestMove = m.searchRootMoves(ctx, rs, depth, alpha, beta)

		if !useAspirationWindow || (bestScore > prevScore-window && bestScore < prevScore+window) {
			// Score is within aspiration window or we're not using aspiration
			return bestScore, bestMove, failLows, failHighs
		}

		// Aspiration window failed - widen and retry with fresh bounds
		if bestScore <= prevScore-window {
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
}

// searchRootMoves makes one pass over the root moves with the given window,
// filling rs.scores and rs.exact, and returns the best score and move found.
func (m *MinimaxEngine) searchRootMoves(ctx context.Context, rs *rootSearch, depth int, alpha, beta eval.EvaluationScore) (bestScore eval.EvaluationScore, bestMove board.Move) {
	b := m.searchState.board
	pv := m.searchState.pv
	maxTime := m.searchState.config.MaxTime

	bestScore = -eval.MateScore - 1
	searched := 0

	for i := range rs.scores {
		rs.scores[i] = -eval.MateScore - 1
		rs.exact[i] = false
	}

	for moveIdx := range rs.moves.Count {
		move := rs.moves.Moves[moveIdx]
		if m.searchState.searchCancelled {
			break
		}

		undo, ok := m.tryMove(move, rs.player)
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
		// full re-search) can then read and promote into pv[0].
		pv.clearPly(1)

		windowAlpha := alpha
		score, fullWindow := m.searchRootMove(ctx, rs.opponent, depth, alpha, beta, searched == 0)

		b.UnmakeMove(undo)

		m.removeHistory()

		rs.scores[moveIdx] = score
		rs.exact[moveIdx] = fullWindow && score > windowAlpha && score < beta

		if score > bestScore {
			bestScore = score
			bestMove = move

			if score > alpha {
				alpha = score
			}

			pv.store(0, move)
		}

		searched++

		if maxTime > 0 && time.Since(rs.startTime) >= maxTime {
			break
		}
	}

	return bestScore, bestMove
}

// searchRootMove scores one root move, already made on the board, from the
// root player's point of view. The first move gets the full window; later ones
// a null-window scout, re-searched with the full window if they land inside
// (alpha, beta). fullWindow reports whether the returned score came from a
// full-window search.
//
// The root move has been made, so these negamax nodes sit one ply below the
// root: ply == 1.
func (m *MinimaxEngine) searchRootMove(ctx context.Context, opponent movegen.Player, depth int, alpha, beta eval.EvaluationScore, first bool) (score eval.EvaluationScore, fullWindow bool) {
	if first {
		return -m.negamax(ctx, opponent, depth-1, -beta, -alpha, 1, true), true
	}

	score = -m.negamax(ctx, opponent, depth-1, -alpha-1, -alpha, 1, true)
	if score > alpha && score < beta {
		return -m.negamax(ctx, opponent, depth-1, -beta, -alpha, 1, true), true
	}
	return score, false
}

// recordIteration appends a completed iteration to stats and prepares the next
// one: it logs the top root moves and principal variation, and re-orders the
// root moves by this iteration's scores.
func (m *MinimaxEngine) recordIteration(stats *Stats, rs *rootSearch, depth int, bestScore eval.EvaluationScore, bestMove board.Move, failLows, failHighs int) {
	stats.Iterations = append(stats.Iterations, Iteration{
		Depth:    depth,
		Score:    bestScore,
		BestMove: bestMove,
		Nodes:    m.searchState.searchStats.NodesSearched,
		Elapsed:  time.Since(rs.startTime),

		AspirationFailLows:  failLows,
		AspirationFailHighs: failHighs,
	})
	stats.RootMoves = topRootMoves(rs.moves, rs.scores, rs.exact, MaxLoggedRootMoves)

	// Re-order root moves by this iteration's scores so the best move
	// is searched first at the next depth - PVS gives only the first
	// move a full window and aspiration assumes the score carries over
	reorderRootMoves(rs.moves, rs.scores)

	stats.PrincipalVariation = m.searchState.pv.lineAt(0)
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
