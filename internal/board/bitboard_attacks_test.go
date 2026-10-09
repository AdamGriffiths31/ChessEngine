package board

import "testing"

// One white piece of each type, then spot-checks of which squares are attacked.
func TestIsSquareAttackedByColor(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	board.SetPiece(1, 1, WhitePawn)   // b2
	board.SetPiece(2, 2, WhiteKnight) // c3
	board.SetPiece(0, 0, WhiteRook)   // a1
	board.SetPiece(3, 3, WhiteBishop) // d4
	board.SetPiece(4, 4, WhiteQueen)  // e5
	board.SetPiece(0, 4, WhiteKing)   // e1

	tests := []struct {
		square string
		reason string
	}{
		{"c3", "pawn on b2"},
		{"a3", "pawn on b2"},
		{"b1", "knight on c3"},
		{"d5", "knight on c3"},
		{"a2", "rook on a1"},
		{"c1", "rook on a1"},
		{"a8", "rook on a1"},
		{"f2", "bishop on d4"},
		{"c5", "bishop on d4"},
		{"e6", "queen on e5"},
		{"a5", "queen on e5"},
		{"h8", "queen on e5"},
		{"d1", "king on e1"},
		{"f1", "king on e1"},
	}
	for _, tt := range tests {
		if !board.IsSquareAttackedByColor(StringToSquare(tt.square), BitboardWhite) {
			t.Errorf("%s should be attacked by White (%s)", tt.square, tt.reason)
		}
	}
}

// Sliders attack up to and including the first blocker, and nothing beyond it.
func TestSlidingPieceAttacksWithBlockers(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	board.SetPiece(0, 0, WhiteRook) // a1
	board.SetPiece(0, 3, BlackPawn) // d1, blocker

	for _, square := range []string{"b1", "c1", "d1"} {
		if !board.IsSquareAttackedByColor(StringToSquare(square), BitboardWhite) {
			t.Errorf("rook on a1 should attack %s", square)
		}
	}
	for _, square := range []string{"e1", "f1", "g1", "h1"} {
		if board.IsSquareAttackedByColor(StringToSquare(square), BitboardWhite) {
			t.Errorf("rook on a1 should not attack %s (beyond the blocker)", square)
		}
	}
}

func TestGetAttackersToSquare(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	board.SetPiece(2, 3, WhitePawn)   // d3 attacks e4
	board.SetPiece(2, 2, WhiteKnight) // c3 attacks e4
	board.SetPiece(3, 0, WhiteRook)   // a4 attacks e4
	board.SetPiece(1, 1, WhiteBishop) // b2 does not attack e4

	got := board.GetAttackersToSquare(StringToSquare("e4"), BitboardWhite)
	want := bitboardOf(StringToSquare("d3"), StringToSquare("c3"), StringToSquare("a4"))
	if got != want {
		t.Errorf("attackers of e4 = %v, want %v", got.BitList(), want.BitList())
	}
}

func TestIsInCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		setup     func(*Board)
		whiteInCk bool
		blackInCk bool
	}{
		{"empty board", func(*Board) {}, false, false},
		{"white king on a rook's file", func(b *Board) {
			b.SetPiece(0, 4, WhiteKing)
			b.SetPiece(7, 4, BlackRook)
		}, true, false},
		{"black king on a bishop's diagonal", func(b *Board) {
			b.SetPiece(7, 4, BlackKing)
			b.SetPiece(6, 3, WhiteBishop)
		}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setup(board)
			if got := board.IsInCheck(BitboardWhite); got != tt.whiteInCk {
				t.Errorf("IsInCheck(White) = %v, want %v", got, tt.whiteInCk)
			}
			if got := board.IsInCheck(BitboardBlack); got != tt.blackInCk {
				t.Errorf("IsInCheck(Black) = %v, want %v", got, tt.blackInCk)
			}
		})
	}
}

func TestGetPieceAttacks(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	e4 := StringToSquare("e4")

	if got, want := board.GetPieceAttacks(WhitePawn, e4), bitboardOf(StringToSquare("d5"), StringToSquare("f5")); got != want {
		t.Errorf("white pawn on e4 attacks %v, want %v", got.BitList(), want.BitList())
	}
	if got := board.GetPieceAttacks(WhiteKnight, e4).PopCount(); got != 8 {
		t.Errorf("knight on e4 has %d attacks, want 8", got)
	}
	if got := board.GetPieceAttacks(Empty, e4); got != 0 {
		t.Errorf("Empty piece attacks %v, want none", got.BitList())
	}
}

func TestGetAllAttackedSquares(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	board.SetPiece(1, 4, WhitePawn)   // e2
	board.SetPiece(2, 1, WhiteKnight) // b3
	board.SetPiece(0, 4, WhiteKing)   // e1

	attacks := board.GetAllAttackedSquares(BitboardWhite)
	for _, square := range []string{"d3", "f3" /* pawn */, "d4", "a5" /* knight */, "d1", "f1" /* king */} {
		if !attacks.HasBit(StringToSquare(square)) {
			t.Errorf("%s should be among White's attacked squares", square)
		}
	}
}

// Square accessors tolerate out-of-range indices instead of panicking.
func TestSquareAccessors(t *testing.T) {
	t.Parallel()
	board := NewBoard()
	e4 := StringToSquare("e4")

	if !board.IsSquareEmptyBitboard(e4) || board.GetPieceOnSquare(e4) != Empty {
		t.Error("e4 should be empty on a new board")
	}
	board.SetPiece(3, 4, WhiteQueen)
	if board.IsSquareEmptyBitboard(e4) || board.GetPieceOnSquare(e4) != WhiteQueen {
		t.Error("e4 should hold the white queen")
	}
	for _, square := range []int{-1, 64} {
		if board.IsSquareEmptyBitboard(square) {
			t.Errorf("IsSquareEmptyBitboard(%d) = true, want false", square)
		}
		if got := board.GetPieceOnSquare(square); got != Empty {
			t.Errorf("GetPieceOnSquare(%d) = %c, want Empty", square, got)
		}
	}
}

// The benchmarks below cover the hot paths used by move legality checks.
func benchBoard(b *testing.B) *Board {
	b.Helper()
	board, err := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		b.Fatal(err)
	}
	return board
}

func BenchmarkIsSquareAttackedByColor(b *testing.B) {
	board, square := benchBoard(b), StringToSquare("e4")
	b.ResetTimer()
	for range b.N {
		_ = board.IsSquareAttackedByColor(square, BitboardWhite)
	}
}

func BenchmarkIsInCheck(b *testing.B) {
	board := benchBoard(b)
	b.ResetTimer()
	for range b.N {
		_ = board.IsInCheck(BitboardWhite)
	}
}
