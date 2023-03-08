package moveGen

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func MakeMove(move int, pos *data.Board) bool {
	//TODO Validate the move
	board.CheckBoard(pos)

	from := data.FromSquare(move)
	to := data.ToSquare(move)
	side := pos.Side

	if !validate.SquareOnBoard(from) {
		fmt.Printf("\nMakeMove from sqaure Move: %v\n", io.PrintMove(move))
		io.PrintBoard(pos)
	}

	if !validate.SquareOnBoard(to) {
		fmt.Printf("\nMakeMove to sqaure Move: %v\n", io.PrintMove(move))
		io.PrintBoard(pos)
	}

	pos.History[pos.HistoryPlay].PositionKey = pos.PositionKey

	if (move & data.MFLAGEP) != 0 {
		if side == data.White {
			ClearPiece(to-10, pos)
		} else {
			ClearPiece(to+10, pos)
		}
	} else if (move & data.MFLAGGCA) != 0 {
		switch to {
		case data.C1:
			MovePiece(data.A1, data.D1, pos)
		case data.C8:
			MovePiece(data.A8, data.D8, pos)
		case data.G1:
			MovePiece(data.H1, data.F1, pos)
		case data.G8:
			MovePiece(data.H8, data.F8, pos)
		default:
			panic(fmt.Errorf("MakeMove: castle error %v %v", from, to))
		}
	}

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}

	hashCastle(pos)

	pos.History[pos.HistoryPlay].Move = move
	pos.History[pos.HistoryPlay].FiftyMove = pos.FiftyMove
	pos.History[pos.HistoryPlay].EnPas = pos.EnPas
	pos.History[pos.HistoryPlay].CastlePermission = pos.CastlePermission

	pos.CastlePermission &= data.CastlePerm[from]
	pos.CastlePermission &= data.CastlePerm[to]
	pos.EnPas = data.NoSquare
	hashCastle(pos)

	captured := data.Captured(move)
	pos.FiftyMove++

	if captured != data.Empty {
		ClearPiece(to, pos)
		pos.FiftyMove = 0
	}

	pos.HistoryPlay++
	pos.Play++

	if data.PiecePawn[pos.Pieces[from]] == data.True {
		pos.FiftyMove = 0
		if (move & data.MFLAGPS) != 0 {
			if side == data.White {
				pos.EnPas = from + 10
			} else {
				pos.EnPas = from - 10
			}
			hashEnPas(pos)
		}
	}

	MovePiece(from, to, pos)

	promotedPiece := data.Promoted(move)
	if promotedPiece != 0 {

		ClearPiece(to, pos)
		AddPiece(to, promotedPiece, pos)
	}

	if data.PieceKing[pos.Pieces[to]] == data.True {
		pos.KingSquare[pos.Side] = to
	}

	pos.Side ^= 1

	hashSide(pos)

	if attack.SquareAttacked(pos.KingSquare[side], pos.Side, pos) {
		TakeMoveBack(pos)
		return false
	}

	return true
}

func TakeMoveBack(pos *data.Board) {
	board.CheckBoard(pos)

	pos.HistoryPlay--
	pos.Play--

	move := pos.History[pos.HistoryPlay].Move
	from := data.FromSquare(move)
	to := data.ToSquare(move)

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}

	hashCastle(pos)
	pos.CastlePermission = pos.History[pos.HistoryPlay].CastlePermission
	pos.FiftyMove = pos.History[pos.HistoryPlay].FiftyMove
	pos.EnPas = pos.History[pos.HistoryPlay].EnPas

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}
	hashCastle(pos)

	pos.Side ^= 1
	hashSide(pos)

	if (move & data.MFLAGEP) != 0 {
		if pos.Side == data.White {
			AddPiece(to-10, data.BP, pos)
		} else {
			AddPiece(to+10, data.WP, pos)
		}
	} else if (move & data.MFLAGGCA) != 0 {
		switch to {
		case data.C1:
			MovePiece(data.D1, data.A1, pos)
		case data.C8:
			MovePiece(data.D8, data.A8, pos)
		case data.G1:
			MovePiece(data.F1, data.H1, pos)
		case data.G8:
			MovePiece(data.F8, data.H8, pos)
		default:
			panic(fmt.Errorf("TakeMoveBack: castle error %v %v", from, to))
		}
	}

	MovePiece(to, from, pos)

	if data.PieceKing[pos.Pieces[from]] == data.True {
		pos.KingSquare[pos.Side] = from
	}

	captured := data.Captured(move)
	if captured != data.Empty {
		AddPiece(to, captured, pos)
	}

	if data.Promoted(move) != data.Empty {
		if data.PiecePawn[data.Promoted(move)] == data.True {
			panic(fmt.Errorf("MakeMove: can't promote to a pawn"))
		}
		ClearPiece(from, pos)
		col := data.PieceCol[data.Promoted(move)]
		if col == data.White {
			AddPiece(from, data.WP, pos)
		} else {
			AddPiece(from, data.BP, pos)
		}
	}
}

