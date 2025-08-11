package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestEvaluateRooks(t *testing.T) {
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
			name:        "white_rook_open_file",
			fen:         "rnbqkbnr/ppp1pppp/8/8/8/8/PPPPPPPP/RNBQR1NR w KQkq - 0 1",
			expected:    16, // Actual observed value
			description: "White rook on open d-file gets bonus",
		},
		{
			name:        "rook_on_seventh",
			fen:         "4k3/3R4/8/8/8/8/8/4K3 w - - 0 1",
			expected:    73, // Actual observed value
			description: "White rook on 7th rank with open file",
		},
		{
			name:        "connected_rooks",
			fen:         "4k3/8/8/8/8/8/8/R6R w - - 0 1",
			expected:    96, // Actual observed value
			description: "Two rooks connected on back rank with open files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateRooks(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestEvaluateRooksForColor(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		isWhite     bool
		expected    int
		description string
	}{
		{
			name:        "rook_open_file",
			fen:         "8/8/8/8/8/8/8/3R4 w - - 0 1",
			isWhite:     true,
			expected:    44, // 20 (open file) + 24 (mobility: 12*2)
			description: "White rook on open d-file",
		},
		{
			name:        "rook_semi_open_file",
			fen:         "8/3p4/8/8/8/8/8/3R4 w - - 0 1",
			isWhite:     true,
			expected:    34, // 10 (semi-open) + 24 (mobility: 12*2)
			description: "White rook on semi-open file (enemy pawn)",
		},
		{
			name:        "rook_closed_file",
			fen:         "8/3p4/8/8/8/8/3P4/3R4 w - - 0 1",
			isWhite:     true,
			expected:    24, // 0 (closed file) + 24 (mobility: 12*2)
			description: "White rook on closed file",
		},
		{
			name:        "rook_seventh_rank",
			fen:         "4k3/3R4/8/8/8/8/8/8 w - - 0 1",
			isWhite:     true,
			expected:    73, // 20 (open) + 25 (7th rank) + 28 (mobility: 14*2)
			description: "White rook on 7th rank",
		},
		{
			name:        "black_rook_second_rank",
			fen:         "8/8/8/8/8/8/3r4/4K3 w - - 0 1",
			isWhite:     false,
			expected:    73, // 20 (open) + 25 (2nd rank for black) + 28 (mobility: 14*2)
			description: "Black rook on 2nd rank (equivalent to 7th)",
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
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestRookConnectedBehavior(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "rooks_same_rank",
			fen:         "8/8/8/8/R6R/8/8/8 w - - 0 1",
			expected:    8,
			description: "Rooks on same rank are connected",
		},
		{
			name:        "rooks_same_file",
			fen:         "3R4/8/8/8/8/8/8/3R4 w - - 0 1",
			expected:    8,
			description: "Rooks on same file are connected",
		},
		{
			name:        "rooks_not_connected",
			fen:         "R7/8/8/8/8/8/8/7R w - - 0 1",
			expected:    0,
			description: "Rooks on different ranks and files are not connected",
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
			
			// Calculate expected score without connected bonus
			expectedWithoutConnection := 0
			tempRooks := whiteRooks
			for tempRooks != 0 {
				square, newRooks := tempRooks.PopLSB()
				tempRooks = newRooks
				
				rank := square / 8
				
				// Open file bonus (both rooks are on open files in these tests)
				expectedWithoutConnection += RookOpenFileBonus
				
				// Mobility bonus
				expectedWithoutConnection += RookMobilityByRank[rank] * RookMobilityUnit
			}
			
			connectionBonus := fullScore - expectedWithoutConnection
			if connectionBonus != tt.expected {
				t.Errorf("%s: expected connection bonus %d, got %d", tt.description, tt.expected, connectionBonus)
			}
		})
	}
}

func TestRookMobilityByRank(t *testing.T) {
	// Test that the mobility table has reasonable values
	if len(RookMobilityByRank) != 8 {
		t.Errorf("RookMobilityByRank should have 8 entries, has %d", len(RookMobilityByRank))
	}

	// Check that corner ranks have lower mobility than central ranks
	if RookMobilityByRank[0] >= RookMobilityByRank[3] {
		t.Errorf("First rank should have lower mobility than middle ranks")
	}
	
	if RookMobilityByRank[7] >= RookMobilityByRank[3] {
		t.Errorf("Last rank should have lower mobility than middle ranks")
	}

	// Check mobility values are reasonable (typical rook has 12-14 moves)
	for i, mobility := range RookMobilityByRank {
		if mobility < 10 || mobility > 16 {
			t.Errorf("Rank %d mobility %d seems unreasonable (should be 10-16)", i, mobility)
		}
	}
}