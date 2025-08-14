// Package main benchmarks chess engine performance with and without transposition tables.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// BenchmarkPosition represents a chess position for performance testing.
type BenchmarkPosition struct {
	Name        string
	FEN         string
	Description string
}

// BenchmarkResult contains performance metrics from testing a position.
type BenchmarkResult struct {
	Position      BenchmarkPosition
	BestMove      board.Move
	Score         ai.EvaluationScore
	NodesSearched int64
	Depth         int
	Time          time.Duration
	TTHits        uint64
	TTMisses      uint64
	TTHitRate     float64
}

var standardPositions = []BenchmarkPosition{
	{
		Name:        "Starting Position",
		FEN:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Description: "Standard chess starting position",
	},
	{
		Name:        "Middlegame Position",
		FEN:         "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 4",
		Description: "Italian Game opening middlegame",
	},
	{
		Name:        "Tactical Position",
		FEN:         "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 4",
		Description: "Position with tactical opportunities",
	},
	{
		Name:        "Endgame Position",
		FEN:         "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		Description: "King and pawn endgame",
	},
}

func main() {
	depth := flag.Int("depth", 5, "Search depth for benchmarking")
	ttSize := flag.Int("ttsize", 256, "Transposition table size in MB")
	customPos := flag.String("fen", "", "Custom position FEN (optional)")
	timeout := flag.Int("timeout", 30, "Search timeout in seconds")
	flag.Parse()

	fmt.Printf("Chess Engine Benchmark\n")
	fmt.Printf("======================\n")
	fmt.Printf("Search Depth: %d\n", *depth)
	fmt.Printf("TT Size: %d MB\n", *ttSize)
	fmt.Printf("Timeout: %d seconds\n\n", *timeout)

	positions := standardPositions
	if *customPos != "" {
		positions = []BenchmarkPosition{
			{
				Name:        "Custom Position",
				FEN:         *customPos,
				Description: "User-provided position",
			},
		}
	}

	fmt.Println("Running benchmark WITHOUT transposition tables...")
	fmt.Println("------------------------------------------------")
	resultsWithoutTT := runBenchmark(positions, *depth, *timeout, 0)

	fmt.Println("\nRunning benchmark WITH transposition tables...")
	fmt.Println("---------------------------------------------")
	resultsWithTT := runBenchmark(positions, *depth, *timeout, *ttSize)

	compareResults(resultsWithoutTT, resultsWithTT)
}

func runBenchmark(positions []BenchmarkPosition, depth, timeoutSecs, ttSizeMB int) []BenchmarkResult {
	results := make([]BenchmarkResult, 0, len(positions))

	for _, pos := range positions {
		fmt.Printf("Testing: %s\n", pos.Name)

		engine := search.NewMinimaxEngine()
		if ttSizeMB > 0 {
			engine.SetTranspositionTableSize(ttSizeMB)
		}

		b, err := board.FromFEN(pos.FEN)
		if err != nil {
			fmt.Printf("Failed to parse FEN for position %s: %v\n", pos.Name, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
		defer cancel()

		config := ai.SearchConfig{
			MaxDepth:       depth,
			UseOpeningBook: false,
			DebugMode:      false,
		}

		startTime := time.Now()
		result := engine.FindBestMove(ctx, b, moves.White, config)
		searchTime := time.Since(startTime)

		var ttHits, ttMisses uint64
		var ttHitRate float64
		if ttSizeMB > 0 {
			ttHits, ttMisses, _, ttHitRate = engine.GetTranspositionTableStats()
		}

		benchResult := BenchmarkResult{
			Position:      pos,
			BestMove:      result.BestMove,
			Score:         result.Score,
			NodesSearched: result.Stats.NodesSearched,
			Depth:         result.Stats.Depth,
			Time:          searchTime,
			TTHits:        ttHits,
			TTMisses:      ttMisses,
			TTHitRate:     ttHitRate,
		}
		results = append(results, benchResult)

		fmt.Printf("  Best Move: %s%s\n", result.BestMove.From.String(), result.BestMove.To.String())
		fmt.Printf("  Score: %d\n", result.Score)
		fmt.Printf("  Nodes: %d\n", result.Stats.NodesSearched)
		fmt.Printf("  Time: %v\n", searchTime)
		if ttSizeMB > 0 {
			fmt.Printf("  TT Hit Rate: %.1f%%\n", ttHitRate)
		}
		fmt.Println()
	}

	return results
}

func compareResults(withoutTT, withTT []BenchmarkResult) {
	fmt.Println("\nComparison Results")
	fmt.Println("==================")

	if len(withoutTT) != len(withTT) {
		fmt.Println("Error: Result sets have different lengths")
		return
	}

	var totalNodesWithoutTT, totalNodesWithTT int64
	var totalTimeWithoutTT, totalTimeWithTT time.Duration

	fmt.Printf("%-20s %12s %12s %12s %12s %12s\n",
		"Position", "Nodes (No TT)", "Nodes (TT)", "Time (No TT)", "Time (TT)", "TT Hit Rate")
	fmt.Println("-------------------------------------------------------------------------------------")

	for i := 0; i < len(withoutTT); i++ {
		noTT := withoutTT[i]
		withTTResult := withTT[i]

		totalNodesWithoutTT += noTT.NodesSearched
		totalNodesWithTT += withTTResult.NodesSearched
		totalTimeWithoutTT += noTT.Time
		totalTimeWithTT += withTTResult.Time

		fmt.Printf("%-20s %12d %12d %12v %12v %11.1f%%\n",
			truncateString(noTT.Position.Name, 20),
			noTT.NodesSearched,
			withTTResult.NodesSearched,
			noTT.Time.Round(time.Millisecond),
			withTTResult.Time.Round(time.Millisecond),
			withTTResult.TTHitRate)
	}

	fmt.Println("-------------------------------------------------------------------------------------")
	fmt.Printf("%-20s %12d %12d %12v %12v\n",
		"TOTAL",
		totalNodesWithoutTT,
		totalNodesWithTT,
		totalTimeWithoutTT.Round(time.Millisecond),
		totalTimeWithTT.Round(time.Millisecond))

	nodeReduction := float64(totalNodesWithoutTT-totalNodesWithTT) / float64(totalNodesWithoutTT) * 100
	timeImprovement := float64(totalTimeWithoutTT-totalTimeWithTT) / float64(totalTimeWithoutTT) * 100

	fmt.Printf("\nPerformance Impact:\n")
	fmt.Printf("- Node reduction: %.1f%%\n", nodeReduction)
	fmt.Printf("- Time improvement: %.1f%%\n", timeImprovement)

	if nodeReduction > 0 {
		fmt.Printf("- Effective branching factor reduction: %.2fx\n",
			float64(totalNodesWithoutTT)/float64(totalNodesWithTT))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
