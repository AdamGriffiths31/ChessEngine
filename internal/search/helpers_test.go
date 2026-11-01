package search

import (
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

func sq(file, rank int) board.Square {
	return board.NewSquare(file, rank)
}

func playerFromSide(side string) movegen.Player {
	if side == "w" {
		return movegen.White
	}
	return movegen.Black
}

func opposite(p movegen.Player) movegen.Player {
	if p == movegen.White {
		return movegen.Black
	}
	return movegen.White
}

func moveString(move board.Move) string {
	s := move.From.String() + move.To.String()
	if move.Promotion != board.Empty {
		s += strings.ToLower(string(move.Promotion))
	}
	return s
}

func assertLegalMove(t *testing.T, engine *MinimaxEngine, b *board.Board, player movegen.Player, move board.Move) {
	t.Helper()
	if !move.IsValid() {
		t.Fatalf("move %+v is not valid", move)
	}
	moves := engine.generator.GenerateAllMoves(b, player)
	defer movegen.ReleaseMoveList(moves)
	for i := range moves.Count {
		if moves.Moves[i] == move {
			return
		}
	}
	t.Errorf("move %s is not legal in this position", moveString(move))
}
