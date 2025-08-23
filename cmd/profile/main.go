// Package main provides profiling utilities for the chess engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/epd"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func main() {
	var (
		stsFile     = flag.String("file", "", "STS EPD file to benchmark")
		position    = flag.Int("pos", 1, "Position number in the file (1-based)")
		searchTime  = flag.Duration("time", 10*time.Second, "Search time per position")
		cpuProfile  = flag.String("cpuprofile", "", "Write CPU profile to file")
		memProfile  = flag.String("memprofile", "", "Write memory profile to file")
		ttSize      = flag.Int("tt", 256, "Transposition table size in MB")
		showDetails = flag.Bool("details", false, "Show detailed search information")
	)
	flag.Parse()

	if *stsFile == "" {
		log.Fatal("Please specify an STS file with -file")
	}

	// Load EPD file
	content, err := os.ReadFile(*stsFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	positions, err := epd.ParseEPDFile(string(content))
	if err != nil {
		log.Fatalf("Failed to parse EPD file: %v", err)
	}

	if *position < 1 || *position > len(positions) {
		log.Fatalf("Position %d out of range (1-%d)", *position, len(positions))
	}

	// Get the requested position
	epdPos := positions[*position-1]
	fmt.Printf("\nBenchmarking position %d/%d from %s\n", *position, len(positions), *stsFile)
	fmt.Printf("ID: %s\n", epdPos.ID)
	if epdPos.BestMove != "" {
		fmt.Printf("Best move: %s\n", epdPos.BestMove)
	}
	if len(epdPos.MoveScores) > 0 {
		fmt.Printf("Move scores: ")
		for _, ms := range epdPos.MoveScores {
			fmt.Printf("%s(%d) ", ms.Move, ms.Points)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("Search time: %v, TT: %dMB\n\n", *searchTime, *ttSize)

	// Use the parsed board
	b := epdPos.Board

	// Determine side to move from board state
	player := moves.White
	if b.GetSideToMove() == "b" {
		player = moves.Black
	}

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatalf("Failed to create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Create engine
	engine := search.NewMinimaxEngine()
	engine.SetTranspositionTableSize(*ttSize)

	// Configure search
	config := ai.SearchConfig{
		MaxDepth: 999,
		MaxTime:  *searchTime,
	}

	// Run the search
	ctx := context.Background()
	startTime := time.Now()

	result := engine.FindBestMove(ctx, b, player, config)

	searchDuration := time.Since(startTime)

	// Display results
	fmt.Printf("\nSearch Results:\n")
	fmt.Printf("Best move: %s%s\n", result.BestMove.From.String(), result.BestMove.To.String())
	fmt.Printf("Score: %d\n", result.Score)
	fmt.Printf("Depth: %d\n", result.Stats.Depth)
	fmt.Printf("Time: %v\n", searchDuration)
	fmt.Printf("Nodes: %d\n", result.Stats.NodesSearched)
	fmt.Printf("NPS: %.0f\n", float64(result.Stats.NodesSearched)/searchDuration.Seconds())

	if *showDetails {
		fmt.Printf("\nDetailed Statistics:\n")
		fmt.Printf("Null moves: %d\n", result.Stats.NullMoves)
		fmt.Printf("Null cutoffs: %d\n", result.Stats.NullCutoffs)
		fmt.Printf("LMR reductions: %d\n", result.Stats.LMRReductions)
		fmt.Printf("LMR re-searches: %d\n", result.Stats.LMRReSearches)

		// Get TT stats
		hits, misses, collisions, hitRate := engine.GetTranspositionTableStats()
		fmt.Printf("\nTransposition Table:\n")
		fmt.Printf("Hits: %d\n", hits)
		fmt.Printf("Misses: %d\n", misses)
		fmt.Printf("Collisions: %d\n", collisions)
		fmt.Printf("Hit rate: %.1f%%\n", hitRate)
	}

	// Check if the move matches expected best moves
	moveStr := fmt.Sprintf("%s%s", result.BestMove.From.String(), result.BestMove.To.String())

	// Calculate STS score based on move scores
	score := 0
	for _, ms := range epdPos.MoveScores {
		if moveStr == ms.Move {
			score = ms.Points
			break
		}
	}

	if len(epdPos.MoveScores) > 0 {
		if score > 0 {
			fmt.Printf("\n✓ STS Score: %d/10\n", score)
		} else {
			fmt.Printf("\n✗ Move not in STS scoring list\n")
		}
	} else if epdPos.BestMove != "" {
		if moveStr == epdPos.BestMove {
			fmt.Printf("\n✓ Found best move!\n")
		} else {
			fmt.Printf("\n✗ Expected: %s\n", epdPos.BestMove)
		}
	}

	// Memory profiling if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatalf("Failed to create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("Failed to write memory profile: %v", err)
		}
	}

	// Print profiling instructions
	if *cpuProfile != "" {
		fmt.Printf("\nCPU profile written to %s\n", *cpuProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", *cpuProfile)
	}
	if *memProfile != "" {
		fmt.Printf("\nMemory profile written to %s\n", *memProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", *memProfile)
	}
}
