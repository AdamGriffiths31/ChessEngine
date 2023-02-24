package movegen

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func PrintMoveList(moveList *data.MoveList) {
	fmt.Println("\nPrinting move list:")
	for i := 0; i < moveList.Count; i++ {
		fmt.Printf("Move %v: %v (score: %v)\n", i+1, io.PrintMove(moveList.Moves[i].Move), moveList.Moves[i].Score)
	}
	fmt.Printf("Printed %v total moves.\n", moveList.Count)
}

// MakeMoveInt Builds the move int
func MakeMoveInt(f, t, ca, pro, fl int) int {
	return f | t<<7 | ca<<14 | pro<<20 | fl
}

func GenerateAllMoves(pos *data.Board, moveList *data.MoveList) {
	board.CheckBoard(pos)
	moveList.Count = 0

	if pos.Side == data.White {
		generateWhitePawnMoves(pos, moveList)
		generateWhiteCastleMoves(pos, moveList)
	} else {
		generateBlackPawnMoves(pos, moveList)
		generateBlackCastleMoves(pos, moveList)
	}
	generateSliderMoves(pos, moveList)
	generateNonSliderMoves(pos, moveList)
}

func GenerateAllCaptures(pos *data.Board, moveList *data.MoveList) {
	board.CheckBoard(pos)
	moveList.Count = 0

	if pos.Side == data.White {
		for pieceNum := 0; pieceNum < pos.PieceNumber[data.WP]; pieceNum++ {
			sq := pos.PieceList[data.WP][pieceNum]

			if validate.SqaureOnBoard(sq+9) && data.PieceCol[pos.Pieces[sq+9]] == data.Black {
				addWhitePawnCaptureMove(pos, moveList, sq, sq+9, pos.Pieces[sq+9])
			}

			if validate.SqaureOnBoard(sq+11) && data.PieceCol[pos.Pieces[sq+11]] == data.Black {
				addWhitePawnCaptureMove(pos, moveList, sq, sq+11, pos.Pieces[sq+11])
			}
			if pos.EnPas != data.NoSqaure {
				if sq+9 == pos.EnPas {
					addEnPasMove(pos, MakeMoveInt(sq, sq+9, data.Empty, data.Empty, data.MFLAGEP), moveList)
				}

				if sq+11 == pos.EnPas {
					addEnPasMove(pos, MakeMoveInt(sq, sq+11, data.Empty, data.Empty, data.MFLAGEP), moveList)
				}
			}
		}

	} else {
		for pieceNum := 0; pieceNum < pos.PieceNumber[data.BP]; pieceNum++ {
			sq := pos.PieceList[data.BP][pieceNum]

			if validate.SqaureOnBoard(sq-9) && data.PieceCol[pos.Pieces[sq-9]] == data.White {
				addBlackPawnCaptureMove(pos, moveList, sq, sq-9, pos.Pieces[sq-9])
			}

			if validate.SqaureOnBoard(sq-11) && data.PieceCol[pos.Pieces[sq-11]] == data.White {
				addBlackPawnCaptureMove(pos, moveList, sq, sq-11, pos.Pieces[sq-11])
			}

			if pos.EnPas != data.NoSqaure {
				if sq-9 == pos.EnPas {
					addEnPasMove(pos, MakeMoveInt(sq, sq-9, data.Empty, data.Empty, data.MFLAGEP), moveList)
				}

				if sq-11 == pos.EnPas {
					addEnPasMove(pos, MakeMoveInt(sq, sq-11, data.Empty, data.Empty, data.MFLAGEP), moveList)
				}
			}
		}
	}

	pieceIndex := data.LoopSlideIndex[pos.Side]
	piece := data.LoopSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				for validate.SqaureOnBoard(tempSq) {
					if pos.Pieces[tempSq] != data.Empty {
						if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
							addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
						}
						break
					}
					tempSq += dir
				}
			}
		}
		piece = data.LoopSlidePiece[pieceIndex]
		pieceIndex++
	}

	pieceIndex = data.LoopNonSlideIndex[pos.Side]
	piece = data.LoopNonSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				if !validate.SqaureOnBoard(tempSq) {
					continue
				}

				if pos.Pieces[tempSq] != data.Empty {
					if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
						addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
					}
					continue
				}
			}
		}
		piece = data.LoopNonSlidePiece[pieceIndex]
		pieceIndex++
	}
}

