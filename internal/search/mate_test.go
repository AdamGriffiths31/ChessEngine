package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// mateSearchDepth, mateTTSizeMB: every mate test searches at fixed depth 6
// with a 64MB transposition table and no opening book, so scores are exact
// and deterministic (same preconditions as TestSearchGolden).
const (
	mateSearchDepth = 6
	mateTTSizeMB    = 64
)

// Expected mate scores are derived from the engine's own conventions, not
// from a generic "MateScore - plies" formula:
//
//   - handleNoLegalMoves returns -MateScore + pliesFromRoot at the mated
//     node, where pliesFromRoot == ply - extended. `ply` counts plies along
//     the unextended spine, and the check extension (negamax: inCheck &&
//     ply > 0) holds a child's ply constant instead of incrementing it.
//   - A checkmate node is always in check, so the extension always fires
//     there; it also fires at every in-check interior node of the mating
//     line (i.e. after every checking move).
//
// Therefore the root score for a forced mate is
//
//	MateScore - (truePlies - checkExtensionsOnLine)
//
// where truePlies is the real length of the mating line in plies and
// checkExtensionsOnLine counts the moves in the line that give check
// (including the mating move itself). Concretely, for the suite below:
//
//   - mate in 1:               1 ply,  1 check  -> 30000 - (1-1) = 30000
//   - mate in 2, quiet + mate: 3 plies, 1 check -> 30000 - (3-1) = 29998
//   - mate in 3, all checks:   5 plies, 3 checks-> 30000 - (5-3) = 29998
//
// (A quiet-then-mate mate-in-2 and an all-check mate-in-3 legitimately
// share the score 29998 under this convention; mate-in-1 is always exactly
// MateScore.) Every FEN below was verified by running the fixed engine at
// depth 6: the reported score matches the derivation for the documented
// line, and the reported best move is the documented key move.
type mateCase struct {
	name string
	fen  string
	// line documents the forced mating line the expected score was derived
	// from (truePlies / checking-move count as per the comment above).
	line      string
	wantScore eval.EvaluationScore
	// wantMove is the exact expected best move (coordinate notation) for
	// mate-in-1 cases, where the mating move is unique and must be chosen.
	// Empty for longer mates: several key moves of equal mate distance can
	// exist (e.g. either rook may start a ladder), so only legality and the
	// exact score are asserted there.
	wantMove string
}

