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