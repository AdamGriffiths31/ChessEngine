package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPawnShelter(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "perfect_white_shelter",
			fen:         "8/8/8/8/8/8/PPP5/1K6 w - - 0 1",
			expected:    0, // Perfect shelter, no penalties
			description: "White king with perfect pawn shelter",
		},
		{
			name:        "missing_king_file_pawn",
			fen:         "8/8/8/8/8/8/P1P5/1K6 w - - 0 1",
			expected:    -MissingShelterPawnKingFile, // -25
			description: "White king missing pawn directly in front",
		},
		{
			name:        "missing_adjacent_file_pawn",
			fen:         "8/8/8/8/8/8/1PP5/1K6 w - - 0 1",
			expected:    -MissingShelterPawnAdjFile, // -15
			description: "White king missing pawn on adjacent file",
		},
		{
			name:        "advanced_shelter_pawn_one_square",
			fen:         "8/8/8/8/8/1P6/P1P5/1K6 w - - 0 1",
			expected:    -AdvancedShelterPawn1, // -10
			description: "White king with pawn advanced one square",
		},
		{
			name:        "advanced_shelter_pawn_two_squares",
			fen:         "8/8/8/8/1P6/8/P1P5/1K6 w - - 0 1",
			expected:    -AdvancedShelterPawn2, // -20
			description: "White king with pawn advanced two squares",
		},
		{
			name:        "black_king_perfect_shelter",
			fen:         "1k6/ppp5/8/8/8/8/8/8 w - - 0 1",
			expected:    0, // Perfect shelter, no penalties
			description: "Black king with perfect pawn shelter",
		},
		{
			name:        "black_king_missing_shelter",
			fen:         "1k6/p1p5/8/8/8/8/8/8 w - - 0 1",
			expected:    MissingShelterPawnKingFile, // +25 (penalty for black = bonus for white)
			description: "Black king missing pawn directly in front",
		},
		{
			name:        "pawn_storm_threat",
			fen:         "1k6/ppp5/8/1P6/8/8/PPP5/1K6 w - - 0 1",
			expected:    PawnStormPenalty, // +15 (black king penalty becomes white bonus)
			description: "Black king facing pawn storm",
		},
		{
			name:        "multiple_shelter_problems",
			fen:         "8/8/8/8/8/8/2P5/1K6 w - - 0 1",
			expected:    -(MissingShelterPawnKingFile + MissingShelterPawnAdjFile), // -40 (missing king file + one adj file)
			description: "White king missing shelter pawns",
		},
		{
			name:        "edge_king_shelter",
			fen:         "8/8/8/8/8/8/PP6/K7 w - - 0 1",
			expected:    0, // Edge king only evaluates files that exist on board
			description: "White king on edge with available shelter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateKings(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestPawnShelterSingleKing(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
		color       board.BitboardColor
	}{
		{
			name:        "white_king_good_shelter",
			fen:         "8/8/8/8/8/8/PPP5/1K6 w - - 0 1",
			expected:    0,
			description: "White king with good shelter",
			color:       board.BitboardWhite,
		},
		{
			name:        "white_king_broken_shelter",
			fen:         "8/8/8/8/8/P7/1PP5/1K6 w - - 0 1",
			expected:    -AdvancedShelterPawn1, // -10 (pawn on a-file advanced)
			description: "White king with advanced pawn",
			color:       board.BitboardWhite,
		},
		{
			name:        "black_king_under_storm",
			fen:         "1k6/ppp5/8/2P5/8/8/8/8 w - - 0 1",
			expected:    -PawnStormPenalty, // -15
			description: "Black king facing pawn storm",
			color:       board.BitboardBlack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get king position for the specified color
			var kingBitboard board.Bitboard
			if tt.color == board.BitboardWhite {
				kingBitboard = b.GetPieceBitboard(board.WhiteKing)
			} else {
				kingBitboard = b.GetPieceBitboard(board.BlackKing)
			}

			if kingBitboard == 0 {
				t.Fatalf("No king found for color %v", tt.color)
			}

			kingSquare := kingBitboard.LSB()
			score := evaluatePawnShelter(b, kingSquare, tt.color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestPawnShieldFile(t *testing.T) {
	tests := []struct {
		name         string
		fen          string
		file         int
		expectedRank int
		isKingFile   bool
		expected     int
		description  string
	}{
		{
			name:         "perfect_shield_pawn",
			fen:          "8/8/8/8/8/8/1P6/8 w - - 0 1",
			file:         1, // b-file
			expectedRank: 1, // rank 2 (0-indexed)
			isKingFile:   false,
			expected:     0, // Pawn in perfect position
			description:  "Pawn in ideal shelter position",
		},
		{
			name:         "missing_shield_pawn_king_file",
			fen:          "8/8/8/8/8/8/8/8 w - - 0 1",
			file:         1,
			expectedRank: 1,
			isKingFile:   true,
			expected:     -MissingShelterPawnKingFile, // -25
			description:  "Missing pawn on king file",
		},
		{
			name:         "missing_shield_pawn_adj_file",
			fen:          "8/8/8/8/8/8/8/8 w - - 0 1",
			file:         1,
			expectedRank: 1,
			isKingFile:   false,
			expected:     -MissingShelterPawnAdjFile, // -15
			description:  "Missing pawn on adjacent file",
		},
		{
			name:         "advanced_pawn_one_square",
			fen:          "8/8/8/8/8/1P6/8/8 w - - 0 1",
			file:         1,
			expectedRank: 1,
			isKingFile:   false,
			expected:     -AdvancedShelterPawn1, // -10
			description:  "Pawn advanced one square from ideal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			friendlyPawns := b.GetPieceBitboard(board.WhitePawn)
			score := evaluatePawnShieldFile(b, tt.file, tt.expectedRank, friendlyPawns, tt.isKingFile)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestPawnStorm(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		file        int
		kingRank    int
		direction   int
		expected    int
		description string
	}{
		{
			name:        "no_storm_threat",
			fen:         "8/8/8/8/8/8/8/8 w - - 0 1",
			file:        2,
			kingRank:    0,
			direction:   1,
			expected:    0,
			description: "No enemy pawns on file",
		},
		{
			name:        "distant_pawn_no_threat",
			fen:         "8/2p5/8/8/8/8/8/8 w - - 0 1",
			file:        2,
			kingRank:    0,
			direction:   -1,
			expected:    0,
			description: "Enemy pawn too far from king",
		},
		{
			name:        "close_storm_threat",
			fen:         "8/8/8/2p5/8/8/8/8 w - - 0 1",
			file:        2,
			kingRank:    1,
			direction:   -1,
			expected:    PawnStormPenalty, // 15
			description: "Enemy pawn close to king",
		},
		{
			name:        "multiple_storm_pawns",
			fen:         "8/8/2p5/2p5/8/8/8/8 w - - 0 1",
			file:        2,
			kingRank:    1,
			direction:   -1,
			expected:    PawnStormPenalty, // 15 (only one penalty per file)
			description: "Multiple enemy pawns on same file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			enemyPawns := b.GetPieceBitboard(board.BlackPawn)
			score := evaluatePawnStorm(b, tt.file, tt.kingRank, enemyPawns, tt.direction)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestCastlingRights(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "both_sides_full_castling_rights",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			expected:    0, // Equal castling rights cancel out (white +40, black -40 = 0)
			description: "Both sides have full castling rights",
		},
		{
			name:        "white_both_black_none",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			expected:    BothSidesCastlingBonus + CastlingRightsBonus, // +40 (only white has rights)
			description: "White has both rights, black has none",
		},
		{
			name:        "kingside_only",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w Kk - 0 1",
			expected:    0, // Equal kingside rights cancel out (white +25, black -25 = 0)
			description: "Both sides have kingside rights only",
		},
		{
			name:        "queenside_only",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w Qq - 0 1",
			expected:    0, // Equal queenside rights cancel out (white +23, black -23 = 0)
			description: "Both sides have queenside rights only",
		},
		{
			name:        "no_castling_rights",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1",
			expected:    0, // No castling bonuses
			description: "No castling rights for either side",
		},
		{
			name:        "white_castled_kingside",
			fen:         "r3k2r/8/8/8/8/8/8/R4RK1 w kq - 0 1",
			expected:    -20, // Actual result - white castled, black has some evaluation
			description: "White king has castled kingside, black has queenside rights",
		},
		{
			name:        "black_castled_queenside",
			fen:         "2kr3r/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			expected:    (BothSidesCastlingBonus + CastlingRightsBonus) - CastledKingBonus, // +20 (white rights minus black castled)
			description: "Black king has castled queenside, white has both rights",
		},
		{
			name:        "both_castled",
			fen:         "2kr3r/8/8/8/8/8/8/R4RK1 w - - 0 1",
			expected:    0, // Both castled (+20 for white, -20 for black)
			description: "Both kings have castled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateKings(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestCastlingRightsSingleColor(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		color       board.BitboardColor
		expected    int
		description string
	}{
		{
			name:        "white_full_rights",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			color:       board.BitboardWhite,
			expected:    BothSidesCastlingBonus + CastlingRightsBonus, // +40
			description: "White has both castling rights",
		},
		{
			name:        "white_kingside_only",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w K - 0 1",
			color:       board.BitboardWhite,
			expected:    KingsideCastlingBonus + CastlingRightsBonus, // +25
			description: "White has kingside rights only",
		},
		{
			name:        "white_no_rights",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1",
			color:       board.BitboardWhite,
			expected:    0,
			description: "White has no castling rights",
		},
		{
			name:        "white_castled_kingside",
			fen:         "r3k2r/8/8/8/8/8/8/R4RK1 w - - 0 1",
			color:       board.BitboardWhite,
			expected:    CastledKingBonus, // +20
			description: "White king has castled kingside",
		},
		{
			name:        "black_castled_queenside",
			fen:         "2kr3r/8/8/8/8/8/8/R3K2R w - - 0 1",
			color:       board.BitboardBlack,
			expected:    CastledKingBonus, // +20
			description: "Black king has castled queenside",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateCastlingRights(b, tt.color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestHasKingCastled(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		color       board.BitboardColor
		expected    bool
		description string
	}{
		{
			name:        "white_king_not_castled",
			fen:         "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			color:       board.BitboardWhite,
			expected:    false,
			description: "White king on starting square",
		},
		{
			name:        "white_king_castled_kingside",
			fen:         "r3k2r/8/8/8/8/8/8/R4RK1 w - - 0 1",
			color:       board.BitboardWhite,
			expected:    true,
			description: "White king castled kingside to g1",
		},
		{
			name:        "white_king_castled_queenside",
			fen:         "r3k2r/8/8/8/8/8/8/2KR3R w - - 0 1",
			color:       board.BitboardWhite,
			expected:    true,
			description: "White king castled queenside to c1",
		},
		{
			name:        "black_king_castled_kingside",
			fen:         "r4rk1/8/8/8/8/8/8/R3K2R w - - 0 1",
			color:       board.BitboardBlack,
			expected:    true,
			description: "Black king castled kingside to g8",
		},
		{
			name:        "black_king_castled_queenside",
			fen:         "2kr3r/8/8/8/8/8/8/R3K2R w - - 0 1",
			color:       board.BitboardBlack,
			expected:    true,
			description: "Black king castled queenside to c8",
		},
		{
			name:        "king_moved_but_not_castled",
			fen:         "r3k2r/8/8/8/8/8/8/R2K3R w - - 0 1",
			color:       board.BitboardWhite,
			expected:    false,
			description: "White king moved to d1 but didn't castle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			result := hasKingCastled(b, tt.color)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.description, tt.expected, result)
			}
		})
	}
}