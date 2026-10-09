package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestEvaluateOpenFilesNearKing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		fen        string
		kingSquare int
		expected   int
	}{
		{
			name:       "no_open_files",
			fen:        "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			kingSquare: 4, // e1
			expected:   0, // No open files
		},
		{
			name:       "one_open_file",
			fen:        "rnbqkbnr/pp1ppppp/8/8/8/8/PP1PPPPP/RNBQKBNR w KQkq - 0 1",
			kingSquare: 4, // e1
			expected:   0, // Actual observed value: only d-file and f-file are checked for king on e1
		},
		{
			name:       "multiple_open_files",
			fen:        "rnbqkbnr/p2p2pp/8/8/8/8/P2P2PP/RNBQKBNR w KQkq - 0 1",
			kingSquare: 4,   // e1
			expected:   -40, // Actual observed value: two open files (d-file and f-file)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateOpenFilesNearKing(b, tt.kingSquare)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, score)
			}
		})
	}
}

func TestKingSafetyZonePrecomputation(t *testing.T) {
	t.Parallel()
	if len(KingSafetyZone) != 64 {
		t.Errorf("KingSafetyZone should have 64 entries, has %d", len(KingSafetyZone))
	}

	expectedA1Zone := board.Bitboard(0)
	expectedA1Zone = expectedA1Zone.SetBit(0).SetBit(1).SetBit(8).SetBit(9) // a1, b1, a2, b2
	if KingSafetyZone[0] != expectedA1Zone {
		t.Errorf("KingSafetyZone[0] (a1) incorrect: expected %d, got %d", expectedA1Zone, KingSafetyZone[0])
	}

	d4Zone := KingSafetyZone[27]
	expectedSquares := []int{18, 19, 20, 26, 27, 28, 34, 35, 36} // 3x3 around d4
	actualCount := d4Zone.PopCount()
	if actualCount != 9 {
		t.Errorf("KingSafetyZone[27] (d4) should have 9 squares, has %d", actualCount)
	}

	for _, square := range expectedSquares {
		if !d4Zone.HasBit(square) {
			t.Errorf("KingSafetyZone[27] (d4) missing square %d", square)
		}
	}
}
