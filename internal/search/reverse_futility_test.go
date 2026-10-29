package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

const rfpStartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func TestTryReverseFutility_Fires(t *testing.T) {
	m := NewMinimaxEngine()
	margin := m.searchState.searchParams.RFPMargin

	t.Run("white to move", func(t *testing.T) {
		b := testutil.MustFromFEN(t, rfpStartFEN)
		score, done := m.tryReverseFutility(b, movegen.White, 2, 99, 100, 600, false)
		want := 600 - 2*margin
		if !done || score != want {
			t.Errorf("= (%d, done=%v), want (%d, true)", score, done, want)
		}
		if m.searchState.searchStats.RFPCutoffs != 1 {
			t.Errorf("RFPCutoffs = %d, want 1", m.searchState.searchStats.RFPCutoffs)
		}
	})

	t.Run("black to move", func(t *testing.T) {
		b := testutil.MustFromFEN(t, rfpStartFEN)
		score, done := m.tryReverseFutility(b, movegen.Black, 2, 99, 100, 600, false)
		if !done || score != 600-2*margin {
			t.Errorf("= (%d, done=%v), want (%d, true)", score, done, 600-2*margin)
		}
	})
}

func TestTryReverseFutility_Declines(t *testing.T) {
	tests := []struct {
		name       string
		fen        string
		depth      int
		alpha      eval.EvaluationScore
		beta       eval.EvaluationScore
		staticEval eval.EvaluationScore
		inCheck    bool
		disable    bool
	}{
		{"in check", rfpStartFEN, 2, 99, 100, 600, true, false},
		{"PV window", rfpStartFEN, 2, 0, 200, 600, false, false},
		{"too deep", rfpStartFEN, 99, 99, 100, 600, false, false},
		{"depth zero", rfpStartFEN, 0, 99, 100, 600, false, false},
		{"eval margin too small", rfpStartFEN, 2, 99, 100, 150, false, false},
		{"beta near mate", rfpStartFEN, 2, eval.MateScore - 1, eval.MateScore, 600, false, false},
		{"no non-pawn material", "6k1/8/8/8/8/8/P7/K7 w - - 0 1", 2, 99, 100, 600, false, false},
		{"disabled", rfpStartFEN, 2, 99, 100, 600, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMinimaxEngine()
			m.searchState.searchParams.RFPEnabled = !tt.disable
			b := testutil.MustFromFEN(t, tt.fen)
			score, done := m.tryReverseFutility(b, movegen.White, tt.depth, tt.alpha, tt.beta, tt.staticEval, tt.inCheck)
			if done || score != 0 {
				t.Errorf("= (%d, done=%v), want it to decline", score, done)
			}
			if m.searchState.searchStats.RFPCutoffs != 0 {
				t.Errorf("RFPCutoffs = %d, want 0", m.searchState.searchStats.RFPCutoffs)
			}
		})
	}
}
