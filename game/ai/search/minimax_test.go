package search

import (
	"context"
	"strings"
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
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}
	config := ai.SearchConfig{
		MaxDepth: 1,
		MaxTime:  time.Second,
	}

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
	threadState := engine.getThreadLocalState()
	engine.orderMoves(b, legalMoves, 0, board.Move{}, threadState)

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
	config := ai.SearchConfig{
		MaxDepth: 2,
		MaxTime:  time.Second,
	}

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
		MaxDepth:        4,
		MaxTime:         time.Second,
		DisableNullMove: false,
	}

	startTime := time.Now()
	resultWithNull := engine.FindBestMove(ctx, b, moves.White, configWithNull)
	timeWithNull := time.Since(startTime)

	// Test with null moves disabled
	configWithoutNull := ai.SearchConfig{
		MaxDepth:        4,
		MaxTime:         time.Second,
		DisableNullMove: true,
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
		MaxDepth:        3,
		MaxTime:         time.Second,
		DisableNullMove: false,
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
		MaxDepth:        4, // Deep enough to trigger adaptive null move reduction
		DisableNullMove: false,
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
		MaxDepth:        4,
		DisableNullMove: false,
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

// TestEngineWithValidation creates an engine configured like uci/engine.go and validates best moves
func TestEngineWithValidation(t *testing.T) {
	tests := []struct {
		name         string
		fen          string
		player       moves.Player
		timeout      time.Duration
		expectedMove string // UCI format move (e.g., "e2e4")
		description  string
	}{
		{
			name:         "Complex endgame position - 5s",
			fen:          "1r6/3pk1P1/4pp2/p1p1n3/4P1P1/1P1P4/B1P1RRK1/3r4 w - - 5 35",
			player:       moves.White,
			timeout:      5 * time.Second,
			expectedMove: "", // Let the engine decide the best move
			description:  "Complex endgame with tactical possibilities (5 second search)",
		},
		{
			name:         "Complex endgame position - 10s",
			fen:          "1r6/3pk1P1/4pp2/p1p1n3/4P1P1/1P1P4/B1P1RRK1/3r4 w - - 5 35",
			player:       moves.White,
			timeout:      10 * time.Second,
			expectedMove: "", // Let the engine decide the best move
			description:  "Complex endgame with tactical possibilities (10 second search)",
		},
		{
			name:         "Alpha-beta debug - depth 1",
			fen:          "1r6/3pk1P1/4pp2/p1p1n3/4P1P1/1P1P4/B1P1RRK1/3r4 w - - 5 35",
			player:       moves.White,
			timeout:      50 * time.Millisecond,
			expectedMove: "", // Let the engine decide the best move
			description:  "Debug alpha-beta pruning at depth 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and configure engine like uci/engine.go does
			engine := NewMinimaxEngine()
			// Enable TT - our PVS fix should handle TT interference properly
			engine.SetTranspositionTableSize(256) // Same as UCI engine

			// Clear any previous search state to avoid stale TT data
			engine.ClearSearchState()

			// Create board from FEN
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			// Configure search like uci/engine.go
			maxDepth := 999 // Let time limit control depth
			if strings.Contains(tt.name, "debug") {
				maxDepth = 1 // Force depth 1 for debug test
			}

			config := ai.SearchConfig{
				MaxDepth:            maxDepth,
				MaxTime:             tt.timeout,
				DebugMode:           false,
				UseOpeningBook:      false, // Disable for deterministic testing
				BookSelectMode:      ai.BookSelectWeightedRandom,
				BookWeightThreshold: 1,
				LMRMinDepth:         3,
				LMRMinMoves:         4,
				NumThreads:          1, // Single threaded for reproducible results
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Find best move
			startTime := time.Now()
			result := engine.FindBestMove(ctx, b, tt.player, config)
			searchDuration := time.Since(startTime)

			// Validate result
			if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
				t.Errorf("Engine failed to find a valid move")
				return
			}

			// Convert to UCI format for comparison
			converter := NewMoveConverter()
			actualMoveUCI := converter.ToUCI(result.BestMove)

			// Log detailed results
			t.Logf("Position: %s", tt.fen)
			t.Logf("Description: %s", tt.description)
			t.Logf("Found move: %s (score: %d, depth: %d, nodes: %d, time: %.3fs)",
				actualMoveUCI, result.Score, result.Stats.Depth, result.Stats.NodesSearched, searchDuration.Seconds())

			// Validate specific expected moves if provided
			if tt.expectedMove != "" {
				if actualMoveUCI != tt.expectedMove {
					t.Errorf("Expected move %s, got %s", tt.expectedMove, actualMoveUCI)
				}
			}

			// Validate search completed successfully
			if result.Stats.NodesSearched == 0 {
				t.Errorf("Expected to search some nodes, got %d", result.Stats.NodesSearched)
			}

			if result.Stats.Depth == 0 {
				t.Errorf("Expected search depth > 0, got %d", result.Stats.Depth)
			}

			// Validate timeout wasn't exceeded significantly
			if searchDuration > tt.timeout+100*time.Millisecond {
				t.Errorf("Search took too long: %v (timeout was %v)", searchDuration, tt.timeout)
			}
		})
	}
}

// NewMoveConverter creates a UCI move converter (simplified version for testing)
func NewMoveConverter() *MoveConverter {
	return &MoveConverter{}
}

// MoveConverter handles UCI move format conversion
type MoveConverter struct{}

// ToUCI converts a board.Move to UCI notation
func (mc *MoveConverter) ToUCI(move board.Move) string {
	from := move.From.String()
	to := move.To.String()

	result := from + to

	// Add promotion
	if move.Promotion != board.Empty {
		switch move.Promotion {
		case board.WhiteQueen, board.BlackQueen:
			result += "q"
		case board.WhiteRook, board.BlackRook:
			result += "r"
		case board.WhiteBishop, board.BlackBishop:
			result += "b"
		case board.WhiteKnight, board.BlackKnight:
			result += "n"
		}
	}

	return result
}
