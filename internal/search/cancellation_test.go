package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// A depth-99 search would run effectively forever. With a 500ms context it must
// return promptly with a legal move from its last completed iteration. The 2s
// budget is deliberately loose: it catches a search that ignores cancellation,
// not exact polling latency, which would be flaky under CI load.
func TestCancellationResponsiveness(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	player := playerFromSide(b.GetSideToMove())

	cfg := Config{MaxDepth: 99, MaxTime: 0, UseOpeningBook: false}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	result := engine.FindBestMove(ctx, b, player, cfg)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("search did not honor cancellation promptly: elapsed=%v (budget 2s)", elapsed)
	}

	assertLegalMove(t, engine, b, player, result.BestMove)
}
