package book

import (
	"errors"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// testBookConfig is a weighted-random config with no book files.
func testBookConfig() Config {
	return Config{Enabled: true, SelectionMode: SelectWeightedRandom, WeightThreshold: 1}
}

func TestBookManager(t *testing.T) {
	manager := NewManager()

	testBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	book1 := &MockBook{loaded: true, moves: map[uint64][]Move{
		0x123: {{Weight: 100}, {Weight: 50}},
	}}
	book2 := &MockBook{loaded: true, moves: map[uint64][]Move{
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
	if !errors.Is(err, ErrPositionNotFound) {
		t.Errorf("Expected ErrPositionNotFound, got %v", err)
	}
}

func TestMoveSelection(t *testing.T) {
	config := testBookConfig()
	service := NewLookupService(config)

	bookMoves := []Move{
		{Weight: 100},
		{Weight: 200},
		{Weight: 50},
	}

	service.config.SelectionMode = SelectBest
	selected := service.selectMove(bookMoves)
	if selected.Weight != 200 {
		t.Errorf("Expected best move with weight 200, got %d", selected.Weight)
	}

	service.config.SelectionMode = SelectRandom
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

	singleMove := []Move{{Weight: 100}}
	selected = service.selectMove(singleMove)
	if selected.Weight != 100 {
		t.Errorf("Expected single move with weight 100, got %d", selected.Weight)
	}
}

func TestWeightedRandomSelection(t *testing.T) {
	config := testBookConfig()
	config.SelectionMode = SelectWeightedRandom
	config.RandomSeed = 12345 // Fixed seed for reproducible tests

	service := NewLookupService(config)

	tests := []struct {
		name      string
		bookMoves []Move
	}{
		{
			name: "equal weights",
			bookMoves: []Move{
				{Weight: 100},
				{Weight: 100},
				{Weight: 100},
			},
		},
		{
			name: "different weights",
			bookMoves: []Move{
				{Weight: 10},
				{Weight: 90},
			},
		},
		{
			name: "zero weights",
			bookMoves: []Move{
				{Weight: 0},
				{Weight: 0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Run multiple selections to ensure it doesn't crash
			for range 100 {
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

type MockBook struct {
	loaded bool
	moves  map[uint64][]Move
	info   Info
}

func (mb *MockBook) LookupMove(hash uint64, _ *board.Board) ([]Move, error) {
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

func (mb *MockBook) GetBookInfo() Info {
	return mb.info
}
