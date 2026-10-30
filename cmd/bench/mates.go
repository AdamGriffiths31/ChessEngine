package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
	"github.com/AdamGriffiths31/ChessEngine/internal/uci"
)

const defaultMatesFile = "testdata/matetrack.epd"

type mateResult struct {
	pos        bench.MatePosition
	engineMate int
	outcome    bench.MateOutcome
	elapsed    time.Duration
	err        error
}

func runMates(args []string) error {
	fs := flag.NewFlagSet("bench mates", flag.ExitOnError)
	file := fs.String("file", defaultMatesFile, "matetrack-format EPD (see README.md for where to get it)")
	moveTime := fs.Duration("time", 200*time.Millisecond, "search time per position")
	maxPositions := fs.Int("max", 0, "maximum number of positions to test (0 = all)")
	concurrency := fs.Int("concurrency", 1, "positions searched in parallel, one engine each")
	ttSize := fs.Int("ttsize", 16, "transposition table size per engine in MB")
	verbose := fs.Bool("verbose", false, "print every position that is not a best mate")
	if err := fs.Parse(args); err != nil {
		return err
	}

	f, err := os.Open(*file)
	if err != nil {
		return fmt.Errorf("open mate suite: %w", err)
	}
	positions, err := bench.ParseMatetrack(f)
	_ = f.Close()
	if err != nil {
		return fmt.Errorf("parse mate suite %s: %w", *file, err)
	}
	if *maxPositions > 0 && *maxPositions < len(positions) {
		positions = positions[:*maxPositions]
	}
	if *concurrency < 1 {
		*concurrency = 1
	}

	fmt.Printf("Mate suite: %s, %d positions, %v per position, concurrency %d\n",
		*file, len(positions), *moveTime, *concurrency)

	start := time.Now()
	results := make([]mateResult, len(positions))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range *concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine := search.NewMinimaxEngine()
			engine.SetTranspositionTableSize(*ttSize)
			config := search.Config{MaxDepth: uci.DefaultMaxDepth, MaxTime: *moveTime}
			for i := range jobs {
				results[i] = searchMate(engine, positions[i], config)
			}
		}()
	}
	for i := range positions {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	var summary bench.MateSummary
	var slowest mateResult
	for _, r := range results {
		if r.err != nil {
			return r.err
		}
		summary.Add(r.outcome)
		if r.elapsed > slowest.elapsed {
			slowest = r
		}
		if r.outcome == bench.MateWrong || (*verbose && r.outcome != bench.MateBest) {
			fmt.Printf("%-8s engine mate %4d, best %4d: %s\n", outcomeLabel(r.outcome), r.engineMate, r.pos.BestMate, r.pos.FEN)
		}
	}

	pct := func(n int) float64 { return 100 * float64(n) / float64(max(summary.Positions, 1)) }
	fmt.Printf("\nPositions: %d\n", summary.Positions)
	fmt.Printf("Found mates: %d (%.1f%%)\n", summary.Found, pct(summary.Found))
	fmt.Printf("Best mates:  %d (%.1f%%)\n", summary.Best, pct(summary.Best))
	fmt.Printf("Wrong mates: %d (shorter than proven or wrong side; should be 0)\n", summary.Wrong)
	fmt.Printf("Total time:  %v\n", time.Since(start).Round(time.Second))
	fmt.Printf("Slowest:     %v (budget %v): %s\n", slowest.elapsed.Round(time.Millisecond), *moveTime, slowest.pos.FEN)

	if summary.Wrong > 0 {
		return fmt.Errorf("%d impossible mate scores reported", summary.Wrong)
	}
	return nil
}

func searchMate(engine *search.MinimaxEngine, pos bench.MatePosition, config search.Config) mateResult {
	b, err := board.FromFEN(pos.FEN)
	if err != nil {
		return mateResult{err: fmt.Errorf("mate suite FEN %q: %w", pos.FEN, err)}
	}
	player := movegen.White
	if b.GetSideToMove() == "b" {
		player = movegen.Black
	}
	engine.ClearSearchState()
	start := time.Now()
	result := engine.FindBestMove(context.Background(), b, player, config)
	elapsed := time.Since(start)
	mate := mateMoves(int(result.Score))
	return mateResult{pos: pos, engineMate: mate, outcome: bench.ClassifyMate(mate, pos.BestMate), elapsed: elapsed}
}

func mateMoves(score int) int {
	mate := int(eval.MateScore)
	switch {
	case score >= mate-search.MateDistanceThreshold:
		return (mate - score + 1) / 2
	case score <= -mate+search.MateDistanceThreshold:
		return -((mate + score) / 2)
	default:
		return 0
	}
}

func outcomeLabel(o bench.MateOutcome) string {
	switch o {
	case bench.MateWrong:
		return "WRONG"
	case bench.MateFound:
		return "longer"
	case bench.MateNotFound:
		return "missed"
	default:
		return "best"
	}
}
