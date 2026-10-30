package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

func TestMateScoreMatchesPVLength(t *testing.T) {
	fens := []string{
		"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
		"2R5/6K1/8/8/8/3R4/8/6k1 w - - 0 1",
		"8/3k4/1Q6/7R/8/1K6/8/8 w - - 0 1",
		"K7/7k/8/8/8/4R3/8/5R2 w - - 0 1",
		"8/8/6k1/1R6/7K/8/4R3/8 w - - 0 1",
		"7k/2K5/8/1R6/1Q6/8/8/8 w - - 0 1",
	}
	gen := movegen.NewGenerator()
	for _, fen := range fens {
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(0)
		b := testutil.MustFromFEN(t, fen)
		player := playerFromSide(b.GetSideToMove())
		r := engine.FindBestMove(context.Background(), b, player, Config{MaxDepth: 8})

		if r.Score < eval.MateScore-1000 {
			t.Errorf("%s: score %d is not a mate score", fen, r.Score)
			continue
		}
		claimed := int(eval.MateScore - r.Score)

		walk, _ := board.FromFEN(fen)
		p := player
		plies := 0
		for _, mv := range r.Stats.PrincipalVariation {
			if _, err := walk.MakeMoveWithUndo(mv); err != nil {
				t.Fatalf("%s: PV move %v unplayable: %v", fen, mv, err)
			}
			p = opposite(p)
			plies++
		}
		ml := gen.GenerateAllMoves(walk, p)
		mated := ml.Count == 0 && gen.IsKingInCheck(walk, p)
		movegen.ReleaseMoveList(ml)
		if !mated {
			t.Errorf("%s: PV (%d plies) does not end in checkmate", fen, plies)
			continue
		}
		if claimed != plies {
			t.Errorf("%s: score claims mate in %d plies but PV mates in %d", fen, claimed, plies)
		}
	}
}

func TestMateDistancePruningStopsAtShortestMate(t *testing.T) {
	const (
		fen      = "8/p4P1K/rb1N2Bp/p3P3/N3R1P1/1P1k1P2/1Pbrnp2/2R2Q2 w - - 0 1"
		maxNodes = 1_000_000
	)

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)
	b := testutil.MustFromFEN(t, fen)
	r := engine.FindBestMove(context.Background(), b, playerFromSide(b.GetSideToMove()), Config{MaxDepth: 8})

	if want := eval.MateScore - 3; r.Score != want {
		t.Errorf("score = %d, want %d (mate in 2)", r.Score, want)
	}
	if got := moveString(r.BestMove); got != "h7h6" {
		t.Errorf("best move = %s, want h7h6", got)
	}
	if r.Stats.NodesSearched > maxNodes {
		t.Errorf("searched %d nodes, want at most %d", r.Stats.NodesSearched, maxNodes)
	}
}
