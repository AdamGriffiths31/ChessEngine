package ai

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// MockEngine implements Engine interface for testing
type MockEngine struct {
	name string
}

func (m *MockEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config SearchConfig) SearchResult {
	// Return a simple pawn move for testing
	var move board.Move
	if player == moves.White {
		// e2e4 for white
		move = board.Move{
			From: board.Square{File: 4, Rank: 1}, // e2
			To:   board.Square{File: 4, Rank: 3}, // e4
		}
	} else {
		// e7e5 for black
		move = board.Move{
			From: board.Square{File: 4, Rank: 6}, // e7  
			To:   board.Square{File: 4, Rank: 4}, // e5
		}
	}
	
	return SearchResult{
		BestMove: move,
		Score:    EvaluationScore(0),
		Stats: SearchStats{
			NodesSearched: 100,
			Depth:        config.MaxDepth, // Return the configured depth
			Time:         10 * time.Millisecond,
		},
	}
}

func (m *MockEngine) SetEvaluator(evaluator Evaluator) {
	// No-op for mock
}

func (m *MockEngine) GetName() string {
	return m.name
}

func NewMockEngine() *MockEngine {
	return &MockEngine{name: "Mock Engine"}
}

func TestNewComputerPlayer(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
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
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}
	
	player := NewComputerPlayer("Test Computer", engine, config)
	
	// Create starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}
	
	// Get move for white
	move, err := player.GetMove(b, moves.White, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}
	
	// Should return a valid move
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
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}
	
	player := NewComputerPlayer("Black Computer", engine, config)
	
	// Position where it's black to move
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Get move for black
	move, err := player.GetMove(b, moves.Black, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}
	
	// Should return a valid move
	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move for black")
	}
	
	// Mock engine returns e7e5 for black, which should be a black pawn
	expectedFrom := board.Square{File: 4, Rank: 6} // e7
	if move.From != expectedFrom {
		t.Errorf("Expected move from e7, got %s", move.From.String())
	}
}

func TestSetDifficultyEasy(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 4,
		MaxTime:  3 * time.Second,
	}
	
	player := NewComputerPlayer("Test Computer", engine, config)
	
	// Set to easy difficulty
	player.SetDifficulty("easy")
	
	// Check that difficulty description is correct
	expectedDiff := "Easy (depth 2, 1s think time)"
	if player.GetDifficulty() != expectedDiff {
		t.Errorf("Expected difficulty '%s', got '%s'", expectedDiff, player.GetDifficulty())
	}
}

func TestSetDifficultyMedium(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}
	
	player := NewComputerPlayer("Test Computer", engine, config)
	
	// Set to medium difficulty
	player.SetDifficulty("medium")
	
	// Check that difficulty description is correct
	expectedDiff := "Medium (depth 4, 3s think time)"
	if player.GetDifficulty() != expectedDiff {
		t.Errorf("Expected difficulty '%s', got '%s'", expectedDiff, player.GetDifficulty())
	}
}

func TestSetDifficultyHard(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}
	
	player := NewComputerPlayer("Test Computer", engine, config)
	
	// Set to hard difficulty
	player.SetDifficulty("hard")
	
	// Check that difficulty description is correct
	expectedDiff := "Hard (depth 6, 5s think time)"
	if player.GetDifficulty() != expectedDiff {
		t.Errorf("Expected difficulty '%s', got '%s'", expectedDiff, player.GetDifficulty())
	}
}

func TestSetDifficultyUnknown(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  1 * time.Second,
	}
	
	player := NewComputerPlayer("Test Computer", engine, config)
	
	// Set to unknown difficulty - should default to medium
	player.SetDifficulty("unknown")
	
	// Check that it defaults to medium
	expectedDiff := "Medium (depth 4, 3s think time)"
	if player.GetDifficulty() != expectedDiff {
		t.Errorf("Expected default difficulty '%s', got '%s'", expectedDiff, player.GetDifficulty())
	}
}

func TestGetMoveWithTimeout(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 1, // Keep depth low for faster execution
		MaxTime:  100 * time.Millisecond,
	}
	
	player := NewComputerPlayer("Fast Computer", engine, config)
	
	// Create starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}
	
	// Get move with very short timeout
	start := time.Now()
	move, err := player.GetMove(b, moves.White, 50*time.Millisecond)
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}
	
	// Should return a move within reasonable time
	if duration > 500*time.Millisecond {
		t.Errorf("GetMove took too long: %v", duration)
	}
	
	// Should still return a valid move
	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move even with short timeout")
	}
}

