package io

import (
	"fmt"

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

func PrintBoard(pos *data.Board) {
	fmt.Println("Printing board...")

	// Print the board grid
	for rank := data.Rank8; rank >= data.Rank1; rank-- {
		fmt.Printf("%d ", rank+1)
		for file := data.FileA; file <= data.FileH; file++ {
			sq := data.FileRankToSquare(file, rank)
			piece := pos.Pieces[sq]
			fmt.Printf("%3s", data.PceChar[piece])
		}
		fmt.Println()
	}

	// Print file labels
	fmt.Print("  ")
	for file := data.FileA; file <= data.FileH; file++ {
		fmt.Printf("%3c", 'a'+file)
	}
	fmt.Println()

	// Print side, en passant, and position key information
	fmt.Printf("Side: %s\n", data.SideChar[pos.Side])
	fmt.Printf("EnPas: %s\n", SquareString(pos.EnPas))
	fmt.Printf("PosKey: %11X (%v)\n", pos.PositionKey, pos.PositionKey)
}