func MoveExists(pos *data.Board, move int) bool {
	ml := data.MoveList{}
	GenerateAllMoves(pos, &ml)

	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		if !MakeMove(ml.Moves[moveNum].Move, pos) {
			continue
		}
		TakeMoveBack(pos)
		if ml.Moves[moveNum].Move == move {
			return true
		}
	}

	return false
}

func addQuiteMove(pos *data.Board, move int, moveList *data.MoveList) {

	moveList.Moves[moveList.Count].Move = move

	if pos.SearchKillers[0][pos.Play] == move {
		moveList.Moves[moveList.Count].Score = 900000
	} else if pos.SearchKillers[1][pos.Play] == move {
		moveList.Moves[moveList.Count].Score = 800000
	} else {
		moveList.Moves[moveList.Count].Score = pos.SearchHistory[pos.Pieces[data.FromSquare(move)]][pos.Pieces[data.ToSqaure(move)]]
	}

	moveList.Count++
}

func addCaptureMove(pos *data.Board, move int, moveList *data.MoveList) {

	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = data.MvvLvaScores[data.Captured(move)][pos.Pieces[data.FromSquare(move)]] + 1000000
	moveList.Count++
}

func addEnPasMove(pos *data.Board, move int, moveList *data.MoveList) {

	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 105 + 1000000
	moveList.Count++
}

func addWhitePawnCaptureMove(pos *data.Board, moveList *data.MoveList, from, to, cap int) {
	if data.RanksBoard[from] == data.Rank7 {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.WQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.WR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.WB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.WN, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.Empty, 0), moveList)
	}
}

func addWhitePawnMove(pos *data.Board, from, to int, moveList *data.MoveList) {
	if data.RanksBoard[from] == data.Rank7 {
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.WQ, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.WR, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.WB, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.WN, 0), moveList)
	} else {
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.Empty, 0), moveList)
	}
}

func addBlackPawnCaptureMove(pos *data.Board, moveList *data.MoveList, from, to, cap int) {
	if data.RanksBoard[from] == data.Rank2 {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.BQ, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.BR, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.BB, 0), moveList)
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.BN, 0), moveList)
	} else {
		addCaptureMove(pos, MakeMoveInt(from, to, cap, data.Empty, 0), moveList)
	}
}

func addBlackPawnMove(pos *data.Board, from, to int, moveList *data.MoveList) {
	if data.RanksBoard[from] == data.Rank2 {
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.BQ, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.BR, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.BB, 0), moveList)
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.BN, 0), moveList)
	} else {
		addQuiteMove(pos, MakeMoveInt(from, to, data.Empty, data.Empty, 0), moveList)
	}
}

func generateWhitePawnMoves(pos *data.Board, moveList *data.MoveList) {
	for pieceNum := 0; pieceNum < pos.PieceNumber[data.WP]; pieceNum++ {
		sq := pos.PieceList[data.WP][pieceNum]
		if pos.Pieces[sq+10] == data.Empty {
			addWhitePawnMove(pos, sq, sq+10, moveList)
			if data.RanksBoard[sq] == data.Rank2 && pos.Pieces[sq+20] == data.Empty {
				addQuiteMove(pos, MakeMoveInt(sq, sq+20, data.Empty, data.Empty, data.MFLAGPS), moveList)
			}
		}

		if validate.SqaureOnBoard(sq+9) && data.PieceCol[pos.Pieces[sq+9]] == data.Black {
			addWhitePawnCaptureMove(pos, moveList, sq, sq+9, pos.Pieces[sq+9])
		}

		if validate.SqaureOnBoard(sq+11) && data.PieceCol[pos.Pieces[sq+11]] == data.Black {
			addWhitePawnCaptureMove(pos, moveList, sq, sq+11, pos.Pieces[sq+11])
		}
		if pos.EnPas != data.NoSqaure {
			if sq+9 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq+9, data.Empty, data.Empty, data.MFLAGEP), moveList)
			}

			if sq+11 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq+11, data.Empty, data.Empty, data.MFLAGEP), moveList)
			}
		}
	}
}

