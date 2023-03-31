package engine

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

var PreCalculatedKnightMoves [64]uint64
var PreCalculatedKingMoves [64]uint64
var index64 [64]int

func init() {
	preCalculatedMoves()
	for sq := 0; sq < 64; sq++ {
		b := uint64(1) << uint(sq)
		index64[(((b-1)^b)*0x03f79d71b4cb0a89)>>58] = sq
	}
}

// FirstSquare returns the first square in a bitboard
func FirstSquare(b uint64) int {
	return index64[(((b-1)^b)*0x03f79d71b4cb0a89)>>58]
}

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
	case data.WP:
		SetBit(&b.WhitePawn, sq64)
	case data.WN:
		SetBit(&b.WhiteKnight, sq64)
	case data.WB:
		SetBit(&b.WhiteBishop, sq64)
	case data.WR:
		SetBit(&b.WhiteRook, sq64)
	case data.WQ:
		SetBit(&b.WhiteQueen, sq64)
	case data.WK:
		SetBit(&b.WhiteKing, sq64)
	case data.BP:
		SetBit(&b.BlackPawn, sq64)
	case data.BN:
		SetBit(&b.BlackKnight, sq64)
	case data.BB:
		SetBit(&b.BlackBishop, sq64)
	case data.BR:
		SetBit(&b.BlackRook, sq64)
	case data.BQ:
		SetBit(&b.BlackQueen, sq64)
	case data.BK:
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
	case data.WP:
		ClearBit(&b.WhitePawn, sq64)
	case data.WN:
		ClearBit(&b.WhiteKnight, sq64)
	case data.WB:
		ClearBit(&b.WhiteBishop, sq64)
	case data.WR:
		ClearBit(&b.WhiteRook, sq64)
	case data.WQ:
		ClearBit(&b.WhiteQueen, sq64)
	case data.WK:
		ClearBit(&b.WhiteKing, sq64)
	case data.BP:
		ClearBit(&b.BlackPawn, sq64)
	case data.BN:
		ClearBit(&b.BlackKnight, sq64)
	case data.BB:
		ClearBit(&b.BlackBishop, sq64)
	case data.BR:
		ClearBit(&b.BlackRook, sq64)
	case data.BQ:
		ClearBit(&b.BlackQueen, sq64)
	case data.BK:
		ClearBit(&b.BlackKing, sq64)
	}
}

