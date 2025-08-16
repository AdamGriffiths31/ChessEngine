package board

import (
	"testing"
)

func TestTablesInitialization(t *testing.T) {
	// Attack tables are automatically initialized at package load time
	// Test that they have non-zero values
	if FileMasks[0] == 0 {
		t.Error("FileMasks should be initialized")
	}
	if KnightAttacks[0] == 0 {
		t.Error("KnightAttacks should be initialized")
	}
}

func TestKnightAttackPatterns(t *testing.T) {

	testCases := []struct {
		square   int
		expected []int // Expected attack squares
	}{
		{
			square:   E4, // e4
			expected: []int{D2, F2, C3, G3, C5, G5, D6, F6},
		},
		{
			square:   A1, // a1 - corner case
			expected: []int{B3, C2},
		},
		{
			square:   H8, // h8 - corner case
			expected: []int{F7, G6},
		},
		{
			square:   D1, // d1 - edge case
			expected: []int{B2, F2, C3, E3},
		},
	}

	for _, tc := range testCases {
		attacks := GetKnightAttacks(tc.square)

		// Check that expected squares are attacked
		for _, expectedSquare := range tc.expected {
			if !attacks.HasBit(expectedSquare) {
				t.Errorf("Knight on %s should attack %s", SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}

		// Check that the number of attacks is correct
		if attacks.PopCount() != len(tc.expected) {
			t.Errorf("Knight on %s should have %d attacks, got %d", SquareToString(tc.square), len(tc.expected), attacks.PopCount())
		}
	}
}

func TestKingAttackPatterns(t *testing.T) {

	testCases := []struct {
		square   int
		expected []int // Expected attack squares
	}{
		{
			square:   E4, // e4
			expected: []int{D3, E3, F3, D4, F4, D5, E5, F5},
		},
		{
			square:   A1, // a1 - corner case
			expected: []int{A2, B1, B2},
		},
		{
			square:   H8, // h8 - corner case
			expected: []int{G7, G8, H7},
		},
		{
			square:   E1, // e1 - edge case
			expected: []int{D1, F1, D2, E2, F2},
		},
	}

	for _, tc := range testCases {
		attacks := GetKingAttacks(tc.square)

		// Check that expected squares are attacked
		for _, expectedSquare := range tc.expected {
			if !attacks.HasBit(expectedSquare) {
				t.Errorf("King on %s should attack %s", SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}

		// Check that the number of attacks is correct
		if attacks.PopCount() != len(tc.expected) {
			t.Errorf("King on %s should have %d attacks, got %d", SquareToString(tc.square), len(tc.expected), attacks.PopCount())
		}
	}
}

func TestPawnAttackPatterns(t *testing.T) {

	testCases := []struct {
		square   int
		color    BitboardColor
		expected []int
	}{
		{
			square:   E4, // White pawn on e4
			color:    BitboardWhite,
			expected: []int{D5, F5},
		},
		{
			square:   E4, // Black pawn on e4
			color:    BitboardBlack,
			expected: []int{D3, F3},
		},
		{
			square:   A2, // White pawn on a2 (edge case)
			color:    BitboardWhite,
			expected: []int{B3},
		},
		{
			square:   H7, // Black pawn on h7 (edge case)
			color:    BitboardBlack,
			expected: []int{G6},
		},
		{
			square:   E8, // White pawn on 8th rank (impossible but test edge case)
			color:    BitboardWhite,
			expected: []int{}, // No attacks from 8th rank
		},
		{
			square:   E1, // Black pawn on 1st rank (impossible but test edge case)
			color:    BitboardBlack,
			expected: []int{}, // No attacks from 1st rank
		},
	}

	for _, tc := range testCases {
		attacks := GetPawnAttacks(tc.square, tc.color)

		// Check that expected squares are attacked
		for _, expectedSquare := range tc.expected {
			if !attacks.HasBit(expectedSquare) {
				t.Errorf("%s pawn on %s should attack %s",
					colorString(tc.color), SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}

		// Check that the number of attacks is correct
		if attacks.PopCount() != len(tc.expected) {
			t.Errorf("%s pawn on %s should have %d attacks, got %d",
				colorString(tc.color), SquareToString(tc.square), len(tc.expected), attacks.PopCount())
		}
	}
}

func TestPawnPushPatterns(t *testing.T) {

	testCases := []struct {
		square   int
		color    BitboardColor
		expected []int
	}{
		{
			square:   E2, // White pawn on starting rank
			color:    BitboardWhite,
			expected: []int{E3},
		},
		{
			square:   E7, // Black pawn on starting rank
			color:    BitboardBlack,
			expected: []int{E6},
		},
		{
			square:   E4, // Pawn in middle
			color:    BitboardWhite,
			expected: []int{E5},
		},
		{
			square:   E8, // White pawn on 8th rank
			color:    BitboardWhite,
			expected: []int{}, // No pushes from 8th rank
		},
		{
			square:   E1, // Black pawn on 1st rank
			color:    BitboardBlack,
			expected: []int{}, // No pushes from 1st rank
		},
	}

	for _, tc := range testCases {
		pushes := GetPawnPushes(tc.square, tc.color)

		// Check that expected squares can be pushed to
		for _, expectedSquare := range tc.expected {
			if !pushes.HasBit(expectedSquare) {
				t.Errorf("%s pawn on %s should be able to push to %s",
					colorString(tc.color), SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}

		// Check that the number of pushes is correct
		if pushes.PopCount() != len(tc.expected) {
			t.Errorf("%s pawn on %s should have %d pushes, got %d",
				colorString(tc.color), SquareToString(tc.square), len(tc.expected), pushes.PopCount())
		}
	}
}

func TestPawnDoublePushPatterns(t *testing.T) {

	testCases := []struct {
		square   int
		color    BitboardColor
		expected []int
	}{
		{
			square:   E2, // White pawn on starting rank
			color:    BitboardWhite,
			expected: []int{E4},
		},
		{
			square:   E7, // Black pawn on starting rank
			color:    BitboardBlack,
			expected: []int{E5},
		},
		{
			square:   E3, // White pawn not on starting rank
			color:    BitboardWhite,
			expected: []int{}, // No double push
		},
		{
			square:   E6, // Black pawn not on starting rank
			color:    BitboardBlack,
			expected: []int{}, // No double push
		},
	}

	for _, tc := range testCases {
		doublePushes := GetPawnDoublePushes(tc.square, tc.color)

		// Check that expected squares can be double-pushed to
		for _, expectedSquare := range tc.expected {
			if !doublePushes.HasBit(expectedSquare) {
				t.Errorf("%s pawn on %s should be able to double push to %s",
					colorString(tc.color), SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}

		// Check that the number of double pushes is correct
		if doublePushes.PopCount() != len(tc.expected) {
			t.Errorf("%s pawn on %s should have %d double pushes, got %d",
				colorString(tc.color), SquareToString(tc.square), len(tc.expected), doublePushes.PopCount())
		}
	}
}

func TestDistanceCalculation(t *testing.T) {

	testCases := []struct {
		sq1      int
		sq2      int
		expected int
	}{
		{A1, A1, 0},  // Same square
		{A1, B1, 1},  // Adjacent horizontally
		{A1, A2, 1},  // Adjacent vertically
		{A1, B2, 2},  // Diagonal adjacent
		{A1, H8, 14}, // Opposite corners
		{E4, E6, 2},  // Same file, 2 ranks apart
		{C3, F3, 3},  // Same rank, 3 files apart
	}

	for _, tc := range testCases {
		distance := GetDistance(tc.sq1, tc.sq2)
		if distance != tc.expected {
			t.Errorf("Distance from %s to %s should be %d, got %d",
				SquareToString(tc.sq1), SquareToString(tc.sq2), tc.expected, distance)
		}

		// Distance should be symmetric
		reverseDistance := GetDistance(tc.sq2, tc.sq1)
		if reverseDistance != tc.expected {
			t.Errorf("Distance should be symmetric: %s to %s = %d, but %s to %s = %d",
				SquareToString(tc.sq1), SquareToString(tc.sq2), distance,
				SquareToString(tc.sq2), SquareToString(tc.sq1), reverseDistance)
		}
	}
}

func TestBetweenSquares(t *testing.T) {

	testCases := []struct {
		sq1      int
		sq2      int
		expected []int
	}{
		{A1, A5, []int{A2, A3, A4}}, // Same file
		{A1, E1, []int{B1, C1, D1}}, // Same rank
		{A1, D4, []int{B2, C3}},     // Diagonal
		{H8, E5, []int{G7, F6}},     // Diagonal
		{A1, B3, []int{}},           // Not on same line
		{A1, A1, []int{}},           // Same square
		{A1, A2, []int{}},           // Adjacent squares
	}

	for _, tc := range testCases {
		between := GetBetween(tc.sq1, tc.sq2)

		// Check that expected squares are between
		for _, expectedSquare := range tc.expected {
			if !between.HasBit(expectedSquare) {
				t.Errorf("Square %s should be between %s and %s",
					SquareToString(expectedSquare), SquareToString(tc.sq1), SquareToString(tc.sq2))
			}
		}

		// Check that the number of between squares is correct
		if between.PopCount() != len(tc.expected) {
			t.Errorf("Should have %d squares between %s and %s, got %d",
				len(tc.expected), SquareToString(tc.sq1), SquareToString(tc.sq2), between.PopCount())
		}
	}
}

func TestLineSquares(t *testing.T) {

	testCases := []struct {
		sq1      int
		sq2      int
		expected []int
	}{
		{A1, A3, []int{A1, A2, A3}}, // Same file
		{A1, C1, []int{A1, B1, C1}}, // Same rank
		{A1, C3, []int{A1, B2, C3}}, // Diagonal
		{A1, B3, []int{}},           // Not on same line
	}

	for _, tc := range testCases {
		line := GetLine(tc.sq1, tc.sq2)

		// Check that expected squares are on the line
		for _, expectedSquare := range tc.expected {
			if !line.HasBit(expectedSquare) {
				t.Errorf("Square %s should be on line between %s and %s",
					SquareToString(expectedSquare), SquareToString(tc.sq1), SquareToString(tc.sq2))
			}
		}

		// Check that the number of line squares is correct
		if line.PopCount() != len(tc.expected) {
			t.Errorf("Should have %d squares on line between %s and %s, got %d",
				len(tc.expected), SquareToString(tc.sq1), SquareToString(tc.sq2), line.PopCount())
		}
	}
}

func TestDiagonalMasks(t *testing.T) {

	testCases := []struct {
		square   int
		diagonal []int // Squares on the same main diagonal
	}{
		{
			square:   A1,
			diagonal: []int{A1, B2, C3, D4, E5, F6, G7, H8},
		},
		{
			square:   E4,
			diagonal: []int{B1, C2, D3, E4, F5, G6, H7},
		},
		{
			square:   H1,
			diagonal: []int{H1},
		},
	}

	for _, tc := range testCases {
		diagonal := GetDiagonalMask(tc.square)

		// Check that expected squares are on the diagonal
		for _, expectedSquare := range tc.diagonal {
			if !diagonal.HasBit(expectedSquare) {
				t.Errorf("Square %s should be on diagonal with %s",
					SquareToString(expectedSquare), SquareToString(tc.square))
			}
		}

		// Check that the number of diagonal squares is correct
		if diagonal.PopCount() != len(tc.diagonal) {
			t.Errorf("Should have %d squares on diagonal with %s, got %d",
				len(tc.diagonal), SquareToString(tc.square), diagonal.PopCount())
		}
	}
}

func TestAntiDiagonalMasks(t *testing.T) {

	testCases := []struct {
		square       int
		antiDiagonal []int // Squares on the same anti-diagonal
	}{
		{
			square:       A8,
			antiDiagonal: []int{A8, B7, C6, D5, E4, F3, G2, H1},
		},
		{
			square:       E4,
			antiDiagonal: []int{A8, B7, C6, D5, E4, F3, G2, H1},
		},
		{
			square:       A1,
			antiDiagonal: []int{A1},
		},
	}

	for _, tc := range testCases {
		antiDiagonal := GetAntiDiagonalMask(tc.square)

		// Check that expected squares are on the anti-diagonal
		for _, expectedSquare := range tc.antiDiagonal {
			if !antiDiagonal.HasBit(expectedSquare) {
				t.Errorf("Square %s should be on anti-diagonal with %s",
					SquareToString(expectedSquare), SquareToString(tc.square))
			}
		}

		// Check that the number of anti-diagonal squares is correct
		if antiDiagonal.PopCount() != len(tc.antiDiagonal) {
			t.Errorf("Should have %d squares on anti-diagonal with %s, got %d",
				len(tc.antiDiagonal), SquareToString(tc.square), antiDiagonal.PopCount())
		}
	}
}

func TestInvalidInputs(t *testing.T) {

	// Test invalid square indices
	invalidSquares := []int{-1, 64, 100}

	for _, square := range invalidSquares {
		if GetKnightAttacks(square) != 0 {
			t.Errorf("GetKnightAttacks(%d) should return 0 for invalid square", square)
		}
		if GetKingAttacks(square) != 0 {
			t.Errorf("GetKingAttacks(%d) should return 0 for invalid square", square)
		}
		if GetPawnAttacks(square, BitboardWhite) != 0 {
			t.Errorf("GetPawnAttacks(%d, White) should return 0 for invalid square", square)
		}
		if GetDistance(square, E4) != -1 {
			t.Errorf("GetDistance(%d, E4) should return -1 for invalid square", square)
		}
		if GetBetween(square, E4) != 0 {
			t.Errorf("GetBetween(%d, E4) should return 0 for invalid square", square)
		}
	}
}

// Helper function for color string representation
func colorString(color BitboardColor) string {
	if color == BitboardWhite {
		return "White"
	}
	return "Black"
}

// Benchmark tests
func BenchmarkKnightAttacks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetKnightAttacks(E4)
	}
}

func BenchmarkKingAttacks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetKingAttacks(E4)
	}
}

func BenchmarkPawnAttacks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetPawnAttacks(E4, BitboardWhite)
	}
}

func BenchmarkDistance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetDistance(A1, H8)
	}
}

func BenchmarkBetween(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBetween(A1, H8)
	}
}
