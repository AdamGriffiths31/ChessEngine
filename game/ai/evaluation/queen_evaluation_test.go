package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestEarlyQueenDevelopment(t *testing.T) {
	tests := []struct {
		fen         string
		expected    int
		description string
	}{
		{
			fen:         "rnbqkbnr/pppppppp/8/8/8/3Q4/PPPPPPPP/RNB1KBNR w KQkq - 0 1",
			description: "White queen developed before minor pieces",
			expected:    EarlyQueenDevelopmentPenalty, // Penalty for early queen development
		},
		{
			fen:         "r1bqkb1r/pppppppp/2n2n2/8/8/B1NQPN2/PPPP1PPP/R3KB1R w KQkq - 0 1",
			description: "Queen developed after minor pieces",
		},
		{
			fen:         "rnb1kbnr/pppppppp/3q4/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1",
			description: "Black queen developed early",
			expected:    EarlyQueenDevelopmentPenalty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateEarlyDevelopment(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenBatteries(t *testing.T) {
	tests := []struct {
		fen         string
		description string
		expected    int
	}{
		{
			fen:         "8/8/8/8/Q6R/8/8/8 w - - 0 1",
			description: "White queen-rook battery on rank",
			expected:    QueenRookBatteryBonus,
		},
		{
			fen:         "8/8/8/8/3Q4/8/8/B7 w - - 0 1",
			description: "White queen-bishop battery on diagonal",
			expected:    QueenBishopBatteryBonus,
		},
		{
			fen:         "8/8/8/8/Q2p3R/8/8/8 w - - 0 1",
			description: "Battery blocked by piece",
		},
		{
			fen:         "4k3/8/8/8/4Q3/4R3/8/8 w - - 0 1",
			description: "Battery pointing at enemy king",
			expected:    QueenRookBatteryBonus + 10, // Bonus for attacking king
		},
		{
			fen:         "8/7k/8/8/4Q3/R2B4/8/8 w - - 0 1",
			description: "Battery pointing at enemy king",
			expected:    QueenBishopBatteryBonus + 10, // Bonus for attacking king
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateQueenBatteries(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenCentralization(t *testing.T) {
	tests := []struct {
		fen         string
		description string
		expected    int
	}{
		{
			fen:         "8/8/8/8/3Q4/8/8/8 w - - 0 1",
			description: "Queen on central square d4",
			expected:    QueenCentralizationBonus, // Bonus for central queen
		},
		{
			fen:         "8/8/8/2Q5/8/8/8/8 w - - 0 1",
			description: "Queen in extended center c5",
			expected:    QueenExtendedCenterBonus, // Bonus for extended center
		},
		{
			fen:         "8/8/8/8/8/8/8/Q7 w - - 0 1",
			description: "Queen on edge a1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
			}

			score := evaluateQueenCentralization(queenSquare)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenAttackers(t *testing.T) {
	tests := []struct {
		fen         string
		description string
		expected    int
	}{
		{
			fen:         "8/8/8/3ppp2/3pQp2/3ppp2/8/8 w - - 0 1",
			description: "Queen attacking multiple pawns",
			expected:    QueenMultipleAttacks,
		},
		{
			fen:         "4k3/4p3/8/8/4Q3/8/8/8 w - - 0 1",
			description: "Queen attacking enemy king",
			expected:    QueenAttackingKingZone,
		},
		{
			fen:         "4k3/4p3/8/8/3Q4/8/8/8 w - - 0 1",
			description: "Queen attacking enemy king zone",
			expected:    QueenAttackingKingZone,
		},
		{
			fen:         "4k3/4p3/8/8/8/6Q1/8/8 w - - 0 1",
			description: "Queen not attacking any pieces",
			expected:    0, // No attacks
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateQueenAttacks(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenSafety(t *testing.T) {
	tests := []struct {
		fen         string
		expected    int
		description string
	}{
		{
			fen:         "8/8/8/3q4/4P3/8/8/8 w - - 0 1",
			description: "Queen attacked by pawn",
			expected:    QueenAttackedByPawnPenalty,
		},
		{
			fen:         "8/8/8/3q4/8/4N3/8/8 w - - 0 1",
			description: "Queen attacked by minor piece",
			expected:    QueenAttackedByMinorPenalty,
		},
		{
			fen:         "8/8/1PPPP3/1N1Qr3/1PPPP3/8/8/8 w - - 0 1",
			description: "Queen trapped by pawns",
			expected:    QueenTrappedPenalty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateQueenSafety(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenMobility(t *testing.T) {
	tests := []struct {
		fen         string
		expected    int
		description string
	}{
		{
			fen:         "8/8/8/3Q4/8/8/8/8 w - - 0 1",
			description: "Queen has maximum mobility",
			expected:    76,
		},
		{
			fen:         "8/8/2PPP3/2PQP3/8/8/8/8 w - - 0 1",
			description: "Queen has restricted mobility",
			expected:    32,
		},
		{
			fen:         "8/8/2PPP3/2PQP3/8/2p1p3/2p1p3/2p1p3 w - - 0 1",
			description: "Queen has restricted mobility by pawns",
			expected:    30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateQueenMobility(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenPins(t *testing.T) {
	tests := []struct {
		fen         string
		expected    int
		description string
	}{
		{
			fen:         "1k6/8/1n6/8/8/1Q6/8/8 w - - 0 1",
			description: "Queen pins knight to king",
			expected:    30,
		},
		{
			fen:         "1r5k/8/1n6/8/8/1Q6/8/8 w - - 0 1",
			description: "Queen pins knight to rook",
			expected:    20,
		},
		{
			fen:         "1r5k/5r2/1n6/3n4/8/1Q2n2q/1n6/1q6 w - - 0 1",
			description: "Queen pins multiple pieces",
			expected:    55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get queen position and evaluate open files directly
			whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
			blackQueens := b.GetPieceBitboard(board.BlackQueen)

			var queenSquare int
			var color board.BitboardColor

			if whiteQueens != 0 {
				queenSquare, _ = whiteQueens.PopLSB()
				color = board.BitboardWhite
			} else if blackQueens != 0 {
				queenSquare, _ = blackQueens.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateQueenPins(b, queenSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}
