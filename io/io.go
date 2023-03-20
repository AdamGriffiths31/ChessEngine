package io

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func SquareString(square int) string {
	if !validate.SquareOnBoard(square) {
		return "Square not on board"
	}
	file := data.FilesBoard[square] + 'a'
	rank := data.RanksBoard[square] + '1'
	return string([]byte{byte(file), byte(rank)})
}

func PrintMove(move int) string {
	if move == data.NoMove {
		return "NoMove"
	}
	fromFile := data.FilesBoard[data.FromSquare(move)] + 'a'
	fromRank := data.RanksBoard[data.FromSquare(move)] + '1'
	toFile := data.FilesBoard[data.ToSquare(move)] + 'a'
	toRank := data.RanksBoard[data.ToSquare(move)] + '1'

	promotedStr := "q"
	promoted := data.Promoted(move)
	if promoted != 0 {
		if data.PieceKnight[promoted] == data.True {
			promotedStr = "n"
		} else if data.PieceRookQueen[promoted] == data.True && data.PieceBishopQueen[promoted] == data.False {
			promotedStr = "r"
		} else if data.PieceRookQueen[promoted] == data.False && data.PieceBishopQueen[promoted] == data.True {
			promotedStr = "b"
		}
		return string([]byte{byte(fromFile), byte(fromRank), byte(toFile), byte(toRank)}) + promotedStr
	}

	return string([]byte{byte(fromFile), byte(fromRank), byte(toFile), byte(toRank)})
}
