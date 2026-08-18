package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/epd"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// runProfile provides CPU and memory profiling utilities for the chess engine
// (formerly cmd/profile).
//
// NOTE: the three log.Fatalf calls below (CPU profile start failure, memory
// profile create failure, memory profile write failure) are intentionally
// NOT converted to returned errors. By the time those calls execute, one or
// two defers (closing the CPU profile file / stopping the CPU profiler) have
// already been registered earlier in this function. Converting those three
// call sites to `return err` would make this function return normally,
// which would run those pending defers -- a behavior change from the
// original (os.Exit skips defers) that goes beyond a trivial mechanical
// conversion. The four earlier log.Fatal/log.Fatalf calls (missing -file,
// read failure, parse failure, out-of-range position) and the CPU-profile
// os.Create failure all occur before any defer is registered in this
// function, so those were converted to returned errors as trivial.
//
//nolint:gocyclo // linear sequence of flag/mode branches in a CLI entrypoint;
// see the defer-ordering note above for why it isn't split further
func runProfile(args []string) error {
	fs := flag.NewFlagSet("bench profile", flag.ExitOnError)
	var (
		stsFile     = fs.String("file", "", "STS EPD file to benchmark")
		position    = fs.Int("pos", 1, "Position number in the file (1-based)")
		count       = fs.Int("count", 1, "Number of positions to sample evenly across the file, profiled as one combined capture (ignores -pos when > 1)")
		searchTime  = fs.Duration("time", 10*time.Second, "Search time per position")
		cpuProfile  = fs.String("cpuprofile", "", "Write CPU profile to file")
		memProfile  = fs.String("memprofile", "", "Write memory profile to file")
		ttSize      = fs.Int("tt", 256, "Transposition table size in MB")
		showDetails = fs.Bool("details", false, "Show detailed search information")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *stsFile == "" {
		return fmt.Errorf("please specify an STS file with -file")
	}

	content, err := os.ReadFile(*stsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	positions, err := epd.ParseEPDFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse EPD file: %v", err)
	}

	if *count > 1 {
		return runProfileSample(positions, *stsFile, *count, *searchTime, *cpuProfile, *memProfile, *ttSize, *showDetails)
	}

	if *position < 1 || *position > len(positions) {
		return fmt.Errorf("position %d out of range (1-%d)", *position, len(positions))
	}

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

	b := epdPos.Board

	player := movegen.White
	if b.GetSideToMove() == "b" {
		player = movegen.Black
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			return fmt.Errorf("failed to create CPU profile: %v", err)
		}
		defer func() { _ = f.Close() }()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	engine := search.NewMinimaxEngine()
	engine.SetTranspositionTableSize(*ttSize)

	config := search.SearchConfig{
		MaxDepth: 999,
		MaxTime:  *searchTime,
	}

	ctx := context.Background()
	startTime := time.Now()

	result := engine.FindBestMove(ctx, b, player, config)

	searchDuration := time.Since(startTime)

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

		ttStats := engine.TTStats()
		fmt.Printf("\nTransposition Table:\n")
		fmt.Printf("Hits: %d\n", ttStats.Hits)
		fmt.Printf("Misses: %d\n", ttStats.Misses)
		fmt.Printf("Collisions: %d\n", ttStats.Collisions)
		fmt.Printf("Hit rate: %.1f%%\n", ttStats.HitRate)
	}

	moveStr := fmt.Sprintf("%s%s", result.BestMove.From.String(), result.BestMove.To.String())

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

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatalf("Failed to create memory profile: %v", err)
		}
		defer func() { _ = f.Close() }()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("Failed to write memory profile: %v", err)
		}
	}

	if *cpuProfile != "" {
		fmt.Printf("\nCPU profile written to %s\n", *cpuProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", *cpuProfile)
	}
	if *memProfile != "" {
		fmt.Printf("\nMemory profile written to %s\n", *memProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", *memProfile)
	}
	return nil
}

// runProfileSample profiles requestedCount positions sampled evenly across
// positions (by index) as one combined CPU/memory capture, so a single
// profile reflects a spread of position types rather than one tree shape.
// Each position gets a fresh engine (no shared TT/killer/history state
// across unrelated positions).
func runProfileSample(positions []*epd.Position, fileName string, requestedCount int, searchTime time.Duration, cpuProfile, memProfile string, ttSize int, showDetails bool) error {
	indices := evenlySpacedIndices(requestedCount, len(positions))
	fmt.Printf("\nProfiling %d position(s) sampled from %s (of %d total)\n", len(indices), fileName, len(positions))
	fmt.Printf("Search time per position: %v, TT: %dMB\n\n", searchTime, ttSize)

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return fmt.Errorf("failed to create CPU profile: %v", err)
		}
		defer func() { _ = f.Close() }()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
	}

	var totalNodes int64
	var totalDuration time.Duration
	var totalNullMoves, totalNullCutoffs, totalLMRReductions, totalLMRReSearches int64
	var totalTTHits, totalTTMisses uint64
	scored, matched := 0, 0

	for i, idx := range indices {
		epdPos := positions[idx]
		b := epdPos.Board
		player := movegen.White
		if b.GetSideToMove() == "b" {
			player = movegen.Black
		}

		engine := search.NewMinimaxEngine()
		engine.SetTranspositionTableSize(ttSize)
		config := search.SearchConfig{MaxDepth: 999, MaxTime: searchTime}

		start := time.Now()
		result := engine.FindBestMove(context.Background(), b, player, config)
		elapsed := time.Since(start)

		moveStr := fmt.Sprintf("%s%s", result.BestMove.From.String(), result.BestMove.To.String())
		nps := float64(result.Stats.NodesSearched) / elapsed.Seconds()

		status := ""
		if len(epdPos.MoveScores) > 0 {
			scored++
			for _, ms := range epdPos.MoveScores {
				if moveStr == ms.Move {
					matched++
					status = fmt.Sprintf("check %d/10", ms.Points)
					break
				}
			}
			if status == "" {
				status = "x not scored"
			}
		} else if epdPos.BestMove != "" {
			scored++
			if moveStr == epdPos.BestMove {
				matched++
				status = "check"
			} else {
				status = fmt.Sprintf("x expected %s", epdPos.BestMove)
			}
		}

		fmt.Printf("[%d/%d] pos %d (%s): move=%s depth=%d nodes=%d nps=%.0f %s\n",
			i+1, len(indices), idx+1, epdPos.ID, moveStr, result.Stats.Depth, result.Stats.NodesSearched, nps, status)

		totalNodes += result.Stats.NodesSearched
		totalDuration += elapsed
		totalNullMoves += result.Stats.NullMoves
		totalNullCutoffs += result.Stats.NullCutoffs
		totalLMRReductions += result.Stats.LMRReductions
		totalLMRReSearches += result.Stats.LMRReSearches
		ttStats := engine.TTStats()
		totalTTHits += ttStats.Hits
		totalTTMisses += ttStats.Misses
	}

	fmt.Printf("\nAggregate: %d positions, %d nodes, %.0f avg NPS, %v total time\n",
		len(indices), totalNodes, float64(totalNodes)/totalDuration.Seconds(), totalDuration)
	if scored > 0 {
		fmt.Printf("Matched best/scored move: %d/%d\n", matched, scored)
	}
	if showDetails {
		fmt.Printf("\nAggregate Statistics:\n")
		fmt.Printf("Null moves: %d\n", totalNullMoves)
		fmt.Printf("Null cutoffs: %d\n", totalNullCutoffs)
		fmt.Printf("LMR reductions: %d\n", totalLMRReductions)
		fmt.Printf("LMR re-searches: %d\n", totalLMRReSearches)
		if totalTTHits+totalTTMisses > 0 {
			fmt.Printf("TT hit rate: %.1f%%\n", 100*float64(totalTTHits)/float64(totalTTHits+totalTTMisses))
		}
	}

	if cpuProfile != "" {
		pprof.StopCPUProfile()
		fmt.Printf("\nCPU profile written to %s\n", cpuProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", cpuProfile)
	}

	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatalf("Failed to create memory profile: %v", err)
		}
		defer func() { _ = f.Close() }()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("Failed to write memory profile: %v", err)
		}
		fmt.Printf("Memory profile written to %s\n", memProfile)
		fmt.Printf("Analyze with: go tool pprof -http=:8080 %s\n", memProfile)
	}

	return nil
}

// evenlySpacedIndices returns up to count indices spread evenly across
// [0, total). If count >= total, all indices are returned. Rounding can
// produce duplicate indices when count is large relative to total, so the
// result may contain fewer than count entries.
func evenlySpacedIndices(count, total int) []int {
	if total <= 0 {
		return nil
	}
	if count >= total {
		indices := make([]int, total)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}
	if count == 1 {
		return []int{0}
	}
	seen := make(map[int]bool, count)
	indices := make([]int, 0, count)
	step := float64(total-1) / float64(count-1)
	for i := range count {
		idx := int(math.Round(float64(i) * step))
		if !seen[idx] {
			seen[idx] = true
			indices = append(indices, idx)
		}
	}
	return indices
}
