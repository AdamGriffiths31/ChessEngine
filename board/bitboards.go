package board

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

// PrintBitboard displays a visual representation of the board
func PrintBitboard(bitBoard uint64) {
	var shiftMe uint64 = 1
	fmt.Printf("bitBoard:%v\n", bitBoard)

	for rank := data.Rank8; rank >= data.Rank1; rank-- {
		for file := data.FileA; file <= data.FileH; file++ {
			sq := data.FileRankToSquare(file, rank)
			sq64 := data.Square120ToSquare64[sq]
			if ((shiftMe << sq64) & bitBoard) == 0 {
				x := "-"
				fmt.Printf("%3s", x)

			} else {
				fmt.Printf("%3c", 'a'+file)
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
}

// PopBit returns the index of the first set bit and converts it to 0
func PopBitCopy(bb *uint64, count int) int {
	copy := *bb
	var res int
	for i := 0; i <= count; i++ {
		b := copy ^ (copy - uint64(1))
		fold := uint32((b & 0xffffffff) ^ (b >> 32))
		copy &= (copy - uint64(1))
		res = data.BitTable[(fold*0x783a9b23)>>26]
	}

	return res
}

// PopBitCopyReturn120 returns the index of the bit at the given count
func PopBitCopyReturn120(bb *uint64, count int) int {
	copy := *bb
	var res int
	for i := 0; i <= count; i++ {
		b := copy ^ (copy - uint64(1))
		fold := uint32((b & 0xffffffff) ^ (b >> 32))
		copy &= (copy - uint64(1))
		res = data.BitTable[(fold*0x783a9b23)>>26]
	}

	return data.Square64ToSquare120[res]
}

// PopBit returns the index of the first set bit and converts it to 0
func PopBit(bb *uint64) int {
	b := *bb ^ (*bb - uint64(1))
	fold := uint32((b & 0xffffffff) ^ (b >> 32))
	*bb &= (*bb - uint64(1))

	return data.BitTable[(fold*0x783a9b23)>>26]
}

// CountBits returns the count of '1' in the bitboard
func CountBits(b uint64) int {
	r := 0
	for ; b > 0; r++ {
		b &= b - 1
	}
	return r
}

// ClearBit clears the bit at the given square
func ClearBit(bb *uint64, square int) {
	*bb &= data.ClearMask[square]
}

// SetBit sets the bit at the given square
func SetBit(bb *uint64, square int) {
	*bb |= data.SetMask[square]
}
