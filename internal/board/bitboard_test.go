package board

import (
	"testing"
)

func TestBitboardBasicOperations(t *testing.T) {
	t.Parallel()
	var bb Bitboard
	bb = bb.SetBit(0)  // a1
	bb = bb.SetBit(63) // h8
	bb = bb.SetBit(28) // e4

	if !bb.HasBit(0) {
		t.Error("Expected bit 0 (a1) to be set")
	}
	if !bb.HasBit(63) {
		t.Error("Expected bit 63 (h8) to be set")
	}
	if !bb.HasBit(28) {
		t.Error("Expected bit 28 (e4) to be set")
	}
	if bb.HasBit(1) {
		t.Error("Expected bit 1 (b1) to not be set")
	}

	bb = bb.ClearBit(28)
	if bb.HasBit(28) {
		t.Error("Expected bit 28 (e4) to be cleared")
	}
	if !bb.HasBit(0) || !bb.HasBit(63) {
		t.Error("Other bits should remain set")
	}

	bb = bb.ToggleBit(28) // Should set it again
	if !bb.HasBit(28) {
		t.Error("Expected bit 28 (e4) to be set after toggle")
	}
	bb = bb.ToggleBit(28) // Should clear it
	if bb.HasBit(28) {
		t.Error("Expected bit 28 (e4) to be cleared after second toggle")
	}
}

func TestBitboardBitScanning(t *testing.T) {
	t.Parallel()
	var empty Bitboard
	if empty.LSB() != -1 {
		t.Error("LSB of empty bitboard should be -1")
	}
	if empty.MSB() != -1 {
		t.Error("MSB of empty bitboard should be -1")
	}

	bb := Bitboard(0).SetBit(28) // e4
	if bb.LSB() != 28 {
		t.Errorf("Expected LSB to be 28, got %d", bb.LSB())
	}
	if bb.MSB() != 28 {
		t.Errorf("Expected MSB to be 28, got %d", bb.MSB())
	}

	bb = bb.SetBit(0).SetBit(63)
	if bb.LSB() != 0 {
		t.Errorf("Expected LSB to be 0, got %d", bb.LSB())
	}
	if bb.MSB() != 63 {
		t.Errorf("Expected MSB to be 63, got %d", bb.MSB())
	}

	square, newBB := bb.PopLSB()
	if square != 0 {
		t.Errorf("Expected popped square to be 0, got %d", square)
	}
	if newBB.HasBit(0) {
		t.Error("Bit 0 should be cleared after PopLSB")
	}
	if !newBB.HasBit(28) || !newBB.HasBit(63) {
		t.Error("Other bits should remain set")
	}
}

func TestBitboardShifts(t *testing.T) {
	t.Parallel()
	// e4 square (bit 28)
	bb := Bitboard(0).SetBit(28)

	// North shift should move to e5 (bit 36)
	north := bb.ShiftNorth()
	if !north.HasBit(36) {
		t.Error("ShiftNorth from e4 should set e5")
	}
	if north.HasBit(28) {
		t.Error("ShiftNorth should clear original bit")
	}

	// South shift should move to e3 (bit 20)
	south := bb.ShiftSouth()
	if !south.HasBit(20) {
		t.Error("ShiftSouth from e4 should set e3")
	}

	// East shift should move to f4 (bit 29)
	east := bb.ShiftEast()
	if !east.HasBit(29) {
		t.Error("ShiftEast from e4 should set f4")
	}

	// West shift should move to d4 (bit 27)
	west := bb.ShiftWest()
	if !west.HasBit(27) {
		t.Error("ShiftWest from e4 should set d4")
	}

	// Test edge cases - h-file should not wrap to a-file
	hFile := Bitboard(0).SetBit(H4)
	eastH := hFile.ShiftEast()
	if eastH != 0 {
		t.Error("ShiftEast from h-file should result in empty bitboard")
	}

	// Test edge cases - a-file should not wrap to h-file
	aFile := Bitboard(0).SetBit(A4)
	westA := aFile.ShiftWest()
	if westA != 0 {
		t.Error("ShiftWest from a-file should result in empty bitboard")
	}
}
