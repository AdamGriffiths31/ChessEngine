package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestIsFileOpen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fen         string
		file        int
		description string
		expected    bool
	}{
		{
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			file:        3, // d-file
			description: "d-file closed in starting position",
			expected:    false,
		},
		{
			fen:         "rnbqkbnr/ppp1pppp/8/8/8/8/PPP1PPPP/RNBQKBNR w KQkq - 0 1",
			file:        3, // d-file
			description: "d-file open after both pawns moved",
			expected:    true,
		},
		{
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPP1PPPP/RNBQKBNR w KQkq - 0 1",
			file:        3, // d-file
			description: "d-file semi-open (only white pawn missing)",
			expected:    false,
		},
		{
			fen:         "rnbqkbnr/8/8/8/8/8/8/RNBQKBNR w KQkq - 0 1",
			file:        4, // e-file
			description: "e-file completely open",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			result := isFileOpen(b, tt.file)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.description, tt.expected, result)
			}
		})
	}
}

// abs function is already defined in evaluator.go