// PieceAt returns the piece at a given square
func (b *Bitboard) PieceAt(sq64 int) int {
	if sq64 == data.NoSquare || sq64 < 0 || sq64 >= 64 {
		return data.Empty
	}
	mask := data.SquareMask[sq64]
	if b.WhitePieces&mask != 0 {
		if b.WhitePieces&b.WhitePawn&mask != 0 {
			return data.WP
		} else if b.WhitePieces&b.WhiteKnight&mask != 0 {
			return data.WN
		} else if b.WhitePieces&b.WhiteBishop&mask != 0 {
			return data.WB
		} else if b.WhitePieces&b.WhiteRook&mask != 0 {
			return data.WR
		} else if b.WhitePieces&b.WhiteQueen&mask != 0 {
			return data.WQ
		} else if b.WhitePieces&b.WhiteKing&mask != 0 {
			return data.WK
		}
	}
	if b.BlackPieces&mask != 0 {
		if b.BlackPieces&b.BlackPawn&mask != 0 {
			return data.BP
		} else if b.BlackPieces&b.BlackKnight&mask != 0 {
			return data.BN
		} else if b.BlackPieces&b.BlackBishop&mask != 0 {
			return data.BB
		} else if b.BlackPieces&b.BlackRook&mask != 0 {
			return data.BR
		} else if b.BlackPieces&b.BlackQueen&mask != 0 {
			return data.BQ
		} else if b.BlackPieces&b.BlackKing&mask != 0 {
			return data.BK
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

const m0 = 0x5555555555555555
const m1 = 0x3333333333333333
const m2 = 0x0f0f0f0f0f0f0f0f

func (b *Bitboard) CountBits(bb uint64) int {
	const m = 1<<64 - 1
	bb = bb>>1&(m0&m) + bb&(m0&m)
	bb = bb>>2&(m1&m) + bb&(m1&m)
	bb = (bb>>4 + bb) & (m2 & m)
	bb += bb >> 8
	bb += bb >> 16
	bb += bb >> 32
	return int(bb) & (1<<7 - 1)
}

// GetBitboardForPiece returns the piece bitboard for the given
// piece
func (b *Bitboard) GetBitboardForPiece(piece int) uint64 {
	switch piece {
	case data.WP:
		return b.WhitePawn
	case data.WN:
		return b.WhiteKnight
	case data.WB:
		return b.WhiteBishop
	case data.WR:
		return b.WhiteRook
	case data.WQ:
		return b.WhiteQueen
	case data.WK:
		return b.WhiteKing
	case data.BP:
		return b.BlackPawn
	case data.BN:
		return b.BlackKnight
	case data.BB:
		return b.BlackBishop
	case data.BR:
		return b.BlackRook
	case data.BQ:
		return b.BlackQueen
	case data.BK:
		return b.BlackKing
	}

	panic(fmt.Errorf("GetBitboardForPiece: could not find bitboard for %v", piece))
}

// GetPiecesBitboard returns the bitboard for the given side
func (b *Bitboard) GetPiecesBitboard(side int) uint64 {
	if side == data.White {
		return b.WhitePieces
	} else if side == data.Black {
		return b.BlackPieces
	}

	panic(fmt.Errorf("GetPiecesBitboard: could not find bitboard for %v", side))
}

// GetPieces returns the piece bitboard for the given side
func (b *Bitboard) GetPieces(side, piece int) uint64 {
	if side == data.White {
		switch piece {
		case data.WP:
			return b.WhitePawn
		case data.WN:
			return b.WhiteKnight
		case data.WB:
			return b.WhiteBishop
		case data.WR:
			return b.WhiteRook
		case data.WQ:
			return b.WhiteQueen
		case data.WK:
			return b.WhiteKing
		}
	} else if side == data.Black {
		switch piece {
		case data.WP:
			return b.BlackPawn
		case data.WN:
			return b.BlackKnight
		case data.WB:
			return b.BlackBishop
		case data.WR:
			return b.BlackRook
		case data.WQ:
			return b.BlackQueen
		case data.WK:
			return b.BlackKing
		}
	}
	panic(fmt.Errorf("GetPieces: could not find bitboard for %v %v", piece, side))
}

// AllWhitePawnAttacks returns all white pawn attacks for the given bitboard
func (b *Bitboard) AllWhitePawnAttacks(bitboard uint64) uint64 {
	return ((bitboard & ^data.FileAMask) << 7) | ((bitboard & ^data.FileHMask) << 9)
}

// AllBlackPawnAttacks returns all black pawn attacks for the given bitboard
func (b *Bitboard) AllBlackPawnAttacks(bitboard uint64) uint64 {
	return ((bitboard & ^data.FileAMask) >> 9) | ((bitboard & ^data.FileHMask) >> 7)
}

// PrintBitboard visual representation of the given bitboard
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

// PrintBoard visual representation of the current position
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

// copy returns a copy of the current bitboard data
func (b *Bitboard) copy() Bitboard {
	return Bitboard{
		Pieces:      b.Pieces,
		BlackPieces: b.BlackPieces,
		BlackPawn:   b.BlackPawn,
		BlackKnight: b.BlackKnight,
		BlackBishop: b.BlackBishop,
		BlackRook:   b.BlackRook,
		BlackQueen:  b.BlackQueen,
		BlackKing:   b.BlackKing,
		WhitePieces: b.WhitePieces,
		WhitePawn:   b.WhitePawn,
		WhiteKnight: b.WhiteKnight,
		WhiteBishop: b.WhiteBishop,
		WhiteRook:   b.WhiteRook,
		WhiteQueen:  b.WhiteQueen,
		WhiteKing:   b.WhiteKing,
	}
}

// preCalculatedMoves pre-calculates all possible knight and king moves
func preCalculatedMoves() {
	for i := 0; i < 64; i++ {
		PreCalculatedKnightMoves[i] = calculateKnightMoves(i)
		PreCalculatedKingMoves[i] = calculateKingMoves(i)
	}
}

// calculateKingMoves calculates all possible king moves for the given square
func calculateKingMoves(sq64 int) uint64 {
	var attacks uint64
	sq120 := data.Square64ToSquare120[sq64]

	for i := 0; i < data.NumDir[data.WK]; i++ {
		dir := data.PieceDir[data.WK][i]
		tempSq := sq120 + dir
		if !validate.SquareOnBoard(tempSq) {
			continue
		}
		SetBit(&attacks, data.Square120ToSquare64[tempSq])
	}
	return attacks
}

// calculateKnightMoves calculates all possible knight moves for the given square
func calculateKnightMoves(sq64 int) uint64 {
	var attacks uint64
	sq120 := data.Square64ToSquare120[sq64]

	for i := 0; i < data.NumDir[data.WN]; i++ {
		dir := data.PieceDir[data.BN][i]
		tempSq := sq120 + dir
		if !validate.SquareOnBoard(tempSq) {
			continue
		}
		SetBit(&attacks, data.Square120ToSquare64[tempSq])
	}
	return attacks
}
