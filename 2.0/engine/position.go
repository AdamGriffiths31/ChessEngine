package engine

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func (p *Position) CheckBitboard() {
	return
	newBB := Bitboard{}
	for sq := 0; sq < 64; sq++ {
		piece := p.Board.PieceAt(sq)
		if sq != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			newBB.SetPieceAtSquare(sq, piece)
		}
	}
	for piece := data.WP; piece <= data.BK; piece++ {
		original := p.Board.CountBits(p.Board.GetBitboardForPiece(piece))
		copy := newBB.CountBits(p.Board.GetBitboardForPiece(piece))
		if original != copy {
			panic("err")
		}
	}

	if p.PositionKey != p.GeneratePositionKey() {
		fmt.Printf("%v - %v\n", p.PositionKey, p.GeneratePositionKey())
		panic("err")
	}
}

func (p *Position) TakeNullMoveBack(enPas int, castlePerm int) {
	p.Play--
	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}
	p.EnPassant = enPas
	p.CastlePermission = castlePerm
	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}
	p.Side ^= 1
	p.hashSide()
}

func (p *Position) MakeNullMove() (bool, int, int) {
	p.CheckBitboard()
	p.Play++
	enPas := p.EnPassant
	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}
	p.EnPassant = data.NoSquare
	p.Side ^= 1
	p.hashSide()
	return true, enPas, p.CastlePermission
}

func (p *Position) MakeMove(move int) (bool, int, int) {
	p.CheckBitboard()
	from := data.FromSquare(move)
	to := data.ToSquare(move)
	side := p.Side
	castlePerm := p.CastlePermission
	enPas := p.EnPassant
	if (move & data.MFLAGEP) != 0 {
		if side == data.White {
			p.ClearPiece(data.Square120ToSquare64[to-10])
		} else {
			p.ClearPiece(data.Square120ToSquare64[to+10])
		}
	} else if (move & data.MFLAGGCA) != 0 {
		switch to {
		case data.C1:
			p.MovePiece(data.A1, data.D1)
		case data.C8:
			p.MovePiece(data.A8, data.D8)
		case data.G1:
			p.MovePiece(data.H1, data.F1)
		case data.G8:
			p.MovePiece(data.H8, data.F8)
		default:
			panic(fmt.Errorf("TakeMoveBack: castle error %v %v", from, to))
		}
	}
	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}

	p.hashCastle()
	p.CastlePermission &= data.CastlePerm[from]
	p.CastlePermission &= data.CastlePerm[to]
	p.EnPassant = data.NoSquare
	p.hashCastle()

	captured := data.Captured(move)
	if captured != data.Empty {
		p.ClearPiece(data.Square120ToSquare64[to])
	}

	p.Play++

	piece := p.Board.PieceAt(data.Square120ToSquare64[from])
	if piece == WP || piece == BP {
		if (move & data.MFLAGPS) != 0 {
			if side == data.White {
				p.EnPassant = from + 10
			} else {
				p.EnPassant = from - 10
			}
			p.hashEnPas()
		}
	}
	p.MovePiece(from, to)

	promotedPiece := data.Promoted(move)
	if promotedPiece != 0 {
		p.ClearPiece(data.Square120ToSquare64[to])
		p.AddPiece(data.Square120ToSquare64[to], promotedPiece)
	}

	p.Side ^= 1

	p.hashSide()

	if p.IsKingAttacked(p.Side) {
		p.TakeMoveBack(move, p.EnPassant, castlePerm)
		return false, p.EnPassant, castlePerm
	}
	//p.History[p.PositionKey]++
	return true, enPas, castlePerm
}

func (p *Position) TakeMoveBack(move int, enPas int, castlePerm int) {
	p.CheckBitboard()
	p.Play--
	from := data.FromSquare(move)
	to := data.ToSquare(move)

	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}

	p.hashCastle()

	p.CastlePermission = castlePerm
	p.EnPassant = enPas

	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		p.hashEnPas()
	}

	p.hashCastle()

	p.Side ^= 1
	p.hashSide()
	if (move & data.MFLAGEP) != 0 {
		if p.Side == data.White {
			p.AddPiece(data.Square120ToSquare64[to-10], data.BP)
		} else {
			p.AddPiece(data.Square120ToSquare64[to+10], data.WP)
		}
	}
	if (move & data.MFLAGGCA) != 0 {
		switch to {
		case data.C1:
			p.MovePiece(data.D1, data.A1)
		case data.C8:
			p.MovePiece(data.D8, data.A8)
		case data.G1:
			p.MovePiece(data.F1, data.H1)
		case data.G8:
			p.MovePiece(data.F8, data.H8)
		default:
			panic(fmt.Errorf("TakeMoveBack: castle error %v %v", from, to))
		}
	}
	p.MovePiece(to, from)

	captured := data.Captured(move)
	if captured != data.Empty {
		p.AddPiece(data.Square120ToSquare64[to], captured)
	}

	if data.Promoted(move) != data.Empty {
		p.ClearPiece(data.Square120ToSquare64[from])
		if p.Side == data.White {
			p.AddPiece(data.Square120ToSquare64[from], data.WP)
		} else {
			p.AddPiece(data.Square120ToSquare64[from], data.BP)
		}
	}
	//p.History[p.PositionKey]--
}

func (p *Position) MovePiece(from, to int) {
	piece := p.Board.PieceAt(data.Square120ToSquare64[from])
	p.hashPiece(piece, from)
	p.Board.RemovePieceAtSquare(data.Square120ToSquare64[from], piece)
	p.hashPiece(piece, to)
	p.Board.SetPieceAtSquare(data.Square120ToSquare64[to], piece)
}

func (p *Position) ClearPiece(sq64 int) {
	p.hashPiece(p.Board.PieceAt(sq64), data.Square64ToSquare120[sq64])
	p.Board.RemovePieceAtSquare(sq64, p.Board.PieceAt(sq64))
}

func (p *Position) AddPiece(sq, piece int) {
	p.hashPiece(piece, data.Square64ToSquare120[sq])
	p.Board.SetPieceAtSquare(sq, piece)
}

func (p *Position) hashPiece(piece, square int) {
	p.PositionKey ^= data.PieceKeys[piece][square]
}

func (p *Position) hashCastle() {
	p.PositionKey ^= data.CastleKeys[p.CastlePermission]
}

func (p *Position) hashSide() {
	p.PositionKey ^= data.SideKey
}

func (p *Position) hashEnPas() {
	p.PositionKey ^= data.PieceKeys[data.Empty][p.EnPassant]
}

func (p *Position) GeneratePositionKey() uint64 {
	var finalKey uint64 = 0
	piece := data.Empty

	for sq := 0; sq < 64; sq++ {
		piece = p.Board.PieceAt(sq)
		if sq != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			finalKey ^= data.PieceKeys[piece][data.Square64ToSquare120[sq]]
		}
	}

	if p.Side == data.White {
		finalKey ^= data.SideKey
	}

	if p.EnPassant != data.NoSquare && p.EnPassant != Empty {
		finalKey ^= data.PieceKeys[data.Empty][p.EnPassant]
	}

	finalKey ^= data.CastleKeys[p.CastlePermission]
	return finalKey
}

func (p *Position) IsEndGame() bool {
	if p.Side == data.White {
		return p.Board.WhiteQueen+p.Board.WhiteRook+p.Board.WhiteBishop+p.Board.WhiteKnight == 0
	} else {
		return p.Board.BlackQueen+p.Board.BlackRook+p.Board.BlackBishop+p.Board.BlackKnight == 0
	}
}
