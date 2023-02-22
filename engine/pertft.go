package engine

import (
	"fmt"
	"time"
)

var leafNodes int64 = 0

func Perft(depth int, pos *Board) {
	CheckBoard(pos)

	if depth == 0 {
		leafNodes++
		return
	}

	ml := MoveList{}
	GenerateAllMoves(pos, &ml)

	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		if !MakeMove(ml.Moves[moveNum].Move, pos) {
			continue
		}

		Perft(depth-1, pos)

		TakeMoveBack(pos)

	}
}

func PerftTest(depth int, fen string) int64 {
	defer TimeTrackNano(time.Now(), "time: ")

	b := Board{}
	ParseFEN(fen, &b)

	CheckBoard(&b)
	fmt.Printf("Starting test to depth %d\n", depth)

	leafNodes = 0
	ml := MoveList{}
	GenerateAllMoves(&b, &ml)
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		move := ml.Moves[moveNum].Move
		if !MakeMove(move, &b) {
			continue
		}

		Perft(depth-1, &b)
		TakeMoveBack(&b)
		//total := leafNodes
		//oldNodes := leafNodes - total
		//fmt.Printf("move %d : %s : %v\n", moveNum+1, PrintMove(move), oldNodes)
	}
	fmt.Printf("\nTest Complete : %d nodes\n", leafNodes)
	return leafNodes
}
