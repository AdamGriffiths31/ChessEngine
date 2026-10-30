package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

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

type cancelAfterEvals struct {
	eval.Evaluator
	remaining int
	cancel    context.CancelFunc
}

func (c *cancelAfterEvals) Evaluate(b *board.Board) eval.EvaluationScore {
	c.remaining--
	if c.remaining == 0 {
		c.cancel()
	}
	return c.Evaluator.Evaluate(b)
}

func TestInterruptedIterationKeepsFinishedRootMove(t *testing.T) {
	const fen = "8/p4P1K/rb1N2Bp/p3P3/N3R1P1/1P1k1P2/1Pbrnp2/2R2Q2 w - - 0 1"

	for evals := 20_000; evals <= 200_000; evals += 5_000 {
		ctx, cancel := context.WithCancel(context.Background())
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(64)
		engine.SetEvaluator(&cancelAfterEvals{Evaluator: eval.NewEvaluator(), remaining: evals, cancel: cancel})

		b := testutil.MustFromFEN(t, fen)
		player := playerFromSide(b.GetSideToMove())
		result := engine.FindBestMove(ctx, b, player, Config{MaxDepth: 99})
		cancel()

		iterations := result.Stats.Iterations
		if len(iterations) == 0 || iterations[len(iterations)-1].Score >= eval.MateScore-MateDistanceThreshold {
			continue
		}
		if result.Score >= eval.MateScore-MateDistanceThreshold {
			assertLegalMove(t, engine, b, player, result.BestMove)
			if pv := result.Stats.PrincipalVariation; len(pv) == 0 || pv[0] != result.BestMove {
				t.Errorf("PV %v does not start with best move %s", pv, moveString(result.BestMove))
			}
			return
		}
	}
	t.Fatal("no interruption point reported the mate proven by a finished root move")
}
