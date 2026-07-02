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
	if err := engine.MakeMove(move); err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

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

func TestEngineDetectsCheckmate(t *testing.T) {
	engine := NewEngine()
	// Fool's mate: 1.f3 e5 2.g4 Qh4# - White to move, checkmated.
	// Engine only updates GameOver inside MakeMove, so load the position
	// one ply before the mating move (queen still on d8) and play Qh4#
	// so MakeMove's detection actually runs.
	if err := engine.LoadFromFEN("rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq - 0 2"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}
	queenMove := board.Move{
		From:      board.Square{File: 3, Rank: 7}, // d8
		To:        board.Square{File: 7, Rank: 3}, // h4
		Promotion: board.Empty,
	}
	if err := engine.MakeMove(queenMove); err != nil {
		t.Fatalf("expected legal move, got error: %v", err)
	}

	state := engine.GetState()
	if !state.GameOver {
		t.Fatal("expected GameOver to be true after checkmate")
	}
	if state.IsDraw {
		t.Fatal("expected IsDraw to be false for a checkmate")
	}
	if state.Winner != Black {
		t.Errorf("expected Black to win, got %v", state.Winner)
	}
}

func TestEngineDetectsStalemate(t *testing.T) {
	engine := NewEngine()
	// Classic K+Q vs K stalemate: White king h1, White queen b6, Black
	// king a8, Black to move, no legal moves, not in check. Load the
	// position one ply before (queen still on b5, not yet on b6) and
	// play the move that delivers stalemate so MakeMove's detection runs.
	if err := engine.LoadFromFEN("k7/8/8/1Q6/8/8/8/7K w - - 0 1"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}
	queenMove := board.Move{
		From:      board.Square{File: 1, Rank: 4}, // b5
		To:        board.Square{File: 1, Rank: 5}, // b6
		Promotion: board.Empty,
	}
	if err := engine.MakeMove(queenMove); err != nil {
		t.Fatalf("expected legal move, got error: %v", err)
	}

	state := engine.GetState()
	if !state.GameOver {
		t.Fatal("expected GameOver to be true after stalemate")
	}
	if !state.IsDraw {
		t.Fatal("expected IsDraw to be true for a stalemate")
	}
}

func TestEngineDetects50MoveRule(t *testing.T) {
	engine := NewEngine()
	// Two lone kings, halfmove clock already at 99: one more quiet move
	// pushes it to 100, triggering the 50-move rule.
	if err := engine.LoadFromFEN("8/8/4k3/8/8/4K3/8/8 w - - 99 50"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}

	kingMove := board.Move{
		From:      board.Square{File: 4, Rank: 2}, // e3
		To:        board.Square{File: 4, Rank: 3}, // e4
		Promotion: board.Empty,
	}
	if err := engine.MakeMove(kingMove); err != nil {
		t.Fatalf("expected legal move, got error: %v", err)
	}

	state := engine.GetState()
	if !state.GameOver {
		t.Fatal("expected GameOver to be true after the 50-move rule triggers")
	}
	if !state.IsDraw {
		t.Fatal("expected IsDraw to be true for the 50-move rule")
	}
}

func TestEngineDetectsThreefoldRepetition(t *testing.T) {
	engine := NewEngine()
	if err := engine.LoadFromFEN("8/8/4k3/8/8/4K3/8/8 w - - 0 1"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}

	// Shuffle kings back to the starting position twice more (3 total
	// occurrences of the starting position's hash).
	shuffle := []board.Move{
		{From: board.Square{File: 4, Rank: 2}, To: board.Square{File: 4, Rank: 3}, Promotion: board.Empty}, // Ke3-e4
		{From: board.Square{File: 4, Rank: 5}, To: board.Square{File: 4, Rank: 6}, Promotion: board.Empty}, // Ke6-e7
		{From: board.Square{File: 4, Rank: 3}, To: board.Square{File: 4, Rank: 2}, Promotion: board.Empty}, // Ke4-e3
		{From: board.Square{File: 4, Rank: 6}, To: board.Square{File: 4, Rank: 5}, Promotion: board.Empty}, // Ke7-e6 (2nd occurrence)
		{From: board.Square{File: 4, Rank: 2}, To: board.Square{File: 4, Rank: 3}, Promotion: board.Empty}, // Ke3-e4
		{From: board.Square{File: 4, Rank: 5}, To: board.Square{File: 4, Rank: 6}, Promotion: board.Empty}, // Ke6-e7
		{From: board.Square{File: 4, Rank: 3}, To: board.Square{File: 4, Rank: 2}, Promotion: board.Empty}, // Ke4-e3
		{From: board.Square{File: 4, Rank: 6}, To: board.Square{File: 4, Rank: 5}, Promotion: board.Empty}, // Ke7-e6 (3rd occurrence)
	}

	var state *State
	for i, move := range shuffle {
		if err := engine.MakeMove(move); err != nil {
			t.Fatalf("move %d failed: %v", i, err)
		}
		state = engine.GetState()
	}

	if !state.GameOver {
		t.Fatal("expected GameOver to be true after threefold repetition")
	}
	if !state.IsDraw {
		t.Fatal("expected IsDraw to be true for threefold repetition")
	}
}
