package search

import (
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

func iterations(moves []board.Move, scores []eval.EvaluationScore) []Iteration {
	its := make([]Iteration, len(moves))
	for i := range moves {
		its[i] = Iteration{Depth: i + 1, BestMove: moves[i], Score: scores[i]}
	}
	return its
}

func tcMove(from, to int) board.Move {
	return board.Move{From: board.Square{File: from, Rank: 1}, To: board.Square{File: to, Rank: 3}}
}

func TestSoftTimeScaleTooFewIterations(t *testing.T) {
	m := tcMove(0, 0)
	got := softTimeScale(iterations([]board.Move{m, m, m}, []eval.EvaluationScore{0, 0, 0}))
	if got != 1 {
		t.Errorf("scale = %v, want 1 before depth %d", got, softTimeMinDepth)
	}
}

func TestSoftTimeScaleStableMoveShrinks(t *testing.T) {
	m := tcMove(0, 0)
	flat := make([]eval.EvaluationScore, 8)
	stable := softTimeScale(iterations([]board.Move{m, m, m, m, m, m, m, m}, flat))
	unstable := softTimeScale(iterations([]board.Move{tcMove(1, 1), m, tcMove(2, 2), m, tcMove(3, 3), m, tcMove(4, 4), m}, flat))
	if stable >= 1 {
		t.Errorf("stable scale = %v, want < 1", stable)
	}
	if unstable <= stable {
		t.Errorf("unstable scale %v should exceed stable %v", unstable, stable)
	}
}

func TestSoftTimeScaleScoreDropGrows(t *testing.T) {
	m := tcMove(0, 0)
	moves := []board.Move{m, m, m, m, m}
	steady := softTimeScale(iterations(moves, []eval.EvaluationScore{50, 50, 50, 50, 50}))
	falling := softTimeScale(iterations(moves, []eval.EvaluationScore{50, 50, 30, 10, -20}))
	if falling <= steady {
		t.Errorf("falling-score scale %v should exceed steady %v", falling, steady)
	}
}

func TestSoftTimeScaleBounds(t *testing.T) {
	m := tcMove(0, 0)
	moves := make([]board.Move, 20)
	scores := make([]eval.EvaluationScore, 20)
	for i := range moves {
		moves[i] = m
		scores[i] = eval.EvaluationScore(-i * 100)
	}
	if got := softTimeScale(iterations(moves, scores)); got > 1.20*1.25+1e-9 {
		t.Errorf("scale = %v above upper bound", got)
	}
	for i := range scores {
		scores[i] = eval.EvaluationScore(i * 100)
	}
	if got := softTimeScale(iterations(moves, scores)); got < 0.80*0.75-1e-9 {
		t.Errorf("scale = %v below lower bound", got)
	}
}

func TestShouldStopOnSoftTime(t *testing.T) {
	m := tcMove(0, 0)
	its := iterations([]board.Move{m, m, m, m, m, m}, make([]eval.EvaluationScore, 6))

	if shouldStopOnSoftTime(0, time.Hour, its) {
		t.Error("zero soft limit must never stop the search")
	}
	if shouldStopOnSoftTime(time.Second, time.Hour, its[:2]) {
		t.Error("must not stop before the minimum depth")
	}
	if shouldStopOnSoftTime(time.Second, 10*time.Millisecond, its) {
		t.Error("stopped with almost no time used")
	}
	if !shouldStopOnSoftTime(time.Second, 900*time.Millisecond, its) {
		t.Error("should stop once past the soft share")
	}
}
