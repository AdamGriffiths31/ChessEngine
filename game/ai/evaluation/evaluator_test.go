package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	if evaluator == nil {
		t.Fatal("NewEvaluator should not return nil")
	}

	if evaluator.GetName() != "Evaluator" {
		t.Errorf("Expected name 'Evaluator', got '%s'", evaluator.GetName())
	}
}

func TestEvaluateEmptyBoard(t *testing.T) {
	evaluator := NewEvaluator()
	b := board.NewBoard()

	// Empty board should have score 0 (always from White's perspective)
	score := evaluator.Evaluate(b)

	if score != 0 {
		t.Errorf("Expected score 0 for empty board, got %d", score)
	}
}

func TestEvaluateStartingPosition(t *testing.T) {
	evaluator := NewEvaluator()

	// Create board from starting position FEN
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from starting FEN: %v", err)
	}

	// Starting position should be roughly equal
	score := evaluator.Evaluate(b)

	if score != 0 {
		t.Errorf("Expected score 0 in starting position, got %d", score)
	}
}
