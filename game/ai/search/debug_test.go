package search

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMinimaxEngineDebugInfo(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test debug info with opening book
	bookPath := "../../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}
	
	// Test with debug enabled
	configWithDebug := ai.SearchConfig{
		MaxDepth:            2,
		MaxTime:             100 * time.Millisecond,
		UseOpeningBook:      true,
		BookFiles:           []string{bookPath},
		BookSelectMode:      ai.BookSelectBest,
		BookWeightThreshold: 1,
		DebugMode:           true,
	}
	
	result := engine.FindBestMove(context.Background(), b, moves.White, configWithDebug)
	
	// Should have debug information
	if len(result.Stats.DebugInfo) == 0 {
		t.Error("Expected debug info when DebugMode is enabled")
	}
	
	// Should indicate if book move was used
	t.Logf("Book move used: %t", result.Stats.BookMoveUsed)
	t.Logf("Debug info:")
	for i, msg := range result.Stats.DebugInfo {
		t.Logf("  [%d] %s", i, msg)
	}
	
	// Test without debug mode
	configWithoutDebug := configWithDebug
	configWithoutDebug.DebugMode = false
	
	result2 := engine.FindBestMove(context.Background(), b, moves.White, configWithoutDebug)
	
	// Should not have debug information
	if len(result2.Stats.DebugInfo) != 0 {
		t.Error("Expected no debug info when DebugMode is disabled")
	}
	
	// But should still have BookMoveUsed flag
	if result.Stats.BookMoveUsed != result2.Stats.BookMoveUsed {
		t.Error("BookMoveUsed should be consistent regardless of debug mode")
	}
}

func TestMinimaxEngineDebugWithBookMove(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Create a mock book that will return a specific move
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Use a non-existent book file to test error handling
	config := ai.SearchConfig{
		MaxDepth:            2,
		UseOpeningBook:      true,
		BookFiles:           []string{"/nonexistent/file.bin"},
		DebugMode:           true,
	}
	
	result := engine.FindBestMove(context.Background(), b, moves.White, config)
	
	// Should have debug info about book loading failure
	found := false
	for _, msg := range result.Stats.DebugInfo {
		if strings.Contains(msg, "Opening book initialization failed") {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected debug message about book initialization failure")
	}
	
	// Should fall back to regular search
	if result.Stats.BookMoveUsed {
		t.Error("Should not use book move when initialization fails")
	}
	
	if result.Stats.NodesSearched == 0 {
		t.Error("Should perform regular search when book fails")
	}
	
	t.Logf("Debug messages for failed book loading:")
	for i, msg := range result.Stats.DebugInfo {
		t.Logf("  [%d] %s", i, msg)
	}
}