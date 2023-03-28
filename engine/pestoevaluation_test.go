package engine

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
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
