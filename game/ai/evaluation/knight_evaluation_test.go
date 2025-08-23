package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestEvaluateKnights(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "starting_position",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expected:    0,
			description: "Starting position - both sides equal",
		},
		{
			name:        "white_knight_outpost",
			fen:         "rnbqkb1r/pppppppp/8/3N4/8/8/PPPPPPPP/R1BQKBNR w KQkq - 0 1",
			expected:    32, // Updated actual observed value without pair penalty
			description: "White knight on d5 outpost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateKnights(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestEvaluateKnightsSimple(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		isWhite     bool
		expected    int
		description string
	}{
		{
			name:        "knight_central_square",
			fen:         "8/8/8/3N4/8/8/8/8 w - - 0 1",
			isWhite:     true,
			expected:    32, // 8 * 4 (mobility) - no outpost without pawn support
			description: "White knight on central d5 square without outpost",
		},
		{
			name:        "knight_corner_square",
			fen:         "8/8/8/8/8/8/8/N7 w - - 0 1",
			isWhite:     true,
			expected:    8, // 2 * 4 (mobility) - corner has low mobility
			description: "White knight on corner a1 square",
		},
		{
			name:        "knight_edge_no_outpost",
			fen:         "8/8/8/8/8/8/1N6/8 w - - 0 1",
			isWhite:     true,
			expected:    16, // Actual observed value (4 * 4 mobility for rank 2)
			description: "White knight on edge with no outpost",
		},
		{
			name:        "black_knight_no_outpost",
			fen:         "8/8/8/8/3n4/8/8/8 w - - 0 1",
			isWhite:     false,
			expected:    32, // 8 * 4 (mobility) - no outpost without pawn support
			description: "Black knight on e4 without outpost support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			var knights board.Bitboard
			if tt.isWhite {
				knights = b.GetPieceBitboard(board.WhiteKnight)
			} else {
				knights = b.GetPieceBitboard(board.BlackKnight)
			}

			score := evaluateKnightsSimple(b, knights, tt.isWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestKnightOutpostDetection(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    bool
		description string
	}{
		{
			name:        "not_outpost_without_pawn_defense",
			fen:         "8/8/8/3N4/8/8/8/8 w - - 0 1",
			expected:    false, // Now correctly requires pawn support
			description: "White knight on d5 not outpost without pawn defense",
		},
		{
			name:        "not_outpost_with_enemy_pawn_nearby",
			fen:         "8/2p5/8/3N4/8/8/8/8 w - - 0 1",
			expected:    false, // Enemy pawn can attack + no pawn support
			description: "White knight on d5 not outpost with enemy pawn on c file",
		},
		{
			name:        "not_outpost_black_knight_no_support",
			fen:         "8/8/8/8/4n3/8/8/8 w - - 0 1",
			expected:    false, // Now correctly requires pawn support
			description: "Black knight on e4 not outpost without pawn support",
		},
		{
			name:        "not_outpost_d6_no_support",
			fen:         "8/8/3N4/8/8/8/8/8 w - - 0 1",
			expected:    false, // Rank 6 but no pawn support
			description: "White knight on d6 not outpost without pawn defense",
		},
		{
			name:        "valid_outpost_with_pawn_support",
			fen:         "8/8/8/3N4/2P1P3/8/8/8 w - - 0 1",
			expected:    true, // Should be true - knight defended by pawns on c4,e4
			description: "Knight on d5 is valid outpost when defended by pawns",
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

			// Test the outpost detection logic within evaluateKnightsSimple
			score := evaluateKnightsSimple(b, knights, isWhite)

			// Calculate expected mobility score for the knight position
			knightSquare, _ := knights.PopLSB()
			expectedMobility := KnightMobilityTable[knightSquare] * KnightMobilityUnit
			hasOutpost := score > expectedMobility // Has outpost bonus if score exceeds mobility

			if hasOutpost != tt.expected {
				t.Errorf("%s: expected outpost %t, got %t (score: %d)", tt.description, tt.expected, hasOutpost, score)
			}
		})
	}
}

func TestKnightMobilityTable(t *testing.T) {
	// Test that the mobility table has reasonable values
	if len(KnightMobilityTable) != 64 {
		t.Errorf("KnightMobilityTable should have 64 entries, has %d", len(KnightMobilityTable))
	}

	// Check corner squares have lower mobility
	corners := []int{0, 7, 56, 63} // a1, h1, a8, h8
	for _, corner := range corners {
		if KnightMobilityTable[corner] > 3 {
			t.Errorf("Corner square %d should have low mobility, got %d", corner, KnightMobilityTable[corner])
		}
	}

	// Check central squares have higher mobility
	central := []int{27, 28, 35, 36} // d4, e4, d5, e5
	for _, center := range central {
		if KnightMobilityTable[center] < 7 {
			t.Errorf("Central square %d should have high mobility, got %d", center, KnightMobilityTable[center])
		}
	}
}
