package perft2

import (
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

var leafNodes int64 = 0

func Perft2(depth int, e engine.Engine) {
	if depth == 0 {
		leafNodes++
		return
	}

	// mlOld := data.MoveList{}
	// movegen.GenerateAllMoves(pos, &mlOld)

	ml := &engine.MoveList{}
	e.Position.GenerateAllMoves(ml)
	// for _, v := range mlOld.Moves {
	// 	found := false
	// 	for _, s := range ml.Moves {
	// 		if v.Move == s.Move {
	// 			found = true
	// 		}
	// 	}
	// 	if !found {
	// 		e.Position.PrintMoveList()
	// 		io.PrintBoard(pos)
	// 		e.Position.Board.PrintBitboard(e.Position.Board.BlackPawn)
	// 		fmt.Printf("%v was not found for new (%v-%v)\n", io.PrintMove(v.Move), pos.Side, e.Position.Side)
	// 		panic("err")
	// 	}
	// }
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		//allowed2 := movegen.MakeMove(ml.Moves[moveNum].Move, pos)
		allowed, enpas, castle := e.Position.MakeMove(ml.Moves[moveNum].Move)

		// if allowed != allowed2 {
		// 	fmt.Printf("new said %v old said %v for %v \n", allowed, allowed2, io.PrintMove(ml.Moves[moveNum].Move))
		// 	io.PrintBoard(pos)
		// 	e.Position.Board.PrintBitboard(e.Position.Board.Pieces)
		// 	panic("err")
		// }
		if !allowed {
			continue
		}
		Perft2(depth-1, e)
		e.Position.TakeMoveBack(ml.Moves[moveNum].Move, enpas, castle)
		//movegen.TakeMoveBack(pos)
	}
}

func PerftTest2(depth int, fen string) int64 {
	defer util.TimeTrackNano(time.Now(), "time: ")

	// bOld := data.Board{}
	// board.ParseFEN(fen, &bOld)
	// mlOld := data.MoveList{}
	// movegen.GenerateAllMoves(&bOld, &mlOld)

	board := engine.Bitboard{}
	pos := &engine.Position{Board: board}
	b := engine.Engine{Position: pos}
	b.Position.ParseFen(fen)
	fmt.Printf("\nStarting test to depth %d\n", depth)

	leafNodes = 0
	ml := &engine.MoveList{}
	b.Position.GenerateAllMoves(ml)
	// fmt.Printf("new %v - old %v\n", ml.Count, mlOld.Count)
	// for _, v := range mlOld.Moves {
	// 	found := false
	// 	for _, s := range ml.Moves {
	// 		if v.Move == s.Move {
	// 			found = true
	// 		}
	// 	}
	// 	if !found {
	// 		fmt.Printf("%v was not found for new", v)
	// 		panic("err")
	// 	}
	// }

	// for _, v := range ml.Moves {
	// 	found := false
	// 	for _, s := range mlOld.Moves {
	// 		if v.Move == s.Move {
	// 			found = true
	// 		}
	// 	}
	// 	if !found {
	// 		fmt.Printf("%v was not found for old", io.PrintMove(v.Move))
	// 		panic("err")
	// 	}
	// }

	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		move := ml.Moves[moveNum].Move
		// allowedOld := movegen.MakeMove(ml.Moves[moveNum].Move, &bOld)
		allowed, enpas, castle := b.Position.MakeMove(move)
		// if allowed != allowedOld {
		// 	fmt.Printf("%v was not found for old", io.PrintMove(ml.Moves[moveNum].Move))
		// 	b.Position.Board.PrintBitboard(b.Position.Board.Pieces)
		// 	panic("err")
		// }
		if !allowed {
			continue
		}

		Perft2(depth-1, b)
		b.Position.TakeMoveBack(move, enpas, castle)
		//movegen.TakeMoveBack(&bOld)

	}
	fmt.Printf("\nTest Complete : %d nodes\n", leafNodes)
	return leafNodes
}
