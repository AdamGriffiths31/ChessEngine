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

// BenchmarkSearchRealistic benchmarks the search with realistic UCI-like settings
func BenchmarkSearchRealistic(b *testing.B) {
	positions := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Starting position
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", // Complex middlegame
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", // Endgame position
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", // Tactical position
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", // Balanced middlegame
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(128) // 128MB transposition table

	// Realistic UCI-style configuration with all optimizations
	config := ai.SearchConfig{
		MaxDepth:            8,               // Reasonable depth for benchmarking
		MaxTime:             5 * time.Second, // 5 second search
		DebugMode:           false,
		DisableNullMove:     false,           // Enable null move pruning
		LMRMinDepth:         3,               // Enable LMR at depth 3+
		LMRMinMoves:         4,               // Start reductions after 4 moves
		UseOpeningBook:      false,           // Disable book for pure search benchmark
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, fen := range positions {
			board, err := board.FromFEN(fen)
			if err != nil {
				b.Fatalf("Invalid FEN: %s", fen)
			}

			ctx, cancel := context.WithTimeout(context.Background(), config.MaxTime)
			result := engine.FindBestMove(ctx, board, moves.White, config)
			cancel()

			// Ensure we got a valid result
			if result.BestMove.From == result.BestMove.To {
				b.Fatalf("Invalid move returned for position: %s", fen)
			}
		}
	}
}

// BenchmarkSearchDepthComparison benchmarks different search depths
func BenchmarkSearchDepthComparison(b *testing.B) {
	// Complex middlegame position for consistent benchmarking
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	testBoard, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Invalid FEN: %s", fen)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(128)

	depths := []int{4, 6, 8, 10}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth_%d", depth), func(b *testing.B) {
			config := ai.SearchConfig{
				MaxDepth:        depth,
				MaxTime:         30 * time.Second, // Long timeout to let depth limit control
				DebugMode:       false,
				DisableNullMove: false,
				LMRMinDepth:     3,
				LMRMinMoves:     4,
				UseOpeningBook:  false,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), config.MaxTime)
				result := engine.FindBestMove(ctx, testBoard, moves.White, config)
				cancel()

				if result.BestMove.From == result.BestMove.To {
					b.Fatalf("Invalid move returned")
				}
			}
		})
	}
}

// BenchmarkOptimizationComparison compares performance with/without optimizations
func BenchmarkOptimizationComparison(b *testing.B) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	testBoard, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Invalid FEN: %s", fen)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(128)

	configs := map[string]ai.SearchConfig{
		"all_optimizations": {
			MaxDepth:        6,
			MaxTime:         10 * time.Second,
			DebugMode:       false,
			DisableNullMove: false, // Null move enabled
			LMRMinDepth:     3,     // LMR enabled
			LMRMinMoves:     4,
			UseOpeningBook:  false,
		},
		"no_null_move": {
			MaxDepth:        6,
			MaxTime:         10 * time.Second,
			DebugMode:       false,
			DisableNullMove: true, // Null move disabled
			LMRMinDepth:     3,
			LMRMinMoves:     4,
			UseOpeningBook:  false,
		},
		"no_lmr": {
			MaxDepth:        6,
			MaxTime:         10 * time.Second,
			DebugMode:       false,
			DisableNullMove: false,
			LMRMinDepth:     999, // LMR effectively disabled
			LMRMinMoves:     999,
			UseOpeningBook:  false,
		},
		"no_optimizations": {
			MaxDepth:        6,
			MaxTime:         10 * time.Second,
			DebugMode:       false,
			DisableNullMove: true, // Both disabled
			LMRMinDepth:     999,
			LMRMinMoves:     999,
			UseOpeningBook:  false,
		},
	}

	for name, config := range configs {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), config.MaxTime)
				result := engine.FindBestMove(ctx, testBoard, moves.White, config)
				cancel()

				if result.BestMove.From == result.BestMove.To {
					b.Fatalf("Invalid move returned")
				}
			}
		})
	}
}