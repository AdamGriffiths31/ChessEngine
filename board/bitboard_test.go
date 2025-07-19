package board

import (
	"testing"
)

func TestBitboardBasicOperations(t *testing.T) {
	// Test SetBit
	var bb Bitboard = 0
	bb = bb.SetBit(0)   // a1
	bb = bb.SetBit(63)  // h8
	bb = bb.SetBit(28)  // e4

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

	// Test ClearBit
	bb = bb.ClearBit(28)
	if bb.HasBit(28) {
		t.Error("Expected bit 28 (e4) to be cleared")
	}
	if !bb.HasBit(0) || !bb.HasBit(63) {
		t.Error("Other bits should remain set")
	}

	// Test ToggleBit
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
	// Test empty bitboard
	var empty Bitboard = 0
	if empty.LSB() != -1 {
		t.Error("LSB of empty bitboard should be -1")
	}
	if empty.MSB() != -1 {
		t.Error("MSB of empty bitboard should be -1")
	}

	// Test single bit
	bb := Bitboard(0).SetBit(28) // e4
	if bb.LSB() != 28 {
		t.Errorf("Expected LSB to be 28, got %d", bb.LSB())
	}
	if bb.MSB() != 28 {
		t.Errorf("Expected MSB to be 28, got %d", bb.MSB())
	}

	// Test multiple bits
	bb = bb.SetBit(0).SetBit(63)
	if bb.LSB() != 0 {
		t.Errorf("Expected LSB to be 0, got %d", bb.LSB())
	}
	if bb.MSB() != 63 {
		t.Errorf("Expected MSB to be 63, got %d", bb.MSB())
	}

	// Test PopLSB
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

func TestBitboardPopCount(t *testing.T) {
	var bb Bitboard = 0

	// Empty bitboard
	if bb.PopCount() != 0 {
		t.Errorf("Expected pop count of empty bitboard to be 0, got %d", bb.PopCount())
	}

	// Single bit
	bb = bb.SetBit(28)
	if bb.PopCount() != 1 {
		t.Errorf("Expected pop count to be 1, got %d", bb.PopCount())
	}

	// Multiple bits
	bb = bb.SetBit(0).SetBit(63).SetBit(32)
	if bb.PopCount() != 4 {
		t.Errorf("Expected pop count to be 4, got %d", bb.PopCount())
	}

	// All bits set
	bb = Bitboard(0xFFFFFFFFFFFFFFFF)
	if bb.PopCount() != 64 {
		t.Errorf("Expected pop count to be 64, got %d", bb.PopCount())
	}
}

func TestCoordinateConversion(t *testing.T) {
	testCases := []struct {
		square int
		file   int
		rank   int
		str    string
	}{
		{0, 0, 0, "a1"},   // a1
		{7, 7, 0, "h1"},   // h1
		{56, 0, 7, "a8"},  // a8
		{63, 7, 7, "h8"},  // h8
		{28, 4, 3, "e4"},  // e4
		{35, 3, 4, "d5"},  // d5
	}

	for _, tc := range testCases {
		// Test FileRankToSquare
		square := FileRankToSquare(tc.file, tc.rank)
		if square != tc.square {
			t.Errorf("FileRankToSquare(%d, %d) = %d, expected %d", tc.file, tc.rank, square, tc.square)
		}

		// Test SquareToFileRank
		file, rank := SquareToFileRank(tc.square)
		if file != tc.file || rank != tc.rank {
			t.Errorf("SquareToFileRank(%d) = (%d, %d), expected (%d, %d)", tc.square, file, rank, tc.file, tc.rank)
		}

		// Test SquareToString
		str := SquareToString(tc.square)
		if str != tc.str {
			t.Errorf("SquareToString(%d) = %s, expected %s", tc.square, str, tc.str)
		}

		// Test StringToSquare
		square = StringToSquare(tc.str)
		if square != tc.square {
			t.Errorf("StringToSquare(%s) = %d, expected %d", tc.str, square, tc.square)
		}
	}

	// Test invalid inputs
	if StringToSquare("") != -1 {
		t.Error("StringToSquare(\"\") should return -1")
	}
	if StringToSquare("i1") != -1 {
		t.Error("StringToSquare(\"i1\") should return -1")
	}
	if StringToSquare("a9") != -1 {
		t.Error("StringToSquare(\"a9\") should return -1")
	}
	if SquareToString(-1) != "invalid" {
		t.Error("SquareToString(-1) should return \"invalid\"")
	}
	if SquareToString(64) != "invalid" {
		t.Error("SquareToString(64) should return \"invalid\"")
	}
}

func TestFileAndRankMasks(t *testing.T) {
	// Test file masks
	for file := 0; file < 8; file++ {
		mask := FileMask(file)
		for rank := 0; rank < 8; rank++ {
			square := FileRankToSquare(file, rank)
			if !mask.HasBit(square) {
				t.Errorf("FileMask(%d) should have bit %d set", file, square)
			}
		}
		// Check that no other bits are set
		if mask.PopCount() != 8 {
			t.Errorf("FileMask(%d) should have exactly 8 bits set, got %d", file, mask.PopCount())
		}
	}

	// Test rank masks
	for rank := 0; rank < 8; rank++ {
		mask := RankMask(rank)
		for file := 0; file < 8; file++ {
			square := FileRankToSquare(file, rank)
			if !mask.HasBit(square) {
				t.Errorf("RankMask(%d) should have bit %d set", rank, square)
			}
		}
		// Check that no other bits are set
		if mask.PopCount() != 8 {
			t.Errorf("RankMask(%d) should have exactly 8 bits set, got %d", rank, mask.PopCount())
		}
	}
}

func TestBitboardShifts(t *testing.T) {
	// Test with e4 square (bit 28)
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

func TestBitboardDisplay(t *testing.T) {
	// Test empty bitboard
	var empty Bitboard = 0
	str := empty.String()
	if len(str) == 0 {
		t.Error("Empty bitboard string should not be empty")
	}

	// Test debug output
	bb := Bitboard(0).SetBit(0).SetBit(63)
	debug := bb.Debug()
	if len(debug) == 0 {
		t.Error("Debug string should not be empty")
	}

	// Test hex output
	hex := bb.Hex()
	if len(hex) == 0 {
		t.Error("Hex string should not be empty")
	}
}

func TestBitList(t *testing.T) {
	// Test empty bitboard
	var empty Bitboard = 0
	list := empty.BitList()
	if len(list) != 0 {
		t.Error("Empty bitboard should return empty list")
	}

	// Test single bit
	bb := Bitboard(0).SetBit(28)
	list = bb.BitList()
	if len(list) != 1 || list[0] != 28 {
		t.Errorf("Single bit bitboard should return [28], got %v", list)
	}

	// Test multiple bits
	bb = bb.SetBit(0).SetBit(63)
	list = bb.BitList()
	if len(list) != 3 {
		t.Errorf("Expected 3 bits in list, got %d", len(list))
	}
	// Should be in ascending order due to PopLSB
	expected := []int{0, 28, 63}
	for i, expected := range expected {
		if list[i] != expected {
			t.Errorf("Expected bit %d at index %d, got %d", expected, i, list[i])
		}
	}
}

func TestIsEmptyNotEmpty(t *testing.T) {
	var empty Bitboard = 0
	if !empty.IsEmpty() {
		t.Error("Empty bitboard should return true for IsEmpty()")
	}
	if empty.IsNotEmpty() {
		t.Error("Empty bitboard should return false for IsNotEmpty()")
	}

	bb := Bitboard(0).SetBit(28)
	if bb.IsEmpty() {
		t.Error("Non-empty bitboard should return false for IsEmpty()")
	}
	if !bb.IsNotEmpty() {
		t.Error("Non-empty bitboard should return true for IsNotEmpty()")
	}
}

func TestBitboardPieceColor(t *testing.T) {
	// Test white pieces
	whitePieces := []Piece{WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing}
	for _, piece := range whitePieces {
		if GetBitboardColor(piece) != BitboardWhite {
			t.Errorf("Expected %c to be white", piece)
		}
		if !IsWhitePiece(piece) {
			t.Errorf("Expected %c to be identified as white piece", piece)
		}
		if IsBlackPiece(piece) {
			t.Errorf("Expected %c to not be identified as black piece", piece)
		}
	}

	// Test black pieces
	blackPieces := []Piece{BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing}
	for _, piece := range blackPieces {
		if GetBitboardColor(piece) != BitboardBlack {
			t.Errorf("Expected %c to be black", piece)
		}
		if !IsBlackPiece(piece) {
			t.Errorf("Expected %c to be identified as black piece", piece)
		}
		if IsWhitePiece(piece) {
			t.Errorf("Expected %c to not be identified as white piece", piece)
		}
	}
}

func TestOppositeBitboardColor(t *testing.T) {
	if OppositeBitboardColor(BitboardWhite) != BitboardBlack {
		t.Error("Opposite of BitboardWhite should be BitboardBlack")
	}
	if OppositeBitboardColor(BitboardBlack) != BitboardWhite {
		t.Error("Opposite of BitboardBlack should be BitboardWhite")
	}
}

func TestColorConversion(t *testing.T) {
	// Test PieceColor to BitboardColor
	if ConvertToBitboardColor(WhiteColor) != BitboardWhite {
		t.Error("WhiteColor should convert to BitboardWhite")
	}
	if ConvertToBitboardColor(BlackColor) != BitboardBlack {
		t.Error("BlackColor should convert to BitboardBlack")
	}

	// Test BitboardColor to PieceColor
	if ConvertFromBitboardColor(BitboardWhite) != WhiteColor {
		t.Error("BitboardWhite should convert to WhiteColor")
	}
	if ConvertFromBitboardColor(BitboardBlack) != BlackColor {
		t.Error("BitboardBlack should convert to BlackColor")
	}
}

// Benchmark tests for performance
func BenchmarkBitboardOperations(b *testing.B) {
	bb := Bitboard(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bb = bb.SetBit(i % 64)
		bb = bb.ClearBit(i % 64)
		_ = bb.HasBit(i % 64)
	}
}

func BenchmarkPopCount(b *testing.B) {
	bb := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bb.PopCount()
	}
}

func BenchmarkBitScanning(b *testing.B) {
	bb := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bb.LSB()
		_ = bb.MSB()
	}
}

func BenchmarkPopLSB(b *testing.B) {
	bb := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		temp := bb
		for temp != 0 {
			_, temp = temp.PopLSB()
		}
	}
}

func BenchmarkShiftOperations(b *testing.B) {
	bb := Bitboard(0x123456789ABCDEF0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bb.ShiftNorth()
		_ = bb.ShiftSouth()
		_ = bb.ShiftEast()
		_ = bb.ShiftWest()
	}
}