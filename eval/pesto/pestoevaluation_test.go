package eval

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func TestFlipA1(t *testing.T) {
	result := flip(0) //A1

	if result != data.Square120ToSquare64[data.A8] {
		t.Errorf("Expected %v but got %v", io.SquareString(data.A8), io.SquareString(data.Square64ToSquare120[result]))
	}
}

func TestFlipH8(t *testing.T) {
	result := flip(63) //H8

	if result != data.Square120ToSquare64[data.H1] {
		t.Errorf("Expected %v but got %v", io.SquareString(data.H1), io.SquareString(data.Square64ToSquare120[result]))
	}
}

func TestCombineBishops(t *testing.T) {
	game := engine.ParseFen(data.StartFEN)
	combinedPosition := combineBishops(game.Position().Board.BlackBishop, game.Position().Board.WhiteBishop)

	if game.Position().Board.CountBits(combinedPosition) != 4 {
		t.Errorf("Expected 4 bishops but got %v", game.Position().Board.CountBits(combinedPosition))
	}
}

func TestWeights(t *testing.T) {
	var w = &Weights{}
	w.init()

	printPst("Pawn", w.PST[data.White][data.WP-1])
	printPst("Knight", w.PST[data.White][data.WN-1])
	printPst("Bishop", w.PST[data.White][data.WB-1])
	printPst("Rook", w.PST[data.White][data.WR-1])
	printPst("Queen", w.PST[data.White][data.WQ-1])
	printPst("King", w.PST[data.White][data.WK-1])

	//t.Fatal("Test failed")
}

func printPst(name string, source [64]Score) {
	fmt.Println("PST", name)
	for i := 0; i < 64; i++ {
		sq := flip(i)
		fmt.Printf("%+v", source[sq])
	}
	fmt.Println()

}
