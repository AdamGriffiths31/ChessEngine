package perft2

import (
	"github.com/AdamGriffiths31/ChessEngine/engine"
	"github.com/AdamGriffiths31/ChessEngine/search"
)

var leafNodes int64 = 0

func Perft(depth int, e search.Engine) {
	if depth == 0 {
		leafNodes++
		return
	}

	ml := &engine.MoveList{}
	e.Position.GenerateAllMoves(ml)
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		allowed, enpas, castle, fifty := e.Position.MakeMove(ml.Moves[moveNum].Move)

		if !allowed {
			continue
		}

		Perft(depth-1, e)
		e.Position.TakeMoveBack(ml.Moves[moveNum].Move, enpas, castle, fifty)
	}
}

func PerftTest(depth int, fen string) int64 {

	board := engine.Bitboard{}
	pos := &engine.Position{Board: board}
	b := search.Engine{Position: pos}
	b.Position.ParseFen(fen)

	leafNodes = 0
	ml := &engine.MoveList{}
	b.Position.GenerateAllMoves(ml)

	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		move := ml.Moves[moveNum].Move
		allowed, enpas, castle, fifty := b.Position.MakeMove(move)
		if !allowed {
			continue
		}

		Perft(depth-1, b)
		b.Position.TakeMoveBack(move, enpas, castle, fifty)
	}
	return leafNodes
}
