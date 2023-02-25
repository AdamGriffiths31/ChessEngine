package moveGen

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
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
	generateSliderMoves(pos, moveList, true)
	generateNonSliderMoves(pos, moveList, true)
}

func GenerateAllCaptures(pos *data.Board, moveList *data.MoveList) {
	board.CheckBoard(pos)
	moveList.Count = 0

	if pos.Side == data.White {
		generateWhitePawnEnPassantMoves(pos, moveList)
		generateWhitePawnCaptureMoves(pos, moveList)

	} else {
		generateBlackPawnCaptureMoves(pos, moveList)
		generateBlackPawnEnPassantMoves(pos, moveList)
	}

	generateSliderMoves(pos, moveList, false)
	generateNonSliderMoves(pos, moveList, false)
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

	switch move {
	case pos.SearchKillers[0][pos.Play]:
		moveList.Moves[moveList.Count].Score = 900000
	case pos.SearchKillers[1][pos.Play]:
		moveList.Moves[moveList.Count].Score = 800000
	default:
		fromSq := data.FromSquare(move)
		toSq := data.ToSqaure(move)
		moveList.Moves[moveList.Count].Score = pos.SearchHistory[pos.Pieces[fromSq]][pos.Pieces[toSq]]
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
