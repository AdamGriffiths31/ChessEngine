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
		MaxDepth:        2,
		UseOpeningBook:  false, // Disable opening book for this test
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
			MaxDepth:        depth,
			UseOpeningBook:  false, // Disable opening book for this test
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
			MaxDepth:        maxDepth,
			UseOpeningBook:  false, // Disable opening book for consistent testing
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

// TestMinimaxVsDirectEvaluation checks that depth-1 search matches direct evaluation
func TestMinimaxVsDirectEvaluation(t *testing.T) {
	engine := NewMinimaxEngine()
	
	testPositions := []struct {
		name string
		fen  string
		description string
	}{
		{
			name: "starting_position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Starting position should be roughly equal",
		},
		{
			name: "white_up_queen",
			fen:  "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Black queen missing
			description: "White up a queen (+900) should have positive evaluation",
		},
		{
			name: "black_up_queen", 
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1K2R w KQkq - 0 1", // White queen missing
			description: "Black up a queen (-900) should have negative evaluation",
		},
		{
			name: "problematic_position",
			fen:  "1r2kr2/2p1b2p/2Pp2pP/1p2pbP1/p2n4/P4p2/4nP2/RNB2RK1 w - - 5 27",
			description: "Position where White is significantly behind in material",
		},
	}
	
	for _, pos := range testPositions {
		t.Run(pos.name, func(t *testing.T) {
			b, err := board.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", pos.fen, err)
			}
			
			// Get direct evaluation
			directEval := engine.evaluator.Evaluate(b)
			
			// Get depth-1 search for White
			ctx := context.Background()
			config := ai.SearchConfig{
				MaxDepth:       1,
				UseOpeningBook: false,
			}
			whiteResult := engine.FindBestMove(ctx, b, moves.White, config)
			
			t.Logf("%s:", pos.description)
			t.Logf("  Direct evaluation: %d", directEval)
			t.Logf("  White depth-1 search: %d", whiteResult.Score)
			
			// For depth-1, the search score should closely match direct evaluation
			// Allow small differences due to best move selection, but sign should match
			if (directEval > 0 && whiteResult.Score < 0) || (directEval < 0 && whiteResult.Score > 0) {
				t.Errorf("🚨 SIGN MISMATCH: Direct eval %d vs search %d", directEval, whiteResult.Score)
			}
			
			// Check that the difference isn't too large (shouldn't be more than a pawn value apart)
			diff := abs(int(directEval) - int(whiteResult.Score))
			if diff > 200 { // More than 2 pawns difference
				t.Errorf("Large difference between direct eval (%d) and search (%d): %d", 
					directEval, whiteResult.Score, diff)
			}
		})
	}
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
		name string
		fen  string
		expectedSign int // 1 for positive (White better), -1 for negative (Black better)
		description string
	}{
		{
			name: "white_up_rook",
			fen:  "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Black h8 rook missing
			expectedSign: 1,
			description: "White up a rook should be positive",
		},
		{
			name: "black_up_rook", 
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK1NR w KQkq - 0 1", // White h1 rook missing
			expectedSign: -1,
			description: "Black up a rook should be negative",
		},
		{
			name: "white_up_knight",
			fen:  "rnbqkb1r/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Black knight missing
			expectedSign: 1,
			description: "White up a knight should be positive",
		},
		{
			name: "black_up_knight",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R1BQKBNR w KQkq - 0 1", // White knight missing  
			expectedSign: -1,
			description: "Black up a knight should be negative",
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

// TestMinimaxDepthProgression tests how scores change with increasing depth
func TestMinimaxDepthProgression(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Simple position where White is up a rook  
	fen := "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" // Black h8 rook missing
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Get direct evaluation first
	directEval := engine.evaluator.Evaluate(b)
	t.Logf("Direct evaluation: %d", directEval)
	
	ctx := context.Background()
	
	// Test different depths and see how scores change
	for depth := 1; depth <= 4; depth++ {
		config := ai.SearchConfig{
			MaxDepth:       depth,
			UseOpeningBook: false,
			DebugMode:      depth <= 2, // Only debug for shallow depths
		}
		
		result := engine.FindBestMove(ctx, b, moves.White, config)
		
		t.Logf("Depth %d: score=%d, move=%s%s", 
			depth, result.Score, result.BestMove.From.String(), result.BestMove.To.String())
		
		// Print debug info for shallow depths
		if depth <= 2 && len(result.Stats.DebugInfo) > 0 {
			for _, info := range result.Stats.DebugInfo {
				t.Logf("  DEBUG: %s", info)
			}
		}
		
		// For White playing, score should remain positive (White is up material)
		if result.Score <= 0 {
			t.Errorf("Depth %d: White should have positive score (up a rook), got %d", depth, result.Score)
		}
	}
}

// TestMinimaxSignConsistency tests the sign consistency issue directly
func TestMinimaxSignConsistency(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Position where Black is clearly better (up a queen)
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1K2R w KQkq - 0 1" // White queen missing
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Direct evaluation should be negative (Black better)
	directEval := engine.evaluator.Evaluate(b)
	t.Logf("Direct evaluation: %d (should be negative - Black up a queen)", directEval)
	
	ctx := context.Background()
	
	// Test depth 2 with debug info
	config := ai.SearchConfig{
		MaxDepth:       2,
		UseOpeningBook: false,
		DebugMode:      true,
	}
	
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	t.Logf("Search result: %d (should be negative - Black up a queen)", result.Score)
	t.Logf("Debug info:")
	for _, info := range result.Stats.DebugInfo {
		t.Logf("  %s", info)
	}
	
	// Both should have the same sign (negative - Black is better)
	if directEval < 0 && result.Score > 0 {
		t.Errorf("🚨 SIGN INCONSISTENCY: Direct eval %d vs search %d", directEval, result.Score)
	}
	if directEval > 0 && result.Score < 0 {
		t.Errorf("🚨 SIGN INCONSISTENCY: Direct eval %d vs search %d", directEval, result.Score)
	}
}

func TestFindBestMoveOneSecondSearch(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	b, err := board.FromFEN("rn2r1k1/pp2qppp/2p2n2/3p1b2/PbPp3P/1P3PP1/3BP3/RN1QKBNR w KQ - 1 11")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	config := ai.SearchConfig{
		MaxDepth:       6, // Set max depth to 6
		UseOpeningBook: false,
	}
	
	// Get direct evaluation first for comparison
	directEval := engine.evaluator.Evaluate(b)
	t.Logf("Direct evaluation: %d", directEval)
	
	result := engine.FindBestMove(ctx, b, moves.White, config)

	// Should return a valid move
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move after 1 second search")
	}

	// Should have taken approximately 1 second (with some tolerance)
	if result.Stats.Time < 900*time.Millisecond || result.Stats.Time > 1200*time.Millisecond {
		t.Logf("Search time: %v (expected ~1s)", result.Stats.Time)
	}

	// Should have searched some positions
	if result.Stats.NodesSearched == 0 {
		t.Error("Should have evaluated at least one position in 1 second")
	}

	t.Logf("1-second search results:")
	t.Logf("  Best move: %s%s", result.BestMove.From.String(), result.BestMove.To.String())
	t.Logf("  Score: %d", result.Score)
	t.Logf("  Nodes searched: %d", result.Stats.NodesSearched)
	t.Logf("  Time taken: %v", result.Stats.Time)
	t.Logf("  Depth reached: %d", result.Stats.Depth)
}