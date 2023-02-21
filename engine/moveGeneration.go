package engine

import (
	"fmt"
)

func PrintMoveList(moveList *MoveList) {
	fmt.Println("\nPrinting move list:")
	for i := 0; i < moveList.Count; i++ {
		fmt.Printf("Move %v: %v (score: %v)\n", i+1, PrintMove(moveList.Moves[i].Move), moveList.Moves[i].Score)
	}
	fmt.Printf("Printed %v total moves.\n", moveList.Count)
}

// MakeMoveInt Builds the move int
func MakeMoveInt(f, t, ca, pro, fl int) int {
	return f | t<<7 | ca<<14 | pro<<20 | fl
}

func GenerateAllMoves(pos *Board, moveList *MoveList) {
	CheckBoard(pos)
	moveList.Count = 0

	if pos.Side == White {
		for pieceNum := 0; pieceNum < pos.PieceNumber[WP]; pieceNum++ {
			sq := pos.PieceList[WP][pieceNum]
			if !SqaureOnBoard(sq) {
				panic(fmt.Errorf("GenerateAllMoves: white pawn number %v was at square %v", pieceNum, sq))
			}
			if pos.Pieces[sq+10] == Empty {
				addWhitePawnMove(pos, sq, sq+10, moveList)
				if RanksBoard[sq] == Rank2 && pos.Pieces[sq+20] == Empty {
					addQuiteMove(pos, MakeMoveInt(sq, sq+20, Empty, Empty, MFLAGPS), moveList)
				}
			}

			if SqaureOnBoard(sq+9) && PieceCol[pos.Pieces[sq+9]] == Black {
				addWhitePawnCaptureMove(pos, moveList, sq, sq+9, pos.Pieces[sq+9])
			}

			if SqaureOnBoard(sq+11) && PieceCol[pos.Pieces[sq+11]] == Black {
				addWhitePawnCaptureMove(pos, moveList, sq, sq+11, pos.Pieces[sq+11])
			}

			if sq+9 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq+9, Empty, Empty, MFLAGEP), moveList)
			}

			if sq+11 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq+11, Empty, Empty, MFLAGEP), moveList)
			}
		}

		if (pos.CastlePermission & WhiteKingCastle) != 0 {
			if pos.Pieces[F1] == Empty && pos.Pieces[G1] == Empty {
				if !SquareAttacked(E1, Black, pos) && !SquareAttacked(F1, Black, pos) {
					addQuiteMove(pos, MakeMoveInt(E1, G1, Empty, Empty, MFLAGGCA), moveList)
				}
			}
		}
		if (pos.CastlePermission & WhiteQueenCastle) != 0 {
			if pos.Pieces[D1] == Empty && pos.Pieces[C1] == Empty && pos.Pieces[B1] == Empty {
				if !SquareAttacked(E1, Black, pos) && !SquareAttacked(D1, Black, pos) {
					addQuiteMove(pos, MakeMoveInt(E1, C1, Empty, Empty, MFLAGGCA), moveList)
				}
			}
		}
	} else {
		for pieceNum := 0; pieceNum < pos.PieceNumber[BP]; pieceNum++ {
			sq := pos.PieceList[BP][pieceNum]
			if !SqaureOnBoard(sq) {
				panic(fmt.Errorf("GenerateAllMoves: black pawn number %v was at square %v", pieceNum, sq))
			}
			if pos.Pieces[sq-10] == Empty {
				addBlackPawnMove(pos, sq, sq-10, moveList)
				if RanksBoard[sq] == Rank7 && pos.Pieces[sq-20] == Empty {
					addQuiteMove(pos, MakeMoveInt(sq, sq-20, Empty, Empty, MFLAGPS), moveList)
				}
			}
			if SqaureOnBoard(sq-9) && PieceCol[pos.Pieces[sq-9]] == White {
				addBlackPawnCaptureMove(pos, moveList, sq, sq-9, pos.Pieces[sq-9])
			}

			if SqaureOnBoard(sq-11) && PieceCol[pos.Pieces[sq-11]] == White {
				addBlackPawnCaptureMove(pos, moveList, sq, sq-11, pos.Pieces[sq-11])
			}

			if sq-9 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq-9, Empty, Empty, MFLAGEP), moveList)
			}

			if sq-11 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq-11, Empty, Empty, MFLAGEP), moveList)
			}
		}
		if (pos.CastlePermission & BlackKingCastle) != 0 {
			if pos.Pieces[F8] == Empty && pos.Pieces[G8] == Empty {
				if !SquareAttacked(E8, White, pos) && !SquareAttacked(F8, White, pos) {
					addQuiteMove(pos, MakeMoveInt(E8, G8, Empty, Empty, MFLAGGCA), moveList)
				}
			}
		}
		if (pos.CastlePermission & BlackQueenCastle) != 0 {
			if pos.Pieces[D8] == Empty && pos.Pieces[C8] == Empty && pos.Pieces[B8] == Empty {
				if !SquareAttacked(E8, White, pos) && !SquareAttacked(D8, White, pos) {
					addQuiteMove(pos, MakeMoveInt(E8, C8, Empty, Empty, MFLAGGCA), moveList)
				}
			}
		}
	}

	pieceIndex := LoopSlideIndex[pos.Side]
	piece := LoopSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]

			for i := 0; i < NumDir[piece]; i++ {
				dir := PieceDir[piece][i]
				tempSq := sq + dir

				for SqaureOnBoard(tempSq) {
					if pos.Pieces[tempSq] != Empty {
						if PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
							addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], Empty, 0), moveList)
						}
						break
					}
					addQuiteMove(pos, MakeMoveInt(sq, tempSq, Empty, Empty, 0), moveList)
					tempSq += dir
				}
			}
		}
		piece = LoopSlidePiece[pieceIndex]
		pieceIndex++
	}

	pieceIndex = LoopNonSlideIndex[pos.Side]
	piece = LoopNonSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]

			for i := 0; i < NumDir[piece]; i++ {
				dir := PieceDir[piece][i]
				tempSq := sq + dir

				if !SqaureOnBoard(tempSq) {
					continue
				}

				if pos.Pieces[tempSq] != Empty {
					if PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
						addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], Empty, 0), moveList)
					}
					continue
				}
				addQuiteMove(pos, MakeMoveInt(sq, tempSq, Empty, Empty, 0), moveList)
			}
		}
		piece = LoopNonSlidePiece[pieceIndex]
		pieceIndex++
	}

}

func addQuiteMove(pos *Board, move int, moveList *MoveList) {
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 0
	moveList.Count++
}

func addCaptureMove(pos *Board, move int, moveList *MoveList) {
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 0
	moveList.Count++
}

func addEnPasMove(pos *Board, move int, moveList *MoveList) {
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 0
	moveList.Count++
}

func addWhitePawnCaptureMove(pos *Board, moveList *MoveList, from, to, cap int) {
	if RanksBoard[from] == Rank7 {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, WQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, WR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, WB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, WN, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, Empty, 0), moveList)
	}
}

func addWhitePawnMove(pos *Board, from, to int, moveList *MoveList) {
	if RanksBoard[from] == Rank7 {
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, WQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, WR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, WB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, WN, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, Empty, 0), moveList)
	}
}

func addBlackPawnCaptureMove(pos *Board, moveList *MoveList, from, to, cap int) {
	if RanksBoard[from] == Rank2 {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, BQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, BR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, BB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, B2, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, Empty, 0), moveList)
	}
}

func addBlackPawnMove(pos *Board, from, to int, moveList *MoveList) {
	if RanksBoard[from] == Rank2 {
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, BQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, BR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, BB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, BN, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, Empty, Empty, 0), moveList)
	}
}
