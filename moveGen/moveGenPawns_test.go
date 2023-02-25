package moveGen

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func TestGenerateWhitePawnQuietMoves(t *testing.T) {
	pos := data.NewBoardPos()
	board.ParseFEN(data.StartFEN, pos)
	moveList := &data.MoveList{}
	generateBlackPawnQuietMoves(pos, moveList)
	if moveList.Count != 110 {
		t.Errorf("got %d, want %d", moveList.Count, 0)
	}
	PrintMoveList(moveList)
	fmt.Printf("\n\nCount %v\n", moveList.Count)

}

// func BenchmarkGenerateWhitePawnQuietMove1(b *testing.B) {
// 	pos := data.NewBoardPos()
// 	board.ParseFEN(data.StartFEN, pos)

// 	for n := 0; n < 300000000; n++ {
// 		moveList := &data.MoveList{}
// 		generateWhitePawnCaptureMoves(pos, moveList)
// 	}
// }

// func BenchmarkGenerateWhitePawnQuietMove(b *testing.B) {
// 	pos := data.NewBoardPos()
// 	board.ParseFEN(data.StartFEN, pos)

// 	for n := 0; n < 300000000; n++ {
// 		moveList := &data.MoveList{}
// 		generateWhitePawnCaptureMoves(pos, moveList)

// 	}
// }
