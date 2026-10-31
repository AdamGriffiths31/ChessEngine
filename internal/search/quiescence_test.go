package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// White has three captures here, and Black has none after Qxb7, so a child
// sharing the parent's move-order buffer would empty it mid-loop.
func TestQuiescenceStopsAtTheLastBufferPly(t *testing.T) {
	m := NewMinimaxEngine()
	b := testutil.MustFromFEN(t, "4k3/1p2p2p/8/8/4Q3/8/8/4K3 w - - 0 1")
	m.searchState.board = b

	got := m.quiescence(context.Background(), movegen.White, -scoreInfinity, scoreInfinity, MaxKillerDepth-1)

	if want := m.evaluator.Evaluate(b); got != want {
		t.Errorf("quiescence at ply %d = %d, want static eval %d", MaxKillerDepth-1, got, want)
	}
}

func TestQuiescenceStopsBeyondTheLastBufferPly(t *testing.T) {
	m := NewMinimaxEngine()
	b := testutil.MustFromFEN(t, "4k3/1p2p2p/8/8/4Q3/8/8/4K3 b - - 0 1")
	m.searchState.board = b

	got := m.quiescence(context.Background(), movegen.Black, -scoreInfinity, scoreInfinity, MaxKillerDepth)

	if want := -m.evaluator.Evaluate(b); got != want {
		t.Errorf("quiescence at ply %d = %d, want static eval %d", MaxKillerDepth, got, want)
	}
}
