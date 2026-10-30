package game

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
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

func TestEngineHashHistoryTracksRealMoves(t *testing.T) {
	engine := NewEngine()

	if got := len(engine.HashHistory()); got != 1 {
		t.Fatalf("HashHistory() length = %d after NewEngine, want 1 (just the starting position)", got)
	}
	startHash := engine.HashHistory()[0]

	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	if err := engine.MakeMove(move); err != nil {
		t.Fatalf("MakeMove failed: %v", err)
	}

	history := engine.HashHistory()
	if len(history) != 2 {
		t.Fatalf("HashHistory() length = %d after one move, want 2", len(history))
	}
	if history[0] != startHash {
		t.Errorf("HashHistory()[0] changed after a move: got %d, want the original start hash %d", history[0], startHash)
	}
	wantCurrent := book.GetPolyglotHash().HashPosition(engine.GetState().Board)
	if history[1] != wantCurrent {
		t.Errorf("HashHistory()[1] = %d, want %d (hash of the current position)", history[1], wantCurrent)
	}
}

func TestEngineReset(t *testing.T) {
	engine := NewEngine()

	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	if err := engine.MakeMove(move); err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

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

	piece := state.Board.GetPiece(1, 4)
	if piece != board.WhitePawn {
		t.Errorf("Expected white pawn at e2 after reset, got %c", piece)
	}
}

func TestEngineDetectsCheckmate(t *testing.T) {
	engine := NewEngine()
	if err := engine.LoadFromFEN("rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq - 0 2"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}
	queenMove := board.Move{
		From:      board.Square{File: 3, Rank: 7},
		To:        board.Square{File: 7, Rank: 3},
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
	if err := engine.LoadFromFEN("k7/8/8/1Q6/8/8/8/7K w - - 0 1"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}
	queenMove := board.Move{
		From:      board.Square{File: 1, Rank: 4},
		To:        board.Square{File: 1, Rank: 5},
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
	if err := engine.LoadFromFEN("8/8/4k3/8/8/4K3/8/8 w - - 99 50"); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}

	kingMove := board.Move{
		From:      board.Square{File: 4, Rank: 2},
		To:        board.Square{File: 4, Rank: 3},
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

	shuffle := []board.Move{
		{From: board.Square{File: 4, Rank: 2}, To: board.Square{File: 4, Rank: 3}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 5}, To: board.Square{File: 4, Rank: 6}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 3}, To: board.Square{File: 4, Rank: 2}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 6}, To: board.Square{File: 4, Rank: 5}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 2}, To: board.Square{File: 4, Rank: 3}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 5}, To: board.Square{File: 4, Rank: 6}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 3}, To: board.Square{File: 4, Rank: 2}, Promotion: board.Empty},
		{From: board.Square{File: 4, Rank: 6}, To: board.Square{File: 4, Rank: 5}, Promotion: board.Empty},
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
