package search

import (
	"context"
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// determinismTestFENs reuses a diverse subset of zobristTestFENs (defined in
// zobrist_updates_test.go): the standard starting position plus busy
// middlegame and sparse endgame positions with varying castling rights.
var determinismTestFENs = []string{
	zobristTestFENs[0], // standard starting position: full castling rights
	zobristTestFENs[1], // "Kiwipete": full castling rights, busy middlegame
	zobristTestFENs[2], // perft position 4: partial (black-only) castling rights
	zobristTestFENs[4], // perft "position 6": no castling rights, busy middlegame
	zobristTestFENs[9], // sparse late endgame (king and pawn vs king)
}

// determinismSearchConfig is a fixed-depth configuration with no time
// cutoff and no opening book, so the search is bounded only by depth.
//
// MaxTime is set to 0, which iterative_deepening.go treats as "no time
// limit" (the cutoff checks are all guarded by `config.MaxTime > 0`). Using
// any positive MaxTime here would reintroduce a wall-clock race as an
// independent source of nondeterminism (a run that gets slightly less CPU
// time could stop one iteration earlier than the other), which would defeat
// the purpose of this test. Depth 5 with the transposition table enabled
// completes well under a second per position, so there is no need for a
// time budget.
//
// LMR parameters now live in search.Params (see params.go's getParams), and
// its defaults (LMRMinDepth: 3, LMRMinMoves: 4) already match the values the
// UCI adapter uses in production (see uci/adapter.go), so this test
// exercises the same search configuration used in actual play rather than an
// artificial one without needing to override anything here.
var determinismSearchConfig = SearchConfig{
	MaxDepth:       5,
	MaxTime:        0,
	UseOpeningBook: false,
}

// TestSearchDeterminism_ClearSearchStateResetsAllState runs FindBestMove
// twice per position on the *same* engine instance, calling
// ClearSearchState() between the two runs, and asserts that the resulting
// (BestMove, Score, NodesSearched) triple is identical both times.
//
// This is the only way a leftover-state bug in ClearSearchState (engine.go)
// would surface: stale killer moves, history heuristic scores,
// transposition table entries, or repetition-history state from a previous
// search are not observable in isolation — they only show up as a
// divergence between two searches of the same position that should
// otherwise be indistinguishable. A fresh board is built from FEN for each
// run (a searched board carries mutated hash-updater state), but the engine
// itself — and therefore its killer table, history table, transposition
// table, and repetition history — is shared across both runs and across all
// positions, with ClearSearchState() the only thing standing between them.
func TestSearchDeterminism_ClearSearchStateResetsAllState(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)

	for idx, fen := range determinismTestFENs {
		t.Run(fmt.Sprintf("fen%d", idx), func(t *testing.T) {
			engine.ClearSearchState()
			board1 := testutil.MustFromFEN(t, fen)
			player := playerFromSide(board1.GetSideToMove())
			result1 := engine.FindBestMove(context.Background(), board1, player, determinismSearchConfig)

			engine.ClearSearchState()
			board2 := testutil.MustFromFEN(t, fen)
			result2 := engine.FindBestMove(context.Background(), board2, player, determinismSearchConfig)

			if result1.BestMove != result2.BestMove || result1.Score != result2.Score || result1.Stats.NodesSearched != result2.Stats.NodesSearched {
				t.Fatalf(
					"fen=%q: nondeterministic search across ClearSearchState()\n run1: BestMove=%s Score=%d NodesSearched=%d\n run2: BestMove=%s Score=%d NodesSearched=%d",
					fen,
					moveString(result1.BestMove), result1.Score, result1.Stats.NodesSearched,
					moveString(result2.BestMove), result2.Score, result2.Stats.NodesSearched,
				)
			}
		})
	}
}

