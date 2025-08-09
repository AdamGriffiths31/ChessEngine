package board

import (
	"testing"
)

func TestMagicBitboardsInitialization(t *testing.T) {
	// Magic bitboards are automatically initialized at package load time
	// Check that attack tables have been allocated
	if len(RookAttacks) == 0 {
		t.Error("RookAttacks table should be allocated")
	}
	if len(BishopAttacks) == 0 {
		t.Error("BishopAttacks table should be allocated")
	}
}

func TestRookAttacksEmptyBoard(t *testing.T) {
	
	testCases := []struct {
		square   int
		expected []int // Expected attack squares on empty board
	}{
		{
			square: A1,
			expected: []int{
				A2, A3, A4, A5, A6, A7, A8, // File
				B1, C1, D1, E1, F1, G1, H1, // Rank
			},
		},
		{
			square: E4,
			expected: []int{
				E1, E2, E3, E5, E6, E7, E8, // File
				A4, B4, C4, D4, F4, G4, H4, // Rank
			},
		},
		{
			square: H8,
			expected: []int{
				H1, H2, H3, H4, H5, H6, H7, // File
				A8, B8, C8, D8, E8, F8, G8, // Rank
			},
		},
	}
	
	for _, tc := range testCases {
		var emptyBoard Bitboard = 0
		attacks := GetRookAttacks(tc.square, emptyBoard)
		
		// Check that expected squares are attacked
		for _, expectedSquare := range tc.expected {
			if !attacks.HasBit(expectedSquare) {
				t.Errorf("Rook on %s should attack %s on empty board", 
					SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}
		
		// Check that the number of attacks is correct
		if attacks.PopCount() != len(tc.expected) {
			t.Errorf("Rook on %s should have %d attacks on empty board, got %d", 
				SquareToString(tc.square), len(tc.expected), attacks.PopCount())
		}
		
		// Check that the rook doesn't attack its own square
		if attacks.HasBit(tc.square) {
			t.Errorf("Rook on %s should not attack its own square", SquareToString(tc.square))
		}
	}
}

func TestBishopAttacksEmptyBoard(t *testing.T) {
	
	testCases := []struct {
		square   int
		expected []int // Expected attack squares on empty board
	}{
		{
			square: A1,
			expected: []int{B2, C3, D4, E5, F6, G7, H8}, // Only one diagonal from corner
		},
		{
			square: E4,
			expected: []int{
				// Main diagonal (northwest-southeast)
				A8, B7, C6, D5, F3, G2, H1,
				// Anti-diagonal (northeast-southwest)
				B1, C2, D3, F5, G6, H7,
			},
		},
		{
			square: H8,
			expected: []int{A1, B2, C3, D4, E5, F6, G7}, // Only one diagonal from corner
		},
		{
			square: D1,
			expected: []int{
				// Northeast
				E2, F3, G4, H5,
				// Northwest
				C2, B3, A4,
			},
		},
	}
	
	for _, tc := range testCases {
		var emptyBoard Bitboard = 0
		attacks := GetBishopAttacks(tc.square, emptyBoard)
		
		// Check that expected squares are attacked
		for _, expectedSquare := range tc.expected {
			if !attacks.HasBit(expectedSquare) {
				t.Errorf("Bishop on %s should attack %s on empty board", 
					SquareToString(tc.square), SquareToString(expectedSquare))
			}
		}
		
		// Check that the number of attacks is correct
		if attacks.PopCount() != len(tc.expected) {
			t.Errorf("Bishop on %s should have %d attacks on empty board, got %d", 
				SquareToString(tc.square), len(tc.expected), attacks.PopCount())
		}
		
		// Check that the bishop doesn't attack its own square
		if attacks.HasBit(tc.square) {
			t.Errorf("Bishop on %s should not attack its own square", SquareToString(tc.square))
		}
	}
}

func TestQueenAttacksEmptyBoard(t *testing.T) {
	
	square := E4
	var emptyBoard Bitboard = 0
	
	rookAttacks := GetRookAttacks(square, emptyBoard)
	bishopAttacks := GetBishopAttacks(square, emptyBoard)
	queenAttacks := GetQueenAttacks(square, emptyBoard)
	
	expectedQueenAttacks := rookAttacks | bishopAttacks
	
	if queenAttacks != expectedQueenAttacks {
		t.Errorf("Queen attacks should be the union of rook and bishop attacks")
	}
	
	// Check specific squares
	expectedSquares := []int{
		// Rook attacks (horizontal and vertical)
		E1, E2, E3, E5, E6, E7, E8, // File
		A4, B4, C4, D4, F4, G4, H4, // Rank
		// Bishop attacks (diagonals)
		A8, B7, C6, D5, F3, G2, H1, // Main diagonal
		B1, C2, D3, F5, G6, H7,     // Anti-diagonal
	}
	
	for _, expectedSquare := range expectedSquares {
		if !queenAttacks.HasBit(expectedSquare) {
			t.Errorf("Queen on %s should attack %s on empty board", 
				SquareToString(square), SquareToString(expectedSquare))
		}
	}
}

func TestRookAttacksWithOccupancy(t *testing.T) {
	
	// Test rook on e4 with pieces blocking some squares
	square := E4
	occupancy := Bitboard(0).SetBit(E6).SetBit(C4) // Pieces on e6 and c4
	
	attacks := GetRookAttacks(square, occupancy)
	
	// Should attack up to and including e6
	expectedNorth := []int{E5, E6}
	for _, sq := range expectedNorth {
		if !attacks.HasBit(sq) {
			t.Errorf("Rook should attack %s (blocked by piece on e6)", SquareToString(sq))
		}
	}
	
	// Should NOT attack e7, e8 (blocked by piece on e6)
	blockedNorth := []int{E7, E8}
	for _, sq := range blockedNorth {
		if attacks.HasBit(sq) {
			t.Errorf("Rook should NOT attack %s (blocked by piece on e6)", SquareToString(sq))
		}
	}
	
	// Should attack up to and including c4
	expectedWest := []int{D4, C4}
	for _, sq := range expectedWest {
		if !attacks.HasBit(sq) {
			t.Errorf("Rook should attack %s (up to blocker on c4)", SquareToString(sq))
		}
	}
	
	// Should NOT attack b4, a4 (blocked by piece on c4)
	blockedWest := []int{B4, A4}
	for _, sq := range blockedWest {
		if attacks.HasBit(sq) {
			t.Errorf("Rook should NOT attack %s (blocked by piece on c4)", SquareToString(sq))
		}
	}
	
	// Should attack unblocked directions normally
	expectedSouth := []int{E3, E2, E1}
	for _, sq := range expectedSouth {
		if !attacks.HasBit(sq) {
			t.Errorf("Rook should attack %s (unblocked direction)", SquareToString(sq))
		}
	}
	
	expectedEast := []int{F4, G4, H4}
	for _, sq := range expectedEast {
		if !attacks.HasBit(sq) {
			t.Errorf("Rook should attack %s (unblocked direction)", SquareToString(sq))
		}
	}
}

func TestBishopAttacksWithOccupancy(t *testing.T) {
	
	// Test bishop on e4 with pieces blocking some squares
	square := E4
	occupancy := Bitboard(0).SetBit(G6).SetBit(C2) // Pieces on g6 and c2
	
	attacks := GetBishopAttacks(square, occupancy)
	
	// Should attack up to and including g6
	expectedNorthEast := []int{F5, G6}
	for _, sq := range expectedNorthEast {
		if !attacks.HasBit(sq) {
			t.Errorf("Bishop should attack %s (up to blocker on g6)", SquareToString(sq))
		}
	}
	
	// Should NOT attack h7 (blocked by piece on g6)
	if attacks.HasBit(H7) {
		t.Errorf("Bishop should NOT attack h7 (blocked by piece on g6)")
	}
	
	// Should attack up to and including c2
	expectedSouthWest := []int{D3, C2}
	for _, sq := range expectedSouthWest {
		if !attacks.HasBit(sq) {
			t.Errorf("Bishop should attack %s (up to blocker on c2)", SquareToString(sq))
		}
	}
	
	// Should NOT attack b1 (blocked by piece on c2)
	if attacks.HasBit(B1) {
		t.Errorf("Bishop should NOT attack b1 (blocked by piece on c2)")
	}
	
	// Should attack unblocked directions normally
	expectedNorthWest := []int{D5, C6, B7, A8}
	for _, sq := range expectedNorthWest {
		if !attacks.HasBit(sq) {
			t.Errorf("Bishop should attack %s (unblocked direction)", SquareToString(sq))
		}
	}
	
	expectedSouthEast := []int{F3, G2, H1}
	for _, sq := range expectedSouthEast {
		if !attacks.HasBit(sq) {
			t.Errorf("Bishop should attack %s (unblocked direction)", SquareToString(sq))
		}
	}
}

func TestSlidingPieceAttacksConsistency(t *testing.T) {
	
	// Test that magic bitboard results are consistent with on-the-fly calculation
	// for various occupancy patterns
	testSquares := []int{A1, E4, H8, D5}
	
	for _, square := range testSquares {
		// Test with several different occupancy patterns
		occupancies := []Bitboard{
			0,                                    // Empty board
			Bitboard(0x0F0F0F0F0F0F0F0F),        // Alternating pattern
			Bitboard(0x8142241818244281),        // Random pattern
			Bitboard(0xFFFFFFFFFFFFFFFF),        // Full board
		}
		
		for _, occupancy := range occupancies {
			// Test rook attacks
			magicRookAttacks := GetRookAttacks(square, occupancy)
			onTheFlyRookAttacks := rookAttacksOnTheFly(square, occupancy)
			
			if magicRookAttacks != onTheFlyRookAttacks {
				t.Errorf("Rook magic attacks don't match on-the-fly calculation for square %s", 
					SquareToString(square))
			}
			
			// Test bishop attacks
			magicBishopAttacks := GetBishopAttacks(square, occupancy)
			onTheFlyBishopAttacks := bishopAttacksOnTheFly(square, occupancy)
			
			if magicBishopAttacks != onTheFlyBishopAttacks {
				t.Errorf("Bishop magic attacks don't match on-the-fly calculation for square %s", 
					SquareToString(square))
			}
		}
	}
}

func TestInvalidSquareInputs(t *testing.T) {
	
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
	// Test that relevant occupancy masks exclude edges properly
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
		
		// Verify that edges are excluded for rook masks
		if rookMask.HasBit(tc.square) {
			t.Errorf("Rook relevant occupancy should not include the piece's own square")
		}
		
		// Verify that edges are excluded for bishop masks
		if bishopMask.HasBit(tc.square) {
			t.Errorf("Bishop relevant occupancy should not include the piece's own square")
		}
	}
}

// Benchmark tests
func BenchmarkRookAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRookAttacks(E4, occupancy)
	}
}

func BenchmarkBishopAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBishopAttacks(E4, occupancy)
	}
}

func BenchmarkQueenAttacks(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetQueenAttacks(E4, occupancy)
	}
}

func BenchmarkRookAttacksOnTheFly(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rookAttacksOnTheFly(E4, occupancy)
	}
}

func BenchmarkBishopAttacksOnTheFly(b *testing.B) {
	occupancy := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bishopAttacksOnTheFly(E4, occupancy)
	}
}