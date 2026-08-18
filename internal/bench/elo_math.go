package bench

import "math"

// computeAnchors returns count Elo levels centered on center and spaced by
// step, clamped to [minElo, maxElo].
func computeAnchors(center, count, step, minElo, maxElo int) []int {
	halfSpan := (count / 2) * step
	start := center - halfSpan
	if start < minElo {
		start = minElo
	}

	maxStart := maxElo - (count-1)*step
	if maxStart < minElo {
		maxStart = minElo
	}
	if start > maxStart {
		start = maxStart
	}

	anchors := make([]int, count)
	for i := 0; i < count; i++ {
		anchors[i] = start + i*step
	}
	return anchors
}

// expectedScore returns the logistic-curve expected score for a player
// rated playerRating against an opponent rated opponentRating.
func expectedScore(playerRating, opponentRating float64) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, (opponentRating-playerRating)/400.0))
}

// estimateRating binary-searches [400, 3500] for the rating whose expected
// score against anchorElos best matches the actual achieved scores.
func estimateRating(anchorElos []int, scores []float64) float64 {
	actualScore := 0.0
	for _, s := range scores {
		actualScore += s
	}

	lo, hi := 400.0, 3500.0
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2.0
		expectedTotal := 0.0
		for _, anchor := range anchorElos {
			expectedTotal += expectedScore(mid, float64(anchor))
		}
		if expectedTotal < actualScore {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2.0
}

// gameAnchorsAndScores builds the parallel anchor/score arrays estimateRating
// requires - one entry per game, not one entry per distinct anchor level.
// Passing computeAnchors' distinct-level output directly to estimateRating
// alongside a per-game scores slice silently caps expectedTotal at the
// number of anchor levels, forcing the estimate to saturate at the search
// ceiling for any real multi-game-per-anchor run.
func gameAnchorsAndScores(games []EloGameResult) (anchors []int, scores []float64) {
	anchors = make([]int, len(games))
	scores = make([]float64, len(games))
	for i, g := range games {
		anchors[i] = g.AnchorElo
		scores[i] = g.Result
	}
	return anchors, scores
}

// allocateTime returns the milliseconds to spend on a single move, given
// the remaining clock time and increment for one side.
func allocateTime(remainingMs, incMs int) int {
	allocated := remainingMs/30 + incMs
	if maxAllowed := remainingMs / 3; allocated > maxAllowed {
		allocated = maxAllowed
	}
	if allocated < 50 {
		allocated = 50
	}
	return allocated
}
