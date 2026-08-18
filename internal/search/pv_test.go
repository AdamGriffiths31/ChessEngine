package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// pvMove builds a distinct, non-zero move for PV tests, identified by its
// destination file so lines are easy to read back.
func pvMove(file int) board.Move {
	return board.Move{
		From: board.Square{File: 0, Rank: 1},
		To:   board.Square{File: file, Rank: 3},
	}
}

func linesEqual(got, want []board.Move) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestPVTable_StoresAndPropagatesLine builds a 3-ply line from the deepest ply
// up, exactly as negamax does on successive alpha improvements, and checks the
// root line reads back in order.
func TestPVTable_StoresAndPropagatesLine(t *testing.T) {
	pt := newPVTable(10)

	a, b, c := pvMove(4), pvMove(5), pvMove(6)

	// Deepest ply first, then propagate upward.
	pt.store(2, c)
	pt.store(1, b)
	pt.store(0, a)

	if got, want := pt.lineAt(2), []board.Move{c}; !linesEqual(got, want) {
		t.Errorf("lineAt(2) = %v, want %v", got, want)
	}
	if got, want := pt.lineAt(1), []board.Move{b, c}; !linesEqual(got, want) {
		t.Errorf("lineAt(1) = %v, want %v", got, want)
	}
	if got, want := pt.lineAt(0), []board.Move{a, b, c}; !linesEqual(got, want) {
		t.Errorf("lineAt(0) = %v, want %v", got, want)
	}
}

// TestPVTable_TruncatesAtRowCapacity checks that copying a child line that
// would overflow the destination row stops at capacity and drops the tail,
// matching the inline `if pvLen >= len(pv[ply]) { break }` behavior.
func TestPVTable_TruncatesAtRowCapacity(t *testing.T) {
	pt := newPVTable(2) // rows can hold at most 2 moves

	c, d := pvMove(6), pvMove(7)
	// Fill the child row (ply 1) to capacity with two real moves.
	pt.lines[1][0] = c
	pt.lines[1][1] = d

	b := pvMove(5)
	pt.store(0, b)

	// Room for b plus only the first child move; d is truncated and there is
	// no room for a terminator.
	if got, want := pt.lineAt(0), []board.Move{b, c}; !linesEqual(got, want) {
		t.Errorf("lineAt(0) = %v, want %v (d should be truncated)", got, want)
	}
}

// TestPVTable_ClearPly clears a single ply's line without disturbing others.
func TestPVTable_ClearPly(t *testing.T) {
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
}

// TestPVTable_Reset terminates every line.
func TestPVTable_Reset(t *testing.T) {
	pt := newPVTable(10)

	pt.store(2, pvMove(6))
	pt.store(1, pvMove(5))
	pt.store(0, pvMove(4))

	pt.reset()

	for ply := 0; ply < 3; ply++ {
		if got := pt.lineAt(ply); len(got) != 0 {
			t.Errorf("lineAt(%d) after reset = %v, want empty", ply, got)
		}
	}
}

// TestPVTable_OutOfRangeIsNoOp confirms the bounds guards that replaced the
// former inline `ply < len(pv)` / `ply+1 < len(pv)` checks.
func TestPVTable_OutOfRangeIsNoOp(t *testing.T) {
	pt := newPVTable(3)

	// Must not panic.
	pt.store(3, pvMove(4))
	pt.store(-1, pvMove(4))
	pt.clearPly(3)
	pt.clearPly(-1)

	if got := pt.lineAt(3); got != nil {
		t.Errorf("lineAt(3) out of range = %v, want nil", got)
	}
}
