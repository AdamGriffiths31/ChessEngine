package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestKnightOutpostDetection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		expected bool
	}{
		{
			name:     "not_outpost_without_pawn_defense",
			fen:      "8/8/8/3N4/8/8/8/8 w - - 0 1",
			expected: false,
		},
		{
			name:     "not_outpost_with_enemy_pawn_nearby",
			fen:      "8/2p5/8/3N4/8/8/8/8 w - - 0 1",
			expected: false,
		},
		{
			name:     "not_outpost_black_knight_no_support",
			fen:      "8/8/8/8/4n3/8/8/8 w - - 0 1",
			expected: false,
		},
		{
			name:     "not_outpost_d6_no_support",
			fen:      "8/8/3N4/8/8/8/8/8 w - - 0 1",
			expected: false,
		},
		{
			name:     "valid_outpost_with_pawn_support",
			fen:      "8/8/8/3N4/2P1P3/8/8/8 w - - 0 1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			var knights board.Bitboard
			var isWhite bool
			whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
			blackKnights := b.GetPieceBitboard(board.BlackKnight)

			if whiteKnights != 0 {
				knights = whiteKnights
				isWhite = true
			} else {
				knights = blackKnights
				isWhite = false
			}

			score := evaluateKnightsSimple(b, knights, isWhite)

			knightSquare, _ := knights.PopLSB()
			expectedMobility := knightMobilityScore(knightSquare, mobilityArea(b, isWhite))
			hasOutpost := score > expectedMobility

			if hasOutpost != tt.expected {
				t.Errorf("%s: expected outpost %t, got %t (score: %d)", tt.name, tt.expected, hasOutpost, score)
			}
		})
	}
}
