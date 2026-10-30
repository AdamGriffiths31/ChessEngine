package bench

import (
	"strings"
	"testing"
)

func TestParseMatetrack(t *testing.T) {
	input := strings.Join([]string{
		"5K2/8/2qk4/2nPp3/3r4/6B1/B7/3R4 w - e6 bm #1; ep; 00:00;",
		`8/8/8/8/8/8/8/k1K5 b - - bm #-3; c0 "from a game";`,
		"8/8/8/8/8/8/8/k1K5 w - - ce 0;", // no bm: skipped
		"",
	}, "\n")

	got, err := ParseMatetrack(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMatetrack: %v", err)
	}
	want := []MatePosition{
		{FEN: "5K2/8/2qk4/2nPp3/3r4/6B1/B7/3R4 w - e6 0 1", BestMate: 1},
		{FEN: "8/8/8/8/8/8/8/k1K5 b - - 0 1", BestMate: -3},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d positions, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestClassifyMate(t *testing.T) {
	tests := []struct {
		name             string
		engine, bestMate int
		want             MateOutcome
	}{
		{"no mate reported", 0, 5, MateNotFound},
		{"exact best", 5, 5, MateBest},
		{"longer than best", 7, 5, MateFound},
		{"shorter than proven best", 3, 5, MateWrong},
		{"wrong side", -5, 5, MateWrong},
		{"mated: exact", -4, -4, MateBest},
		{"mated: delays longer than best", -6, -4, MateFound},
		{"mated: sooner than proven", -2, -4, MateWrong},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyMate(tc.engine, tc.bestMate); got != tc.want {
				t.Errorf("ClassifyMate(%d, %d) = %v, want %v", tc.engine, tc.bestMate, got, tc.want)
			}
		})
	}
}

func TestMateSummaryAdd(t *testing.T) {
	var s MateSummary
	for _, o := range []MateOutcome{MateBest, MateFound, MateNotFound, MateWrong, MateBest} {
		s.Add(o)
	}
	want := MateSummary{Positions: 5, Found: 3, Best: 2, Wrong: 1}
	if s != want {
		t.Errorf("summary = %+v, want %+v", s, want)
	}
}
