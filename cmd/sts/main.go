// Package main implements the STS (Strategic Test Suite) benchmark for chess engine evaluation.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/epd"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
)

const (
	defaultSTSFile = "testdata/STS2.epd"
)

func main() {
	depth := flag.Int("depth", 5, "Search depth for each position")
	timeout := flag.Int("timeout", 5, "Timeout per position in seconds")
	epdFile := flag.String("file", defaultSTSFile, "Path to EPD file")
	maxPositions := flag.Int("max", 0, "Maximum number of positions to test (0 = all)")
	verbose := flag.Bool("verbose", false, "Show detailed results for each position")
	ttSize := flag.Int("ttsize", 256, "Transposition table size in MB")
	clearTT := flag.Bool("clear-tt", true, "Clear transposition table between positions (recommended for EPD benchmarks)")
	threads := flag.Int("threads", 1, "Number of search threads for parallel search")
	flag.Parse()

	fmt.Printf("STS (Strategic Test Suite) Benchmark\n")
	fmt.Printf("====================================\n")
	fmt.Printf("Search Depth: %d\n", *depth)
	fmt.Printf("Timeout per position: %d seconds\n", *timeout)
	fmt.Printf("Transposition Table: %d MB\n", *ttSize)
	fmt.Printf("Clear TT between positions: %v\n", *clearTT)
	fmt.Printf("Search Threads: %d\n", *threads)
	if *maxPositions > 0 {
		fmt.Printf("Max positions: %d\n", *maxPositions)
	}
	fmt.Printf("Verbose output: %v\n\n", *verbose)

	fmt.Printf("Loading EPD file: %s\n", *epdFile)
	content, err := os.ReadFile(*epdFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read EPD file %s: %v\n", *epdFile, err)
		os.Exit(1)
	}
	epdContent := string(content)

	positions, err := epd.ParseEPDFile(epdContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse EPD file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d positions\n", len(positions))

	if *maxPositions > 0 && *maxPositions < len(positions) {
		positions = positions[:*maxPositions]
		fmt.Printf("Testing first %d positions\n", *maxPositions)
	}

	engine := search.NewMinimaxEngine()
	evaluator := evaluation.NewEvaluator()
	engine.SetEvaluator(evaluator)

	if *ttSize > 0 {
		engine.SetTranspositionTableSize(*ttSize)
		fmt.Printf("Initialized transposition table: %d MB\n", *ttSize)
	}

	searchConfig := ai.SearchConfig{
		MaxDepth:    *depth,
		MaxTime:     time.Duration(*timeout) * time.Second,
		DebugMode:   false,
		LMRMinDepth: 3,
		LMRMinMoves: 4,
		NumThreads:  *threads,
	}

	scorer := epd.NewSTSScorerWithTTClear(engine, searchConfig, *verbose, *clearTT)

	fmt.Printf("\nRunning STS benchmark...\n")
	fmt.Printf("------------------------\n")

	if *verbose {
		fmt.Printf("%-4s %-10s %-10s %-6s %-8s %-8s %-6s %s\n",
			"#", "Best Move", "Engine", "Score", "Time", "Nodes", "Depth", "Comment")
		fmt.Printf("-------------------------------------------------------------------------------------------\n")
	}

	startTime := time.Now()
	ctx := context.Background()
	results := scorer.ScoreSuite(ctx, positions, "STS Benchmark")
	totalTime := time.Since(startTime)

	displayResults(results, *verbose, totalTime)
}

// displayResults shows the benchmark results in a formatted way
func displayResults(results epd.STSSuiteResult, verbose bool, totalTime time.Duration) {
	fmt.Printf("\nSTS Benchmark Results\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Suite: %s\n", results.SuiteName)
	fmt.Printf("Positions tested: %d\n", results.PositionCount)
	fmt.Printf("Total score: %d/%d (%.1f%%)\n",
		results.TotalScore, results.MaxScore, results.ScorePercent)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Average time per position: %v\n", totalTime/time.Duration(results.PositionCount))

	stsRating := calculateSTSRating(results.ScorePercent)
	fmt.Printf("Approximate STS Rating: %d\n", stsRating)

	showPerformanceCategory(stsRating)

	if verbose {
		fmt.Printf("\nDetailed Position Results:\n")
		fmt.Printf("--------------------------\n")
		fmt.Printf("%-4s %-10s %-10s %-6s %-8s %-8s %-6s %s\n",
			"#", "Best Move", "Engine", "Score", "Time", "Nodes", "Depth", "Comment")
		fmt.Printf("-------------------------------------------------------------------------------------------\n")

		for i, result := range results.Results {
			display := ""
			if result.Position.ID != "" {
				display = result.Position.ID
			}
			if result.Position.Comment != "" {
				if display != "" {
					display += " | " + result.Position.Comment
				} else {
					display = result.Position.Comment
				}
			}

			if display != "" {
				display += " | FEN: " + result.Position.Board.ToFEN()
			} else {
				display = "FEN: " + result.Position.Board.ToFEN()
			}

			nodesStr := formatNodes(result.SearchResult.Stats.NodesSearched)

			fmt.Printf("%-4d %-10s %-10s %-6d %-8v %-8s %-6d %s\n",
				i+1,
				result.Position.BestMove,
				result.EngineMoveStr,
				result.Score,
				result.TestDuration.Round(time.Millisecond),
				nodesStr,
				result.SearchResult.Stats.Depth,
				display)
		}
	}

	fmt.Printf("\nSummary Statistics:\n")
	fmt.Printf("-------------------\n")
	correctMoves := 0
	totalNodes := int64(0)
	totalDepth := 0
	for _, result := range results.Results {
		if result.Score == 10 {
			correctMoves++
		}
		totalNodes += result.SearchResult.Stats.NodesSearched
		totalDepth += result.SearchResult.Stats.Depth
	}
	fmt.Printf("Correct moves (10 points): %d/%d (%.1f%%)\n",
		correctMoves, results.PositionCount,
		float64(correctMoves)/float64(results.PositionCount)*100.0)

	partialCredit := results.TotalScore - (correctMoves * 10)
	fmt.Printf("Partial credit points: %d\n", partialCredit)

	avgScore := float64(results.TotalScore) / float64(results.PositionCount)
	fmt.Printf("Average score per position: %.2f/10\n", avgScore)

	avgDepth := float64(totalDepth) / float64(results.PositionCount)
	nps := float64(totalNodes) / totalTime.Seconds()
	fmt.Printf("Average depth: %.1f\n", avgDepth)
	fmt.Printf("Total nodes: %s\n", formatNodes(totalNodes))
	fmt.Printf("Nodes per second: %s\n", formatNodes(int64(nps)))
}

// calculateSTSRating estimates STS rating based on percentage score.
// This is a simplified approximation based on known engine performance.
func calculateSTSRating(scorePercent float64) int {

	if scorePercent >= 90 {
		return 3400 + int((scorePercent-90)*20)
	} else if scorePercent >= 80 {
		return 3200 + int((scorePercent-80)*20)
	} else if scorePercent >= 70 {
		return 3000 + int((scorePercent-70)*20)
	} else if scorePercent >= 60 {
		return 2700 + int((scorePercent-60)*30)
	} else if scorePercent >= 50 {
		return 2400 + int((scorePercent-50)*30)
	}
	return int(2000 + scorePercent*8)
}

func showPerformanceCategory(rating int) {
	fmt.Printf("Performance Category: ")

	switch {
	case rating >= 3400:
		fmt.Printf("Elite (GM+ level)\n")
	case rating >= 3200:
		fmt.Printf("Very Strong (Strong GM level)\n")
	case rating >= 3000:
		fmt.Printf("Strong (IM+ level)\n")
	case rating >= 2700:
		fmt.Printf("Good (Expert+ level)\n")
	case rating >= 2400:
		fmt.Printf("Decent (Club level)\n")
	default:
		fmt.Printf("Weak (Beginner level)\n")
	}

	fmt.Printf("Reference: Stockfish 8+ typically scores 3300-3400+\n")
}

func formatNodes(nodes int64) string {
	if nodes >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(nodes)/1000000.0)
	} else if nodes >= 1000 {
		return fmt.Sprintf("%.1fK", float64(nodes)/1000.0)
	}
	return fmt.Sprintf("%d", nodes)
}
