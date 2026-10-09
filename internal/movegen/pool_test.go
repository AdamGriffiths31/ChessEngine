package movegen

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func fillMoves(ml *MoveList, n int) {
	for i := range n {
		ml.AddMove(board.Move{
			From: board.Square{File: i % 8, Rank: 0},
			To:   board.Square{File: i % 8, Rank: 1},
		})
	}
}

func TestMoveListPool(t *testing.T) {
	ml := GetMoveList()
	if ml.Count != 0 || len(ml.Moves) != 0 {
		t.Fatalf("new list has Count=%d len=%d, want both 0", ml.Count, len(ml.Moves))
	}

	fillMoves(ml, 10)
	if ml.Count != 10 {
		t.Fatalf("Count = %d after 10 adds, want 10", ml.Count)
	}

	ml.Clear()
	if ml.Count != 0 || len(ml.Moves) != 0 {
		t.Errorf("after Clear: Count=%d len=%d, want both 0", ml.Count, len(ml.Moves))
	}
	if cap(ml.Moves) == 0 {
		t.Error("Clear should keep the backing array")
	}

	fillMoves(ml, 3)
	ReleaseMoveList(ml)
	if again := GetMoveList(); again.Count != 0 {
		t.Errorf("list from the pool has Count=%d, want it reset", again.Count)
	}
}

// Oversized lists are dropped on release rather than kept in the pool.
func TestMoveListPoolDropsOversizedLists(t *testing.T) {
	ml := GetMoveList()
	fillMoves(ml, maxPooledCapacity+1)
	grown := cap(ml.Moves)
	ReleaseMoveList(ml)

	if got := cap(GetMoveList().Moves); got >= grown {
		t.Errorf("pool returned capacity %d, want less than the oversized %d", got, grown)
	}
}
