package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMinimaxEngineWithOpeningBook(t *testing.T) {
	// Test the engine with opening book integration
	engine := NewMinimaxEngine()
	
	// Create a test board in starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}
	
	// Check if performance.bin exists for testing
	bookPath := "../../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found at", bookPath)
		return
	}
	
	// Test with opening book disabled
	configWithoutBook := ai.SearchConfig{
		MaxDepth:            3,
		MaxTime:             time.Second,
		UseOpeningBook:      false,
		DebugMode:           true,
	}
	
	ctx := context.Background()
	resultWithoutBook := engine.FindBestMove(ctx, b, moves.White, configWithoutBook)
	
	if resultWithoutBook.Stats.NodesSearched == 0 {
		t.Error("Expected nodes to be searched when book is disabled")
	}
	if resultWithoutBook.Stats.Time == 0 {
		t.Error("Expected search time when book is disabled")
	}
	
	// Test with opening book enabled
	configWithBook := ai.SearchConfig{
		MaxDepth:            3,
		MaxTime:             time.Second,
		UseOpeningBook:      true,
		BookFiles:           []string{bookPath},
		BookSelectMode:      ai.BookSelectWeightedRandom,
		BookWeightThreshold: 1,
		DebugMode:           true,
	}
	
	resultWithBook := engine.FindBestMove(ctx, b, moves.White, configWithBook)
	
	// The book may or may not have the starting position, but the search should complete
	if resultWithBook.BestMove.From.File == -1 && resultWithBook.BestMove.From.Rank == -1 {
		t.Error("Expected a valid move from engine")
	}
	
	t.Logf("Without book: %d nodes, %v time", resultWithoutBook.Stats.NodesSearched, resultWithoutBook.Stats.Time)
	t.Logf("With book: %d nodes, %v time", resultWithBook.Stats.NodesSearched, resultWithBook.Stats.Time)
}

func TestMinimaxEngineBookMoveSelection(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test different book selection modes
	bookPath := "../../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}
	
	testCases := []struct {
		name string
		mode ai.BookSelectionMode
	}{
		{"Best Move", ai.BookSelectBest},
		{"Random Move", ai.BookSelectRandom},
		{"Weighted Random", ai.BookSelectWeightedRandom},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := ai.SearchConfig{
				MaxDepth:            2,
				MaxTime:             100 * time.Millisecond,
				UseOpeningBook:      true,
				BookFiles:           []string{bookPath},
				BookSelectMode:      tc.mode,
				BookWeightThreshold: 1,
			}
			
			ctx := context.Background()
			result := engine.FindBestMove(ctx, b, moves.White, config)
			
			// Should get a valid move regardless of whether it comes from book or search
			if result.BestMove.From.File < 0 || result.BestMove.From.File > 7 ||
			   result.BestMove.From.Rank < 0 || result.BestMove.From.Rank > 7 {
				t.Errorf("Invalid move returned: %+v", result.BestMove)
			}
			
			t.Logf("Mode %s: Move %c%d-%c%d, Nodes: %d, Time: %v", 
				tc.name,
				'a'+result.BestMove.From.File, result.BestMove.From.Rank+1,
				'a'+result.BestMove.To.File, result.BestMove.To.Rank+1,
				result.Stats.NodesSearched, result.Stats.Time)
		})
	}
}

func TestMinimaxEngineBookInitialization(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test with non-existent book file
	configBadFile := ai.SearchConfig{
		UseOpeningBook: true,
		BookFiles:      []string{"/nonexistent/file.bin"},
		MaxDepth:       2,
	}
	
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ctx := context.Background()
	
	// Should not crash and fall back to regular search
	result := engine.FindBestMove(ctx, b, moves.White, configBadFile)
	if result.Stats.NodesSearched == 0 {
		t.Error("Expected fallback to regular search when book loading fails")
	}
	
	// Test with empty book files
	configEmptyFiles := ai.SearchConfig{
		UseOpeningBook: true,
		BookFiles:      []string{},
		MaxDepth:       2,
	}
	
	result2 := engine.FindBestMove(ctx, b, moves.White, configEmptyFiles)
	if result2.Stats.NodesSearched == 0 {
		t.Error("Expected regular search when no book files provided")
	}
}

func TestMinimaxEngineBookDisabled(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test with book explicitly disabled
	config := ai.SearchConfig{
		UseOpeningBook: false,
		BookFiles:      []string{"some_file.bin"}, // Should be ignored
		MaxDepth:       3,
		MaxTime:        time.Second,
	}
	
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ctx := context.Background()
	
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should perform regular search
	if result.Stats.NodesSearched == 0 {
		t.Error("Expected nodes to be searched when book is disabled")
	}
	if result.Stats.Depth == 0 {
		t.Error("Expected positive search depth when book is disabled")
	}
}

