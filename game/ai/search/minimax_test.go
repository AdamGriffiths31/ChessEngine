package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestNewMinimaxEngine(t *testing.T) {
	engine := NewMinimaxEngine()

	if engine == nil {
		t.Fatal("NewMinimaxEngine should not return nil")
	}

	if engine.GetName() != "Minimax Engine" {
		t.Errorf("Expected name 'Minimax Engine', got '%s'", engine.GetName())
	}
}

func TestSetEvaluator(t *testing.T) {
	evaluator := evaluation.NewEvaluator()
	engine := NewMinimaxEngine()

	// Change evaluator
	engine.SetEvaluator(evaluator)

	// We can't directly check if evaluator changed, but we can verify it doesn't crash
	ctx := context.Background()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	config := ai.SearchConfig{MaxDepth: 1}

	result := engine.FindBestMove(ctx, b, moves.White, config)
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move after SetEvaluator")
	}
}

func TestHistoryHeuristicMoveOrdering(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetDebugMoveOrdering(true)

	// Create a position where we can test move ordering
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	// First, let's manually add some history scores to specific moves
	// Simulate that Nf3 has been successful in the past
	nf3 := board.Move{
		From:      board.Square{File: 6, Rank: 0}, // g1
		To:        board.Square{File: 5, Rank: 2}, // f3
		Piece:     board.WhiteKnight,
		Captured:  board.Empty,
		IsCapture: false,
	}
	engine.historyTable.UpdateHistory(nf3, 3) // Add significant history

	// Generate moves and check ordering
	generator := moves.NewGenerator()
	legalMoves := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(legalMoves)

	// Order moves using the engine's method
	engine.orderMoves(b, legalMoves, 0, board.Move{})

	// Get the ordered moves for debugging
	orderedMoves := engine.GetLastMoveOrder()

	// Find position of Nf3 in the ordered list
	nf3Position := -1
	for i, move := range orderedMoves {
		if move.From == nf3.From && move.To == nf3.To {
			nf3Position = i
			break
		}
	}

	if nf3Position == -1 {
		t.Error("Nf3 should be in the move list")
		return
	}

	// Nf3 should be relatively early in the list since it has history
	// It won't be first (no TT move, no captures), but should be before moves with no history
	totalMoves := len(orderedMoves)
	if nf3Position > totalMoves/2 {
		t.Errorf("Nf3 with history should be in first half of moves, found at position %d out of %d", nf3Position, totalMoves)
	}
}

func TestHistoryTableIntegrationWithSearch(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Simple position where we can force some beta cutoffs
	b, err := board.FromFEN("4k3/8/8/8/8/8/4P3/4K3 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	config := ai.SearchConfig{MaxDepth: 2}

	// Run a search - this should populate the history table
	result := engine.FindBestMove(ctx, b, moves.White, config)
	if result.BestMove.From.File == -1 {
		t.Error("Should find a valid move")
		return
	}

	// Check that history table has some entries
	hasHistoryEntries := false
	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			for toFile := 0; toFile < 8; toFile++ {
				for toRank := 0; toRank < 8; toRank++ {
					move := board.Move{
						From: board.Square{File: file, Rank: rank},
						To:   board.Square{File: toFile, Rank: toRank},
					}
					if engine.historyTable.GetHistoryScore(move) > 0 {
						hasHistoryEntries = true
						break
					}
				}
				if hasHistoryEntries {
					break
				}
			}
			if hasHistoryEntries {
				break
			}
		}
		if hasHistoryEntries {
			break
		}
	}

	if !hasHistoryEntries {
		t.Error("Expected history table to have some entries after search")
	}
}

func TestClearSearchStateWithHistory(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Add some history
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	engine.historyTable.UpdateHistory(move, 2)
	
	// Verify history exists
	if engine.historyTable.GetHistoryScore(move) == 0 {
		t.Error("Expected non-zero history score before clear")
	}
	
	// Clear search state
	engine.ClearSearchState()
	
	// Verify history is cleared
	if engine.historyTable.GetHistoryScore(move) != 0 {
		t.Error("Expected history to be cleared after ClearSearchState")
	}
}

