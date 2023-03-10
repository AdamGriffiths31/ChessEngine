package engine

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

// SetPieceAtSquare updates the bitboard for the corresponding
// piece to an active occupancy
func (b *Bitboard) SetPieceAtSquare(sq64, piece int) {
	switch piece {
	case WP:
		SetBit(&b.WhitePawn, sq64)
	case WN:
		SetBit(&b.WhiteKnight, sq64)
	case WB:
		SetBit(&b.WhiteBishop, sq64)
	case WR:
		SetBit(&b.WhiteRook, sq64)
	case WQ:
		SetBit(&b.WhiteQueen, sq64)
	case WK:
		SetBit(&b.WhiteKing, sq64)
	case BP:
		SetBit(&b.BlackPawn, sq64)
	case BN:
		SetBit(&b.BlackKnight, sq64)
	case BB:
		SetBit(&b.BlackBishop, sq64)
	case BR:
		SetBit(&b.BlackRook, sq64)
	case BQ:
		SetBit(&b.BlackQueen, sq64)
	case BK:
		SetBit(&b.BlackKing, sq64)
	}
}

// SetBit updates the given bitboard at the given square to
// an active occupancy
func SetBit(bb *uint64, square int) {
	*bb |= data.SetMask[square]
}

func (b *Bitboard) PrintBitboard(bitBoard uint64) {
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
