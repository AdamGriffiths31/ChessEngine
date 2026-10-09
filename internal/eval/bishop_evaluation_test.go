package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestEvaluateBishopPairBonus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		expected int
	}{
		{
			name:     "white_bishop_pair",
			fen:      "8/8/8/3B4/8/8/1B6/8 w - - 0 1", // d5 and b2
			expected: BishopPairBonus,
		},
		{
			name:     "black_bishop_pair",
			fen:      "8/1b6/8/2b5/8/8/8/8 w - - 0 1", // b7 (light) and c5 (dark)
			expected: -BishopPairBonus,
		},
		{
			name:     "same_color_bishops",
			fen:      "8/8/2B5/8/8/8/2B5/8 w - - 0 1",
			expected: 0,
		},
		{
			name:     "both_have_pairs",
			fen:      "8/1b6/2b5/8/8/2B5/1B6/8 w - - 0 1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
			blackBishops := b.GetPieceBitboard(board.BlackBishop)

			score := evaluateBishopPairBonus(whiteBishops, blackBishops)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, score)
			}
		})
	}
}

func TestEvaluateBadBishop(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		expected int
	}{
		{
			name:     "bad_light_bishop",
			fen:      "8/8/8/2P1P3/1P1B1P2/2P1P3/8/8 w - - 0 1",
			expected: -48, // Bishop on d4 (dark) with all 6 pawns on dark squares: 6 * -8
		},
		{
			name:     "good_bishop_different_colors",
			fen:      "8/8/8/1p1p4/2B5/1p1p4/8/8 w - - 0 1",
			expected: 0, // No own pawns blocking
		},
		{
			name:     "bad_dark_bishop",
			fen:      "8/8/1p2p3/p1b1p3/1p2p3/8/8/8 w - - 0 1",
			expected: -32, // Bishop on c5 (dark) with pawns b6/a5/e5/b4 on dark squares: 4 * -8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			var bishops board.Bitboard
			var ownPawns board.Bitboard
			isWhite := tt.name != "bad_dark_bishop" // Determine from test name for simplicity

			if isWhite {
				bishops = b.GetPieceBitboard(board.WhiteBishop)
				ownPawns = b.GetPieceBitboard(board.WhitePawn)
			} else {
				bishops = b.GetPieceBitboard(board.BlackBishop)
				ownPawns = b.GetPieceBitboard(board.BlackPawn)
			}

			if bishops == 0 {
				t.Fatalf("No bishop found in position")
			}

			bishopSquare, _ := bishops.PopLSB()
			score := evaluateBadBishop(bishopSquare, ownPawns)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, score)
			}
		})
	}
}

func TestEvaluateFianchetto(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		square   int
		isWhite  bool
		expected int
	}{
		{
			name:     "white_b2_fianchetto",
			square:   9, // b2
			isWhite:  true,
			expected: FianchettoBishopBonus,
		},
		{
			name:     "white_g2_fianchetto",
			square:   14, // g2
			isWhite:  true,
			expected: FianchettoBishopBonus,
		},
		{
			name:     "black_b7_fianchetto",
			square:   49, // b7
			isWhite:  false,
			expected: FianchettoBishopBonus,
		},
		{
			name:     "black_g7_fianchetto",
			square:   54, // g7
			isWhite:  false,
			expected: FianchettoBishopBonus,
		},
		{
			name:     "not_fianchetto_square",
			square:   27, // d4
			isWhite:  true,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := evaluateFianchetto(tt.square, tt.isWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, score)
			}
		})
	}
}
