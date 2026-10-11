package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// Quiet moves are ordered by history alone: a quiet move with learned history
// must outrank a quiet move that merely attacks an enemy piece.
func TestScoreMoves_HistoryBeatsMereAttack(t *testing.T) {
	// White knight on b1 can go to a3 (attacks nothing), c3 (attacks the Black rook on e4) or d2.
	b := testutil.MustFromFEN(t, "7k/8/8/8/4r3/8/8/KN6 w - - 0 1")
	m := NewMinimaxEngine()

	moves := m.generator.GenerateAllMoves(b, movegen.White)
	defer movegen.ReleaseMoveList(moves)

	find := func(to board.Square) board.Move {
		for i := range moves.Count {
			mv := moves.Moves[i]
			if mv.Piece == board.WhiteKnight && mv.To == to {
				return mv
			}
		}
		t.Fatalf("no knight move to %v", to)
		return board.Move{}
	}
	attacking := find(board.Square{File: 2, Rank: 2}) // c3, attacks e4
	learned := find(board.Square{File: 0, Rank: 2})   // a3, no attack

	for range 8 {
		m.historyTable.Bonus(learned, 6)
	}

	m.scoreMoves(b, moves, 1, board.Move{})
	scoreOf := func(want board.Move) int {
		for i := range moves.Count {
			if moves.Moves[i] == want {
				return m.searchState.moveOrderBuffers[1][i].score
			}
		}
		t.Fatalf("move %v not in list", want)
		return 0
	}

	if got, other := scoreOf(learned), scoreOf(attacking); got <= other {
		t.Errorf("learned-history quiet move scored %d, attack-only quiet move scored %d; history must win", got, other)
	}
}
