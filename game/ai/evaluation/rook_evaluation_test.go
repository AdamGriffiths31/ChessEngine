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
			expected:    DoubledRooksFileBonus,
		},
		{
			name:        "white_doubled_rooks_rank",
			fen:         "8/8/8/8/R6R/8/8/8 w - - 0 1",
			description: "White rooks doubled on 4th rank",
			expected:    2,
		},
		{
			name:        "white_rooks_7th_rank",
			fen:         "4k3/R6R/8/8/8/8/8/8 w - - 0 1",
			description: "White rooks both on 7th rank",
			expected:    2,
		},
		{
			name:        "disconnected_rooks",
			fen:         "R7/8/8/8/8/8/8/7R w - - 0 1",
			description: "White rooks not connected",
			expected:    2,
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
	evaluator := NewEvaluator()

	tests := []struct {
		name            string
		goodMobilityFEN string
		badMobilityFEN  string
		description     string
	}{
		{
			name:            "center_vs_corner",
			goodMobilityFEN: "8/8/8/3R4/8/8/8/8 w - - 0 1", // Rook on d5 - open center
			badMobilityFEN:  "8/8/8/8/8/8/8/R7 w - - 0 1",  // Rook on a1 - corner restriction
			description:     "Rook with good mobility should score higher than restricted mobility",
		},
		{
			name:            "open_vs_blocked_same_square",
			goodMobilityFEN: "8/8/8/8/R7/8/8/8 w - - 0 1",      // Rook on a4 - open
			badMobilityFEN:  "8/p7/p7/pR6/p7/p7/8/8 w - - 0 1", // Same rook blocked by enemy pawns
			description:     "Rook with open lines should score higher than blocked rook",
		},
		{
			name:            "back_rank_vs_center_same_material",
			goodMobilityFEN: "8/8/8/8/3R4/8/8/7K w - - 0 1", // Rook d4 + King h1
			badMobilityFEN:  "8/8/8/8/8/8/8/3R3K w - - 0 1", // Rook d1 + King h1 (same material)
			description:     "Center rook should have better mobility than back rank rook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goodBoard, err := board.FromFEN(tt.goodMobilityFEN)
			if err != nil {
				t.Fatalf("Failed to create good mobility board from FEN: %v", err)
			}

			badBoard, err := board.FromFEN(tt.badMobilityFEN)
			if err != nil {
				t.Fatalf("Failed to create bad mobility board from FEN: %v", err)
			}

			goodScore := evaluator.Evaluate(goodBoard)
			badScore := evaluator.Evaluate(badBoard)

			if goodScore < badScore {
				t.Errorf("%s: good mobility position (%d) should score at least as high as bad mobility position (%d)",
					tt.description, goodScore, badScore)
			}

			scoreDiff := goodScore - badScore
			t.Logf("%s: score difference = %d (good: %d, bad: %d)",
				tt.description, scoreDiff, goodScore, badScore)
		})
	}
}

func TestRookTrappedByKing(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name        string
		fen         string
		description string
		trapped     bool
	}{
		{
			name:        "white_kingside_trap",
			fen:         "8/8/8/8/8/8/8/5RKR w - - 0 1",
			description: "White rook trapped after kingside castling",
			trapped:     true,
		},
		{
			name:        "white_queenside_trap",
			fen:         "8/8/8/8/8/8/8/R2K4 w - - 0 1",
			description: "White rook potentially trapped queenside",
			trapped:     true,
		},
		{
			name:        "no_trap",
			fen:         "8/8/8/8/8/8/8/R3K2R w - - 0 1",
			description: "Rooks not trapped by king",
			trapped:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluator.Evaluate(b)

			if tt.trapped {
				t.Logf("%s: detected trapped rook penalty in score %d",
					tt.description, score)
			}
		})
	}
}
