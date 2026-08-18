package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
	"github.com/AdamGriffiths31/ChessEngine/internal/epd"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

const (
	defaultSTSFile = "testdata/STS2.epd"
)

func runSTS(args []string) error {
	fs := flag.NewFlagSet("bench sts", flag.ExitOnError)
	depth := fs.Int("depth", 5, "Search depth for each position")
	timeout := fs.Int("timeout", 5, "Timeout per position in seconds")
	epdFile := fs.String("file", defaultSTSFile, "Path to EPD file")
	maxPositions := fs.Int("max", 0, "Maximum number of positions to test (0 = all)")
	verbose := fs.Bool("verbose", false, "Show detailed results for each position")
	ttSize := fs.Int("ttsize", 256, "Transposition table size in MB")
	clearTT := fs.Bool("clear-tt", true, "Clear transposition table between positions (recommended for EPD benchmarks)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Printf("STS (Strategic Test Suite) Benchmark\n")
	fmt.Printf("====================================\n")
	fmt.Printf("Search Depth: %d\n", *depth)
	fmt.Printf("Timeout per position: %d seconds\n", *timeout)
	fmt.Printf("Transposition Table: %d MB\n", *ttSize)
	fmt.Printf("Clear TT between positions: %v\n", *clearTT)
	if *maxPositions > 0 {
		fmt.Printf("Max positions: %d\n", *maxPositions)
	}
	fmt.Printf("Verbose output: %v\n\n", *verbose)

	fmt.Printf("Loading EPD file: %s\n", *epdFile)
	content, err := os.ReadFile(*epdFile)
	if err != nil {
		return fmt.Errorf("Failed to read EPD file %s: %v", *epdFile, err)
	}
	epdContent := string(content)

	positions, err := epd.ParseEPDFile(epdContent)
	if err != nil {
		return fmt.Errorf("Failed to parse EPD file: %v", err)
	}

	fmt.Printf("Loaded %d positions\n", len(positions))

	if *maxPositions > 0 && *maxPositions < len(positions) {
		positions = positions[:*maxPositions]
		fmt.Printf("Testing first %d positions\n", *maxPositions)
	}

	engine := search.NewMinimaxEngine()
	evaluator := eval.NewEvaluator()
	engine.SetEvaluator(evaluator)

	if *ttSize > 0 {
		engine.SetTranspositionTableSize(*ttSize)
		fmt.Printf("Initialized transposition table: %d MB\n", *ttSize)
	}

	searchConfig := search.SearchConfig{
		MaxDepth:  *depth,
		MaxTime:   time.Duration(*timeout) * time.Second,
		DebugMode: false,
	}

	scorer := epd.NewSTSScorerWithTTClear(engine, searchConfig, *verbose, *clearTT)

	rootPath, err := os.Getwd()
	if err != nil {
		rootPath = "."
	}
	jsonl, err := bench.NewSTSResultsWriter(rootPath)
	if err != nil {
		return fmt.Errorf("failed to open JSONL results file: %v", err)
	}
	defer func() {
		if err := jsonl.Close(); err != nil {
			slog.Warn("failed to close JSONL results file", "path", jsonl.Path, "error", err)
		}
	}()

	onResult := func(r epd.STSResult) {
		if err := jsonl.WriteLine(bench.NewSTSPositionRecord(r)); err != nil {
			slog.Warn("failed to write STS JSONL position record", "error", err)
		}
	}

	fmt.Printf("\nRunning STS benchmark...\n")
	fmt.Printf("------------------------\n")

	if *verbose {
		fmt.Printf("%-4s %-10s %-10s %-6s %-8s %-8s %-6s %s\n",
			"#", "Best Move", "Engine", "Score", "Time", "Nodes", "Depth", "Comment")
		fmt.Printf("-------------------------------------------------------------------------------------------\n")
	}

	startTime := time.Now()
	ctx := context.Background()
	results := scorer.ScoreSuite(ctx, positions, "STS Benchmark", onResult)
	totalTime := time.Since(startTime)

	displayResults(results, *verbose, totalTime)

	stsRating := calculateSTSRating(results.ScorePercent)
	agg := bench.ComputeSTSAggregates(results.Results, totalTime)
	avgTime := time.Duration(0)
	if results.PositionCount > 0 {
		avgTime = totalTime / time.Duration(results.PositionCount)
	}
	summary := bench.STSSummaryRecord{
		Type:            "summary",
		Timestamp:       time.Now().Format("2006-01-02 15:04"),
		GitCommit:       bench.GitCommit(rootPath),
		Files:           []string{*epdFile},
		FileDescription: *epdFile,
		PositionCount:   results.PositionCount,
		TotalScore:      results.TotalScore,
		MaxScore:        results.MaxScore,
		ScorePercent:    results.ScorePercent,
		CorrectMoves:    agg.CorrectMoves,
		STSRating:       stsRating,
		Depth:           *depth,
		TimeoutSeconds:  *timeout,
		AvgTimeMS:       avgTime.Milliseconds(),
		TotalTimeMS:     totalTime.Milliseconds(),
		NodesPerSecond:  agg.NodesPerSec,
		AverageDepth:    agg.AvgDepth,
		Notes:           fmt.Sprintf("depth=%d, timeout=%ds, max=%d, clear-tt=%v", *depth, *timeout, *maxPositions, *clearTT),
	}
	if err := jsonl.WriteLine(summary); err != nil {
		slog.Warn("failed to write STS JSONL summary record", "error", err)
	}

	fmt.Printf("\nJSONL results: %s\n", jsonl.Path)
	return nil
}

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
	agg := bench.ComputeSTSAggregates(results.Results, totalTime)
	fmt.Printf("Correct moves (10 points): %d/%d (%.1f%%)\n",
		agg.CorrectMoves, results.PositionCount,
		float64(agg.CorrectMoves)/float64(results.PositionCount)*100.0)

	partialCredit := results.TotalScore - (agg.CorrectMoves * 10)
	fmt.Printf("Partial credit points: %d\n", partialCredit)

	avgScore := float64(results.TotalScore) / float64(results.PositionCount)
	fmt.Printf("Average score per position: %.2f/10\n", avgScore)

	fmt.Printf("Average depth: %.1f\n", agg.AvgDepth)
	fmt.Printf("Total nodes: %s\n", formatNodes(agg.TotalNodes))
	fmt.Printf("Nodes per second: %s\n", formatNodes(int64(agg.NodesPerSec)))
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
