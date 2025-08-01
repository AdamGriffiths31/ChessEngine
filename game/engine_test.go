package game

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("Expected engine to be non-nil")
	}

	state := engine.GetState()
	if state == nil {
		t.Fatal("Expected state to be non-nil")
	}

	currentPlayer := engine.GetCurrentPlayer()
	if currentPlayer != White {
		t.Errorf("Expected initial turn to be White, got %v", currentPlayer)
	}

	if state.MoveCount != 1 {
		t.Errorf("Expected initial move count to be 1, got %d", state.MoveCount)
	}

	if state.GameOver {
		t.Errorf("Expected game to not be over initially")
	}
}

func TestEnginePlayerString(t *testing.T) {
	testCases := []struct {
		player   Player
		expected string
	}{
		{White, "White"},
		{Black, "Black"},
	}

	for _, tc := range testCases {
		result := tc.player.String()
		if result != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, result)
		}
	}
}

func TestEngineMakeMove(t *testing.T) {
	engine := NewEngine()

	// Test white move
	move := board.Move{
		From:      board.Square{File: 4, Rank: 1}, // e2
		To:        board.Square{File: 4, Rank: 3}, // e4
		Promotion: board.Empty,
	}

	err := engine.MakeMove(move)
	if err != nil {
		t.Errorf("Expected no error making move, got: %v", err)
	}

	state := engine.GetState()
	currentPlayer := engine.GetCurrentPlayer()
	if currentPlayer != Black {
		t.Errorf("Expected turn to switch to Black after white move, got %v", currentPlayer)
	}

	if state.MoveCount != 1 {
		t.Errorf("Expected move count to remain 1 after white move, got %d", state.MoveCount)
	}

	// Test black move
	move = board.Move{
		From:      board.Square{File: 4, Rank: 6}, // e7
		To:        board.Square{File: 4, Rank: 4}, // e5
		Promotion: board.Empty,
	}

	err = engine.MakeMove(move)
	if err != nil {
		t.Errorf("Expected no error making black move, got: %v", err)
	}

	state = engine.GetState()
	currentPlayer = engine.GetCurrentPlayer()
	if currentPlayer != White {
		t.Errorf("Expected turn to switch to White after black move, got %v", currentPlayer)
	}

	if state.MoveCount != 2 {
		t.Errorf("Expected move count to be 2 after black move, got %d", state.MoveCount)
	}
}

func TestEngineReset(t *testing.T) {
	engine := NewEngine()

	// Make a move
	move := board.Move{
		From:      board.Square{File: 4, Rank: 1}, // e2
		To:        board.Square{File: 4, Rank: 3}, // e4
		Promotion: board.Empty,
	}
	engine.MakeMove(move)

	// Reset the engine
	engine.Reset()

	state := engine.GetState()
	currentPlayer := engine.GetCurrentPlayer()
	if currentPlayer != White {
		t.Errorf("Expected turn to be White after reset, got %v", currentPlayer)
	}

	if state.MoveCount != 1 {
		t.Errorf("Expected move count to be 1 after reset, got %d", state.MoveCount)
	}

	if state.GameOver {
		t.Errorf("Expected game to not be over after reset")
	}

	// Check that board is back to initial position
	piece := state.Board.GetPiece(1, 4) // e2
	if piece != board.WhitePawn {
		t.Errorf("Expected white pawn at e2 after reset, got %c", piece)
	}
}
