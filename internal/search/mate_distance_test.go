package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// TestMateScoreMatchesPVLength requires the reported mate distance to equal
// the number of plies in the principal variation that actually reaches
// checkmate. Check extensions and LMR/null-move reductions must not skew the
// distance the score encodes. The TT is disabled so the PV is never cut short
// by a TT cutoff. Positions are multi-move mates that include checking lines.
func TestMateScoreMatchesPVLength(t *testing.T) {
	fens := []string{
		"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1", // mate in 1
		"2R5/6K1/8/8/8/3R4/8/6k1 w - - 0 1",    // mate in 2 (3 plies)
		"8/3k4/1Q6/7R/8/1K6/8/8 w - - 0 1",     // mate in 2 (3 plies)
		"K7/7k/8/8/8/4R3/8/5R2 w - - 0 1",      // mate in 3 (5 plies)
		"8/8/6k1/1R6/7K/8/4R3/8 w - - 0 1",     // longer rook mate
		"7k/2K5/8/1R6/1Q6/8/8/8 w - - 0 1",     // queen+rook mate
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