func MovePiece(from, to int, pos *data.Board) {
	//TODO Validate the move
	piece := pos.Pieces[from]

	hashPiece(piece, from, pos)
	pos.Pieces[from] = data.Empty
	hashPiece(piece, to, pos)
	pos.Pieces[to] = piece

	board.SetPosPieceData(pos, to, piece)
	board.ClearPosPieceData(pos, from, piece)
}

func AddPiece(sq, piece int, pos *data.Board) {
	//TODO Validate the move

	hashPiece(piece, sq, pos)

	pos.Pieces[sq] = piece

	board.SetPosPieceData(pos, sq, piece)
}

func ClearPiece(sq int, pos *data.Board) {
	//TODO Validate the move

	piece := pos.Pieces[sq]
	hashPiece(piece, sq, pos)

	pos.Pieces[sq] = data.Empty

	board.ClearPosPieceData(pos, sq, piece)
}

func MakeNullMove(pos *data.Board) {
	board.CheckBoard(pos)
	pos.Play++
	pos.History[pos.HistoryPlay].PositionKey = pos.PositionKey

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}

	pos.History[pos.HistoryPlay].Move = data.NoMove
	pos.History[pos.HistoryPlay].FiftyMove = pos.FiftyMove
	pos.History[pos.HistoryPlay].EnPas = pos.EnPas
	pos.History[pos.HistoryPlay].CastlePermission = pos.CastlePermission
	pos.EnPas = data.NoSquare

	pos.Side ^= 1
	pos.HistoryPlay++
	hashSide(pos)
}

func TakeBackNullMove(pos *data.Board) {
	board.CheckBoard(pos)

	pos.Play--
	pos.HistoryPlay--

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}

	pos.CastlePermission = pos.History[pos.HistoryPlay].CastlePermission
	pos.FiftyMove = pos.History[pos.HistoryPlay].FiftyMove
	pos.EnPas = pos.History[pos.HistoryPlay].EnPas

	if pos.EnPas != data.NoSquare {
		hashEnPas(pos)
	}

	pos.Side ^= 1
	hashSide(pos)
}

func ParseMove(move []byte, pos *data.Board) int {
	if move[1] > '8' || move[1] < '1' {
		return data.NoMove
	}
	if move[3] > '8' || move[3] < '1' {
		return data.NoMove
	}
	if move[0] > 'h' || move[0] < 'a' {
		return data.NoMove
	}
	if move[2] > 'h' || move[2] < 'a' {
		return data.NoMove
	}

	from := data.FileRankToSquare(int(move[0]-'a'), int(move[1]-'1'))
	to := data.FileRankToSquare(int(move[2]-'a'), int(move[3]-'1'))

	ml := &data.MoveList{}
	GenerateAllMoves(pos, ml)

	for MoveNum := 0; MoveNum < ml.Count; MoveNum++ {
		userMove := ml.Moves[MoveNum].Move
		if data.FromSquare(userMove) == from && data.ToSquare(userMove) == to {
			promPce := data.Promoted(userMove)
			if promPce != data.Empty {
				if data.PieceRookQueen[promPce] == data.True && data.PieceBishopQueen[promPce] == data.False && move[4] == 'r' {
					return userMove
				} else if data.PieceRookQueen[promPce] == data.False && data.PieceBishopQueen[promPce] == data.True && move[4] == 'b' {
					return userMove
				} else if data.PieceRookQueen[promPce] == data.True && data.PieceBishopQueen[promPce] == data.True && move[4] == 'q' {
					return userMove
				} else if data.PieceKnight[promPce] == data.True && move[4] == 'n' {
					return userMove
				}
				continue
			}
			return userMove
		}
	}

	return data.NoMove
}

func hashPiece(piece, square int, pos *data.Board) {
	pos.PositionKey ^= data.PieceKeys[piece][square]
}

func hashCastle(pos *data.Board) {
	pos.PositionKey ^= data.CastleKeys[pos.CastlePermission]
}

func hashSide(pos *data.Board) {
	pos.PositionKey ^= data.SideKey
}

func hashEnPas(pos *data.Board) {
	pos.PositionKey ^= data.PieceKeys[data.Empty][pos.EnPas]
}
