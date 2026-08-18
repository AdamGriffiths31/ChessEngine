package book

import (
	"os"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestProberNoFilesAlwaysMisses(t *testing.T) {
	p := NewProber(nil)

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	if move, ok := p.Probe(b); ok {
		t.Errorf("Expected a miss with no book files configured, got move %+v", move)
	}
}

func TestProberPastMoveLimitAlwaysMisses(t *testing.T) {
	bookPath := "testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	p := NewProber([]string{bookPath})

	// Full-move number 11 is past BookMoveLimit (10).
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 11")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	if move, ok := p.Probe(b); ok {
		t.Errorf("Expected a miss past BookMoveLimit, got move %+v", move)
	}
}

func TestProberFindsOpeningMove(t *testing.T) {
	bookPath := "testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	p := NewProber([]string{bookPath})

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	move, ok := p.Probe(b)
	if !ok {
		t.Fatal("Expected a book move for the starting position")
	}
	if move == nil {
		t.Fatal("Probe reported a hit but returned a nil move")
	}
}

func TestProberMissesUnknownPosition(t *testing.T) {
	bookPath := "testdata/performance.bin"
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skip("Skipping test: performance.bin not found")
		return
	}

	p := NewProber([]string{bookPath})

	// A bare king-vs-king position never arises from opening theory.
	b, err := board.FromFEN("8/8/8/4k3/8/8/4K3/8 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	if move, ok := p.Probe(b); ok {
		t.Errorf("Expected a miss for a position outside the opening book, got move %+v", move)
	}
}

func TestProberLoadFailureIsGracefulAndCached(t *testing.T) {
	p := NewProber([]string{"testdata/does-not-exist.bin"})

	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// First call attempts (and fails) to load; subsequent calls must not
	// panic or retry the load, just keep reporting a miss.
	for i := 0; i < 3; i++ {
		if move, ok := p.Probe(b); ok {
			t.Errorf("Expected a miss when the book file cannot be loaded, got move %+v", move)
		}
	}
	if p.service != nil {
		t.Error("Expected service to remain nil after a failed load")
	}
	if !p.attempted {
		t.Error("Expected attempted to be set after the first load attempt")
	}
}
