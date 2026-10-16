package uci

import "time"

const (
	// DefaultMoveOverhead is the default "Move Overhead": time kept back per move for latency.
	DefaultMoveOverhead = 50 * time.Millisecond

	// suddenDeathHorizon is the moves the clock is assumed to cover without movestogo.
	suddenDeathHorizon = 40
	maxMovesToGo       = 50

	incrementShare  = 0.8 // fraction of the increment added to the soft budget
	softClockCap    = 0.5 // soft budget cap, as a fraction of usable time
	hardToSoftRatio = 4.0
	hardClockCap    = 0.8 // hard limit cap, as a fraction of usable time

	minSoftTime = 10 * time.Millisecond
)

// timeBudget holds the per-move limits: Soft is checked between iterations,
// Hard cancels the search.
type timeBudget struct {
	Soft time.Duration
	Hard time.Duration
}

// allocateTime splits the clock into soft and hard limits for one move.
// movesToGo is zero for sudden death; overhead is subtracted from the clock first.
func allocateTime(timeLeft, increment time.Duration, movesToGo int, overhead time.Duration) timeBudget {
	usable := max(timeLeft-overhead, time.Millisecond)

	horizon := suddenDeathHorizon
	if movesToGo > 0 {
		horizon = min(movesToGo, maxMovesToGo)
	}

	soft := usable/time.Duration(horizon) + time.Duration(float64(increment)*incrementShare)
	soft = min(soft, time.Duration(float64(usable)*softClockCap))
	soft = max(soft, minSoftTime)

	hard := min(time.Duration(float64(soft)*hardToSoftRatio), time.Duration(float64(usable)*hardClockCap))
	hard = max(hard, soft)

	// The minimum floor must not promise more than the clock holds.
	if hard > usable {
		hard = usable
		soft = min(soft, hard)
	}

	return timeBudget{Soft: soft, Hard: hard}
}
