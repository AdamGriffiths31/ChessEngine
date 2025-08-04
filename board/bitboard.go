package board

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// Bitboard represents a chess board using a 64-bit integer
// Each bit corresponds to a square on the 8x8 chess board
// Bit 0 = a1, Bit 1 = b1, ..., Bit 63 = h8
type Bitboard uint64

// BitboardColor represents the color for bitboard operations
type BitboardColor int

const (
	BitboardWhite BitboardColor = iota
	BitboardBlack
)

// Square constants for easy reference
const (
	A1, B1, C1, D1, E1, F1, G1, H1 = 0, 1, 2, 3, 4, 5, 6, 7
	A2, B2, C2, D2, E2, F2, G2, H2 = 8, 9, 10, 11, 12, 13, 14, 15
	A3, B3, C3, D3, E3, F3, G3, H3 = 16, 17, 18, 19, 20, 21, 22, 23
	A4, B4, C4, D4, E4, F4, G4, H4 = 24, 25, 26, 27, 28, 29, 30, 31
	A5, B5, C5, D5, E5, F5, G5, H5 = 32, 33, 34, 35, 36, 37, 38, 39
	A6, B6, C6, D6, E6, F6, G6, H6 = 40, 41, 42, 43, 44, 45, 46, 47
	A7, B7, C7, D7, E7, F7, G7, H7 = 48, 49, 50, 51, 52, 53, 54, 55
	A8, B8, C8, D8, E8, F8, G8, H8 = 56, 57, 58, 59, 60, 61, 62, 63
)

// Center square masks
var (
	CenterFiles = FileMask(3) | FileMask(4) // d and e files
	CenterRanks = RankMask(3) | RankMask(4) // 4th and 5th ranks
)

// Core bitboard operations

// SetBit sets the bit at the given square
func (b Bitboard) SetBit(square int) Bitboard {
	return b | (1 << square)
}

// ClearBit clears the bit at the given square
func (b Bitboard) ClearBit(square int) Bitboard {
	return b &^ (1 << square)
}

// ToggleBit toggles the bit at the given square
func (b Bitboard) ToggleBit(square int) Bitboard {
	return b ^ (1 << square)
}

// HasBit returns true if the bit at the given square is set
func (b Bitboard) HasBit(square int) bool {
	return (b & (1 << square)) != 0
}

// PopCount returns the number of set bits in the bitboard
func (b Bitboard) PopCount() int {
	return bits.OnesCount64(uint64(b))
}

// LSB returns the index of the least significant bit (rightmost 1 bit)
// Returns -1 if the bitboard is empty
func (b Bitboard) LSB() int {
	if b == 0 {
		return -1
	}
	return bits.TrailingZeros64(uint64(b))
}

// MSB returns the index of the most significant bit (leftmost 1 bit)
// Returns -1 if the bitboard is empty
func (b Bitboard) MSB() int {
	if b == 0 {
		return -1
	}
	return 63 - bits.LeadingZeros64(uint64(b))
}

// PopLSB returns the LSB index and a new bitboard with that bit cleared
// Returns -1 and the original bitboard if empty
func (b Bitboard) PopLSB() (int, Bitboard) {
	if b == 0 {
		return -1, b
	}
	lsb := b.LSB()
	return lsb, b.ClearBit(lsb)
}

// IsEmpty returns true if the bitboard has no set bits
func (b Bitboard) IsEmpty() bool {
	return b == 0
}

// IsNotEmpty returns true if the bitboard has at least one set bit
func (b Bitboard) IsNotEmpty() bool {
	return b != 0
}

// Coordinate system and conversion utilities

// FileRankToSquare converts file (0-7) and rank (0-7) to square index (0-63)
func FileRankToSquare(file, rank int) int {
	return rank*8 + file
}

// SquareToFileRank converts square index (0-63) to file (0-7) and rank (0-7)
func SquareToFileRank(square int) (file, rank int) {
	return square % 8, square / 8
}

// SquareToString converts square index to algebraic notation (e.g., 0 -> "a1")
func SquareToString(square int) string {
	if square < 0 || square > 63 {
		return "invalid"
	}
	file, rank := SquareToFileRank(square)
	return string(rune('a'+file)) + string(rune('1'+rank))
}