func TestNullMovePruning(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test position where null move should cause pruning
	// Use a position where one side has a significant advantage
	b, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	
	// Test with null moves enabled
	configWithNull := ai.SearchConfig{
		MaxDepth:    4,
		UseNullMove: true,
	}
	
	startTime := time.Now()
	resultWithNull := engine.FindBestMove(ctx, b, moves.White, configWithNull)
	timeWithNull := time.Since(startTime)
	
	// Test with null moves disabled
	configWithoutNull := ai.SearchConfig{
		MaxDepth:    4,
		UseNullMove: false,
	}
	
	startTime = time.Now()
	resultWithoutNull := engine.FindBestMove(ctx, b, moves.White, configWithoutNull)
	timeWithoutNull := time.Since(startTime)
	
	// Both should find a valid move
	if resultWithNull.BestMove.From.File == -1 {
		t.Error("Search with null moves should find a valid move")
	}
	if resultWithoutNull.BestMove.From.File == -1 {
		t.Error("Search without null moves should find a valid move")
	}
	
	// Null move version should typically be faster (more pruning)
	// Allow some variance but expect significant speedup in most cases
	if timeWithNull > timeWithoutNull*2 {
		t.Logf("Warning: Null move search took %v, without null moves took %v. Expected null moves to be faster.", 
			timeWithNull, timeWithoutNull)
		// Not failing the test as performance can vary, but logging for observation
	}
	
	// Both should search some nodes
	if resultWithNull.Stats.NodesSearched == 0 {
		t.Error("Search with null moves should search some nodes")
	}
	if resultWithoutNull.Stats.NodesSearched == 0 {
		t.Error("Search without null moves should search some nodes")
	}
	
	t.Logf("Null move search: %d nodes in %v", resultWithNull.Stats.NodesSearched, timeWithNull)
	t.Logf("Without null move: %d nodes in %v", resultWithoutNull.Stats.NodesSearched, timeWithoutNull)
}

func TestNullMoveInCheck(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test position where king is in check - null move should not be attempted
	b, err := board.FromFEN("rnb1kbnr/pppp1ppp/8/4p3/6Pq/8/PPPPP1PP/RNBQKBNR w KQkq - 1 3")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	config := ai.SearchConfig{
		MaxDepth:    3,
		UseNullMove: true,
	}
	
	// This should not crash or cause issues, even with null moves enabled
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should find a valid move to get out of check
	if result.BestMove.From.File == -1 {
		t.Error("Should find a move to escape check")
	}
}

func TestNullMoveDeepSearch(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test with deeper search to ensure null move reduction works correctly
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	config := ai.SearchConfig{
		MaxDepth:    4, // Deep enough to trigger adaptive null move reduction
		UseNullMove: true,
	}
	
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should complete without errors and find a reasonable opening move
	if result.BestMove.From.File == -1 {
		t.Error("Deep search with null moves should find a valid move")
	}
	
	// Should reach a reasonable depth
	if result.Stats.Depth < 3 {
		t.Errorf("Expected to reach depth of at least 3, got %d", result.Stats.Depth)
	}
}

func TestNullMoveWithTranspositionTable(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1) // Enable TT with small size
	
	// Test that null moves work correctly with transposition table
	b, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	config := ai.SearchConfig{
		MaxDepth:    4,
		UseNullMove: true,
	}
	
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should find a valid move
	if result.BestMove.From.File == -1 {
		t.Error("Search with null moves and TT should find a valid move")
	}
	
	// Check that TT has some entries
	hits, misses, _, hitRate := engine.GetTranspositionTableStats()
	if hits+misses == 0 {
		t.Error("Expected some transposition table activity")
	}
	
	t.Logf("TT Stats: %d hits, %d misses, %.1f%% hit rate", hits, misses, hitRate*100)
}

func TestFindBestMoveForcedMate(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	// Position where white can capture black queen (White king on e1, Rook on d2, Black queen on d4)
	b, err := board.FromFEN("4k3/8/8/8/3q4/8/3R4/4K3 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	config := ai.SearchConfig{
		MaxDepth:       2,
		UseOpeningBook: false, // Disable opening book for this test
	}
	result := engine.FindBestMove(ctx, b, moves.White, config)

	// Should capture the queen (d2 to d4)
	expectedFrom := board.Square{File: 3, Rank: 1} // d2
	expectedTo := board.Square{File: 3, Rank: 3}   // d4

	if result.BestMove.From != expectedFrom || result.BestMove.To != expectedTo {
		t.Errorf("Expected move d2d4 to capture queen, got %s%s",
			result.BestMove.From.String(), result.BestMove.To.String())
	}
}

func TestOppositePlayer(t *testing.T) {
	if oppositePlayer(moves.White) != moves.Black {
		t.Error("Opposite of White should be Black")
	}
	if oppositePlayer(moves.Black) != moves.White {
		t.Error("Opposite of Black should be White")
	}
}
