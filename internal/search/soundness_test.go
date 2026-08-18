package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// soundnessSearchDepth is the fixed depth used for every pruning-soundness
// comparison: deep enough to engage null-move pruning (depth >= 3), LMR
// (depth >= LMRMinDepth == 3), and razoring (depth <= RazoringMaxDepth == 1,
// reached as the search descends), while staying fast for ~10 positions.
const soundnessSearchDepth = 4

// soundnessTTSizeMB is the transposition table size used by both the pruned
// and unpruned engine in every case. Each test case gets a fresh pair of
// engines, so there is no cross-position or cross-run TT state to worry
// about.
const soundnessTTSizeMB = 16

// soundnessCase pairs a tactical FEN with a human rationale for why its best
// move is uniquely best, so that a same-best-move assertion here cannot pass
// by accident on a position with multiple equally-good moves (which would
// flake as pruning nudges bounds-affected scores around on non-PV moves).
type soundnessCase struct {
	name string
	fen  string
	// why documents the uniqueness rationale for this position's best move,
	// verified either by "it is a forced mate in 1" (only one move ends the
	// game immediately) or by a manual per-root-move score dump showing a
	// large gap between the best and second-best move (see the report for
	// the exact numbers).
	why string
}

// soundnessCases reuses eight forced mate-in-1 positions from mate_test.go's
// mateCases (each already documented there as having a unique mating move -
// see mateCase.wantMove's doc comment) plus two additional tactical
// positions verified separately (a hanging-queen capture and the golden
// suite's mating-net tactic). None of these positions has the side to move
// reduced to a bare/lone king, so the known null-move zugzwang-vs-lone-king
// exception does not apply here.
var soundnessCases = []soundnessCase{
	{
		name: "mate1_white_back_rank_rook",
		fen:  "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
		why:  "forced mate in 1 (1.Ra8#); no other move ends the game, so it is the unique best move at any search depth >= 1",
	},
	{
		name: "mate1_white_scholars_mate",
		fen:  "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
		why:  "forced mate in 1 (1.Qxf7#); unique mating move, same reasoning as above",
	},
	{
		name: "mate1_white_kr_vs_k",
		fen:  "4k3/R7/4K3/8/8/8/8/8 w - - 0 1",
		why:  "forced mate in 1 (1.Ra8#); unique mating move",
	},
	{
		name: "mate1_white_kq_vs_k",
		fen:  "7k/8/6QK/8/8/8/8/8 w - - 0 1",
		why:  "forced mate in 1 (1.Qg7#); unique mating move",
	},
	{
		name: "mate1_white_knight_corner",
		fen:  "r5rk/6pp/7N/8/8/8/8/6KR w - - 0 1",
		why:  "forced mate in 1 (1.Nf7#, g8 blocked by black's own rook); unique mating move",
	},
	{
		name: "mate1_white_queen_back_rank",
		fen:  "3k4/8/3K4/8/8/8/8/7Q w - - 0 1",
		why:  "forced mate in 1 (1.Qh8#); unique mating move",
	},
	{
		name: "mate1_black_back_rank_rook",
		fen:  "r5k1/5ppp/8/8/8/8/5PPP/6K1 b - - 0 1",
		why:  "forced mate in 1 (1...Ra1#); unique mating move",
	},
	{
		name: "mate1_black_kr_vs_k",
		fen:  "8/8/8/8/8/4k3/r7/4K3 b - - 0 1",
		why:  "forced mate in 1 (1...Ra1#); unique mating move",
	},
	{
		name: "hanging_queen_knight_capture",
		fen:  "rnb1kbnr/ppp2ppp/8/3pp3/4P2q/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 0 1",
		why:  "black's queen on h4 hangs to Nf3xh4; a manual per-root-move score dump at depth 4 gave Nxh4 = 875 vs. the second-best move (Bf1-b5, a pin) = 745, a 130cp gap decisively favoring the queen capture",
	},
	{
		name: "mating_net_tactic",
		fen:  "1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
		why:  "black to move, ...Qxd1+ forces mate (reused from goldenTestFENs); a manual per-root-move score dump at depth 4 gave Qxd1 = 29998 (near-mate score) vs. the second-best move = -313, an overwhelming gap",
	},
}

// disablePruning turns off every pruning technique this suite compares
// against: null-move pruning, razoring, and Late Move Reductions. It leaves
// the transposition table, move ordering, and check extensions untouched -
// those affect search efficiency and node order but not soundness the way
// the toggled techniques do, and are not gated by these tests.
func disablePruning(m *MinimaxEngine) {
	m.searchState.searchParams.NullMoveEnabled = false
	m.searchState.searchParams.RazoringEnabled = false
	m.searchState.searchParams.LMREnabled = false
}

// TestPruningSoundness_SameBestMoveAsUnpruned searches each case in
// soundnessCases twice at a fixed depth - once with the default (pruning
// enabled) Params, once with null-move pruning, razoring, and LMR all
// disabled via disablePruning - and asserts both runs choose the same best
// move.
//
// Each run gets its own fresh engine and board (not a shared engine with
// ClearSearchState) so that TT/history/killer state from the pruned run can
// never leak into the unpruned run or vice versa; the only difference
// between the two runs is the Params toggles.
//
// A mismatch here would mean a pruning technique discarded the actual best
// move on a position where the brief's failure protocol applies: stop and
// report the FEN rather than "fixing" the toggle to hide it. The known
// documented exception is null-move pruning misbehaving in king-only
// zugzwang positions; soundnessCases deliberately avoids bare-lone-king
// positions for the side to move (see its doc comment), so that exception
// should never be hit here.
func TestPruningSoundness_SameBestMoveAsUnpruned(t *testing.T) {
	for _, tc := range soundnessCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			prunedBoard := testutil.MustFromFEN(t, tc.fen)
			player := playerFromSide(prunedBoard.GetSideToMove())

			prunedEngine := NewMinimaxEngine()
			prunedEngine.SetTranspositionTableSize(soundnessTTSizeMB)
			prunedResult := prunedEngine.FindBestMove(context.Background(), prunedBoard, player, SearchConfig{
				MaxDepth:       soundnessSearchDepth,
				UseOpeningBook: false,
			})

			unprunedBoard := testutil.MustFromFEN(t, tc.fen)
			unprunedEngine := NewMinimaxEngine()
			unprunedEngine.SetTranspositionTableSize(soundnessTTSizeMB)
			disablePruning(unprunedEngine)
			unprunedResult := unprunedEngine.FindBestMove(context.Background(), unprunedBoard, player, SearchConfig{
				MaxDepth:       soundnessSearchDepth,
				UseOpeningBook: false,
			})

			prunedMove := moveString(prunedResult.BestMove)
			unprunedMove := moveString(unprunedResult.BestMove)
			if prunedMove != unprunedMove {
				t.Errorf("fen=%q: pruning changed the best move: pruned=%s (score=%d) unpruned=%s (score=%d)\n  uniqueness rationale: %s",
					tc.fen, prunedMove, prunedResult.Score, unprunedMove, unprunedResult.Score, tc.why)
			}
		})
	}
}
