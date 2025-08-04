package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestOpenFileDetection(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "rook_on_open_file",
			fen:         "8/8/8/8/8/8/8/3R4 w - - 0 1",
			expected:    RookOpenFileBonus, // 25
			description: "White rook on open d-file",
		},
		{
			name:        "rook_on_semi_open_file",
			fen:         "8/3p4/8/8/8/8/8/3R4 w - - 0 1",
			expected:    RookSemiOpenFileBonus, // 15
			description: "White rook on semi-open file (black pawn)",
		},
		{
			name:        "rook_on_closed_file",
			fen:         "8/3p4/8/8/8/8/3P4/3R4 w - - 0 1",
			expected:    0, // No file bonus
			description: "White rook on closed file",
		},
		{
			name:        "black_rook_open_file",
			fen:         "3r4/8/8/8/8/8/8/8 w - - 0 1",
			expected:    RookOpenFileBonus, // 25 (but will be negated in main evaluation)
			description: "Black rook on open d-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get rook position and evaluate open files directly
			whiteRooks := b.GetPieceBitboard(board.WhiteRook)
			blackRooks := b.GetPieceBitboard(board.BlackRook)

			var rookSquare int
			var friendlyPawns, enemyPawns board.Bitboard

			if whiteRooks != 0 {
				rookSquare, _ = whiteRooks.PopLSB()
				friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
				enemyPawns = b.GetPieceBitboard(board.BlackPawn)
			} else if blackRooks != 0 {
				rookSquare, _ = blackRooks.PopLSB()
				friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
				enemyPawns = b.GetPieceBitboard(board.WhitePawn)
			}

			score := evaluateOpenFiles(rookSquare, friendlyPawns, enemyPawns)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestSeventhRankBonus(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "white_rook_7th_black_king_8th",
			fen:         "3k4/3R4/8/8/8/8/8/8 w - - 0 1",
			description: "White rook on 7th, black king on 8th",
			expected:    RookOnSeventhBonus, // 30
		},
		{
			name:        "white_rook_7th_no_king_8th",
			fen:         "8/3R4/3k4/8/8/8/8/8 w - - 0 1",
			description: "White rook on 7th, black king not on 8th",
			expected:    10,
		},
		{
			name:        "black_rook_2nd_white_king_1st",
			fen:         "3k4/8/8/8/8/8/3r4/K7 w - - 0 1",
			description: "Black rook on 2nd, white king on 1st",
			expected:    RookOnSeventhBonus, // 30
		},
		{
			name:        "rook_not_on_seventh",
			fen:         "3k4/8/3R4/8/8/8/8/8 w - - 0 1",
			description: "White rook not on 7th rank",
			expected:    0, // No bonus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get rook position and evaluate seventh rank directly
			whiteRooks := b.GetPieceBitboard(board.WhiteRook)
			blackRooks := b.GetPieceBitboard(board.BlackRook)

			var rookSquare int
			var enemyKing board.Bitboard
			var color board.BitboardColor

			if whiteRooks != 0 {
				rookSquare, _ = whiteRooks.PopLSB()
				enemyKing = b.GetPieceBitboard(board.BlackKing)
				color = board.BitboardWhite
			} else if blackRooks != 0 {
				rookSquare, _ = blackRooks.PopLSB()
				enemyKing = b.GetPieceBitboard(board.WhiteKing)
				color = board.BitboardBlack
			}

			score := evaluateSeventhRank(rookSquare, enemyKing, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestDoubledRooks(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "white_doubled_rooks_file",
			fen:         "3R4/8/8/8/8/8/8/3R4 w - - 0 1",
			description: "White rooks doubled on d-file",
			expected:    DoubledRooksFileBonus + ConnectedRooksBonus + 10, //Open File
		},
		{
			name:        "white_doubled_rooks_rank",
			fen:         "8/8/8/8/R6R/8/8/8 w - - 0 1",
			description: "White rooks doubled on 4th rank",
			expected:    DoubledRooksRankBonus + 10, //Open File
		},
		{
			name:        "white_rooks_7th_rank",
			fen:         "4k3/R6R/8/8/8/8/8/8 w - - 0 1",
			description: "White rooks both on 7th rank",
			expected:    DoubledRooksRankBonus + RookPairSeventhBonus + ConnectedRooksBonus,
		},
		{
			name:        "disconnected_rooks",
			fen:         "R7/8/8/8/8/8/8/7R w - - 0 1",
			description: "White rooks not connected",
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			whiteRooks := b.GetPieceBitboard(board.WhiteRook)

			score := evaluateRookPairs(b, whiteRooks, board.BitboardWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestRookMobility(t *testing.T) {
	tests := []struct {
		fen         string
		description string
		expected    int
	}{
		{
			fen:         "8/8/8/3R4/8/8/8/8 w - - 0 1",
			description: "Rook with full mobility",
			expected:    42,
		},
		{
			fen:         "8/8/3K4/3R4/8/8/8/8 w - - 0 1",
			description: "Rook with one direction blocked",
			expected:    33,
		},
		{
			fen:         "8/8/3K4/3R4/3N4/8/8/8 w - - 0 1",
			description: "Rook with two directions blocked",
			expected:    21,
		},
		{
			fen:         "8/8/3K4/2NRN3/3N4/8/8/8 w - - 0 1",
			description: "Rook with all directions blocked",
			expected:    RookTrappedPenalty, // -50
		},
		{
			fen:         "8/8/3K4/3R1N2/3N4/8/8/8 w - - 0 1",
			description: "Rook with 4 moves (partially trapped)",
			expected:    RookPartiallyTrapped, // -25
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get rook position and evaluate seventh rank directly
			whiteRooks := b.GetPieceBitboard(board.WhiteRook)
			blackRooks := b.GetPieceBitboard(board.BlackRook)

			var rookSquare int
			var color board.BitboardColor

			if whiteRooks != 0 {
				rookSquare, _ = whiteRooks.PopLSB()
				color = board.BitboardWhite
			} else if blackRooks != 0 {
				rookSquare, _ = blackRooks.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateRookMobility(b, rookSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestRookTrappedByKing(t *testing.T) {
	tests := []struct {
		fen         string
		description string
		expected    int
	}{
		{
			fen:         "8/8/8/8/8/8/8/5QKR w - - 0 1",
			description: "White rook trapped",
			expected:    RookTrappedByKing, // -30
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get rook position and evaluate seventh rank directly
			whiteRooks := b.GetPieceBitboard(board.WhiteRook)
			blackRooks := b.GetPieceBitboard(board.BlackRook)

			var rookSquare int
			var color board.BitboardColor

			if whiteRooks != 0 {
				rookSquare, _ = whiteRooks.PopLSB()
				color = board.BitboardWhite
			} else if blackRooks != 0 {
				rookSquare, _ = blackRooks.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateRookTrappedByKing(b, rookSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}
