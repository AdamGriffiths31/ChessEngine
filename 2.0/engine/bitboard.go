package engine

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

// SetPieceAtSquare updates the bitboard for the corresponding
// piece to an active occupancy
func (b *Bitboard) SetPieceAtSquare(sq64, piece int) {
	SetBit(&b.Pieces, sq64)

	switch data.PieceCol[piece] {
	case data.White:
		SetBit(&b.WhitePieces, sq64)
	case data.Black:
		SetBit(&b.BlackPieces, sq64)
	}

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

// RemovePieceAtSquare updates the bitboard for the corresponding
// piece to an inactive occupancy
func (b *Bitboard) RemovePieceAtSquare(sq64, piece int) {
	ClearBit(&b.Pieces, sq64)

	switch data.PieceCol[piece] {
	case data.White:
		ClearBit(&b.WhitePieces, sq64)
	case data.Black:
		ClearBit(&b.BlackPieces, sq64)
	}

	switch piece {
	case WP:
		ClearBit(&b.WhitePawn, sq64)
	case WN:
		ClearBit(&b.WhiteKnight, sq64)
	case WB:
		ClearBit(&b.WhiteBishop, sq64)
	case WR:
		ClearBit(&b.WhiteRook, sq64)
	case WQ:
		ClearBit(&b.WhiteQueen, sq64)
	case WK:
		ClearBit(&b.WhiteKing, sq64)
	case BP:
		ClearBit(&b.BlackPawn, sq64)
	case BN:
		ClearBit(&b.BlackKnight, sq64)
	case BB:
		ClearBit(&b.BlackBishop, sq64)
	case BR:
		ClearBit(&b.BlackRook, sq64)
	case BQ:
		ClearBit(&b.BlackQueen, sq64)
	case BK:
		ClearBit(&b.BlackKing, sq64)
	}
}

// PieceAt returns the piece at a given square
func (b *Bitboard) PieceAt(sq64 int) int {
	if sq64 == data.NoSquare {
		return data.Empty
	}
	mask := data.SquareMask[sq64]
	if b.WhitePieces&mask != 0 {
		if b.WhitePieces&b.WhitePawn&mask != 0 {
			return WP
		} else if b.WhitePieces&b.WhiteKnight&mask != 0 {
			return WN
		} else if b.WhitePieces&b.WhiteBishop&mask != 0 {
			return WB
		} else if b.WhitePieces&b.WhiteRook&mask != 0 {
			return WR
		} else if b.WhitePieces&b.WhiteQueen&mask != 0 {
			return WQ
		} else if b.WhitePieces&b.WhiteKing&mask != 0 {
			return WK
		}
	}
	if b.BlackPieces&mask != 0 {
		if b.BlackPieces&b.BlackPawn&mask != 0 {
			return BP
		} else if b.BlackPieces&b.BlackKnight&mask != 0 {
			return BN
		} else if b.BlackPieces&b.BlackBishop&mask != 0 {
			return BB
		} else if b.BlackPieces&b.BlackRook&mask != 0 {
			return BR
		} else if b.BlackPieces&b.BlackQueen&mask != 0 {
			return BQ
		} else if b.BlackPieces&b.BlackKing&mask != 0 {
			return BK
		}
	}

	return data.Empty
}

// SetBit updates the given bitboard at the given square to
// an active occupancy
func SetBit(bb *uint64, square int) {
	*bb |= data.SetMask[square]
}

// ClearBit updates the given bitboard at the given square to
// an inactive occupancy
func ClearBit(bb *uint64, square int) {
	*bb &= data.ClearMask[square]
}

func (b *Bitboard) CountBits(bb uint64) int {
	r := 0
	for ; bb > 0; r++ {
		bb &= bb - 1
	}
	return r
}

func (b *Bitboard) GetBitboardForPiece(piece int) uint64 {
	switch piece {
	case WP:
		return b.WhitePawn
	case WN:
		return b.WhiteKnight
	case WB:
		return b.WhiteBishop
	case WR:
		return b.WhiteRook
	case WQ:
		return b.WhiteQueen
	case WK:
		return b.WhiteKing
	case BP:
		return b.BlackPawn
	case BN:
		return b.BlackKnight
	case BB:
		return b.BlackBishop
	case BR:
		return b.BlackRook
	case BQ:
		return b.BlackQueen
	case BK:
		return b.BlackKing
	}

	panic(fmt.Errorf("GetBitboardForPiece: could not find bitboard for %v", piece))
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

func (b *Bitboard) PrintBoard() {
	var shiftMe uint64 = 1
	for rank := data.Rank8; rank >= data.Rank1; rank-- {
		for file := data.FileA; file <= data.FileH; file++ {
			sq := data.FileRankToSquare(file, rank)
			sq64 := data.Square120ToSquare64[sq]
			if ((shiftMe << sq64) & b.Pieces) == 0 {
				x := "-"
				fmt.Printf("%3s", x)

			} else {

				fmt.Printf("%3s", data.PceChar[b.PieceAt(sq64)])
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
}
