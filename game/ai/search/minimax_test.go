package search

import (
	"context"
	"testing"

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
