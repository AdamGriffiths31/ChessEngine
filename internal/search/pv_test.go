package search

import (
	"slices"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// pvMove builds a distinct move, identified by its destination file.
func pvMove(file int) board.Move {
	return board.Move{
		From: board.Square{File: 0, Rank: 1},
		To:   board.Square{File: file, Rank: 3},
	}
}

// A line is built from the deepest ply up, as negamax does on successive alpha
// improvements, and the root line reads back in order.
func TestPVTable_StoresAndPropagatesLine(t *testing.T) {
	pt := newPVTable(10)
	a, b, c := pvMove(4), pvMove(5), pvMove(6)

	pt.store(2, c)
	pt.store(1, b)
	pt.store(0, a)

	for ply, want := range [][]board.Move{{a, b, c}, {b, c}, {c}} {
		if got := pt.lineAt(ply); !slices.Equal(got, want) {
			t.Errorf("lineAt(%d) = %v, want %v", ply, got, want)
		}
	}
}

// Copying a child line that would overflow the row stops at capacity.
func TestPVTable_TruncatesAtRowCapacity(t *testing.T) {
	pt := newPVTable(2)
	c, d := pvMove(6), pvMove(7)
	pt.lines[1][0], pt.lines[1][1] = c, d

	b := pvMove(5)
	pt.store(0, b)

	if got, want := pt.lineAt(0), []board.Move{b, c}; !slices.Equal(got, want) {
		t.Errorf("lineAt(0) = %v, want %v (d truncated)", got, want)
	}
}

func TestPVTable_ClearAndReset(t *testing.T) {
	pt := newPVTable(10)
	pt.store(1, pvMove(5))
	pt.store(0, pvMove(4))

	pt.clearPly(0)
	if got := pt.lineAt(0); len(got) != 0 {
		t.Errorf("lineAt(0) after clearPly(0) = %v, want empty", got)
	}
	if got := pt.lineAt(1); len(got) != 1 {
		t.Errorf("clearPly(0) disturbed ply 1: lineAt(1) = %v", got)
	}

	pt.reset()
	for ply := range 2 {
		if got := pt.lineAt(ply); len(got) != 0 {
			t.Errorf("lineAt(%d) after reset = %v, want empty", ply, got)
		}
	}
}

// Out-of-range plies are ignored instead of panicking.
func TestPVTable_OutOfRangeIsNoOp(t *testing.T) {
	pt := newPVTable(3)

	pt.store(3, pvMove(4))
	pt.store(-1, pvMove(4))
	pt.clearPly(3)
	pt.clearPly(-1)

	if got := pt.lineAt(3); got != nil {
		t.Errorf("lineAt(3) out of range = %v, want nil", got)
	}
}
