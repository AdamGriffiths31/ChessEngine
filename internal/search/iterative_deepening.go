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

type rootSearch struct {
	player, opponent movegen.Player
	moves            *movegen.MoveList

	scores []eval.EvaluationScore
	exact  []bool

	partialScore eval.EvaluationScore
	partialMove  board.Move

	startTime time.Time
}

func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player movegen.Player, config Config, startTime time.Time) Result {
	m.searchState.board = b
	m.searchState.config = config

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer movegen.ReleaseMoveList(pseudoMoves)

	m.orderMovesAtRoot(b, pseudoMoves, m.probeRootTTMove(b))

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
		rs.partialScore = -eval.MateScore - 1

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

			if bestScore >= eval.MateScore-1 || bestScore <= -eval.MateScore+1 {
				finalStats.StopReason = StopReasonMate
				break
			}

			if shouldStopOnSoftTime(config.SoftTime, time.Since(startTime), finalStats.Iterations) {
				finalStats.StopReason = StopReasonTime
				break
			}
		} else if rs.partialScore > lastCompletedScore {
			lastCompletedBestMove = rs.partialMove
			lastCompletedScore = rs.partialScore
			finalStats.PrincipalVariation = []board.Move{rs.partialMove}
		}
	}

	if finalStats.StopReason == "" {
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

func (m *MinimaxEngine) searchWithAspiration(ctx context.Context, rs *rootSearch, depth int, prevScore eval.EvaluationScore) (bestScore eval.EvaluationScore, bestMove board.Move, failLows, failHighs int) {
	useAspirationWindow := depth > 1
	window := eval.EvaluationScore(AspirationWindow)

	for {
		alpha := -eval.MateScore - 1
		beta := eval.MateScore + 1
		if useAspirationWindow {
			alpha = prevScore - window
			beta = prevScore + window
		}

		bestScore, bestMove = m.searchRootMoves(ctx, rs, depth, alpha, beta)

		if !useAspirationWindow || (bestScore > prevScore-window && bestScore < prevScore+window) {
			return bestScore, bestMove, failLows, failHighs
		}

		if bestScore <= prevScore-window {
			failLows++
		} else {
			failHighs++
		}
		window *= 2

		if window > 1000 {
			useAspirationWindow = false
		}
	}
}

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

		pv.clearPly(1)

		windowAlpha := alpha
		score, fullWindow := m.searchRootMove(ctx, rs.opponent, depth, alpha, beta, searched == 0)

		b.UnmakeMove(undo)

		m.removeHistory()

		if !m.searchState.searchCancelled && score > alpha && score > rs.partialScore {
			rs.partialScore, rs.partialMove = score, move
		}

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

	reorderRootMoves(rs.moves, rs.scores)

	stats.PrincipalVariation = m.searchState.pv.lineAt(0)
}

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
