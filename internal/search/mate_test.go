package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

const (
	mateSearchDepth = 6
	mateTTSizeMB    = 64
)

type mateCase struct {
	name      string
	fen       string
	line      string
	wantScore eval.EvaluationScore
	wantMove  string
}

var mateCases = []mateCase{
	{
		name:      "mate1_white_back_rank_rook",
		fen:       "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
		line:      "1.Ra8#",
		wantScore: eval.MateScore - 1,
		wantMove:  "a1a8",
	},
	{
		name:      "mate1_white_scholars_mate",
		fen:       "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
		line:      "1.Qxf7#",
		wantScore: eval.MateScore - 1,
		wantMove:  "h5f7",
	},
	{
		name:      "mate1_white_kr_vs_k",
		fen:       "4k3/R7/4K3/8/8/8/8/8 w - - 0 1",
		line:      "1.Ra8#",
		wantScore: eval.MateScore - 1,
		wantMove:  "a7a8",
	},
	{
		name:      "mate1_white_kq_vs_k",
		fen:       "7k/8/6QK/8/8/8/8/8 w - - 0 1",
		line:      "1.Qg7#",
		wantScore: eval.MateScore - 1,
		wantMove:  "g6g7",
	},
	{
		name:      "mate1_white_knight_corner",
		fen:       "r5rk/6pp/7N/8/8/8/8/6KR w - - 0 1",
		line:      "1.Nf7# (g8 blocked by black's own rook)",
		wantScore: eval.MateScore - 1,
		wantMove:  "h6f7",
	},
	{
		name:      "mate1_white_queen_back_rank",
		fen:       "3k4/8/3K4/8/8/8/8/7Q w - - 0 1",
		line:      "1.Qh8#",
		wantScore: eval.MateScore - 1,
		wantMove:  "h1h8",
	},

	{
		name:      "mate1_black_back_rank_rook",
		fen:       "r5k1/5ppp/8/8/8/8/5PPP/6K1 b - - 0 1",
		line:      "1...Ra1#",
		wantScore: eval.MateScore - 1,
		wantMove:  "a8a1",
	},
	{
		name:      "mate1_black_kr_vs_k",
		fen:       "8/8/8/8/8/4k3/r7/4K3 b - - 0 1",
		line:      "1...Ra1#",
		wantScore: eval.MateScore - 1,
		wantMove:  "a2a1",
	},

	{
		name:      "mate2_white_rook_sac_pawn_mate",
		fen:       "kbK5/pp6/1P6/8/8/8/8/R7 w - - 0 1",
		line:      "1.Ra6! bxa6 2.b7#  (3 plies; zugzwang - every Black move loses)",
		wantScore: eval.MateScore - 3,
	},
	{
		name:      "mate2_white_two_rook_ladder",
		fen:       "7k/8/8/8/8/8/R7/1R4K1 w - - 0 1",
		line:      "1.Ra7 Kg8 2.Rb8#",
		wantScore: eval.MateScore - 3,
	},
	{
		name:      "mate2_white_two_rooks_first_rank",
		fen:       "7k/8/8/8/8/8/8/2R1R2K w - - 0 1",
		line:      "1.Re7 Kg8 2.Rc8#",
		wantScore: eval.MateScore - 3,
	},

	{
		name:      "mate2_black_two_rook_ladder",
		fen:       "1r4k1/r7/8/8/8/8/8/7K b - - 0 1",
		line:      "1...Ra2 2.Kg1 Rb1#",
		wantScore: eval.MateScore - 3,
	},
	{
		name:      "mate2_black_rook_lift_back_rank",
		fen:       "5rk1/8/8/8/8/8/r7/5K2 b - - 0 1",
		line:      "1...Rb8 2.Kg1 Rb1#",
		wantScore: eval.MateScore - 3,
	},

	{
		name:      "mate3_black_king_hunt",
		fen:       "r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1",
		line:      "1...Bc5+ 2.Kxc5 Qb6+ 3.Kd5 Qd6#  (5 plies, 3 checks)",
		wantScore: eval.MateScore - 5,
	},
	{
		name:      "mate3_white_king_hunt",
		fen:       "r1bq1b1r/ppp4p/2n3pn/3kp3/4N3/5Q2/PPPP1PPP/R1B1KB1R w KQ - 0 1",
		line:      "1.Bc4+ Kxc4 2.Qb3+ Kd4 3.Qd3#  (5 plies, 3 checks)",
		wantScore: eval.MateScore - 5,
	},
}

func TestFindBestMove_MateInN(t *testing.T) {
	for _, tc := range mateCases {

		t.Run(tc.name, func(t *testing.T) {
			engine := NewMinimaxEngine()
			engine.SetTranspositionTableSize(mateTTSizeMB)

			b := testutil.MustFromFEN(t, tc.fen)
			player := playerFromSide(b.GetSideToMove())

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result := engine.FindBestMove(ctx, b, player, Config{
				MaxDepth:       mateSearchDepth,
				UseOpeningBook: false,
			})

			if result.Score != tc.wantScore {
				t.Errorf("Score = %d, want %d (line: %s)", result.Score, tc.wantScore, tc.line)
			}

			got := moveString(result.BestMove)
			if tc.wantMove != "" && got != tc.wantMove {
				t.Errorf("BestMove = %s, want %s (line: %s)", got, tc.wantMove, tc.line)
			}

			assertLegalMove(t, engine, b, player, result.BestMove)
		})
	}
}
