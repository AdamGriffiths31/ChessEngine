package player

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// MockEngine implements Engine interface for testing
type MockEngine struct {
	name string
}

func (m *MockEngine) FindBestMove(_ context.Context, _ *board.Board, player movegen.Player, config search.SearchConfig) search.SearchResult {
	var move board.Move
	if player == movegen.White {
		move = board.Move{
			From: board.Square{File: 4, Rank: 1}, // e2
			To:   board.Square{File: 4, Rank: 3}, // e4
		}
	} else {
		move = board.Move{
			From: board.Square{File: 4, Rank: 6}, // e7
			To:   board.Square{File: 4, Rank: 4}, // e5
		}
	}

	return search.SearchResult{
		BestMove: move,
		Score:    eval.EvaluationScore(0),
		Stats: search.SearchStats{
			NodesSearched: 100,
			Depth:         config.MaxDepth,
			Time:          10 * time.Millisecond,
		},
	}
}

func (m *MockEngine) SetEvaluator(_ eval.Evaluator) {
}

func (m *MockEngine) GetName() string {
	return m.name
}

func NewMockEngine() *MockEngine {
	return &MockEngine{name: "Mock Engine"}
}

func TestNewComputerPlayer(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}

	player := NewComputerPlayer("Test Computer", engine, config)

	if player == nil {
		t.Fatal("NewComputerPlayer should not return nil")
	}

	if player.GetName() != "Test Computer" {
		t.Errorf("Expected name 'Test Computer', got '%s'", player.GetName())
	}
}

func TestComputerPlayerGetMove(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}

	player := NewComputerPlayer("Test Computer", engine, config)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	move, err := player.GetMove(b, movegen.White, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}

	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move")
	}

	// Mock engine returns e2e4 for white, which should be a white pawn
	expectedFrom := board.Square{File: 4, Rank: 1} // e2
	if move.From != expectedFrom {
		t.Errorf("Expected move from e2, got %s", move.From.String())
	}
}

func TestComputerPlayerGetMoveBlack(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}

	player := NewComputerPlayer("Black Computer", engine, config)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")

	move, err := player.GetMove(b, movegen.Black, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}

	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move for black")
	}

	// Mock engine returns e7e5 for black, which should be a black pawn
	expectedFrom := board.Square{File: 4, Rank: 6} // e7
	if move.From != expectedFrom {
		t.Errorf("Expected move from e7, got %s", move.From.String())
	}
}

func TestGetMoveWithTimeout(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 1, // Keep depth low for faster execution
		MaxTime:  100 * time.Millisecond,
	}

	player := NewComputerPlayer("Fast Computer", engine, config)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	start := time.Now()
	move, err := player.GetMove(b, movegen.White, 50*time.Millisecond)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}

	if duration > 500*time.Millisecond {
		t.Errorf("GetMove took too long: %v", duration)
	}

	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move even with short timeout")
	}
}

func TestGetMoveForcedCapture(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}

	player := NewComputerPlayer("Tactical Computer", engine, config)

	// Position where white can capture black queen with rook (White king on e1, Rook on d2, Black queen on d4, Black king on e8)
	b := testutil.MustFromFEN(t, "4k3/8/8/8/3q4/8/3R4/4K3 w - - 0 1")

	move, err := player.GetMove(b, movegen.White, 3*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}

	// Mock engine will return e2e4, so just verify we get a valid move
	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move")
	}
}

func TestGetMoveDifferentDifficulties(t *testing.T) {
	b, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	// Depths/times formerly set via the removed SetDifficulty helper
	// (easy/medium/hard), kept here to preserve coverage of GetMove across
	// different search configurations.
	configs := []search.SearchConfig{
		{MaxDepth: 2, MaxTime: 1 * time.Second},
		{MaxDepth: 4, MaxTime: 3 * time.Second},
		{MaxDepth: 6, MaxTime: 5 * time.Second},
	}

	for _, config := range configs {
		engine := NewMockEngine()
		player := NewComputerPlayer("Variable Computer", engine, config)

		move, err := player.GetMove(b, movegen.White, 3*time.Second)
		if err != nil {
			t.Errorf("GetMove failed for depth %d: %v", config.MaxDepth, err)
			continue
		}

		if move.From.File == -1 && move.From.Rank == -1 {
			t.Errorf("Invalid move returned for depth %d", config.MaxDepth)
		}

		// Mock engine returns e2e4 for white
		expectedFrom := board.Square{File: 4, Rank: 1} // e2
		if move.From != expectedFrom {
			t.Errorf("Expected move from e2 for depth %d, got %s", config.MaxDepth, move.From.String())
		}
	}
}

func TestGetMoveWithStats(t *testing.T) {
	engine := NewMockEngine()
	config := search.SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}

	player := NewComputerPlayer("Stats Computer", engine, config)

	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	result, err := player.GetMoveWithStats(b, movegen.White, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMoveWithStats returned error: %v", err)
	}

	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("GetMoveWithStats should return a valid move")
	}

	if result.Stats.NodesSearched == 0 {
		t.Error("GetMoveWithStats should return search statistics")
	}

	if result.Stats.Time <= 0 {
		t.Error("GetMoveWithStats should return positive search time")
	}

	if result.Stats.NodesSearched != 100 {
		t.Errorf("Expected 100 nodes searched, got %d", result.Stats.NodesSearched)
	}

	if result.Stats.Depth != config.MaxDepth {
		t.Errorf("Expected depth %d, got %d", config.MaxDepth, result.Stats.Depth)
	}
}
