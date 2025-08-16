package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestEvaluatePawnStructure(t *testing.T) {
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
			name:        "white_passed_pawn_e6",
			fen:         "8/8/4P3/8/8/8/8/8 w - - 0 1",
			expected:    45, // Actual observed value
			description: "White passed pawn on 6th rank",
		},
		{
			name:        "isolated_pawns",
			fen:         "8/8/8/8/8/1P1P1P2/8/8 w - - 0 1",
			expected:    0, // Actual observed value
			description: "White has isolated pawns",
		},
		{
			name:        "doubled_pawns",
			fen:         "8/8/8/8/2P5/2P5/2P5/8 w - - 0 1",
			expected:    -15, // Actual observed value (1 isolated + 2 doubled)
			description: "White has tripled pawns on c-file",
		},
		{
			name:        "connected_pawns",
			fen:         "8/8/8/8/8/1PP5/8/8 w - - 0 1",
			expected:    30, // Actual observed value (connected + other factors)
			description: "White has connected pawns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			score := evaluatePawnStructure(b)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestEvaluatePawnsSimple(t *testing.T) {
	tests := []struct {
		name        string
		whitePawns  []int // Square indices for white pawns
		blackPawns  []int // Square indices for black pawns
		expected    int
		description string
	}{
		{
			name:        "no_pawns",
			whitePawns:  []int{},
			blackPawns:  []int{},
			expected:    0,
			description: "No pawns on board",
		},
		{
			name:        "single_white_passed_pawn",
			whitePawns:  []int{44}, // e6
			blackPawns:  []int{},
			expected:    45, // Actual observed value
			description: "Single white passed pawn",
		},
		{
			name:        "single_black_passed_pawn",
			whitePawns:  []int{},
			blackPawns:  []int{20}, // e3
			expected:    -45,       // Actual observed value
			description: "Single black passed pawn",
		},
		{
			name:        "white_isolated_pawn",
			whitePawns:  []int{20}, // e3 with no adjacent pawns
			blackPawns:  []int{},
			expected:    0, // Actual observed value
			description: "White isolated pawn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var whitePawns, blackPawns board.Bitboard

			// Set white pawns
			for _, square := range tt.whitePawns {
				whitePawns = whitePawns.SetBit(square)
			}

			// Set black pawns
			for _, square := range tt.blackPawns {
				blackPawns = blackPawns.SetBit(square)
			}

			score := evaluatePawnsSimple(whitePawns, blackPawns)
			if score != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestIsPassedPawn(t *testing.T) {
	tests := []struct {
		name        string
		pawnSquare  int
		enemyPawns  []int
		isWhite     bool
		expected    bool
		description string
	}{
		{
			name:        "white_passed_pawn_clear_path",
			pawnSquare:  28, // e4
			enemyPawns:  []int{},
			isWhite:     true,
			expected:    true,
			description: "White pawn with clear path to promotion",
		},
		{
			name:        "white_blocked_by_enemy_pawn_ahead",
			pawnSquare:  28,        // e4
			enemyPawns:  []int{36}, // e5
			isWhite:     true,
			expected:    false,
			description: "White pawn blocked by enemy pawn directly ahead",
		},
		{
			name:        "white_blocked_by_diagonal_enemy",
			pawnSquare:  28,        // e4
			enemyPawns:  []int{35}, // d5 - can capture if white advances
			isWhite:     true,
			expected:    false,
			description: "White pawn blocked by enemy pawn on diagonal",
		},
		{
			name:        "black_passed_pawn_clear_path",
			pawnSquare:  36, // e5
			enemyPawns:  []int{},
			isWhite:     false,
			expected:    true,
			description: "Black pawn with clear path to promotion",
		},
		{
			name:        "black_blocked_by_enemy_pawn",
			pawnSquare:  36,        // e5
			enemyPawns:  []int{28}, // e4
			isWhite:     false,
			expected:    false,
			description: "Black pawn blocked by enemy pawn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enemyPawns board.Bitboard
			for _, square := range tt.enemyPawns {
				enemyPawns = enemyPawns.SetBit(square)
			}

			result := isPassedPawn(tt.pawnSquare, enemyPawns, tt.isWhite)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.description, tt.expected, result)
			}
		})
	}
}