// TestFindBestMove_StatsAreNotCumulativeAcrossCalls reproduces the reuse
// pattern of a whole game played on one engine instance without
// ClearSearchState between moves (the Elo benchmark runner and the live UCI
// adapter both only call ClearSearchState on a brand new game, never between
// individual moves of the same game -- TT/history/repetition state
// deliberately persist across a game's moves). Regardless of that reuse,
// SearchStats must always describe only the most recent FindBestMove call:
// before FindBestMove reset searchStats itself, NodesSearched (and every
// other counter field) silently accumulated across every move of a game,
// corrupting per-move telemetry (JSONL logs, UCI "info nodes" lines) with a
// running game-long total instead of the current move's real count.
//
// This deliberately searches two DIFFERENT positions rather than the same
// position twice: reusing the identical position would let the second call's
// warm transposition table (populated by the first call, and correctly
// persisting across calls by design) resolve nearly the whole tree from
// cache, making its NodesSearched legitimately far SMALLER than the first
// call's for reasons that have nothing to do with the reset bug being tested
// here. Two unrelated positions share negligible TT state, so a standalone
// (fresh-engine, cold-TT) search of the second position is a valid baseline
// to compare against.
func TestFindBestMove_StatsAreNotCumulativeAcrossCalls(t *testing.T) {
	fen1 := determinismTestFENs[0]
	fen2 := determinismTestFENs[1]

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)
	engine.ClearSearchState() // start-of-game reset only, matching real callers

	board1 := testutil.MustFromFEN(t, fen1)
	player1 := playerFromSide(board1.GetSideToMove())
	result1 := engine.FindBestMove(context.Background(), board1, player1, determinismSearchConfig)
	if result1.Stats.NodesSearched <= 0 {
		t.Fatalf("first call: expected positive NodesSearched, got %d", result1.Stats.NodesSearched)
	}

	// No ClearSearchState here: this is the exact reuse pattern of a second
	// move in the same game.
	board2 := testutil.MustFromFEN(t, fen2)
	player2 := playerFromSide(board2.GetSideToMove())
	reusedResult2 := engine.FindBestMove(context.Background(), board2, player2, determinismSearchConfig)

	// Baseline: the same position searched on a completely fresh engine, so
	// its NodesSearched reflects only that search, uninfluenced by anything
	// that happened on the reused engine.
	baselineEngine := NewMinimaxEngine()
	baselineEngine.SetTranspositionTableSize(64)
	baselineEngine.ClearSearchState()
	baselineBoard2 := testutil.MustFromFEN(t, fen2)
	baselineResult2 := baselineEngine.FindBestMove(context.Background(), baselineBoard2, player2, determinismSearchConfig)

	// A small variance between reusedResult2 and baselineResult2 is expected
	// even with correct behavior: the history heuristic table also
	// intentionally persists across calls within a game (unlike the TT, this
	// is move-ordering influence, not a cache), so historical cutoffs from
	// the first call's position can nudge move ordering -- and therefore the
	// exact node count -- on the second call's unrelated position. A 10%
	// relative tolerance comfortably absorbs that noise (observed to be a
	// fraction of a percent in practice) while still catching the actual
	// bug: without the reset, reusedResult2 would equal roughly
	// baselineResult2 PLUS the first call's entire NodesSearched (counters
	// are summed in place, not replaced), which is a large, fixed
	// distortion, not a small ordering nudge.
	diff := reusedResult2.Stats.NodesSearched - baselineResult2.Stats.NodesSearched
	if diff < 0 {
		diff = -diff
	}
	tolerance := baselineResult2.Stats.NodesSearched / 10
	if diff > tolerance {
		t.Fatalf(
			"reused-engine second call's NodesSearched (%d) should be close to a fresh engine's standalone search of the same position (%d, tolerance %d); the first call's own NodesSearched was %d, and a mismatch this large indicates stats are bleeding across calls instead of describing only the current search",
			reusedResult2.Stats.NodesSearched, baselineResult2.Stats.NodesSearched, tolerance, result1.Stats.NodesSearched,
		)
	}
}