// StringToSquare converts algebraic notation to square index (e.g., "a1" -> 0)
func StringToSquare(square string) int {
	if len(square) != 2 {
		return -1
	}
	file := int(square[0] - 'a')
	rank := int(square[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return -1
	}
	return FileRankToSquare(file, rank)
}

// GetFile returns the file (0-7) of a square
func GetFile(square int) int {
	return square % 8
}

// GetRank returns the rank (0-7) of a square
func GetRank(square int) int {
	return square / 8
}

// File and rank masks for quick operations
var (
	FileA = Bitboard(0x0101010101010101)
	FileB = Bitboard(0x0202020202020202)
	FileC = Bitboard(0x0404040404040404)
	FileD = Bitboard(0x0808080808080808)
	FileE = Bitboard(0x1010101010101010)
	FileF = Bitboard(0x2020202020202020)
	FileG = Bitboard(0x4040404040404040)
	FileH = Bitboard(0x8080808080808080)

	Rank1 = Bitboard(0x00000000000000FF)
	Rank2 = Bitboard(0x000000000000FF00)
	Rank3 = Bitboard(0x0000000000FF0000)
	Rank4 = Bitboard(0x00000000FF000000)
	Rank5 = Bitboard(0x000000FF00000000)
	Rank6 = Bitboard(0x0000FF0000000000)
	Rank7 = Bitboard(0x00FF000000000000)
	Rank8 = Bitboard(0xFF00000000000000)
)

// FileMask returns the file mask for a given file (0-7)
func FileMask(file int) Bitboard {
	return FileA << file
}

// RankMask returns the rank mask for a given rank (0-7)
func RankMask(rank int) Bitboard {
	return Rank1 << (rank * 8)
}

// GetFileBitboard returns a bitboard with all squares in the given file set
func GetFileBitboard(square int) Bitboard {
	return FileMask(GetFile(square))
}

// GetRankBitboard returns a bitboard with all squares in the given rank set
func GetRankBitboard(square int) Bitboard {
	return RankMask(GetRank(square))
}

// Bitboard display and debugging

// String returns a pretty-printed representation of the bitboard
func (b Bitboard) String() string {
	var result strings.Builder
	result.WriteString("  a b c d e f g h\n")
	
	// Print from rank 8 down to rank 1 (top to bottom visually)
	for rank := 7; rank >= 0; rank-- {
		result.WriteString(strconv.Itoa(rank + 1))
		result.WriteString(" ")
		for file := 0; file < 8; file++ {
			square := FileRankToSquare(file, rank)
			if b.HasBit(square) {
				result.WriteString("1 ")
			} else {
				result.WriteString(". ")
			}
		}
		result.WriteString("\n")
	}
	return result.String()
}

// Debug returns a debug representation showing hex value and bit count
func (b Bitboard) Debug() string {
	return fmt.Sprintf("Bitboard(0x%016X) - %d bits set", uint64(b), b.PopCount())
}

// Hex returns the hexadecimal representation of the bitboard
func (b Bitboard) Hex() string {
	return fmt.Sprintf("0x%016X", uint64(b))
}

// BitList returns a slice of square indices where bits are set
func (b Bitboard) BitList() []int {
	var squares []int
	temp := b
	for temp != 0 {
		square, newTemp := temp.PopLSB()
		squares = append(squares, square)
		temp = newTemp
	}
	return squares
}

// Shift operations for efficient bitboard manipulation

// ShiftNorth shifts the bitboard north (towards rank 8)
func (b Bitboard) ShiftNorth() Bitboard {
	return b << 8
}

// ShiftSouth shifts the bitboard south (towards rank 1)
func (b Bitboard) ShiftSouth() Bitboard {
	return b >> 8
}

// ShiftEast shifts the bitboard east (towards h-file), masking off h-file wrap
func (b Bitboard) ShiftEast() Bitboard {
	return (b << 1) &^ FileA
}

// ShiftWest shifts the bitboard west (towards a-file), masking off a-file wrap
func (b Bitboard) ShiftWest() Bitboard {
	return (b >> 1) &^ FileH
}

// ShiftNorthEast shifts the bitboard northeast
func (b Bitboard) ShiftNorthEast() Bitboard {
	return (b << 9) &^ FileA
}

// ShiftNorthWest shifts the bitboard northwest
func (b Bitboard) ShiftNorthWest() Bitboard {
	return (b << 7) &^ FileH
}

// ShiftSouthEast shifts the bitboard southeast
func (b Bitboard) ShiftSouthEast() Bitboard {
	return (b >> 7) &^ FileA
}

// ShiftSouthWest shifts the bitboard southwest
func (b Bitboard) ShiftSouthWest() Bitboard {
	return (b >> 9) &^ FileH
}

// Utility functions for bitboard colors

// GetBitboardColor returns the bitboard color of a piece
func GetBitboardColor(piece Piece) BitboardColor {
	if piece >= 'A' && piece <= 'Z' {
		return BitboardWhite
	}
	return BitboardBlack
}

// OppositeBitboardColor returns the opposite color
func OppositeBitboardColor(color BitboardColor) BitboardColor {
	return color ^ 1
}

// ConvertToBitboardColor converts PieceColor to BitboardColor
func ConvertToBitboardColor(color PieceColor) BitboardColor {
	if color == WhiteColor {
		return BitboardWhite
	}
	return BitboardBlack
}

// ConvertFromBitboardColor converts BitboardColor to PieceColor
func ConvertFromBitboardColor(color BitboardColor) PieceColor {
	if color == BitboardWhite {
		return WhiteColor
	}
	return BlackColor
}

// GetFileForwardFill returns all squares ahead of the given pawns on the same files
// For white pawns, "ahead" means towards rank 8 (north)
// For black pawns, "ahead" means towards rank 1 (south)
func GetFileForwardFill(pawns Bitboard, color BitboardColor) Bitboard {
	var fill Bitboard
	if color == BitboardWhite {
		// White pawns advance north
		temp := pawns.ShiftNorth()
		for temp != 0 {
			fill |= temp
			temp = temp.ShiftNorth()
		}
	} else {
		// Black pawns advance south
		temp := pawns.ShiftSouth()
		for temp != 0 {
			fill |= temp
			temp = temp.ShiftSouth()
		}
	}
	return fill
}

// GetAdjacentFileForwardFill returns all squares ahead on adjacent files
// Used to check if enemy pawns on adjacent files can capture advancing pawns
func GetAdjacentFileForwardFill(pawns Bitboard, color BitboardColor) Bitboard {
	// Get pawns on adjacent files (left and right)
	leftFiles := pawns.ShiftWest()
	rightFiles := pawns.ShiftEast()
	adjacentPawns := leftFiles | rightFiles
	
	// Get forward fill for adjacent file pawns
	return GetFileForwardFill(adjacentPawns, color)
}

// GetPassedPawns returns a bitboard of all passed pawns for the given color
// A passed pawn has no enemy pawns that can stop it from promoting
func GetPassedPawns(friendlyPawns, enemyPawns Bitboard, color BitboardColor) Bitboard {
	// Get all squares that enemy pawns control ahead on same files
	enemyFileControl := GetFileForwardFill(enemyPawns, OppositeBitboardColor(color))
	
	// Get all squares that enemy pawns control ahead on adjacent files
	enemyAdjacentControl := GetAdjacentFileForwardFill(enemyPawns, OppositeBitboardColor(color))
	
	// Combined enemy control (squares enemy pawns can reach)
	enemyControl := enemyFileControl | enemyAdjacentControl
	
	// Passed pawns are friendly pawns that are not blocked by enemy control
	var passedPawns Bitboard
	pawnList := friendlyPawns.BitList()
	
	for _, square := range pawnList {
		squareBit := Bitboard(1) << square
		
		// Get forward fill for this specific pawn
		pawnForwardFill := GetFileForwardFill(squareBit, color)
		
		// If this pawn's forward path doesn't intersect enemy control, it's passed
		if (pawnForwardFill & enemyControl) == 0 {
			passedPawns |= squareBit
		}
	}
	
	return passedPawns
}