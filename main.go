package main

import "github.com/AdamGriffiths31/ChessEngine/engine"

func main() {

	// b := &engine.Board{}
	// engine.ParseFEN(engine.StartFEN, b)

	// engine.CheckBoard(b)
	// engine.PrintBoard(b)

	// ml := &engine.MoveList{}
	// engine.GenerateAllMoves(b, ml)
	// engine.PrintMoveList(ml)

	// for moveNum := 0; moveNum < ml.Count; moveNum++ {
	// 	move := ml.Moves[moveNum].Move
	// 	fmt.Printf("Move = %v\n", engine.PrintMove(move))
	// 	if !engine.MakeMove(move, b) {
	// 		continue
	// 	}
	// 	engine.PrintBoard(b)
	// 	fmt.Printf("Made %s \n", engine.PrintMove(move))
	// 	engine.TakeMoveBack(b)
	// 	fmt.Printf("Take back %s \n", engine.PrintMove(move))
	// 	engine.PrintBoard(b)
	// }

	engine.PerftTest(4, "n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1")
}
