package player

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// MockBookEngine implements the Engine interface for book testing
type MockBookEngine struct {
}

func (m *MockBookEngine) FindBestMove(_ context.Context, _ *board.Board, _ movegen.Player, config search.SearchConfig) search.SearchResult {
	return search.SearchResult{
		BestMove: board.Move{
			From: board.Square{File: 4, Rank: 1}, // e2
			To:   board.Square{File: 4, Rank: 3}, // e4
		},
		Score: 0,
		Stats: search.SearchStats{
			NodesSearched: 100,
			Depth:         config.MaxDepth,
			Time:          time.Millisecond * 10,
		},
	}
}

func (m *MockBookEngine) SetEvaluator(_ eval.Evaluator) {
}

func (m *MockBookEngine) GetName() string {
	return "Mock Book Engine"
}

func TestComputerPlayerOpeningBook(t *testing.T) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth:  3,
		MaxTime:   time.Second,
		DebugMode: false,
	}

	player := NewComputerPlayer("Test Engine", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	if player.IsUsingOpeningBook() {
		t.Error("Expected opening book to be disabled initially")
	}

	player.SetOpeningBook(true, []string{bookPath})

	if !player.IsUsingOpeningBook() {
		t.Error("Expected opening book to be enabled after SetOpeningBook")
	}

	bookFiles := player.GetBookFiles()
	if len(bookFiles) != 1 || bookFiles[0] != bookPath {
		t.Errorf("Expected book files [%s], got %v", bookPath, bookFiles)
	}

	if !player.IsUsingOpeningBook() {
		t.Error("Expected opening book to be enabled")
	}

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}

	move, err := player.GetMove(b, movegen.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move: %v", err)
	}

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
	config := search.SearchConfig{
		MaxDepth:  2,
		MaxTime:   500 * time.Millisecond,
		DebugMode: true,
	}

	player := NewComputerPlayer("Test Engine with Book", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	player.SetOpeningBook(true, []string{bookPath})

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	result, err := player.GetMoveWithStats(b, movegen.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move with stats: %v", err)
	}

	t.Logf("Move: %c%d-%c%d",
		'a'+result.BestMove.From.File, result.BestMove.From.Rank+1,
		'a'+result.BestMove.To.File, result.BestMove.To.Rank+1)
	t.Logf("Stats: %d nodes, depth %d, time %v",
		result.Stats.NodesSearched, result.Stats.Depth, result.Stats.Time)

	// The starting position is in performance.bin, so ComputerPlayer must
	// return the book move directly (BookMoveUsed set, engine never
	// invoked, so NodesSearched stays at its zero value).
	if !result.Stats.BookMoveUsed {
		t.Error("Expected BookMoveUsed to be true for a position in the opening book")
	}
	if result.Stats.NodesSearched != 0 {
		t.Errorf("Expected no search when book move is used, got NodesSearched=%d", result.Stats.NodesSearched)
	}
}

// TestComputerPlayerBookMoveLimit verifies that the opening book is not
// consulted once the position is past book.BookMoveLimit: GetMoveWithStats
// must fall through to the engine, exactly like an ordinary search.
func TestComputerPlayerBookMoveLimit(t *testing.T) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  500 * time.Millisecond,
	}

	player := NewComputerPlayer("Test Engine Past Book Limit", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	player.SetOpeningBook(true, []string{bookPath})

	// Same starting position, but tagged as full-move 15: past
	// book.BookMoveLimit (10), so the book must be skipped regardless of
	// whether the position itself is in the book.
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 15")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	result, err := player.GetMoveWithStats(b, movegen.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move with stats: %v", err)
	}

	if result.Stats.BookMoveUsed {
		t.Error("Expected BookMoveUsed to be false past book.BookMoveLimit")
	}
	if result.Stats.NodesSearched != 100 {
		t.Errorf("Expected the mock engine's search result (NodesSearched=100), got %d", result.Stats.NodesSearched)
	}
}

// TestComputerPlayerBookMiss verifies that a position not present in the
// opening book gracefully falls back to a full search, rather than erroring
// or returning a zero-value move.
func TestComputerPlayerBookMiss(t *testing.T) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  500 * time.Millisecond,
	}

	player := NewComputerPlayer("Test Engine Book Miss", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	player.SetOpeningBook(true, []string{bookPath})

	// A bare king-vs-king position never arises from opening theory, so it
	// cannot be in a Polyglot opening book, regardless of book contents.
	b, err := board.FromFEN("8/8/8/4k3/8/8/4K3/8 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	result, err := player.GetMoveWithStats(b, movegen.White, time.Second)
	if err != nil {
		t.Fatalf("Failed to get move with stats: %v", err)
	}

	if result.Stats.BookMoveUsed {
		t.Error("Expected BookMoveUsed to be false for a position not in the opening book")
	}
	if result.Stats.NodesSearched != 100 {
		t.Errorf("Expected the mock engine's search result (NodesSearched=100), got %d", result.Stats.NodesSearched)
	}
}

func TestComputerPlayerBookModes(t *testing.T) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  time.Second,
	}

	player := NewComputerPlayer("Book Mode Test", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	player.SetOpeningBook(true, []string{bookPath})

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	move, err := player.GetMove(b, movegen.White, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to get move: %v", err)
	}

	if move.From.File < 0 || move.From.File > 7 {
		t.Errorf("Move has invalid from file: %d", move.From.File)
	}
	if move.To.File < 0 || move.To.File > 7 {
		t.Errorf("Move has invalid to file: %d", move.To.File)
	}
}

// Benchmark computer player performance with and without opening book
func BenchmarkComputerPlayerWithBook(b *testing.B) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth: 2,
		MaxTime:  100 * time.Millisecond,
	}

	player := NewComputerPlayer("Benchmark", engine, config)

	bookPath := "../../internal/book/testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		b.Skip("Skipping benchmark: performance.bin not found")
		return
	}

	player.SetOpeningBook(true, []string{bookPath})
	board, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to create board from FEN: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := player.GetMove(board, movegen.White, 50*time.Millisecond)
		if err != nil {
			b.Errorf("GetMove failed: %v", err)
		}
	}
}

func BenchmarkComputerPlayerWithoutBook(b *testing.B) {
	engine := &MockBookEngine{}
	config := search.SearchConfig{
		MaxDepth:       2,
		MaxTime:        100 * time.Millisecond,
		UseOpeningBook: false,
	}

	player := NewComputerPlayer("Benchmark No Book", engine, config)
	board, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to create board from FEN: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := player.GetMove(board, movegen.White, 50*time.Millisecond)
		if err != nil {
			b.Errorf("GetMove failed: %v", err)
		}
	}
}
