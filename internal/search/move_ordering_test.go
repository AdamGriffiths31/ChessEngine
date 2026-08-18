package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// TestGetTacticalBonus_WhiteMoveAttackingBlackPiece guards against a
// regression where getTacticalBonus's side check used
// `move.Piece >= board.WhitePawn && move.Piece <= board.WhiteKing`, which can
// never be true since Piece values are FEN-letter rune codes
// (WhitePawn='P'=80, WhiteKing='K'=75), not an ordered enum. That sent every
// White move down the "Black" branch, scoring it against White's own
// pieces/king instead of Black's - e.g. this exact position spuriously
// awarded a king-zone bonus for approaching White's own king on a1 (whose
// zone includes a2/b1, both knight-attack squares from c3) while missing the
// real threat against Black's rook on e4 entirely.
func TestGetTacticalBonus_WhiteMoveAttackingBlackPiece(t *testing.T) {
	// White knight on c3, Black rook on e4 (a knight-attack square from c3).
	// Kings are far enough from c3's attack squares (a2,a4,b1,b5,d1,d5,e2,e4)
	// that only the "attacks an enemy piece" +100000 bonus should fire.
	b, err := board.FromFEN("7k/8/8/8/4r3/2N5/8/K7 w - - 0 1")
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	m := NewMinimaxEngine()
	move := board.Move{
		Piece: board.WhiteKnight,
		From:  board.Square{File: 1, Rank: 0}, // b1, irrelevant to the check
		To:    board.Square{File: 2, Rank: 2}, // c3
	}

	if got, want := m.getTacticalBonus(b, move), 100000; got != want {
		t.Errorf("getTacticalBonus() = %d, want %d (White knight on c3 attacking Black rook on e4)", got, want)
	}
}
