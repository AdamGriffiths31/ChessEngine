package main

import (
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func main() {

	b := &engine.Board{}
	engine.ParseFEN(engine.StartFEN, b)

	engine.CheckBoard(b)
	engine.PrintBoard(b)

	ml := &engine.MoveList{}
	engine.GenerateAllMoves(b, ml)
	engine.PrintMoveList(ml)

}