func generateWhiteCastleMoves(pos *data.Board, moveList *data.MoveList) {
	if (pos.CastlePermission & data.WhiteKingCastle) != 0 {
		if pos.Pieces[data.F1] == data.Empty && pos.Pieces[data.G1] == data.Empty {
			if !attack.SquareAttacked(data.E1, data.Black, pos) && !attack.SquareAttacked(data.F1, data.Black, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E1, data.G1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (pos.CastlePermission & data.WhiteQueenCastle) != 0 {
		if pos.Pieces[data.D1] == data.Empty && pos.Pieces[data.C1] == data.Empty && pos.Pieces[data.B1] == data.Empty {
			if !attack.SquareAttacked(data.E1, data.Black, pos) && !attack.SquareAttacked(data.D1, data.Black, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E1, data.C1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}

func generateBlackPawnMoves(pos *data.Board, moveList *data.MoveList) {
	for pieceNum := 0; pieceNum < pos.PieceNumber[data.BP]; pieceNum++ {
		sq := pos.PieceList[data.BP][pieceNum]
		if pos.Pieces[sq-10] == data.Empty {
			addBlackPawnMove(pos, sq, sq-10, moveList)
			if data.RanksBoard[sq] == data.Rank7 && pos.Pieces[sq-20] == data.Empty {
				addQuiteMove(pos, MakeMoveInt(sq, sq-20, data.Empty, data.Empty, data.MFLAGPS), moveList)
			}
		}
		if validate.SqaureOnBoard(sq-9) && data.PieceCol[pos.Pieces[sq-9]] == data.White {
			addBlackPawnCaptureMove(pos, moveList, sq, sq-9, pos.Pieces[sq-9])
		}

		if validate.SqaureOnBoard(sq-11) && data.PieceCol[pos.Pieces[sq-11]] == data.White {
			addBlackPawnCaptureMove(pos, moveList, sq, sq-11, pos.Pieces[sq-11])
		}

		if pos.EnPas != data.NoSqaure {
			if sq-9 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq-9, data.Empty, data.Empty, data.MFLAGEP), moveList)
			}

			if sq-11 == pos.EnPas {
				addEnPasMove(pos, MakeMoveInt(sq, sq-11, data.Empty, data.Empty, data.MFLAGEP), moveList)
			}
		}
	}
}

func generateBlackCastleMoves(pos *data.Board, moveList *data.MoveList) {
	if (pos.CastlePermission & data.BlackKingCastle) != 0 {
		if pos.Pieces[data.F8] == data.Empty && pos.Pieces[data.G8] == data.Empty {
			if !attack.SquareAttacked(data.E8, data.White, pos) && !attack.SquareAttacked(data.F8, data.White, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E8, data.G8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (pos.CastlePermission & data.BlackQueenCastle) != 0 {
		if pos.Pieces[data.D8] == data.Empty && pos.Pieces[data.C8] == data.Empty && pos.Pieces[data.B8] == data.Empty {
			if !attack.SquareAttacked(data.E8, data.White, pos) && !attack.SquareAttacked(data.D8, data.White, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E8, data.C8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}

func generateSliderMoves(pos *data.Board, moveList *data.MoveList) {
	pieceIndex := data.LoopSlideIndex[pos.Side]
	piece := data.LoopSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				for validate.SqaureOnBoard(tempSq) {
					if pos.Pieces[tempSq] != data.Empty {
						if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
							addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
						}
						break
					}
					addQuiteMove(pos, MakeMoveInt(sq, tempSq, data.Empty, data.Empty, 0), moveList)
					tempSq += dir
				}
			}
		}
		piece = data.LoopSlidePiece[pieceIndex]
		pieceIndex++
	}
}

func generateNonSliderMoves(pos *data.Board, moveList *data.MoveList) {
	pieceIndex := data.LoopNonSlideIndex[pos.Side]
	piece := data.LoopNonSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				if !validate.SqaureOnBoard(tempSq) {
					continue
				}

				if pos.Pieces[tempSq] != data.Empty {
					if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
						addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
					}
					continue
				}
				addQuiteMove(pos, MakeMoveInt(sq, tempSq, data.Empty, data.Empty, 0), moveList)
			}
		}
		piece = data.LoopNonSlidePiece[pieceIndex]
		pieceIndex++
	}
}
