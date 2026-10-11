package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

func searchScore(t *testing.T, fen string, depth int) int {
	t.Helper()
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(16)
	b := testutil.MustFromFEN(t, fen)
	r := engine.FindBestMove(context.Background(), b, playerFromSide(b.GetSideToMove()), Config{MaxDepth: depth})
	return int(r.Score)
}

func TestInsufficientMaterialScoresDraw(t *testing.T) {
	cases := map[string]string{
		"KvK":  "8/8/4k3/8/8/4K3/8/8 w - - 0 1",
		"KBvK": "8/8/4k3/8/8/4K3/4B3/8 w - - 0 1",
		"KNvK": "8/8/4k3/8/8/4K3/4N3/8 w - - 0 1",
		"KvKB": "8/8/4k3/8/8/4K3/4b3/8 w - - 0 1",
		"KvKN": "8/8/4k3/8/8/4K3/4n3/8 b - - 0 1",
	}
	for name, fen := range cases {
		if got := searchScore(t, fen, 6); got != 0 {
			t.Errorf("%s: score = %d, want 0 (insufficient material)", name, got)
		}
	}
}

func TestSufficientMaterialNotDraw(t *testing.T) {
	cases := map[string]string{
		"KRvK":  "8/8/4k3/8/8/4K3/8/R7 w - - 0 1",
		"KPvK":  "8/8/4k3/8/8/4K3/4P3/8 w - - 0 1",
		"KBBvK": "8/8/4k3/8/8/4K3/3BB3/8 w - - 0 1",
		"KRvKB": "8/8/4k3/8/8/4K3/4R3/5b2 w - - 0 1",
		"KNNvK": "8/8/4k3/8/8/4K3/3NN3/8 w - - 0 1",
	}
	for name, fen := range cases {
		if got := searchScore(t, fen, 6); got == 0 {
			t.Errorf("%s: score = 0, but material is sufficient to keep play going", name)
		}
	}
}

func TestFiftyMoveRuleScoresDraw(t *testing.T) {
	// At clock 99 every non-capturing, non-pawn move reaches clock 100, so a
	// king-and-rook side with no pawn move or capture available is drawn.
	if got := searchScore(t, "8/8/4k3/8/8/4K3/8/R7 w - - 99 100", 6); got != 0 {
		t.Errorf("KRvK at halfmove 99: score = %d, want 0", got)
	}
	if got := searchScore(t, "8/8/4k3/8/8/4K3/8/R7 w - - 0 1", 6); got <= 0 {
		t.Errorf("KRvK at halfmove 0: score = %d, want a winning score", got)
	}
}

func TestFiftyMoveRuleMateTakesPrecedence(t *testing.T) {
	// Back-rank mate in one with the clock at 99: Ra8# is a pawn-less rook
	// move that makes the clock 100 but delivers checkmate, which wins.
	if got := searchScore(t, "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 99 100", 4); got < 29000 {
		t.Errorf("mate on the 100th halfmove: score = %d, want a mate score", got)
	}
}
