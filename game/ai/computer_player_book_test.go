package ai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// MockBookEngine implements the Engine interface for book testing
type MockBookEngine struct {
	name string
}

func (m *MockBookEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config SearchConfig) SearchResult {
	// Return a simple valid move (e2-e4)
	return SearchResult{
		BestMove: board.Move{
			From: board.Square{File: 4, Rank: 1}, // e2
			To:   board.Square{File: 4, Rank: 3}, // e4
		},
		Score: 0,
		Stats: SearchStats{
			NodesSearched: 100,
			Depth:         config.MaxDepth,
			Time:          time.Millisecond * 10,
		},
	}
}

func (m *MockBookEngine) SetEvaluator(eval Evaluator) {
	// No-op for mock
}

func (m *MockBookEngine) GetName() string {
	return "Mock Book Engine"
}

func TestComputerPlayerOpeningBook(t *testing.T) {
	// Create a mock engine for testing
	engine := &MockBookEngine{}
	config := SearchConfig{
		MaxDepth:     3,
		MaxTime:      time.Second,
		UseAlphaBeta: true,
		DebugMode:    false,
	}
	
	player := NewComputerPlayer("Test Engine", engine, config)
	
	// Test book configuration methods
	bookPath := "../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	// Initially no opening book
	if player.IsUsingOpeningBook() {
		t.Error("Expected opening book to be disabled initially")
	}
	
	// Enable opening book
	player.SetOpeningBook(true, []string{bookPath})
	
	if !player.IsUsingOpeningBook() {
		t.Error("Expected opening book to be enabled after SetOpeningBook")
	}
	
	bookFiles := player.GetBookFiles()
	if len(bookFiles) != 1 || bookFiles[0] != bookPath {
		t.Errorf("Expected book files [%s], got %v", bookPath, bookFiles)
	}
	
	// Test selection mode configuration
	player.SetBookSelectionMode(BookSelectBest)
	if player.config.BookSelectMode != BookSelectBest {
		t.Error("Expected book selection mode to be updated")
	}
	
	// Test weight threshold configuration  
	player.SetBookWeightThreshold(50)
	if player.config.BookWeightThreshold != 50 {
		t.Error("Expected book weight threshold to be updated")
	}
	
	// Test that the player can make moves with book enabled
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}
	
	move, err := player.GetMove(b, moves.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move: %v", err)
	}
	
	// Should return a valid move
	if move.From.File < 0 || move.From.File > 7 || 
	   move.From.Rank < 0 || move.From.Rank > 7 ||
	   move.To.File < 0 || move.To.File > 7 ||
	   move.To.Rank < 0 || move.To.Rank > 7 {
		t.Errorf("Invalid move returned: %+v", move)
	}
	
	t.Logf("Computer player with book made move: %c%d-%c%d",
		'a'+move.From.File, move.From.Rank+1,
		'a'+move.To.File, move.To.Rank+1)
}

func TestComputerPlayerBookWithStats(t *testing.T) {
	engine := &MockBookEngine{}
	config := SearchConfig{
		MaxDepth:     2,
		MaxTime:      500 * time.Millisecond,
		UseAlphaBeta: true,
		DebugMode:    true,
	}
	
	player := NewComputerPlayer("Test Engine with Book", engine, config)
	
	bookPath := "../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	player.SetOpeningBook(true, []string{bookPath})
	player.SetBookSelectionMode(BookSelectWeightedRandom)
	
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Get move with statistics
	result, err := player.GetMoveWithStats(b, moves.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move with stats: %v", err)
	}
	
	// Log the results
	t.Logf("Move: %c%d-%c%d", 
		'a'+result.BestMove.From.File, result.BestMove.From.Rank+1,
		'a'+result.BestMove.To.File, result.BestMove.To.Rank+1)
	t.Logf("Stats: %d nodes, depth %d, time %v", 
		result.Stats.NodesSearched, result.Stats.Depth, result.Stats.Time)
	
	// If the move came from book, nodes searched should be 0
	if result.Stats.NodesSearched == 0 {
		t.Log("Move came from opening book (no search needed)")
	} else {
		t.Log("Move came from search (not found in opening book)")
	}
}

func TestComputerPlayerBookDifficulty(t *testing.T) {
	engine := &MockBookEngine{}
	config := SearchConfig{}
	
	player := NewComputerPlayer("Difficulty Test", engine, config)
	
	bookPath := "../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	player.SetOpeningBook(true, []string{bookPath})
	
	// Test that book settings persist across difficulty changes
	player.SetDifficulty("easy")
	if !player.IsUsingOpeningBook() {
		t.Error("Opening book should remain enabled after difficulty change")
	}
	
	player.SetDifficulty("hard")
	if !player.IsUsingOpeningBook() {
		t.Error("Opening book should remain enabled after difficulty change")
	}
	
	// Verify book files are still configured
	bookFiles := player.GetBookFiles()
	if len(bookFiles) != 1 || bookFiles[0] != bookPath {
		t.Error("Book files should persist across difficulty changes")
	}
}

func TestComputerPlayerBookModes(t *testing.T) {
	engine := &MockBookEngine{}
	config := SearchConfig{MaxDepth: 2}
	
	player := NewComputerPlayer("Book Mode Test", engine, config)
	
	bookPath := "../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}
	
	player.SetOpeningBook(true, []string{bookPath})
	
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test different selection modes
	modes := []struct {
		mode BookSelectionMode
		name string
	}{
		{BookSelectBest, "Best"},
		{BookSelectRandom, "Random"},
		{BookSelectWeightedRandom, "WeightedRandom"},
	}
	
	for _, tc := range modes {
		t.Run(tc.name, func(t *testing.T) {
			player.SetBookSelectionMode(tc.mode)
			
			// Make multiple moves to see if there's variation (for random modes)
			testMoves := make([]board.Move, 3)
			for i := 0; i < 3; i++ {
				move, err := player.GetMove(b, moves.White, 100*time.Millisecond)
				if err != nil {
					t.Fatalf("Failed to get move: %v", err)
				}
				testMoves[i] = move
			}
			
			// Just verify we get valid moves
			for i, move := range testMoves {
				if move.From.File < 0 || move.From.File > 7 {
					t.Errorf("Move %d has invalid from file: %d", i, move.From.File)
				}
			}
			
			t.Logf("Mode %s: Got moves %v", tc.name, testMoves)
		})
	}
}

// Benchmark computer player performance with and without opening book
func BenchmarkComputerPlayerWithBook(b *testing.B) {
	engine := &MockBookEngine{}
	config := SearchConfig{
		MaxDepth: 2,
		MaxTime:  100 * time.Millisecond,
	}
	
	player := NewComputerPlayer("Benchmark", engine, config)
	
	bookPath := "../openings/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		b.Skip("Skipping benchmark: performance.bin not found")
		return
	}
	
	player.SetOpeningBook(true, []string{bookPath})
	board, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.GetMove(board, moves.White, 50*time.Millisecond)
	}
}

func BenchmarkComputerPlayerWithoutBook(b *testing.B) {
	engine := &MockBookEngine{}
	config := SearchConfig{
		MaxDepth:       2,
		MaxTime:        100 * time.Millisecond,
		UseOpeningBook: false,
	}
	
	player := NewComputerPlayer("Benchmark No Book", engine, config)
	board, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.GetMove(board, moves.White, 50*time.Millisecond)
	}
}