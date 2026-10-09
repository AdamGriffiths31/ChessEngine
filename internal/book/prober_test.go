package book

import (
	"os"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestProberProbe(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		fen     string
		wantHit bool
	}{
		{"no book files", nil, startFEN, false},
		{"opening move found", []string{performanceBin}, startFEN, true},
		// Move 11 is past BookMoveLimit (10), so the book is not consulted.
		{"past the move limit", []string{performanceBin}, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 11", false},
		// A bare king ending never occurs in opening theory.
		{"position not in the book", []string{performanceBin}, "8/8/8/4k3/8/8/4K3/8 w - - 0 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.files) > 0 {
				if _, err := os.Stat(tt.files[0]); os.IsNotExist(err) {
					t.Skipf("%s not found", tt.files[0])
				}
			}
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}

			move, ok := NewProber(tt.files).Probe(b)
			if ok != tt.wantHit {
				t.Fatalf("Probe hit = %v, want %v (move %+v)", ok, tt.wantHit, move)
			}
			if ok && move == nil {
				t.Error("Probe reported a hit but returned a nil move")
			}
		})
	}
}

// A book that cannot be loaded is a permanent miss: no panic, and the load is
// attempted once and never retried.
func TestProberLoadFailureIsGracefulAndCached(t *testing.T) {
	p := NewProber([]string{"testdata/does-not-exist.bin"})
	b, err := board.FromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}

	for range 3 {
		if move, ok := p.Probe(b); ok {
			t.Errorf("Probe hit with an unloadable book: %+v", move)
		}
	}
	if p.service != nil {
		t.Error("service should stay nil after a failed load")
	}
	if !p.attempted {
		t.Error("attempted should be set after the first load")
	}
}
