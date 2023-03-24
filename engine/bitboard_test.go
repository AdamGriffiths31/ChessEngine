package engine

import "testing"

func TestCountBits(t *testing.T) {
	var b = Bitboard{Pieces: 0x0101010101010101}
	if b.CountBits(b.Pieces) != 8 {
		t.Errorf("Expected 8 but got %v", b.CountBits(b.Pieces))
	}
}

func TestCountBitsEmpty(t *testing.T) {
	var b = Bitboard{Pieces: 0}
	if b.CountBits(b.Pieces) != 0 {
		t.Errorf("Expected 8 but got %v", b.CountBits(b.Pieces))
	}
}