func TestGetMoveForcedCapture(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}
	
	player := NewComputerPlayer("Tactical Computer", engine, config)
	
	// Position where white can capture black queen with rook
	b, err := board.FromFEN("8/8/8/8/3q4/8/3R4/8 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create tactical position: %v", err)
	}
	
	move, err := player.GetMove(b, moves.White, 3*time.Second)
	if err != nil {
		t.Fatalf("GetMove returned error: %v", err)
	}
	
	// Mock engine will return e2e4, so just verify we get a valid move
	if move.From.File == -1 && move.From.Rank == -1 {
		t.Error("GetMove should return a valid move")
	}
}

func TestGetMoveDifferentDifficulties(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 4,
		MaxTime:  3 * time.Second,
	}
	
	player := NewComputerPlayer("Variable Computer", engine, config)
	
	// Create a complex position
	b, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	difficulties := []string{"easy", "medium", "hard"}
	
	// Test that all difficulties return valid moves
	for _, diff := range difficulties {
		player.SetDifficulty(diff)
		
		move, err := player.GetMove(b, moves.White, 3*time.Second)
		if err != nil {
			t.Errorf("GetMove failed for difficulty %s: %v", diff, err)
			continue
		}
		
		if move.From.File == -1 && move.From.Rank == -1 {
			t.Errorf("Invalid move returned for difficulty %s", diff)
		}
		
		// Mock engine returns e2e4 for white
		expectedFrom := board.Square{File: 4, Rank: 1} // e2
		if move.From != expectedFrom {
			t.Errorf("Expected move from e2 for difficulty %s, got %s", diff, move.From.String())
		}
	}
}

// Helper function to check if a piece is white
func isWhitePiece(piece board.Piece) bool {
	return piece == board.WhitePawn || piece == board.WhiteRook || 
		   piece == board.WhiteKnight || piece == board.WhiteBishop || 
		   piece == board.WhiteQueen || piece == board.WhiteKing
}

// Helper function to check if a piece is black  
func isBlackPiece(piece board.Piece) bool {
	return piece == board.BlackPawn || piece == board.BlackRook || 
		   piece == board.BlackKnight || piece == board.BlackBishop || 
		   piece == board.BlackQueen || piece == board.BlackKing
}

func TestSetDebugMode(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}
	
	player := NewComputerPlayer("Debug Computer", engine, config)
	
	// Initially debug should be off
	if player.IsDebugMode() {
		t.Error("Debug mode should be disabled by default")
	}
	
	// Enable debug mode
	player.SetDebugMode(true)
	if !player.IsDebugMode() {
		t.Error("Debug mode should be enabled after calling SetDebugMode(true)")
	}
	
	// Disable debug mode
	player.SetDebugMode(false)
	if player.IsDebugMode() {
		t.Error("Debug mode should be disabled after calling SetDebugMode(false)")
	}
}

func TestGetMoveWithStats(t *testing.T) {
	engine := NewMockEngine()
	config := SearchConfig{
		MaxDepth: 3,
		MaxTime:  2 * time.Second,
	}
	
	player := NewComputerPlayer("Stats Computer", engine, config)
	
	// Create starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}
	
	// Get move with stats
	result, err := player.GetMoveWithStats(b, moves.White, 2*time.Second)
	if err != nil {
		t.Fatalf("GetMoveWithStats returned error: %v", err)
	}
	
	// Should return a valid move
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("GetMoveWithStats should return a valid move")
	}
	
	// Should have stats
	if result.Stats.NodesSearched == 0 {
		t.Error("GetMoveWithStats should return search statistics")
	}
	
	if result.Stats.Time <= 0 {
		t.Error("GetMoveWithStats should return positive search time")
	}
	
	// Mock returns specific values
	if result.Stats.NodesSearched != 100 {
		t.Errorf("Expected 100 nodes searched, got %d", result.Stats.NodesSearched)
	}
	
	// Mock should return the configured depth
	if result.Stats.Depth != config.MaxDepth {
		t.Errorf("Expected depth %d, got %d", config.MaxDepth, result.Stats.Depth)
	}
}