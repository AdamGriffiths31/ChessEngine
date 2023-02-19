package engine

import (
	"fmt"
)

// PrintBitboard displays a visual representation of the board
func PrintBitboard(bitBoard uint64) {
	var shiftMe uint64 = 1
	fmt.Printf("\n")

	for rank := Rank8; rank >= Rank1; rank-- {
		for file := FileA; file <= FileH; file++ {
			sq := FileRankToSquare(file, rank)
			sq64 := Sqaure120ToSquare64[sq]
			if ((shiftMe << sq64) & bitBoard) == 0 {
				fmt.Printf("-")
			} else {
				fmt.Printf("X")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
}

// PopBit returns the index of the first set bit and converts it to 0
func PopBit(bb *uint64) int {
	b := *bb ^ (*bb - 1)
	fold := uint32(b) ^ uint32(b>>32)
	*bb &= (*bb - 1)

	return BitTable[(fold*0x783a9b23)>>26]
}

// CountBits returns the count of '1' in the bitboard
func CountBits(b uint64) int {
	r := 0
	for ; b > 0; r++ {
		b &= b - 1
	}
	return r
}

// ClearBit clears the bit at the given sqaure
func ClearBit(bb *uint64, sqaure int) {
	*bb &= ClearMask[sqaure]
}

// SetBit sets the bit at the given square
func SetBit(bb *uint64, sqaure int) {
	*bb |= SetMask[sqaure]
}
