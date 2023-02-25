package perft

import (
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	movegen "github.com/AdamGriffiths31/ChessEngine/moveGen"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

var leafNodes int64 = 0

func Perft(depth int, pos *data.Board) {
	board.CheckBoard(pos)

	if depth == 0 {
		leafNodes++
		return
	}

	ml := data.MoveList{}
	movegen.GenerateAllMoves(pos, &ml)
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		if !movegen.MakeMove(ml.Moves[moveNum].Move, pos) {
			continue
		}

		Perft(depth-1, pos)

		movegen.TakeMoveBack(pos)

	}
}

func PerftTest(depth int, fen string) int64 {
	defer util.TimeTrackNano(time.Now(), "time: ")

	b := data.Board{}
	board.ParseFEN(fen, &b)

	board.CheckBoard(&b)
	fmt.Printf("Starting test to depth %d\n", depth)

	leafNodes = 0
	ml := data.MoveList{}
	movegen.GenerateAllMoves(&b, &ml)
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		move := ml.Moves[moveNum].Move
		if !movegen.MakeMove(move, &b) {
			continue
		}

		Perft(depth-1, &b)
		movegen.TakeMoveBack(&b)
	}
	fmt.Printf("\nTest Complete : %d nodes\n", leafNodes)
	return leafNodes
}
