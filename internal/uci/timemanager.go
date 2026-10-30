package uci

import "time"

const (
	// DefaultMoveOverhead is the default "Move Overhead": time kept back per move for latency.
	DefaultMoveOverhead = 50 * time.Millisecond

	suddenDeathHorizon = 40
	maxMovesToGo       = 50

	incrementShare  = 0.8
	softClockCap    = 0.5
	hardToSoftRatio = 4.0
	hardClockCap    = 0.8

	minSoftTime = 10 * time.Millisecond
)

type timeBudget struct {
	Soft time.Duration
	Hard time.Duration
}

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

	if hard > usable {
		hard = usable
		soft = min(soft, hard)
	}

	return timeBudget{Soft: soft, Hard: hard}
}
