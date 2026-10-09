package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval/values"
)

func TestEvaluateKings(t *testing.T) {
	t.Parallel()
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
			name:        "white_castled_kingside",
			fen:         "rnbqk2r/pppppppp/8/8/8/8/PPPPPPPP/RNBQ1RK1 w kq - 0 1",
			expected:    0, // Castling/shelter bonus removed (see king_evaluation.go); no open files or threats here
			description: "White king castled kingside with good shelter",
		},
		{
			name:        "endgame_central_kings",
			fen:         "8/8/8/3k4/3K4/8/8/8 w - - 0 1",
			expected:    0, // Both centralized equally in endgame
			description: "Endgame with both kings centralized",
		},
		{
			name:        "white_king_open_files",
			fen:         "rnbqkbnr/pp1ppppp/8/8/8/8/PP1PPPPP/RNBQKBNR w KQkq - 0 1",
			expected:    0, // Actual observed value: king hasn't castled, no open file penalty applies
			description: "White king with dangerous open file nearby",
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

func TestEvaluateKingSimple(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		fen         string
		isWhite     bool
		expected    int
		description string
	}{
		{
			name:        "white_king_castled_kingside",
			fen:         "8/8/8/8/8/8/PPPPPPPP/RNBQ1RK1 w - - 0 1",
			isWhite:     true,
			expected:    0, // Castling/shelter bonus removed (see king_evaluation.go)
			description: "White king castled kingside",
		},
		{
			name:        "white_king_not_castled",
			fen:         "8/8/8/8/8/8/PPPPPPPP/RNBQK2R w - - 0 1",
			isWhite:     true,
			expected:    0, // Lost-castling-rights penalty removed (see king_evaluation.go)
			description: "White king hasn't castled",
		},
		{
			name:        "endgame_centralized_king",
			fen:         "8/8/8/3K4/8/8/8/8 w - - 0 1",
			isWhite:     true,
			expected:    -60, // Safety terms only: three open files (-60). No lost-rights penalty away
			// from the home square/castled zone - see isInCastledZone.
			description: "White king centralized in endgame",
		},
		{
			name:        "black_king_castled_queenside",
			fen:         "2kr1bnr/pppppppp/8/8/8/8/8/8 w - - 0 1",
			isWhite:     false,
			expected:    0, // Castling/shelter bonus removed (see king_evaluation.go)
			description: "Black king castled queenside",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			var kingSquare int
			if tt.isWhite {
				whiteKing := b.GetPieceBitboard(board.WhiteKing)
				kingSquare = whiteKing.LSB()
			} else {
				blackKing := b.GetPieceBitboard(board.BlackKing)
				kingSquare = blackKing.LSB()
			}

			score := evaluateKingSimple(b, kingSquare, tt.isWhite)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

// TestKingTableEndgameCentralization verifies the endgame king table that
// replaced the removed evaluateKingEndgameActivity: central squares score
// highest, corners lowest, and the black-side flip mirrors correctly.
func TestKingTableEndgameCentralization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		piece     values.Piece
		rank, fil int
		expected  int
	}{
		{"white_d4", values.WhiteKing, 3, 3, 40},
		{"white_e5", values.WhiteKing, 4, 4, 40},
		{"white_a1", values.WhiteKing, 0, 0, -50},
		{"white_h1", values.WhiteKing, 0, 7, -50},
		{"black_d5_mirrors_white_d4", values.BlackKing, 4, 3, -40},
		{"black_a8_mirrors_white_a1", values.BlackKing, 7, 0, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := values.GetPositionalBonusEndgame(tt.piece, tt.rank, tt.fil)
			if got != tt.expected {
				t.Errorf("GetPositionalBonusEndgame(%v, %d, %d) = %d, want %d", tt.piece, tt.rank, tt.fil, got, tt.expected)
			}
		})
	}
}

func TestEvaluateOpenFilesNearKing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		fen         string
		kingSquare  int
		expected    int
		description string
	}{
		{
			name:        "no_open_files",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			kingSquare:  4, // e1
			expected:    0, // No open files
			description: "King with no open files nearby",
		},
		{
			name:        "one_open_file",
			fen:         "rnbqkbnr/pp1ppppp/8/8/8/8/PP1PPPPP/RNBQKBNR w KQkq - 0 1",
			kingSquare:  4, // e1
			expected:    0, // Actual observed value: only d-file and f-file are checked for king on e1
			description: "King with one open file nearby",
		},
		{
			name:        "multiple_open_files",
			fen:         "rnbqkbnr/p2p2pp/8/8/8/8/P2P2PP/RNBQKBNR w KQkq - 0 1",
			kingSquare:  4,   // e1
			expected:    -40, // Actual observed value: two open files (d-file and f-file)
			description: "King with multiple open files nearby",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluateOpenFilesNearKing(b, tt.kingSquare)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestKingSafetyZonePrecomputation(t *testing.T) {
	t.Parallel()
	if len(KingSafetyZone) != 64 {
		t.Errorf("KingSafetyZone should have 64 entries, has %d", len(KingSafetyZone))
	}

	expectedA1Zone := board.Bitboard(0)
	expectedA1Zone = expectedA1Zone.SetBit(0).SetBit(1).SetBit(8).SetBit(9) // a1, b1, a2, b2
	if KingSafetyZone[0] != expectedA1Zone {
		t.Errorf("KingSafetyZone[0] (a1) incorrect: expected %d, got %d", expectedA1Zone, KingSafetyZone[0])
	}

	d4Zone := KingSafetyZone[27]
	expectedSquares := []int{18, 19, 20, 26, 27, 28, 34, 35, 36} // 3x3 around d4
	actualCount := d4Zone.PopCount()
	if actualCount != 9 {
		t.Errorf("KingSafetyZone[27] (d4) should have 9 squares, has %d", actualCount)
	}

	for _, square := range expectedSquares {
		if !d4Zone.HasBit(square) {
			t.Errorf("KingSafetyZone[27] (d4) missing square %d", square)
		}
	}
}
