package board

import (
	"testing"
)

// Hand-checked attack sets: each slider stops at (and includes) the first
// blocker in every direction. This is the ground truth that the consistency
// test below compares the magic lookup against.
func TestSlidingAttacksWithBlockers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		attacks   func(square int, occupancy Bitboard) Bitboard
		occupancy []int
		want      []int
	}{
		{
			name:      "rook on e4, blockers e6 and c4",
			attacks:   GetRookAttacks,
			occupancy: []int{E6, C4},
			want:      []int{E5, E6, D4, C4, E3, E2, E1, F4, G4, H4},
		},
		{
			name:      "bishop on e4, blockers g6 and c2",
			attacks:   GetBishopAttacks,
			occupancy: []int{G6, C2},
			want:      []int{F5, G6, D3, C2, D5, C6, B7, A8, F3, G2, H1},
		},
		{
			name:      "queen on e4, blockers e6 and g6",
			attacks:   GetQueenAttacks,
			occupancy: []int{E6, G6},
			want: []int{
				E5, E6, E3, E2, E1, A4, B4, C4, D4, F4, G4, H4, // rook lines
				F5, G6, D5, C6, B7, A8, D3, C2, B1, F3, G2, H1, // bishop lines
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.attacks(E4, bitboardOf(tt.occupancy...)), bitboardOf(tt.want...); got != want {
				t.Errorf("got %v, want %v", got.BitList(), want.BitList())
			}
		})
	}
}

func TestSlidingPieceAttacksConsistency(t *testing.T) {
	t.Parallel()

	// Magic lookups must equal the slow on-the-fly calculation.
	testSquares := []int{A1, E4, H8, D5}

	for _, square := range testSquares {
		occupancies := []Bitboard{
			0,                            // Empty board
			Bitboard(0x0F0F0F0F0F0F0F0F), // Alternating pattern
			Bitboard(0x8142241818244281), // Random pattern
			Bitboard(0xFFFFFFFFFFFFFFFF), // Full board
		}

		for _, occupancy := range occupancies {
			magicRookAttacks := GetRookAttacks(square, occupancy)
			calculatedRookAttacks := calculateRookAttacks(square, occupancy)

			if magicRookAttacks != calculatedRookAttacks {
				t.Errorf("Rook magic attacks don't match calculated attacks for square %s",
					SquareToString(square))
			}

			magicBishopAttacks := GetBishopAttacks(square, occupancy)
			calculatedBishopAttacks := calculateBishopAttacks(square, occupancy)

			if magicBishopAttacks != calculatedBishopAttacks {
				t.Errorf("Bishop magic attacks don't match calculated attacks for square %s",
					SquareToString(square))
			}
		}
	}
}

func TestInvalidSquareInputs(t *testing.T) {
	t.Parallel()

	invalidSquares := []int{-1, 64, 100}
	occupancy := Bitboard(0x123456789ABCDEF0)

	for _, square := range invalidSquares {
		if GetRookAttacks(square, occupancy) != 0 {
			t.Errorf("GetRookAttacks(%d) should return 0 for invalid square", square)
		}
		if GetBishopAttacks(square, occupancy) != 0 {
			t.Errorf("GetBishopAttacks(%d) should return 0 for invalid square", square)
		}
		if GetQueenAttacks(square, occupancy) != 0 {
			t.Errorf("GetQueenAttacks(%d) should return 0 for invalid square", square)
		}
	}
}

func TestRelevantOccupancyMasks(t *testing.T) {
	t.Parallel()
	// Relevant-occupancy masks exclude edge squares and the piece's own square.
	testCases := []struct {
		square      int
		rookBits    int
		bishopBits  int
		description string
	}{
		{A1, 12, 6, "corner square"},
		{E4, 10, 9, "center square"},
		{H8, 12, 6, "corner square"},
		{A4, 11, 5, "edge square"},
		{E1, 11, 5, "edge square"},
	}

	for _, tc := range testCases {
		rookMask := rookRelevantOccupancy(tc.square)
		bishopMask := bishopRelevantOccupancy(tc.square)

		if rookMask.PopCount() != tc.rookBits {
			t.Errorf("Rook relevant occupancy for %s (%s) should have %d bits, got %d",
				SquareToString(tc.square), tc.description, tc.rookBits, rookMask.PopCount())
		}

		if bishopMask.PopCount() != tc.bishopBits {
			t.Errorf("Bishop relevant occupancy for %s (%s) should have %d bits, got %d",
				SquareToString(tc.square), tc.description, tc.bishopBits, bishopMask.PopCount())
		}

		if rookMask.HasBit(tc.square) {
			t.Errorf("Rook relevant occupancy should not include the piece's own square")
		}

		if bishopMask.HasBit(tc.square) {
			t.Errorf("Bishop relevant occupancy should not include the piece's own square")
		}
	}
}

func BenchmarkRookAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for range b.N {
		_ = GetRookAttacks(E4, occupancy)
	}
}

func BenchmarkBishopAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for range b.N {
		_ = GetBishopAttacks(E4, occupancy)
	}
}

func BenchmarkCalculateRookAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for range b.N {
		_ = calculateRookAttacks(E4, occupancy)
	}
}

func BenchmarkCalculateBishopAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for range b.N {
		_ = calculateBishopAttacks(E4, occupancy)
	}
}
