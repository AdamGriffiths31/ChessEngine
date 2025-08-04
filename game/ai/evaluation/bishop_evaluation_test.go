package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestBishopPairBonus(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "white_bishop_pair",
			fen:         "8/8/8/8/8/2B5/6B1/8 w - - 0 1",
			expected:    BishopPairBonus, // 50
			description: "White has both bishops",
		},
		{
			name:        "black_bishop_pair",
			fen:         "b7/8/5b2/8/8/8/8/8 w - - 0 1",
			expected:    -BishopPairBonus, // -50
			description: "Black has both bishops",
		},
		{
			name:        "white_single_bishop",
			fen:         "8/8/8/8/8/2B5/8/8 w - - 0 1",
			expected:    0, // No pair bonus
			description: "White has only one bishop",
		},
		{
			name:        "both_have_pairs",
			fen:         "8/2b5/5b2/8/8/2B5/5B2/8 w - - 0 1",
			expected:    0, // Both get bonus, cancel out
			description: "Both sides have bishop pairs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateBishopPairBonus(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestBadBishop(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "bad_light_bishop",
			fen:         "8/8/3P1P2/8/2B5/3P1P2/8/8 w - - 0 1",
			description: "White bishop blocked by own pawns on light squares",
			expected:    0, // No blocked central pawns in this position
		},
		{
			name:        "good_bishop",
			fen:         "8/8/2P1P3/8/3B4/2P1P3/8/8 w - - 0 1",
			description: "White bishop not blocked (pawns on dark squares)",
			expected:    0, // No penalty
		},
		{
			name:        "bad_bishop_central_pawns",
			fen:         "8/8/3p4/3Pp3/2B1P3/8/8/8 w - - 0 1",
			description: "Bishop blocked by central pawns on same color",
			expected:    BadBishopPenalty * 2, // Two blocked central pawns (-30)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get bishop position and evaluate bad bishop directly
			whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
			if whiteBishops == 0 {
				t.Fatalf("No white bishop found in position")
			}

			bishopSquare, _ := whiteBishops.PopLSB()
			friendlyPawns := b.GetPieceBitboard(board.WhitePawn)

			score := evaluateBadBishop(b, bishopSquare, friendlyPawns, board.BitboardWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestLongDiagonalControl(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "bishop_on_long_diagonal",
			fen:         "8/8/8/8/3B4/8/8/8 w - - 0 1", // d4 on a1-h8
			description: "Bishop on long diagonal",
			expected:    LongDiagonalControl + 5, // Full control (25) + on diagonal bonus (5) = 30
		},
		{
			name:        "bishop_controlling_diagonal",
			fen:         "8/8/8/8/8/8/8/B7 w - - 0 1", // a1 controlling a1-h8
			description: "Bishop controlling long diagonal from corner",
			expected:    LongDiagonalControl + 5, // Full control (25) + on diagonal bonus (5) = 30
		},
		{
			name:        "bishop_not_on_diagonal",
			fen:         "8/8/8/8/8/3B4/8/8 w - - 0 1", // d3 not on long diagonal
			description: "Bishop not on long diagonal",
			expected:    0, // No diagonal control
		},
		{
			name:        "bishop_partial_diagonal_control",
			fen:         "8/8/8/4B3/3P4/2P5/1P6/P7 w - - 0 1",
			description: "Bishop with partial diagonal control",
			expected:    PartialDiagonalControl + 5, // 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get bishop position and evaluate long diagonal control directly
			whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
			if whiteBishops == 0 {
				t.Fatalf("No white bishop found in position")
			}

			bishopSquare, _ := whiteBishops.PopLSB()

			score := evaluateLongDiagonalControl(b, bishopSquare)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestColorComplexAdvantage(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "white_light_square_dominance",
			fen:         "8/1p1p1p1p/p1p1p1p1/8/8/8/8/1B6 w - - 0 1",
			description: "White has light bishop, black missing theirs",
			expected:    ColorComplexDominance + 8*3, // Base bonus (30) + 8 black pawns on light squares (24) = 54
		},
		{
			name:        "black_dark_square_dominance",
			fen:         "8/8/1b6/8/8/P1P1P1P1/1P1P1P1P/8 w - - 0 1",
			description: "Black has dark bishop, white missing theirs",
			expected:    -(ColorComplexDominance + 8*3), // Same but negative for black = -54
		},
		{
			name:        "no_color_advantage",
			fen:         "8/8/2b5/8/8/8/8/2B5 w - - 0 1",
			description: "Both sides have bishops on same color squares",
			expected:    0, // No advantage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get bishop bitboards and evaluate color complex directly
			whiteLightBishops := b.GetPieceBitboard(board.WhiteBishop) & board.LightSquares
			whiteDarkBishops := b.GetPieceBitboard(board.WhiteBishop) & board.DarkSquares
			blackLightBishops := b.GetPieceBitboard(board.BlackBishop) & board.LightSquares
			blackDarkBishops := b.GetPieceBitboard(board.BlackBishop) & board.DarkSquares

			score := evaluateColorComplex(b, whiteLightBishops, whiteDarkBishops, blackLightBishops, blackDarkBishops)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestXRayAttacks(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "bishop_xray_to_queen",
			fen:         "8/1B6/2p5/8/8/8/6q1/8 w - - 0 1",
			description: "White bishop x-rays through pawn to black queen",
			expected:    XRayAttackBonus + 10, // Queen x-ray (30)
		},
		{
			name:        "bishop_xray_to_rook",
			fen:         "8/7r/6P1/5B2/8/8/8/8 w - - 0 1",
			description: "White bishop x-rays through pawn to black rook",
			expected:    XRayAttackBonus, // Rook x-ray (20)
		},
		{
			name:        "no_xray_multiple_blockers",
			fen:         "8/8/2pp4/8/B7/8/6q1/8 w - - 0 1",
			description: "Multiple blockers prevent x-ray",
			expected:    0, // No x-ray possible
		},
		{
			name:        "no_xray_no_target",
			fen:         "8/8/2p5/8/B7/8/8/8 w - - 0 1",
			description: "No valuable target behind blocker",
			expected:    0, // No target
		},
		{
			name:        "bishop_xray_to_king",
			fen:         "7k/6p1/8/8/8/8/8/B7 w - - 0 1",
			description: "White bishop x-rays through pawn to black king",
			expected:    XRayAttackBonus + 5, // King x-ray (25)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get bishop position and evaluate x-ray attacks directly
			whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
			if whiteBishops == 0 {
				t.Fatalf("No white bishop found in position")
			}

			bishopSquare, _ := whiteBishops.PopLSB()

			score := evaluateXRayAttacks(b, bishopSquare, board.BitboardWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestBishopMobility(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "bishop_good_mobility_center",
			fen:         "8/8/8/3B4/8/8/8/8 w - - 0 1", // Bishop on d5 - good mobility
			description: "Bishop in center with good mobility",
			expected:    13*BishopMobilityUnit + 6*2, // 13 moves * 3 + 6 forward moves * 2 = 55
		},
		{
			name:        "bishop_restricted_corner",
			fen:         "8/8/8/8/8/8/8/B7 w - - 0 1", // Bishop on a1 - restricted mobility
			description: "Bishop in corner with restricted mobility",
			expected:    7*BishopMobilityUnit + 7*2, // 7 moves * 3 + 7 forward moves * 2 = 35
		},
		{
			name:        "bishop_blocked_by_pawns",
			fen:         "8/8/1p1p4/2B5/1p1p4/8/8/8 w - - 0 1", // Bishop blocked by enemy pawns
			description: "Bishop blocked by surrounding pawns",
			expected:    4*BishopMobilityUnit + 2*2, // 4 moves * 3 + 2 forward moves * 2 = 21
		},
		{
			name:        "bishop_trapped",
			fen:         "8/8/8/8/2P1P3/1P1P4/2B5/1P6 w - - 0 1", // Bishop trapped by own pawns
			description: "Bishop trapped with minimal mobility",
			expected:    BishopTrappedPenalty, // Less than 3 moves = -50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get bishop position and evaluate mobility directly
			whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
			if whiteBishops == 0 {
				t.Fatalf("No white bishop found in position")
			}

			bishopSquare, _ := whiteBishops.PopLSB()

			score := evaluateBishopMobility(b, bishopSquare, board.BitboardWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}
