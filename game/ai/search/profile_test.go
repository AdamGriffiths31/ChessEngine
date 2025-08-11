package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// BenchmarkSearchProfile is a fast benchmark for CPU profiling
func BenchmarkSearchProfile(b *testing.B) {
	// Complex middlegame position
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	testBoard, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Invalid FEN: %s", fen)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(128)

	// Fast configuration for profiling
	config := ai.SearchConfig{
		MaxDepth:        6,                   // Reasonable depth
		MaxTime:         500 * time.Millisecond, // Short time for fast iterations
		DebugMode:       false,
		DisableNullMove: false, // Enable all optimizations
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
}

// BenchmarkSearchDepth4 - Quick depth 4 search for profiling
func BenchmarkSearchDepth4(b *testing.B) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	testBoard, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Invalid FEN: %s", fen)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)

	config := ai.SearchConfig{
		MaxDepth:        4, // Very fast
		MaxTime:         10 * time.Second, // Let depth control
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
}