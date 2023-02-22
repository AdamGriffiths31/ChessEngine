package engine

import (
	"fmt"
)

func MakeMove(move int, pos *Board) bool {
	//TODO Validate the move
	CheckBoard(pos)

	from := FromSquare(move)
	to := ToSqaure(move)
	side := pos.Side

	if !SqaureOnBoard(from) {
		panic("Err")
	}
	if !SqaureOnBoard(to) {
		panic("Err")
	}

	if side != White && side != Black {
		panic("Err")
	}

	if !PieceValid(pos.Pieces[from]) {
		panic("Err")
	}

	if pos.Play < 0 {
		panic("Err")
	}

	if pos.HistoryPlay < 0 {
		panic("Err")
	}

	if pos.Play > MaxDepth {
		panic("Err")
	}

	pos.History[pos.HistoryPlay].PosistionKey = pos.PosistionKey

	if (move & MFLAGEP) != 0 {
		if side == White {
			ClearPiece(to-10, pos)
		} else {
			ClearPiece(to+10, pos)
		}
	} else if (move & MFLAGGCA) != 0 {
		switch to {
		case C1:
			MovePiece(A1, D1, pos)
		case C8:
			MovePiece(A8, D8, pos)
		case G1:
			MovePiece(H1, F1, pos)
		case G8:
			MovePiece(H8, F8, pos)
		default:
			panic(fmt.Errorf("MakeMove: castle error %v %v", from, to))
		}
	}

	if pos.EnPas != noSqaure {
		hashEnPas(pos)
	}

	hashCastle(pos)

	pos.History[pos.HistoryPlay].Move = move
	pos.History[pos.HistoryPlay].FiftyMove = pos.FiftyMove
	pos.History[pos.HistoryPlay].EnPas = pos.EnPas
	pos.History[pos.HistoryPlay].CastlePermission = pos.CastlePermission

	pos.CastlePermission &= CastlePerm[from]
	pos.CastlePermission &= CastlePerm[to]
	pos.EnPas = noSqaure
	hashCastle(pos)

	captured := Captured(move)
	pos.FiftyMove++

	if captured != Empty {
		ClearPiece(to, pos)
		pos.FiftyMove = 0
	}

	pos.HistoryPlay++
	pos.Play++

	if PiecePawn[pos.Pieces[from]] == True {
		pos.FiftyMove = 0
		if (move & MFLAGPS) != 0 {
			if side == White {
				pos.EnPas = from + 10
				if RanksBoard[pos.EnPas] != Rank3 {
					panic(fmt.Errorf("MakeMove: white pawn enPas error %v", pos.EnPas))
				}
			} else {
				pos.EnPas = from - 10
				if RanksBoard[pos.EnPas] != Rank6 {
					panic(fmt.Errorf("MakeMove: white pawn enPas error %v", pos.EnPas))
				}
			}
			hashEnPas(pos)
		}
	}

	MovePiece(from, to, pos)

	promotedPiece := Promoted(move)
	if promotedPiece != 0 {
		if PiecePawn[promotedPiece] == True {
			panic(fmt.Errorf("MakeMove: can't promote to a pawn"))
		}
		ClearPiece(to, pos)
		AddPiece(to, promotedPiece, pos)
	}

	if PieceKing[pos.Pieces[to]] == True {
		pos.KingSqaure[pos.Side] = to
	}

	pos.Side ^= 1

	hashSide(pos)

	if SquareAttacked(pos.KingSqaure[side], pos.Side, pos) {
		TakeMoveBack(pos)
		return false
	}

	return true
}

func TakeMoveBack(pos *Board) {
	CheckBoard(pos)

	pos.HistoryPlay--
	pos.Play--

	move := pos.History[pos.HistoryPlay].Move
	from := FromSquare(move)
	to := ToSqaure(move)

	if !SqaureOnBoard(from) {
		panic("Err")
	}
	if !SqaureOnBoard(to) {
		panic("Err")
	}

	if pos.EnPas != noSqaure {
		hashEnPas(pos)
	}

	hashCastle(pos)
	pos.CastlePermission = pos.History[pos.HistoryPlay].CastlePermission
	pos.FiftyMove = pos.History[pos.HistoryPlay].FiftyMove
	pos.EnPas = pos.History[pos.HistoryPlay].EnPas

	if pos.EnPas != noSqaure {
		hashEnPas(pos)
	}
	hashCastle(pos)

	pos.Side ^= 1
	hashSide(pos)

	if (move & MFLAGEP) != 0 {
		if pos.Side == White {
			AddPiece(to-10, BP, pos)
		} else {
			AddPiece(to+10, WP, pos)
		}
	} else if (move & MFLAGGCA) != 0 {
		switch to {
		case C1:
			MovePiece(D1, A1, pos)
		case C8:
			MovePiece(D8, A8, pos)
		case G1:
			MovePiece(F1, H1, pos)
		case G8:
			MovePiece(F8, H8, pos)
		default:
			panic(fmt.Errorf("TakeMoveBack: castle error %v %v", from, to))
		}
	}

	MovePiece(to, from, pos)

	if PieceKing[pos.Pieces[from]] == True {
		pos.KingSqaure[pos.Side] = from
	}

	captured := Captured(move)
	if captured != Empty {
		AddPiece(to, captured, pos)
	}

	if Promoted(move) != Empty {
		if PiecePawn[Promoted(move)] == True {
			panic(fmt.Errorf("MakeMove: can't promote to a pawn"))
		}
		ClearPiece(from, pos)
		col := PieceCol[Promoted(move)]
		if col == White {
			AddPiece(from, WP, pos)
		} else {
			AddPiece(from, BP, pos)
		}
	}
}

