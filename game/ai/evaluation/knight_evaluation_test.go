package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestKnightOutpostDetection(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "white_knight_outpost_d5",
			fen:         "8/8/8/3N4/2P5/8/8/8 w - - 0 1",
			expected:    40, // KnightOutpostBase (30) + central bonus (10)
			description: "White knight on d5 supported by c4 pawn",
		},
		{
			name:        "white_knight_advanced_outpost_e6",
			fen:         "8/8/4N3/3P4/8/8/8/8 w - - 0 1",
			expected:    60, // KnightOutpostAdvanced (50) + central bonus (10)
			description: "White knight on e6 supported by d5 pawn",
		},
		{
			name:        "white_knight_not_outpost_attackable",
			fen:         "8/4p3/8/3N4/2P5/8/8/8 w - - 0 1",
			expected:    0, // Can be attacked by f7 pawn
			description: "White knight on d5 can be attacked by f7 pawn",
		},
		{
			name:        "white_knight_not_outpost_unsupported",
			fen:         "8/8/8/3N4/8/8/8/8 w - - 0 1",
			expected:    0, // Not supported by pawn
			description: "White knight on d5 not supported by pawn",
		},
		{
			name:        "black_knight_outpost",
			fen:         "8/8/8/5p2/4n3/8/8/8 w - - 0 1",
			expected:    40, // KnightOutpostBase (30) + central bonus (10)
			description: "Black knight on e4 supported by f5 pawn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get knight position and evaluate outpost directly
			var knightSquare int
			var friendlyPawns, enemyPawns board.Bitboard
			var color board.BitboardColor

			whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
			blackKnights := b.GetPieceBitboard(board.BlackKnight)

			if whiteKnights != 0 {
				knightSquare, _ = whiteKnights.PopLSB()
				friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
				enemyPawns = b.GetPieceBitboard(board.BlackPawn)
				color = board.BitboardWhite
			} else if blackKnights != 0 {
				knightSquare, _ = blackKnights.PopLSB()
				friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
				enemyPawns = b.GetPieceBitboard(board.WhitePawn)
				color = board.BitboardBlack
			}

			score := evaluateKnightOutpost(knightSquare, friendlyPawns, enemyPawns, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestKnightMobility(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "trapped_knight_corner",
			fen:         "8/8/8/8/8/8/PPP5/N7 w - - 0 1",
			description: "Knight trapped in corner by own pawns",
			expected:    KnightTrappedPenalty, // -50
		},
		{
			name:        "knight_good_mobility",
			fen:         "8/8/8/3N4/8/8/8/8 w - - 0 1",
			description: "Knight in center with full mobility",
			expected:    8*KnightMobilityUnit + 4*KnightMobilityCenter, // 8*4 + 4*2 = 40
		},
		{
			name:        "knight_minimal_mobility",
			fen:         "8/2P1P3/1P3P2/3N4/1P3P2/2P1P3/8/8 w - - 0 1",
			description: "Knight with minimal mobility surrounded by pawns",
			expected:    KnightTrappedPenalty, // Less than 3 moves
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get knight position and evaluate mobility directly
			var knightSquare int
			var color board.BitboardColor

			whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
			blackKnights := b.GetPieceBitboard(board.BlackKnight)

			if whiteKnights != 0 {
				knightSquare, _ = whiteKnights.PopLSB()
				color = board.BitboardWhite
			} else if blackKnights != 0 {
				knightSquare, _ = blackKnights.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateKnightMobility(b, knightSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestKnightForks(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
		expected    int
	}{
		{
			name:        "knight_forking_rook_bishop",
			fen:         "8/8/2r1b3/8/3N4/8/8/8 w - - 0 1",
			description: "White knight forking black rook and bishop",
			expected:    KnightForkActive, // 25
		},
		{
			name:        "knight_royal_fork",
			fen:         "8/8/2k1q3/8/3N4/8/8/8 w - - 0 1",
			description: "White knight forking black king and queen",
			expected:    KnightRoyalFork, // 50
		},
		{
			name:        "knight_potential_fork",
			fen:         "8/8/2r1b3/8/8/5N2/8/8 w - - 0 1",
			description: "White knight can fork with one move",
			expected:    KnightForkThreat, // 15
		},
		{
			name:        "no_fork_available",
			fen:         "8/8/8/8/3N4/8/8/8 w - - 0 1",
			description: "No fork opportunities",
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Get knight position and evaluate forks directly
			var knightSquare int
			var color board.BitboardColor

			whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
			blackKnights := b.GetPieceBitboard(board.BlackKnight)

			if whiteKnights != 0 {
				knightSquare, _ = whiteKnights.PopLSB()
				color = board.BitboardWhite
			} else if blackKnights != 0 {
				knightSquare, _ = blackKnights.PopLSB()
				color = board.BitboardBlack
			}

			score := evaluateKnightForks(b, knightSquare, color)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}