var mateCases = []mateCase{
	// --- mate in 1, white to move: score exactly MateScore, unique move ---
	{
		name:      "mate1_white_back_rank_rook",
		fen:       "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
		line:      "1.Ra8#",
		wantScore: eval.MateScore,
		wantMove:  "a1a8",
	},
	{
		name: "mate1_white_scholars_mate",
		// The Task 7.1 bug reproducer: pre-fix, the TT re-probe during the
		// root PVS re-search returned 29999 instead of 30000 here.
		fen:       "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
		line:      "1.Qxf7#",
		wantScore: eval.MateScore,
		wantMove:  "h5f7",
	},
	{
		name:      "mate1_white_kr_vs_k",
		fen:       "4k3/R7/4K3/8/8/8/8/8 w - - 0 1",
		line:      "1.Ra8#",
		wantScore: eval.MateScore,
		wantMove:  "a7a8",
	},
	{
		name:      "mate1_white_kq_vs_k",
		fen:       "7k/8/6QK/8/8/8/8/8 w - - 0 1",
		line:      "1.Qg7#",
		wantScore: eval.MateScore,
		wantMove:  "g6g7",
	},
	{
		name:      "mate1_white_knight_corner",
		fen:       "r5rk/6pp/7N/8/8/8/8/6KR w - - 0 1",
		line:      "1.Nf7# (g8 blocked by black's own rook)",
		wantScore: eval.MateScore,
		wantMove:  "h6f7",
	},
	{
		name:      "mate1_white_queen_back_rank",
		fen:       "3k4/8/3K4/8/8/8/8/7Q w - - 0 1",
		line:      "1.Qh8#",
		wantScore: eval.MateScore,
		wantMove:  "h1h8",
	},

	// --- mate in 1, black to move ---
	{
		name:      "mate1_black_back_rank_rook",
		fen:       "r5k1/5ppp/8/8/8/8/5PPP/6K1 b - - 0 1",
		line:      "1...Ra1#",
		wantScore: eval.MateScore,
		wantMove:  "a8a1",
	},
	{
		name:      "mate1_black_kr_vs_k",
		fen:       "8/8/8/8/8/4k3/r7/4K3 b - - 0 1",
		line:      "1...Ra1#",
		wantScore: eval.MateScore,
		wantMove:  "a2a1",
	},

	// --- mate in 2, white to move: quiet key move, mate on the 2nd ---
	{
		name:      "mate2_white_rook_sac_pawn_mate",
		fen:       "kbK5/pp6/1P6/8/8/8/8/R7 w - - 0 1",
		line:      "1.Ra6! bxa6 2.b7#  (3 plies, only the mating move checks)",
		wantScore: eval.MateScore - 2,
	},
	{
		name:      "mate2_white_two_rook_ladder",
		fen:       "7k/8/8/8/8/8/R7/1R4K1 w - - 0 1",
		line:      "1.Ra7 Kg8 2.Rb8#",
		wantScore: eval.MateScore - 2,
	},
	{
		name:      "mate2_white_two_rooks_first_rank",
		fen:       "7k/8/8/8/8/8/8/2R1R2K w - - 0 1",
		line:      "1.Re7 Kg8 2.Rc8#",
		wantScore: eval.MateScore - 2,
	},

	// --- mate in 2, black to move ---
	{
		name:      "mate2_black_two_rook_ladder",
		fen:       "1r4k1/r7/8/8/8/8/8/7K b - - 0 1",
		line:      "1...Ra2 2.Kg1 Rb1#",
		wantScore: eval.MateScore - 2,
	},
	{
		name:      "mate2_black_rook_lift_back_rank",
		fen:       "5rk1/8/8/8/8/8/r7/5K2 b - - 0 1",
		line:      "1...Rb8 2.Kg1 Rb1#",
		wantScore: eval.MateScore - 2,
	},

	// --- mate in 3: forced king hunts, every mating move checks ---
	{
		name:      "mate3_black_king_hunt",
		fen:       "r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1",
		line:      "1...Bc5+ 2.Kxc5 Qb6+ 3.Kd5 Qd6#  (5 plies, 3 checks)",
		wantScore: eval.MateScore - 2,
	},
	{
		name: "mate3_white_king_hunt",
		// Color-mirror of mate3_black_king_hunt.
		fen:       "r1bq1b1r/ppp4p/2n3pn/3kp3/4N3/5Q2/PPPP1PPP/R1B1KB1R w KQ - 0 1",
		line:      "1.Bc4+ Kxc4 2.Qb3+ Kd4 3.Qd3#  (5 plies, 3 checks)",
		wantScore: eval.MateScore - 2,
	},
}

// TestFindBestMove_MateInN drives FindBestMove over a fixed suite of
// mate-in-1/2/3 positions (both colors) and asserts the exact, convention-
// derived mate score in every case, plus the exact mating move for the
// mate-in-1 cases and best-move legality everywhere.
//
// These exact-score assertions are what guard the transposition-table mate
// round-trip: handleNoLegalMoves must store its mate score with the same
// raw node-entry ply that probeTT later decodes with. When the store used
// the extension-adjusted ply instead, re-probes (e.g. the root PVS
// re-search) came back one ply off and mate-in-1 positions reported 29999
// instead of 30000.
func TestFindBestMove_MateInN(t *testing.T) {
	for _, tc := range mateCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Fresh engine per position: no cross-position TT state, same
			// isolation the golden suite gets via ClearSearchState.
			engine := NewMinimaxEngine()
			engine.SetTranspositionTableSize(mateTTSizeMB)

			b := testutil.MustFromFEN(t, tc.fen)
			player := playerFromSide(b.GetSideToMove())

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result := engine.FindBestMove(ctx, b, player, SearchConfig{
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

			// Best move must be one of the legal moves in the root position
			// (same check as TestCancellationResponsiveness).
			if !result.BestMove.IsValid() {
				t.Fatalf("BestMove %+v is not a valid move", result.BestMove)
			}
			legal := false
			moves := engine.generator.GenerateAllMoves(b, player)
			for i := 0; i < moves.Count; i++ {
				if moves.Moves[i] == result.BestMove {
					legal = true
					break
				}
			}
			if !legal {
				t.Errorf("BestMove %s is not a legal move in this position", got)
			}
		})
	}
}
