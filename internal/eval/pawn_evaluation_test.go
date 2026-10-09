package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestIsPassedPawn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		pawnSquare int
		enemyPawns []int
		isWhite    bool
		expected   bool
	}{
		{
			name:       "white_passed_pawn_clear_path",
			pawnSquare: 28, // e4
			enemyPawns: []int{},
			isWhite:    true,
			expected:   true,
		},
		{
			name:       "white_blocked_by_enemy_pawn_ahead",
			pawnSquare: 28,        // e4
			enemyPawns: []int{36}, // e5
			isWhite:    true,
			expected:   false,
		},
		{
			name:       "white_blocked_by_diagonal_enemy",
			pawnSquare: 28,        // e4
			enemyPawns: []int{35}, // d5 - can capture if white advances
			isWhite:    true,
			expected:   false,
		},
		{
			name:       "black_passed_pawn_clear_path",
			pawnSquare: 36, // e5
			enemyPawns: []int{},
			isWhite:    false,
			expected:   true,
		},
		{
			name:       "black_blocked_by_enemy_pawn",
			pawnSquare: 36,        // e5
			enemyPawns: []int{28}, // e4
			isWhite:    false,
			expected:   false,
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
				t.Errorf("%s: expected %t, got %t", tt.name, tt.expected, result)
			}
		})
	}
}

func TestIsIsolatedPawn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		friendlyPawns []int
		file          int
		expected      bool
	}{
		{
			name:          "isolated_e_file",
			friendlyPawns: []int{20}, // e3
			file:          4,         // e-file
			expected:      true,
		},
		{
			name:          "not_isolated_with_left_neighbor",
			friendlyPawns: []int{20, 19}, // e3, d3
			file:          4,             // e-file
			expected:      false,
		},
		{
			name:          "not_isolated_with_right_neighbor",
			friendlyPawns: []int{20, 21}, // e3, f3
			file:          4,             // e-file
			expected:      false,
		},
		{
			name:          "edge_file_isolated",
			friendlyPawns: []int{16}, // a3
			file:          0,         // a-file
			expected:      true,
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
				t.Errorf("%s: expected %t, got %t", tt.name, tt.expected, result)
			}
		})
	}
}

func TestIsConnectedPawn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		friendlyPawns []int
		pawnSquare    int
		isWhite       bool
		expected      bool
	}{
		{
			name:          "connected_diagonal_support",
			friendlyPawns: []int{20, 11}, // e3, d2
			pawnSquare:    20,            // e3
			isWhite:       true,
			expected:      true,
		},
		{
			name:          "not_connected_no_support",
			friendlyPawns: []int{20}, // e3 only
			pawnSquare:    20,        // e3
			isWhite:       true,
			expected:      false,
		},
		{
			name:          "connected_right_diagonal",
			friendlyPawns: []int{20, 13}, // e3, f2
			pawnSquare:    20,            // e3
			isWhite:       true,
			expected:      true,
		},
		{
			name:          "edge_pawn_no_connection",
			friendlyPawns: []int{16}, // a3
			pawnSquare:    16,        // a3
			isWhite:       true,
			expected:      false,
		},
		{
			name:          "not_connected_forward_left_diagonal",
			friendlyPawns: []int{20, 27}, // e3, d4 (pawn in front diagonally)
			pawnSquare:    20,            // e3
			isWhite:       true,
			expected:      false,
		},
		{
			name:          "not_connected_forward_right_diagonal",
			friendlyPawns: []int{20, 29}, // e3, f4 (right diagonal in front)
			pawnSquare:    20,            // e3
			isWhite:       true,
			expected:      false,
		},
		{
			name:          "black_connected_support_from_above",
			friendlyPawns: []int{44, 53}, // e6, f7 (black: f7 defends e6)
			pawnSquare:    44,            // e6
			isWhite:       false,
			expected:      true,
		},
		{
			name:          "black_not_connected_pawn_in_front",
			friendlyPawns: []int{44, 35}, // e6, d5 (d5 is in front of a black pawn)
			pawnSquare:    44,            // e6
			isWhite:       false,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var friendlyPawns board.Bitboard
			for _, square := range tt.friendlyPawns {
				friendlyPawns = friendlyPawns.SetBit(square)
			}

			result := isConnectedPawn(friendlyPawns, tt.pawnSquare, tt.isWhite)
			if result != tt.expected {
				t.Errorf("%s: expected %t, got %t", tt.name, tt.expected, result)
			}
		})
	}
}

// The passed-pawn bonus grows with every rank advanced.
func TestPassedPawnBonusGrowsTowardPromotion(t *testing.T) {
	t.Parallel()
	for rank := 1; rank < 6; rank++ {
		if PassedPawnBonus[rank] >= PassedPawnBonus[rank+1] {
			t.Errorf("PassedPawnBonus[%d] = %d, not less than rank %d (%d)",
				rank, PassedPawnBonus[rank], rank+1, PassedPawnBonus[rank+1])
		}
	}
}

func TestPawnHashCaching(t *testing.T) {
	// Not parallel: exercises the package-level PawnHashTable cache
	// directly; see TestEvaluatePawnStructure for why this can't run
	// concurrently with other tests that touch the same cache.
	fen := "8/8/4P3/8/8/8/8/8 w - - 0 1"

	b1, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	score1 := evaluatePawnStructure(b1)
	score2 := evaluatePawnStructure(b1)

	if score1 != score2 {
		t.Errorf("Cached result should match original: %d != %d", score1, score2)
	}

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

// Structure scores are compared, not pinned, so retuning the weights does not
// break them: only the direction of each rule matters.
func TestPawnStructureRules(t *testing.T) {
	t.Parallel()
	score := func(white ...int) int {
		return evaluatePawnsSimple(bitboardOf(white...), 0)
	}
	const (
		b3, c3, d3, f3 = 17, 18, 19, 21
		c2, c4         = 10, 26
		e6, e3         = 44, 20
	)

	if got, want := score(e6), score(e3); got <= want {
		t.Errorf("passed pawn on e6 scored %d, not above the same pawn on e3 (%d)", got, want)
	}
	if got, want := score(b3, c3), score(b3, d3); got <= want {
		t.Errorf("connected pawns scored %d, not above unconnected ones (%d)", got, want)
	}
	if got, want := score(c2, c3, c4), score(b3, d3, f3); got >= want {
		t.Errorf("tripled pawns scored %d, not below the same count spread out (%d)", got, want)
	}
}

func bitboardOf(squares ...int) board.Bitboard {
	var bb board.Bitboard
	for _, sq := range squares {
		bb = bb.SetBit(sq)
	}
	return bb
}
