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
			expected: false, // Now correctly requires pawn support
		},
		{
			name:     "not_outpost_with_enemy_pawn_nearby",
			fen:      "8/2p5/8/3N4/8/8/8/8 w - - 0 1",
			expected: false, // Enemy pawn can attack + no pawn support
		},
		{
			name:     "not_outpost_black_knight_no_support",
			fen:      "8/8/8/8/4n3/8/8/8 w - - 0 1",
			expected: false, // Now correctly requires pawn support
		},
		{
			name:     "not_outpost_d6_no_support",
			fen:      "8/8/3N4/8/8/8/8/8 w - - 0 1",
			expected: false, // Rank 6 but no pawn support
		},
		{
			name:     "valid_outpost_with_pawn_support",
			fen:      "8/8/8/3N4/2P1P3/8/8/8 w - - 0 1",
			expected: true, // Should be true - knight defended by pawns on c4,e4
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
			expectedMobility := KnightMobilityTable[knightSquare] * KnightMobilityUnit
			hasOutpost := score > expectedMobility // Has outpost bonus if score exceeds mobility

			if hasOutpost != tt.expected {
				t.Errorf("%s: expected outpost %t, got %t (score: %d)", tt.name, tt.expected, hasOutpost, score)
			}
		})
	}
}
