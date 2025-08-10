package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// BenchmarkLMRvsNoLMR compares search performance with and without LMR
func BenchmarkLMRvsNoLMR(b *testing.B) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(32) // Larger TT for benchmarks
	
	// Use a tactical position for benchmarking
	board, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		b.Fatalf("Failed to load FEN: %v", err)
	}
	
	baseConfig := ai.SearchConfig{
		MaxDepth:         5,
		MaxTime:          time.Second * 30,
		UseNullMove:      true,
	}
	
	ctx := context.Background()
	
	b.Run("WithoutLMR", func(b *testing.B) {
		config := baseConfig
		config.UseLMR = false
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := engine.FindBestMove(ctx, board, moves.White, config)
			if result.BestMove.From.File == -1 {
				b.Error("Expected to find a best move")
			}
		}
	})
	
	b.Run("WithLMR", func(b *testing.B) {
		config := baseConfig
		config.UseLMR = true
		config.LMRMinDepth = 2
		config.LMRMinMoves = 3
		config.LMRReductionBase = 0.75
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := engine.FindBestMove(ctx, board, moves.White, config)
			if result.BestMove.From.File == -1 {
				b.Error("Expected to find a best move")
			}
		}
	})
}

// BenchmarkLMRConfigurations tests different LMR parameter configurations
func BenchmarkLMRConfigurations(b *testing.B) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(16)
	
	board, err := board.FromFEN("rnbqkb1r/pppp1ppp/5n2/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 2 3")
	if err != nil {
		b.Fatalf("Failed to load FEN: %v", err)
	}
	
	baseConfig := ai.SearchConfig{
		MaxDepth:     4,
		MaxTime:      time.Second * 10,
		UseLMR:       true,
	}
	
	configurations := []struct {
		name         string
		minDepth     int
		minMoves     int
		reductionBase float64
	}{
		{"Conservative", 4, 6, 0.5},
		{"Moderate", 3, 4, 0.75},
		{"Aggressive", 2, 2, 1.0},
	}
	
	ctx := context.Background()
	
	for _, config := range configurations {
		b.Run(config.name, func(b *testing.B) {
			searchConfig := baseConfig
			searchConfig.LMRMinDepth = config.minDepth
			searchConfig.LMRMinMoves = config.minMoves
			searchConfig.LMRReductionBase = config.reductionBase
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := engine.FindBestMove(ctx, board, moves.White, searchConfig)
				if result.BestMove.From.File == -1 {
					b.Error("Expected to find a best move")
				}
			}
		})
	}
}

// BenchmarkLMRNodeReduction measures the actual node reduction achieved by LMR
func BenchmarkLMRNodeReduction(b *testing.B) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(32)
	
	// Use multiple positions for a comprehensive test
	positions := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
		"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 1",
	}
	
	config := ai.SearchConfig{
		MaxDepth:         4,
		MaxTime:          time.Second * 15,
		UseNullMove:      true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      3,
		LMRReductionBase: 0.75,
	}
	
	ctx := context.Background()
	
	var totalNodesWithLMR int64
	var totalLMRReductions int64
	var totalLMRReSearches int64
	var totalNodesSkipped int64
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		for _, fen := range positions {
			board, err := board.FromFEN(fen)
			if err != nil {
				b.Fatalf("Failed to load FEN %s: %v", fen, err)
			}
			
			result := engine.FindBestMove(ctx, board, moves.White, config)
			
			totalNodesWithLMR += result.Stats.NodesSearched
			totalLMRReductions += result.Stats.LMRReductions
			totalLMRReSearches += result.Stats.LMRReSearches
			totalNodesSkipped += result.Stats.LMRNodesSkipped
		}
	}
	
	b.StopTimer()
	
	// Report statistics
	avgNodesPerSearch := totalNodesWithLMR / int64(b.N*len(positions))
	avgReductionsPerSearch := totalLMRReductions / int64(b.N*len(positions))
	avgReSearchesPerSearch := totalLMRReSearches / int64(b.N*len(positions))
	avgNodesSkippedPerSearch := totalNodesSkipped / int64(b.N*len(positions))
	
	var reSearchRate float64
	if totalLMRReductions > 0 {
		reSearchRate = float64(totalLMRReSearches) / float64(totalLMRReductions) * 100
	}
	
	b.Logf("LMR Performance Statistics:")
	b.Logf("  Average nodes per search: %d", avgNodesPerSearch)
	b.Logf("  Average reductions per search: %d", avgReductionsPerSearch)
	b.Logf("  Average re-searches per search: %d", avgReSearchesPerSearch)
	b.Logf("  Average nodes skipped per search: %d", avgNodesSkippedPerSearch)
	b.Logf("  Re-search rate: %.2f%%", reSearchRate)
	
	if totalLMRReductions == 0 {
		b.Error("No LMR reductions occurred - check LMR implementation")
	}
}

// BenchmarkLMRDepthEffect tests how LMR effectiveness changes with search depth
func BenchmarkLMRDepthEffect(b *testing.B) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)
	
	board, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		b.Fatalf("Failed to load FEN: %v", err)
	}
	
	baseConfig := ai.SearchConfig{
		MaxTime:          time.Second * 20,
		UseNullMove:      true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      3,
		LMRReductionBase: 0.75,
	}
	
	depths := []int{3, 4, 5}
	ctx := context.Background()
	
	for _, depth := range depths {
		b.Run(fmt.Sprintf("Depth%d", depth), func(b *testing.B) {
			config := baseConfig
			config.MaxDepth = depth
			
			var totalReductions int64
			var totalNodes int64
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := engine.FindBestMove(ctx, board, moves.White, config)
				totalReductions += result.Stats.LMRReductions
				totalNodes += result.Stats.NodesSearched
			}
			b.StopTimer()
			
			avgReductions := float64(totalReductions) / float64(b.N)
			avgNodes := float64(totalNodes) / float64(b.N)
			
			b.Logf("Depth %d - Avg reductions: %.2f, Avg nodes: %.0f", depth, avgReductions, avgNodes)
		})
	}
}