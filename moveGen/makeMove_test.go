package moveGen

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func BenchmarkAddPiecee(b *testing.B) {
	pos := data.NewBoardPos()
	board.ParseFEN(data.StartFEN, pos)

	for n := 0; n < 1000000000; n++ {
		AddPiece(24, 1, pos)
		ClearPiece(24, pos)
	}
}
