package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestEvaluateRooksForColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		isWhite  bool
		expected int
	}{
		{
			name:     "rook_open_file",
			fen:      "8/8/8/8/8/8/8/3R4 w - - 0 1",
			isWhite:  true,
			expected: 44, // 20 (open file) + 24 (mobility: 12*2)
		},
		{
			name:     "rook_semi_open_file",
			fen:      "8/3p4/8/8/8/8/8/3R4 w - - 0 1",
			isWhite:  true,
			expected: 34, // 10 (semi-open) + 24 (mobility: 12*2)
		},
		{
			name:     "rook_closed_file",
			fen:      "8/3p4/8/8/8/8/3P4/3R4 w - - 0 1",
			isWhite:  true,
			expected: 24, // 0 (closed file) + 24 (mobility: 12*2)
		},
		{
			name:     "rook_seventh_rank",
			fen:      "4k3/3R4/8/8/8/8/8/8 w - - 0 1",
			isWhite:  true,
			expected: 73, // 20 (open) + 25 (7th rank) + 28 (mobility: 14*2)
		},
		{
			name:     "black_rook_second_rank",
			fen:      "8/8/8/8/8/8/3r4/4K3 w - - 0 1",
			isWhite:  false,
			expected: 73, // 20 (open) + 25 (2nd rank for black) + 28 (mobility: 14*2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			var rooks board.Bitboard
			if tt.isWhite {
				rooks = b.GetPieceBitboard(board.WhiteRook)
			} else {
				rooks = b.GetPieceBitboard(board.BlackRook)
			}

			score := evaluateRooksForColor(b, rooks, tt.isWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, score)
			}
		})
	}
}

func TestRookConnectedBehavior(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		expected int
	}{
		{
			name:     "rooks_same_rank",
			fen:      "8/8/8/8/R6R/8/8/8 w - - 0 1",
			expected: 8,
		},
		{
			name:     "rooks_same_file",
			fen:      "3R4/8/8/8/8/8/8/3R4 w - - 0 1",
			expected: 8,
		},
		{
			name:     "rooks_not_connected",
			fen:      "R7/8/8/8/8/8/8/7R w - - 0 1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			whiteRooks := b.GetPieceBitboard(board.WhiteRook)

			// Calculate only the connected rooks bonus from full evaluation
			fullScore := evaluateRooksForColor(b, whiteRooks, true)

			expectedWithoutConnection := 0
			tempRooks := whiteRooks
			for tempRooks != 0 {
				square, newRooks := tempRooks.PopLSB()
				tempRooks = newRooks

				rank := square / 8

				// Open file bonus (both rooks are on open files in these tests)
				expectedWithoutConnection += RookOpenFileBonus

				expectedWithoutConnection += RookMobilityByRank[rank] * RookMobilityUnit
			}

			connectionBonus := fullScore - expectedWithoutConnection
			if connectionBonus != tt.expected {
				t.Errorf("%s: expected connection bonus %d, got %d", tt.name, tt.expected, connectionBonus)
			}
		})
	}
}