func MovePiece(from, to int, pos *Board) {
	//TODO Validate the move
	piece := pos.Pieces[from]
	col := PieceCol[piece]

	hashPiece(piece, from, pos)
	pos.Pieces[from] = Empty
	hashPiece(piece, to, pos)
	pos.Pieces[to] = piece

	if PieceBig[piece] == False {
		ClearBit(&pos.Pawns[col], Sqaure120ToSquare64[from])
		ClearBit(&pos.Pawns[Both], Sqaure120ToSquare64[from])
		SetBit(&pos.Pawns[col], Sqaure120ToSquare64[to])
		SetBit(&pos.Pawns[Both], Sqaure120ToSquare64[to])
	}

	found := false
	for i := 0; i < pos.PieceNumber[piece]; i++ {
		if pos.PieceList[piece][i] == from {
			pos.PieceList[piece][i] = to
			found = true
			break
		}
	}

	if !found {
		PrintBoard(pos)
		panic(fmt.Errorf("MovePiece: piece not found at %v [%v] going to %v [%v]", from, SqaureString(from), to, SqaureString(to)))
	}
}

func AddPiece(sq, piece int, pos *Board) {
	//TODO Validate the move

	col := PieceCol[piece]
	hashPiece(piece, sq, pos)

	pos.Pieces[sq] = piece

	if PieceBig[piece] == True {
		pos.BigPiece[col]++
		if PieceMajor[piece] == True {
			pos.MajorPiece[col]++
		} else if PieceMin[piece] == True {
			pos.MinPiece[col]++
		} else {
			panic(fmt.Errorf("AddPiece: PieceBig error for %v", piece))
		}
	} else {
		SetBit(&pos.Pawns[col], Sqaure120ToSquare64[sq])
		SetBit(&pos.Pawns[Both], Sqaure120ToSquare64[sq])
	}

	pos.Material[col] += PieceVal[piece]
	pos.PieceList[piece][pos.PieceNumber[piece]] = sq
	pos.PieceNumber[piece]++
}

func ClearPiece(sq int, pos *Board) {
	//TODO Validate the move

	piece := pos.Pieces[sq]
	col := PieceCol[piece]
	hashPiece(piece, sq, pos)

	pos.Pieces[sq] = Empty
	pos.Material[col] -= PieceVal[piece]

	if PieceBig[piece] == True {
		pos.BigPiece[col]--
		if PieceMajor[piece] == True {
			pos.MajorPiece[col]--
		} else if PieceMin[piece] == True {
			pos.MinPiece[col]--
		} else {
			panic(fmt.Errorf("AddPiece: PieceBig error for %v", piece))
		}
	} else {
		ClearBit(&pos.Pawns[col], Sqaure120ToSquare64[sq])
		ClearBit(&pos.Pawns[Both], Sqaure120ToSquare64[sq])
	}

	tempPiece := -1
	for i := 0; i < pos.PieceNumber[piece]; i++ {
		if pos.PieceList[piece][i] == sq {
			tempPiece = i
			break
		}
	}

	if tempPiece == -1 {
		PrintBoard(pos)
		panic(fmt.Errorf("ClearPiece: could not find piece [%v] at sq %v [%v]", Pieces[piece], sq, SqaureString(sq)))
	}

	pos.PieceNumber[piece]--
	pos.PieceList[piece][tempPiece] = pos.PieceList[piece][pos.PieceNumber[piece]]
}

func hashPiece(piece, square int, pos *Board) {
	pos.PosistionKey ^= PieceKeys[piece][square]
}

func hashCastle(pos *Board) {
	pos.PosistionKey ^= CastleKeys[pos.CastlePermission]
}

func hashSide(pos *Board) {
	pos.PosistionKey ^= SideKey
}

func hashEnPas(pos *Board) {
	pos.PosistionKey ^= PieceKeys[Empty][pos.EnPas]
}
