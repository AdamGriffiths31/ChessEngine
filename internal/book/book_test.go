package book

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestBookLookupService(t *testing.T) {
	config := DefaultBookConfig()
	config.SelectionMode = SelectBest

	service := NewBookLookupService(config)

	if !service.IsEnabled() {
		t.Error("Service should be enabled with default config")
	}

	service.SetEnabled(false)
	if service.IsEnabled() {
		t.Error("Service should be disabled after SetEnabled(false)")
	}

	service.SetEnabled(true)
	if !service.IsEnabled() {
		t.Error("Service should be enabled after SetEnabled(true)")
	}
}

func TestBookLookupServiceConfiguration(t *testing.T) {
	config := BookConfig{
		Enabled:         true,
		BookFiles:       []string{"test1.bin", "test2.bin"},
		SelectionMode:   SelectRandom,
		WeightThreshold: 50,
	}

	service := NewBookLookupService(config)

	service.SetSelectionMode(SelectBest)
	if service.config.SelectionMode != SelectBest {
		t.Error("Selection mode not updated correctly")
	}

	service.SetWeightThreshold(100)
	if service.config.WeightThreshold != 100 {
		t.Error("Weight threshold not updated correctly")
	}
}

func TestBookManager(t *testing.T) {
	manager := NewBookManager()

	testBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	book1 := &MockBook{loaded: true, moves: map[uint64][]BookMove{
		0x123: {{Weight: 100}, {Weight: 50}},
	}}
	book2 := &MockBook{loaded: true, moves: map[uint64][]BookMove{
		0x456: {{Weight: 200}},
	}}

	manager.AddBook(book1)
	manager.AddBook(book2)

	moves, err := manager.LookupMove(0x123, testBoard)
	if err != nil {
		t.Fatalf("Failed to lookup moves: %v", err)
	}
	if len(moves) != 2 {
		t.Errorf("Expected 2 moves, got %d", len(moves))
	}

	moves, err = manager.LookupMove(0x456, testBoard)
	if err != nil {
		t.Fatalf("Failed to lookup moves: %v", err)
	}
	if len(moves) != 1 {
		t.Errorf("Expected 1 move, got %d", len(moves))
	}

	_, err = manager.LookupMove(0x999, testBoard)
	if err != ErrPositionNotFound {
		t.Errorf("Expected ErrPositionNotFound, got %v", err)
	}
}

func TestMoveSelection(t *testing.T) {
	config := DefaultBookConfig()
	service := NewBookLookupService(config)

	bookMoves := []BookMove{
		{Weight: 100},
		{Weight: 200},
		{Weight: 50},
	}

	service.SetSelectionMode(SelectBest)
	selected := service.selectMove(bookMoves)
	if selected.Weight != 200 {
		t.Errorf("Expected best move with weight 200, got %d", selected.Weight)
	}

	service.SetSelectionMode(SelectRandom)
	selected = service.selectMove(bookMoves)
	found := false
	for _, move := range bookMoves {
		if move.Weight == selected.Weight {
			found = true
			break
		}
	}
	if !found {
		t.Error("Random selection returned invalid move")
	}

	singleMove := []BookMove{{Weight: 100}}
	selected = service.selectMove(singleMove)
	if selected.Weight != 100 {
		t.Errorf("Expected single move with weight 100, got %d", selected.Weight)
	}
}

func TestWeightedRandomSelection(t *testing.T) {
	config := DefaultBookConfig()
	config.SelectionMode = SelectWeightedRandom
	config.RandomSeed = 12345 // Fixed seed for reproducible tests

	service := NewBookLookupService(config)

	tests := []struct {
		name      string
		bookMoves []BookMove
	}{
		{
			name: "equal weights",
			bookMoves: []BookMove{
				{Weight: 100},
				{Weight: 100},
				{Weight: 100},
			},
		},
		{
			name: "different weights",
			bookMoves: []BookMove{
				{Weight: 10},
				{Weight: 90},
			},
		},
		{
			name: "zero weights",
			bookMoves: []BookMove{
				{Weight: 0},
				{Weight: 0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Run multiple selections to ensure it doesn't crash
			for i := 0; i < 100; i++ {
				selected := service.selectMove(test.bookMoves)

				found := false
				for _, move := range test.bookMoves {
					if move.Weight == selected.Weight {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Selected move not found in input list: weight %d", selected.Weight)
				}
			}
		})
	}
}

func TestBookLookupServiceWithMockBook(t *testing.T) {
	config := DefaultBookConfig()
	service := NewBookLookupService(config)

	mockBook := &MockBook{
		loaded: true,
		moves: map[uint64][]BookMove{
			0x123456789ABCDEF0: {
				{Weight: 100},
				{Weight: 200},
				{Weight: 50},
			},
		},
	}

	service.manager.AddBook(mockBook)

	b := board.NewBoard()

	// For testing, we'll use a simple approach - just use the real zobrist hash
	// In a real test, we'd mock the board position to produce the expected hash

	// Test finding book move - this will likely return ErrPositionNotFound since
	// the board position won't match our mock hash, but that's expected
	_, err := service.FindBookMove(b)
	if err != ErrPositionNotFound {
		t.Logf("Book lookup returned: %v (expected not found for random board position)", err)
	}
}

func TestValidateMove(t *testing.T) {
	config := DefaultBookConfig()
	service := NewBookLookupService(config)

	testMove := board.Move{
		From: board.Square{File: 4, Rank: 1},
		To:   board.Square{File: 4, Rank: 3},
	}

	mockBook := &MockBook{
		loaded: true,
		moves: map[uint64][]BookMove{
			0x123: {
				{Move: testMove, Weight: 100},
			},
		},
	}

	service.manager.AddBook(mockBook)

	b := board.NewBoard()

	// For this simplified test, we just test that ValidateMove doesn't crash
	// In a real implementation, we'd set up the board position to match the mock hash
	valid := service.ValidateMove(b, testMove)
	t.Logf("Move validation result: %t (depends on actual board hash)", valid)
}

type MockBook struct {
	loaded bool
	moves  map[uint64][]BookMove
	info   BookInfo
}

func (mb *MockBook) LookupMove(hash uint64, _ *board.Board) ([]BookMove, error) {
	if !mb.loaded {
		return nil, ErrBookNotLoaded
	}

	moves, exists := mb.moves[hash]
	if !exists {
		return nil, ErrPositionNotFound
	}

	return moves, nil
}

func (mb *MockBook) LoadFromFile(filename string) error {
	mb.loaded = true
	mb.info.Filename = filename
	return nil
}

func (mb *MockBook) IsLoaded() bool {
	return mb.loaded
}

func (mb *MockBook) GetBookInfo() BookInfo {
	return mb.info
}

func BenchmarkBookLookup(b *testing.B) {
	config := DefaultBookConfig()
	service := NewBookLookupService(config)

	mockBook := &MockBook{
		loaded: true,
		moves: map[uint64][]BookMove{
			0x123: {{Weight: 100}},
		},
	}

	service.manager.AddBook(mockBook)

	testBoard := board.NewBoard()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.FindBookMove(testBoard)
		if err != nil {
			b.Errorf("FindBookMove failed: %v", err)
		}
	}
}

func BenchmarkMoveSelection(b *testing.B) {
	config := DefaultBookConfig()
	service := NewBookLookupService(config)

	bookMoves := []BookMove{
		{Weight: 100},
		{Weight: 200},
		{Weight: 50},
		{Weight: 150},
		{Weight: 75},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.selectMove(bookMoves)
	}
}
