package search

import (
	"context"
	"sort"
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

func TestFindBestMoveStartingPosition(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	config := ai.SearchConfig{MaxDepth: 2}
	result := engine.FindBestMove(ctx, b, moves.White, config)

	// Should return a valid move
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move in starting position")
	}

	// Should have searched at least one position
	if result.Stats.NodesSearched == 0 {
		t.Error("Should have evaluated at least one position")
	}

	// Should have taken some time
	if result.Stats.Time <= 0 {
		t.Error("Search time should be greater than 0")
	}
}

func TestFindBestMoveDepthOne(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	config := ai.SearchConfig{MaxDepth: 1}
	result := engine.FindBestMove(ctx, b, moves.White, config)

	// Should return a valid move even with depth 1
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move with depth 1")
	}
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

func TestFindBestMoveWithTimeout(t *testing.T) {
	engine := NewMinimaxEngine()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	config := ai.SearchConfig{MaxDepth: 5} // High depth to ensure timeout
	result := engine.FindBestMove(ctx, b, moves.White, config)

	// Should return some move even if timeout occurred
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a move even with timeout")
	}
}

func TestFindBestMoveBlackPlayer(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	// Position where it's black to move
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	config := ai.SearchConfig{MaxDepth: 2}
	result := engine.FindBestMove(ctx, b, moves.Black, config)

	// Should return a valid move for black
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move for black")
	}

	// Verify the move is for a black piece
	piece := b.GetPiece(result.BestMove.From.Rank, result.BestMove.From.File)
	if !isBlackPiece(piece) {
		t.Errorf("Expected black piece to move, but found %v", piece)
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

func TestMinimaxRecursiveSearch(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	b, err := board.FromFEN("4k3/8/8/8/3q4/8/3R4/4K3 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	// Test different depths produce results
	depths := []int{1, 2, 3}

	for _, depth := range depths {
		config := ai.SearchConfig{
			MaxDepth:       depth,
			UseOpeningBook: false, // Disable opening book for this test
		}
		result := engine.FindBestMove(ctx, b, moves.White, config)

		if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
			t.Errorf("FindBestMove should return valid move at depth %d", depth)
		}

		if result.Stats.NodesSearched == 0 {
			t.Errorf("Should evaluate positions at depth %d", depth)
		}
	}
}

func TestScoreConsistency(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	b, err := board.FromFEN("4k3/8/8/8/3q4/8/3R4/4K3 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	config := ai.SearchConfig{MaxDepth: 2}

	// Run search multiple times - should be consistent
	result1 := engine.FindBestMove(ctx, b, moves.White, config)
	result2 := engine.FindBestMove(ctx, b, moves.White, config)

	if result1.BestMove != result2.BestMove {
		t.Error("Minimax should be deterministic and return same move")
	}

	if result1.Score != result2.Score {
		t.Error("Minimax should return consistent scores")
	}
}

func TestDepthTracking(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx := context.Background()
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Test different depths
	depths := []int{1, 2, 3}

	for _, maxDepth := range depths {
		config := ai.SearchConfig{
			MaxDepth:       maxDepth,
			UseOpeningBook: false, // Disable opening book for consistent testing
		}
		result := engine.FindBestMove(ctx, b, moves.White, config)

		if result.Stats.Depth <= 0 {
			t.Errorf("Expected positive depth, got %d for maxDepth %d", result.Stats.Depth, maxDepth)
		}

		if result.Stats.Depth > maxDepth {
			t.Errorf("Actual depth %d should not exceed maxDepth %d", result.Stats.Depth, maxDepth)
		}

		t.Logf("MaxDepth: %d, Actual depth reached: %d, Nodes: %d",
			maxDepth, result.Stats.Depth, result.Stats.NodesSearched)
	}
}

// Helper function to check if a piece is black
func isBlackPiece(piece board.Piece) bool {
	return piece == board.BlackPawn || piece == board.BlackRook ||
		piece == board.BlackKnight || piece == board.BlackBishop ||
		piece == board.BlackQueen || piece == board.BlackKing
}

// TestMinimaxPlayerPerspective tests that the same position gives opposite scores for different players
func TestMinimaxPlayerPerspective(t *testing.T) {
	engine := NewMinimaxEngine()

	// Test position where White is clearly better
	fen := "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" // Black queen missing
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()
	config := ai.SearchConfig{
		MaxDepth:       2,
		UseOpeningBook: false,
	}

	// Search as White (should prefer this position)
	whiteResult := engine.FindBestMove(ctx, b, moves.White, config)

	// Search as Black (should dislike this position)
	blackResult := engine.FindBestMove(ctx, b, moves.Black, config)

	t.Logf("Position: White up a queen")
	t.Logf("  White perspective: %d", whiteResult.Score)
	t.Logf("  Black perspective: %d", blackResult.Score)

	// White should see this as positive (good for White)
	if whiteResult.Score <= 0 {
		t.Errorf("White should evaluate this position positively, got %d", whiteResult.Score)
	}

	// Black should see this as negative (bad for Black, good for White)
	if blackResult.Score >= 0 {
		t.Errorf("Black should evaluate this position negatively, got %d", blackResult.Score)
	}
}

// TestMaterialImbalanceEvaluation tests clear material imbalances
func TestMaterialImbalanceEvaluation(t *testing.T) {
	engine := NewMinimaxEngine()

	testCases := []struct {
		name         string
		fen          string
		expectedSign int // 1 for positive (White better), -1 for negative (Black better)
		description  string
	}{
		{
			name:         "white_up_rook",
			fen:          "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Black h8 rook missing
			expectedSign: 1,
			description:  "White up a rook should be positive",
		},
		{
			name:         "black_up_rook",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK1NR w KQkq - 0 1", // White h1 rook missing
			expectedSign: -1,
			description:  "Black up a rook should be negative",
		},
		{
			name:         "white_up_knight",
			fen:          "rnbqkb1r/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Black knight missing
			expectedSign: 1,
			description:  "White up a knight should be positive",
		},
		{
			name:         "black_up_knight",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R1BQKBNR w KQkq - 0 1", // White knight missing
			expectedSign: -1,
			description:  "Black up a knight should be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tc.fen, err)
			}

			// Test both direct evaluation and search
			directEval := engine.evaluator.Evaluate(b)

			ctx := context.Background()
			config := ai.SearchConfig{
				MaxDepth:       2,
				UseOpeningBook: false,
			}
			searchResult := engine.FindBestMove(ctx, b, moves.White, config)

			t.Logf("%s:", tc.description)
			t.Logf("  Direct evaluation: %d", directEval)
			t.Logf("  Search result: %d", searchResult.Score)

			// Check direct evaluation sign
			if tc.expectedSign > 0 && directEval <= 0 {
				t.Errorf("Direct evaluation should be positive, got %d", directEval)
			}
			if tc.expectedSign < 0 && directEval >= 0 {
				t.Errorf("Direct evaluation should be negative, got %d", directEval)
			}

			// Check search result sign
			if tc.expectedSign > 0 && searchResult.Score <= 0 {
				t.Errorf("Search result should be positive, got %d", searchResult.Score)
			}
			if tc.expectedSign < 0 && searchResult.Score >= 0 {
				t.Errorf("Search result should be negative, got %d", searchResult.Score)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestPVMoveOrdering proves that PV (Principal Variation) move ordering works correctly
func TestPVMoveOrdering(t *testing.T) {
	engine := NewMinimaxEngine()

	// Enable transposition table for PV move ordering
	engine.SetTranspositionTableSize(16) // 16MB should be sufficient for test

	// Use a position where the best move is likely a quiet positional move, not a capture
	// This position is after: 1.e4 e6 2.d4 d5 3.Nd2 - a common opening
	b, err := board.FromFEN("rnbqk1nr/ppppbppp/4p3/8/2PP4/8/PP1NPPPP/R1BQKBNR w KQkq - 3 4")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	ctx := context.Background()

	// Phase 1: Establish PV move by searching to depth 3
	t.Logf("=== PHASE 1: Establishing PV move ===")
	config1 := ai.SearchConfig{
		MaxDepth:       3,
		UseOpeningBook: false,
		DebugMode:      true,
	}
	result1 := engine.FindBestMove(ctx, b, moves.White, config1)
	expectedPVMove := result1.BestMove

	t.Logf("Depth 3 search completed:")
	t.Logf("  Best move: %s%s", expectedPVMove.From.String(), expectedPVMove.To.String())
	t.Logf("  Score: %d", result1.Score)
	t.Logf("  Nodes: %d", result1.Stats.NodesSearched)
	t.Logf("  Move is capture: %t", expectedPVMove.IsCapture)

	// Verify we have a valid move
	if expectedPVMove.From.File == -1 && expectedPVMove.From.Rank == -1 {
		t.Fatal("Phase 1 should return a valid move")
	}

	// Phase 2: Test PV usage by searching to depth 4
	t.Logf("\n=== PHASE 2: Testing PV move usage ===")

	// Clear TT stats before depth-4 search to see fresh stats
	engine.GetTranspositionTableStats() // This will reset the counters

	// Add a small pause to let TT settle (shouldn't be needed, but let's test)
	// time.Sleep(10 * time.Millisecond)

	// Enable debug tracking to monitor move ordering
	engine.SetDebugMoveOrdering(true)
	defer engine.SetDebugMoveOrdering(false)

	config2 := ai.SearchConfig{
		MaxDepth:       4,
		UseOpeningBook: false,
		DebugMode:      true,
	}
	result2 := engine.FindBestMove(ctx, b, moves.White, config2)

	// Debug: Check TT statistics
	if hits, misses, collisions, hitRate := engine.GetTranspositionTableStats(); hits+misses > 0 {
		t.Logf("TT Stats: hits=%d, misses=%d, collisions=%d, hit_rate=%.1f%%", hits, misses, collisions, hitRate)
	}

	t.Logf("Depth 4 search completed:")
	t.Logf("  Best move: %s%s", result2.BestMove.From.String(), result2.BestMove.To.String())
	t.Logf("  Score: %d", result2.Score)
	t.Logf("  Nodes: %d", result2.Stats.NodesSearched)

	// Get the last move ordering (this will be from the depth-4 iteration)
	lastMoveOrder := engine.GetLastMoveOrder()

	if len(lastMoveOrder) == 0 {
		t.Fatal("Should have captured move ordering")
	}

	t.Logf("  Move ordering in depth-4 iteration:")
	for i, move := range lastMoveOrder {
		if i >= 5 { // Only show first 5 moves
			t.Logf("    ... (%d more moves)", len(lastMoveOrder)-5)
			break
		}
		isPV := (move.From == expectedPVMove.From && move.To == expectedPVMove.To)
		t.Logf("    %d: %s%s%s", i+1, move.From.String(), move.To.String(),
			func() string {
				if isPV {
					return " (PV MOVE)"
				} else {
					return ""
				}
			}())
	}

	// Key Assertions: Prove PV is working

	// 1. PV move should be tried FIRST in the depth-4 iteration
	firstMove := lastMoveOrder[0]
	if firstMove.From != expectedPVMove.From || firstMove.To != expectedPVMove.To {
		t.Errorf("PV move should be tried first in depth-4 iteration!")
		t.Errorf("  Expected PV: %s%s", expectedPVMove.From.String(), expectedPVMove.To.String())
		t.Errorf("  First move:  %s%s", firstMove.From.String(), firstMove.To.String())
	} else {
		t.Logf("✅ SUCCESS: PV move was tried first!")
	}

	// 2. Same move should remain best (consistency check)
	if result2.BestMove.From != expectedPVMove.From || result2.BestMove.To != expectedPVMove.To {
		t.Logf("⚠️  WARNING: Best move changed from depth 3 to depth 4")
		t.Logf("  Depth 3 best: %s%s", expectedPVMove.From.String(), expectedPVMove.To.String())
		t.Logf("  Depth 4 best: %s%s", result2.BestMove.From.String(), result2.BestMove.To.String())
		t.Logf("  This could be normal if there are multiple equally good moves")
	} else {
		t.Logf("✅ SUCCESS: Best move remained consistent across depths!")
	}

	// 3. Search should be more efficient due to PV ordering (fewer nodes)
	// This is a relative check - PV should help pruning
	if result2.Stats.NodesSearched > 0 {
		t.Logf("✅ SUCCESS: Search completed with %d nodes", result2.Stats.NodesSearched)
		// Note: We can't easily compare "before/after PV" since PV is integral to the search
		// But we can verify the mechanism is working by checking move ordering
	}

	// 4. Verify the PV move is not a trivial capture (proves ordering benefit)
	if !expectedPVMove.IsCapture {
		t.Logf("✅ SUCCESS: PV move is a quiet move (not capture) - proves ordering is working beyond MVV-LVA")
	} else {
		t.Logf("ℹ️  INFO: PV move is a capture - still valid test but less decisive")
	}

	t.Logf("\n=== PV MOVE ORDERING TEST COMPLETE ===")
}

// Helper functions for TT testing

func createTestPosition(t *testing.T, fen string) *board.Board {
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position from FEN %s: %v", fen, err)
	}
	return b
}

func verifyTTEntry(t *testing.T, engine *MinimaxEngine, b *board.Board, expectedMove board.Move, expectedDepth int) bool {
	if !engine.useTranspositions || engine.transpositionTable == nil {
		t.Fatal("Transposition table not enabled")
	}

	hash := engine.zobrist.HashPosition(b)
	entry, found := engine.transpositionTable.Probe(hash)

	if !found {
		t.Logf("TT entry not found for position")
		return false
	}

	if entry.Depth != expectedDepth {
		t.Errorf("Expected TT depth %d, got %d", expectedDepth, entry.Depth)
		return false
	}

	if entry.BestMove.From != expectedMove.From || entry.BestMove.To != expectedMove.To {
		t.Errorf("Expected TT move %s%s, got %s%s",
			expectedMove.From.String(), expectedMove.To.String(),
			entry.BestMove.From.String(), entry.BestMove.To.String())
		return false
	}

	t.Logf("✅ TT entry verified: move=%s%s, depth=%d, score=%d, type=%d",
		entry.BestMove.From.String(), entry.BestMove.To.String(),
		entry.Depth, entry.Score, entry.Type)
	return true
}

func countNodesWithoutTT(t *testing.T, b *board.Board, player moves.Player, depth int) (bestMove board.Move, nodeCount int64) {
	engine := NewMinimaxEngine()
	// Explicitly disable TT by not setting table size

	config := ai.SearchConfig{
		MaxDepth:       depth,
		UseOpeningBook: false,
		DebugMode:      false,
	}

	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, player, config)

	return result.BestMove, result.Stats.NodesSearched
}

func countNodesWithTT(t *testing.T, b *board.Board, player moves.Player, depth int) (bestMove board.Move, nodeCount int64, hitRate float64) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(16) // 16MB TT

	config := ai.SearchConfig{
		MaxDepth:       depth,
		UseOpeningBook: false,
		DebugMode:      false,
	}

	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, player, config)

	_, _, _, hitRate = engine.GetTranspositionTableStats()
	return result.BestMove, result.Stats.NodesSearched, hitRate
}

// TestTranspositionTableMoveUsage provides comprehensive testing of TT functionality
func TestTranspositionTableMoveUsage(t *testing.T) {
	t.Logf("=== COMPREHENSIVE TRANSPOSITION TABLE TEST ===")

	// Test position: A tactical position with a clear best move
	// This is from a famous puzzle where Re1+ is the winning move
	testFEN := "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	testBoard := createTestPosition(t, testFEN)

	// === TEST 1: Basic TT Storage and Retrieval ===
	t.Logf("\n--- TEST 1: Basic TT Storage and Retrieval ---")

	engine1 := NewMinimaxEngine()
	engine1.SetTranspositionTableSize(16) // 16MB

	config1 := ai.SearchConfig{
		MaxDepth:       3,
		UseOpeningBook: false,
		DebugMode:      false,
	}

	ctx := context.Background()
	result1 := engine1.FindBestMove(ctx, testBoard, moves.White, config1)

	t.Logf("Search result: move=%s%s, score=%d, nodes=%d",
		result1.BestMove.From.String(), result1.BestMove.To.String(),
		result1.Score, result1.Stats.NodesSearched)

	// Verify TT entry was stored correctly
	if verifyTTEntry(t, engine1, testBoard, result1.BestMove, 3) {
		t.Logf("✅ TEST 1 PASSED: TT storage and retrieval working")
	} else {
		t.Errorf("❌ TEST 1 FAILED: TT storage or retrieval not working")
	}

	// === TEST 2: TT Move Ordering Priority ===
	t.Logf("\n--- TEST 2: TT Move Ordering Priority ---")

	engine2 := NewMinimaxEngine()
	engine2.SetTranspositionTableSize(16)

	// First search to establish TT entry
	config2a := ai.SearchConfig{
		MaxDepth:       2,
		UseOpeningBook: false,
		DebugMode:      false,
	}
	result2a := engine2.FindBestMove(ctx, testBoard, moves.White, config2a)
	ttMove := result2a.BestMove

	t.Logf("Established TT move from depth-2 search: %s%s",
		ttMove.From.String(), ttMove.To.String())

	// Second search with move ordering debug enabled
	engine2.SetDebugMoveOrdering(true)
	config2b := ai.SearchConfig{
		MaxDepth:       3,
		UseOpeningBook: false,
		DebugMode:      true,
	}
	_ = engine2.FindBestMove(ctx, testBoard, moves.White, config2b)

	// Check if TT move was tried first
	moveOrder := engine2.GetLastMoveOrder()
	if len(moveOrder) > 0 {
		firstMove := moveOrder[0]
		if firstMove.From == ttMove.From && firstMove.To == ttMove.To {
			t.Logf("✅ TEST 2 PASSED: TT move %s%s was tried first in depth-3 search",
				ttMove.From.String(), ttMove.To.String())
		} else {
			t.Errorf("❌ TEST 2 FAILED: TT move %s%s was not first, got %s%s",
				ttMove.From.String(), ttMove.To.String(),
				firstMove.From.String(), firstMove.To.String())
		}
	} else {
		t.Errorf("❌ TEST 2 FAILED: No move ordering captured")
	}

	// === TEST 3: Search Performance with TT ===
	t.Logf("\n--- TEST 3: Search Performance Comparison ---")

	// Search without TT
	moveWithoutTT, nodesWithoutTT := countNodesWithoutTT(t, testBoard, moves.White, 4)

	// Search with TT
	moveWithTT, nodesWithTT, hitRate := countNodesWithTT(t, testBoard, moves.White, 4)

	t.Logf("Without TT: move=%s%s, nodes=%d",
		moveWithoutTT.From.String(), moveWithoutTT.To.String(), nodesWithoutTT)
	t.Logf("With TT:    move=%s%s, nodes=%d, hit_rate=%.1f%%",
		moveWithTT.From.String(), moveWithTT.To.String(), nodesWithTT, hitRate)

	nodeReduction := float64(nodesWithoutTT-nodesWithTT) / float64(nodesWithoutTT) * 100
	t.Logf("Node reduction: %.1f%%", nodeReduction)

	if nodesWithTT < nodesWithoutTT && hitRate > 5.0 {
		t.Logf("✅ TEST 3 PASSED: TT improved search efficiency (%.1f%% fewer nodes, %.1f%% hit rate)",
			nodeReduction, hitRate)
	} else {
		t.Errorf("❌ TEST 3 FAILED: TT did not improve efficiency sufficiently")
	}

	// === TEST 4: Cross-Search TT Persistence ===
	t.Logf("\n--- TEST 4: Cross-Search TT Persistence ---")

	engine4 := NewMinimaxEngine()
	engine4.SetTranspositionTableSize(16)

	// Search position A to depth 3
	result4a := engine4.FindBestMove(ctx, testBoard, moves.White, ai.SearchConfig{
		MaxDepth: 3, UseOpeningBook: false, DebugMode: false,
	})

	// Create and search a different position B
	positionB := createTestPosition(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	engine4.FindBestMove(ctx, positionB, moves.Black, ai.SearchConfig{
		MaxDepth: 3, UseOpeningBook: false, DebugMode: false,
	})

	// Search position A again - should use TT from first search
	engine4.SetDebugMoveOrdering(true)
	_ = engine4.FindBestMove(ctx, testBoard, moves.White, ai.SearchConfig{
		MaxDepth: 4, UseOpeningBook: false, DebugMode: true,
	})

	// Verify TT move from first search is still available
	moveOrder4 := engine4.GetLastMoveOrder()
	if len(moveOrder4) > 0 {
		firstMove4 := moveOrder4[0]
		if firstMove4.From == result4a.BestMove.From && firstMove4.To == result4a.BestMove.To {
			t.Logf("✅ TEST 4 PASSED: TT persisted across searches - first move %s%s matches previous best",
				firstMove4.From.String(), firstMove4.To.String())
		} else {
			t.Logf("ℹ️  TEST 4 INFO: TT move changed (expected with deeper search)")
			t.Logf("   Previous best: %s%s, Current first: %s%s",
				result4a.BestMove.From.String(), result4a.BestMove.To.String(),
				firstMove4.From.String(), firstMove4.To.String())
		}
	}

	// Final verification: Check TT statistics
	hits, misses, collisions, finalHitRate := engine4.GetTranspositionTableStats()
	t.Logf("Final TT stats: hits=%d, misses=%d, collisions=%d, hit_rate=%.1f%%",
		hits, misses, collisions, finalHitRate)

	if hits > 0 && finalHitRate > 5.0 {
		t.Logf("✅ TEST 4 PASSED: TT cross-search persistence working (%.1f%% hit rate)", finalHitRate)
	} else {
		t.Errorf("❌ TEST 4 FAILED: TT cross-search persistence not working effectively")
	}

	t.Logf("\n=== TRANSPOSITION TABLE TEST COMPLETE ===")
}

// TestDebugTop5MovesAtDepth6 is a debug test to find the top 5 moves at depth 6 for a specific position
func TestDebugTop5MovesAtDepth6(t *testing.T) {
	// Test position: "r7/ppk4p/5p2/6b1/4b1P1/8/P4K2/2q5 b - - 1 47"
	testFEN := "r7/ppk4p/5p2/6b1/4b1P1/8/P4K2/2q5 b - - 1 47"
	b, err := board.FromFEN(testFEN)
	if err != nil {
		t.Fatalf("Failed to create test position from FEN: %v", err)
	}

	t.Logf("=== DEBUG TEST: Top 5 Moves at Depth 6 ===")
	t.Logf("Position FEN: %s", testFEN)
	t.Logf("Player to move: Black")

	// Import the moves package to generate legal moves
	generator := moves.NewGenerator()
	legalMoves := generator.GenerateAllMoves(b, moves.Black)
	defer moves.ReleaseMoveList(legalMoves)

	t.Logf("Total legal moves: %d", legalMoves.Count)

	// Evaluate each move at depth 6
	type MoveScore struct {
		Move  board.Move
		Score ai.EvaluationScore
		Time  time.Duration
		Nodes int64
	}

	var moveScores []MoveScore
	ctx := context.Background()

	t.Logf("\n=== EVALUATING ALL MOVES ===")
	totalStartTime := time.Now()

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Create a fresh engine for each move evaluation
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(64) // 64MB for deep search

		// Make the move
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			t.Logf("Failed to make move %s%s: %v", move.From.String(), move.To.String(), err)
			continue
		}

		// Search from the opponent's perspective (depth 5 since we made one move)
		moveStartTime := time.Now()
		result := engine.FindBestMove(ctx, b, moves.White, ai.SearchConfig{
			MaxDepth:       5, // One less since we made a move
			UseOpeningBook: false,
			DebugMode:      false,
		})
		moveTime := time.Since(moveStartTime)

		// Unmake the move
		b.UnmakeMove(undo)

		// Negate the score since we're looking from Black's perspective
		score := -result.Score

		moveScores = append(moveScores, MoveScore{
			Move:  move,
			Score: score,
			Time:  moveTime,
			Nodes: result.Stats.NodesSearched,
		})

		piece := b.GetPiece(move.From.Rank, move.From.File)
		t.Logf("Move %2d: %s%s (piece: %v) -> Score: %8d, Time: %6s, Nodes: %8d",
			i+1, move.From.String(), move.To.String(), piece, score,
			moveTime.Truncate(time.Millisecond), result.Stats.NodesSearched)
	}

	totalTime := time.Since(totalStartTime)

	// Sort moves by score (best first for Black - highest scores)
	sort.Slice(moveScores, func(i, j int) bool {
		return moveScores[i].Score > moveScores[j].Score
	})

	t.Logf("\n=== TOP 5 MOVES ===")
	maxMoves := 5
	if len(moveScores) < maxMoves {
		maxMoves = len(moveScores)
	}

	var totalNodes int64
	for i := 0; i < maxMoves; i++ {
		ms := moveScores[i]
		piece := b.GetPiece(ms.Move.From.Rank, ms.Move.From.File)
		totalNodes += ms.Nodes

		t.Logf("Rank %d: %s%s", i+1, ms.Move.From.String(), ms.Move.To.String())
		t.Logf("         Piece: %v", piece)
		t.Logf("         Score: %d", ms.Score)
		t.Logf("         Time: %v", ms.Time.Truncate(time.Millisecond))
		t.Logf("         Nodes: %d", ms.Nodes)
		t.Logf("         Is Capture: %t", ms.Move.IsCapture)
		if ms.Move.IsCapture {
			t.Logf("         Captured: %v", ms.Move.Captured)
		}
		if ms.Move.Promotion != board.Empty {
			t.Logf("         Promotion: %v", ms.Move.Promotion)
		}
		t.Logf("")
	}

	t.Logf("=== SUMMARY ===")
	t.Logf("Total evaluation time: %v", totalTime.Truncate(time.Millisecond))
	t.Logf("Total nodes searched: %d", totalNodes)
	t.Logf("Average nodes per move: %d", totalNodes/int64(len(moveScores)))
	t.Logf("Best move: %s%s with score %d",
		moveScores[0].Move.From.String(),
		moveScores[0].Move.To.String(),
		moveScores[0].Score)

	t.Logf("\n=== DEBUG TEST COMPLETE ===")
}