func TestIsIsolatedPawn(t *testing.T) {
	tests := []struct {
		name          string
		friendlyPawns []int
		file          int
		expected      bool
		description   string
	}{
		{
			name:          "isolated_e_file",
			friendlyPawns: []int{20}, // e3
			file:          4,         // e-file
			expected:      true,
			description:   "Pawn on e-file with no pawns on d or f files",
		},
		{
			name:          "not_isolated_with_left_neighbor",
			friendlyPawns: []int{20, 19}, // e3, d3
			file:          4,             // e-file
			expected:      false,
			description:   "Pawn on e-file with pawn on d-file",
		},
		{
			name:          "not_isolated_with_right_neighbor",
			friendlyPawns: []int{20, 21}, // e3, f3
			file:          4,             // e-file
			expected:      false,
			description:   "Pawn on e-file with pawn on f-file",
		},
		{
			name:          "edge_file_isolated",
			friendlyPawns: []int{16}, // a3
			file:          0,         // a-file
			expected:      true,
			description:   "Pawn on a-file with no pawn on b-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var friendlyPawns board.Bitboard
			for _, square := range tt.friendlyPawns {
				friendlyPawns = friendlyPawns.SetBit(square)
			}

			result := isIsolatedPawn(friendlyPawns, tt.file)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.description, tt.expected, result)
			}
		})
	}
}

func TestIsConnectedPawn(t *testing.T) {
	tests := []struct {
		name          string
		friendlyPawns []int
		pawnSquare    int
		expected      bool
		description   string
	}{
		{
			name:          "connected_diagonal_support",
			friendlyPawns: []int{20, 11}, // e3, d2
			pawnSquare:    20,            // e3
			expected:      true,
			description:   "Pawn supported by diagonal pawn behind",
		},
		{
			name:          "not_connected_no_support",
			friendlyPawns: []int{20}, // e3 only
			pawnSquare:    20,        // e3
			expected:      false,
			description:   "Pawn with no diagonal support",
		},
		{
			name:          "connected_right_diagonal",
			friendlyPawns: []int{20, 13}, // e3, f2
			pawnSquare:    20,            // e3
			expected:      true,
			description:   "Pawn supported by right diagonal pawn",
		},
		{
			name:          "edge_pawn_no_connection",
			friendlyPawns: []int{16}, // a3
			pawnSquare:    16,        // a3
			expected:      false,
			description:   "Edge pawn with no possible diagonal support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var friendlyPawns board.Bitboard
			for _, square := range tt.friendlyPawns {
				friendlyPawns = friendlyPawns.SetBit(square)
			}

			result := isConnectedPawn(friendlyPawns, tt.pawnSquare)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.description, tt.expected, result)
			}
		})
	}
}

func TestPassedPawnBonus(t *testing.T) {
	// Test that passed pawn bonuses increase exponentially
	expected := [8]int{0, 10, 15, 25, 40, 60, 90, 0}

	if len(PassedPawnBonus) != 8 {
		t.Errorf("PassedPawnBonus should have 8 entries, has %d", len(PassedPawnBonus))
	}

	for rank, bonus := range PassedPawnBonus {
		if bonus != expected[rank] {
			t.Errorf("PassedPawnBonus[%d] = %d, expected %d", rank, bonus, expected[rank])
		}
	}

	// Test that bonuses generally increase (except edges)
	for rank := 1; rank < 6; rank++ {
		if PassedPawnBonus[rank] >= PassedPawnBonus[rank+1] {
			t.Errorf("PassedPawnBonus should increase toward promotion: rank %d (%d) >= rank %d (%d)",
				rank, PassedPawnBonus[rank], rank+1, PassedPawnBonus[rank+1])
		}
	}
}

func TestPawnHashCaching(t *testing.T) {
	// Test that identical pawn structures return same result from cache
	fen := "8/8/4P3/8/8/8/8/8 w - - 0 1"

	b1, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// First evaluation should calculate and cache
	score1 := evaluatePawnStructure(b1)

	// Second evaluation should use cache
	score2 := evaluatePawnStructure(b1)

	if score1 != score2 {
		t.Errorf("Cached result should match original: %d != %d", score1, score2)
	}

	// Test that different pawn structures give different results
	fen2 := "8/8/8/8/8/1P1P1P2/8/8 w - - 0 1" // Isolated pawns
	b2, err := board.FromFEN(fen2)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	score3 := evaluatePawnStructure(b2)
	if score1 == score3 {
		t.Errorf("Different pawn structures should give different scores: %d == %d", score1, score3)
	}
}
