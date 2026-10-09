package search

import (
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// sq builds a board.Square.
func sq(file, rank int) board.Square {
	return board.Square{File: file, Rank: rank}
}

// playerFromSide converts a board's side-to-move ("w"/"b") into a Player.
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

// moveString renders a move in coordinate notation ("e2e4", "a7a8q").
func moveString(move board.Move) string {
	s := move.From.String() + move.To.String()
	if move.Promotion != board.Empty {
		s += strings.ToLower(string(move.Promotion))
	}
	return s
}

// assertLegalMove fails the test unless move is a legal move for player on b.
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
