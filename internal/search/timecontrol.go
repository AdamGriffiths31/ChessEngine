package search

import (
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	softTimeMinDepth = 4

	nextIterationFraction = 0.5

	stabilityMaxIterations = 10
)

func softTimeScale(iters []Iteration) float64 {
	n := len(iters)
	if n < softTimeMinDepth {
		return 1
	}
	last := iters[n-1]

	stable := 0
	for i := n - 2; i >= 0 && stable < stabilityMaxIterations; i-- {
		if !sameMove(iters[i].BestMove, last.BestMove) {
			break
		}
		stable++
	}
	stability := 1.20 - 0.04*float64(stable)

	drop := float64(iters[n-4].Score - last.Score)
	trend := min(max(0.05*drop, 0.75), 1.25)

	return stability * trend
}

func sameMove(a, b board.Move) bool {
	return a.From == b.From && a.To == b.To && a.Promotion == b.Promotion
}

func shouldStopOnSoftTime(soft, elapsed time.Duration, iters []Iteration) bool {
	if soft <= 0 || len(iters) < softTimeMinDepth {
		return false
	}
	budget := float64(soft) * softTimeScale(iters) * nextIterationFraction
	return float64(elapsed) >= budget
}
