package moveGen

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func TestGenerateRookMoves(t *testing.T) {
	pos := data.NewBoardPos()
	board.ParseFEN(data.StartFEN, pos)
	moveList := &data.MoveList{}
	generateRookMoves(pos, moveList, true)
	if moveList.Count != 110 {
		t.Errorf("got %d, want %d", moveList.Count, 0)
	}
	PrintMoveList(moveList)
	fmt.Printf("\n\nCount %v\n", moveList.Count)

}
