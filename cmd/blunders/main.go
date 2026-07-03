// Package main implements the blunder scanner: it replays games recorded by
// the Elo benchmark and uses full-strength Stockfish to flag positions where
// ChessEngine's move lost significant evaluation, emitting them as an EPD
// regression suite consumable by cmd/sts.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/benchmark"
)

func main() {
	input := flag.String("input", "elo_results.jsonl", "JSONL file of recorded Elo runs")
	output := flag.String("out", "blunders.epd", "EPD output path (appended to)")
	threshold := flag.Int("threshold", 150, "centipawn loss to flag as a blunder")
	movetime := flag.Int("movetime", 200, "Stockfish movetime per evaluation in ms")
	flag.Parse()

	rootPath, err := os.Getwd()
	if err != nil {
		rootPath = "."
	}
	manager := benchmark.NewEngineManager(rootPath)
	if err := manager.LoadEngines(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load engines.json: %v\n", err)
		os.Exit(1)
	}
	stockfish, err := manager.FindEngineByCommandSubstring("stockfish")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find stockfish in engines.json: %v\n", err)
		os.Exit(1)
	}

	cfg := benchmark.BlunderScanConfig{ThresholdCP: *threshold, MoveTimeMs: *movetime}

	newEval := func() (benchmark.PositionEval, func(), error) {
		client, err := benchmark.NewStockfishClient(stockfish.Command)
		if err != nil {
			return nil, nil, err
		}
		eval := func(history []string) (string, int, error) {
			return client.EvalPosition(history, *movetime)
		}
		return eval, func() { _ = client.Close() }, nil
	}

	fmt.Printf("Scanning %s (threshold %dcp, movetime %dms)...\n", *input, *threshold, *movetime)
	summary, err := benchmark.ScanRecordedGames(*input, *output, cfg, newEval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nRuns: %d  Games scanned: %d  Skipped (no moves): %d  Blunders: %d\n",
		summary.RunsScanned, summary.GamesScanned, summary.GamesSkipped, summary.BlundersFound)
	fmt.Printf("Output: %s (run it with: tools/bin/sts -file %s)\n", *output, *output)
}
