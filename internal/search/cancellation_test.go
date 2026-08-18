package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// TestCancellationResponsiveness verifies that the periodic (every ~1024
// nodes) cancellation check still stops a deep search promptly when its
// context is cancelled. A depth-99 search of the start position would run
// effectively forever without a cutoff; with a 500ms context it must return
// well before the search would otherwise complete, and it must still return a
// legal move (from the deepest fully-completed iteration).
//
// The assert budget is deliberately generous (2s) rather than tight to the
// 500ms deadline: the point is to catch a search that ignores cancellation
// entirely (which would run for many seconds/minutes), not to measure exact
// polling latency, which would make the test flaky under CI load.
func TestCancellationResponsiveness(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	player := playerFromSide(b.GetSideToMove())

	cfg := SearchConfig{MaxDepth: 99, MaxTime: 0, UseOpeningBook: false}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	result := engine.FindBestMove(ctx, b, player, cfg)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("search did not honor cancellation promptly: elapsed=%v (budget 2s)", elapsed)
	}

	if !result.BestMove.IsValid() {
		t.Fatalf("cancelled search returned no legal move: %+v", result.BestMove)
	}

	legal := false
	moves := engine.generator.GenerateAllMoves(b, player)
	for i := 0; i < moves.Count; i++ {
		if moves.Moves[i] == result.BestMove {
			legal = true
			break
		}
	}
	if !legal {
		t.Fatalf("cancelled search returned an illegal move %s", moveString(result.BestMove))
	}
}
