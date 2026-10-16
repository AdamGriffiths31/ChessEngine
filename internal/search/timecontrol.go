package search

import (
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// softTimeMinDepth is the first completed depth at which the soft limit may stop the search.
	softTimeMinDepth = 4

	// nextIterationFraction: an iteration costs about as much as all earlier
	// ones combined, so one started past this share of the soft limit rarely finishes.
	nextIterationFraction = 0.5

	stabilityMaxIterations = 10
)

// softTimeScale returns the factor applied to Config.SoftTime after the last
// iteration in iters: a stability factor (1.20 down to 0.80 as the best move
// repeats) times a score-trend factor (0.75 to 1.25, higher when the score fell
// over the last three iterations). Same signals as Ethereal.
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

// shouldStopOnSoftTime reports whether to skip the next iteration.
func shouldStopOnSoftTime(soft, elapsed time.Duration, iters []Iteration) bool {
	if soft <= 0 || len(iters) < softTimeMinDepth {
		return false
	}
	budget := float64(soft) * softTimeScale(iters) * nextIterationFraction
	return float64(elapsed) >= budget
}
