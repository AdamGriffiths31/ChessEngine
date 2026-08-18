package search

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// update, when passed as `go test ./internal/search -run TestSearchGolden
// -update`, regenerates testdata/search_golden.json from the engine's
// current output instead of comparing against it. There is no other
// `flag.` declaration in this package's tests (checked before adding this
// one), so "-update" cannot collide with an existing flag.
var update = flag.Bool("update", false, "regenerate the golden file (testdata/search_golden.json) from current engine output")

// goldenFilePath is the location of the golden JSON file, relative to the
// package directory (which is also `go test`'s working directory).
const goldenFilePath = "testdata/search_golden.json"

// goldenDepth is the fixed search depth used for every golden position.
// Depth 5 with the transposition table enabled keeps the full ~20-position
// suite well under the 30s budget (see TestSearchGolden's doc comment).
const goldenDepth = 5

// goldenSearchConfig is shared by every position in the golden suite: no
// time cutoff (depth is the only bound, per the same reasoning as
// determinismSearchConfig in determinism_test.go) and no opening book (so the
// engine always searches rather than returning a book move). LMR parameters
// now live in search.Params (see params.go's getParams), and its defaults
// (LMRMinDepth: 3, LMRMinMoves: 4) already match the UCI engine's production
// values, so there is nothing to override here.
var goldenSearchConfig = SearchConfig{
	MaxDepth:       goldenDepth,
	MaxTime:        0,
	UseOpeningBook: false,
}

// goldenRecord is the on-disk (and in-memory) representation of one golden
// search result. Field order here also fixes the JSON key order.
type goldenRecord struct {
	FEN      string `json:"fen"`
	Depth    int    `json:"depth"`
	BestMove string `json:"bestmove"`
	Score    int    `json:"score"`
	Nodes    int64  `json:"nodes"`
}

// goldenTestFENs is a fixed, ordered set of ~20 legal positions covering the
// opening, middlegame, endgame, and tactical phases. It reuses the 10
// positions from zobristTestFENs (zobrist_updates_test.go) - which already
// give full/partial/no castling rights, both-side en passant, both-side
// imminent promotion, and a sparse endgame - and adds 10 more chosen for
// additional phase coverage: real opening theory, well-known tactical test
// positions, plain endgames, an in-check (but not mated) position, and a
// one-move-from-mate position.
//
// This slice, and therefore the golden file, is intentionally never
// reordered: TestSearchGolden asserts index-for-index against the golden
// file's FEN field, so reordering would look like every position "moved"
// rather than a clean diff. Append new positions at the end and regenerate
// with -update instead.
var goldenTestFENs = []string{
	// --- reused from zobristTestFENs (see zobrist_updates_test.go) ---
	zobristTestFENs[0], // standard starting position: full castling rights
	zobristTestFENs[1], // "Kiwipete": full castling rights, busy middlegame
	zobristTestFENs[2], // perft position 4: partial (black-only) castling rights
	zobristTestFENs[3], // partial castling rights (white kingside only)
	zobristTestFENs[4], // perft "position 6": no castling rights, busy middlegame
	zobristTestFENs[5], // white en passant capture available
	zobristTestFENs[6], // black en passant capture available (black to move)
	zobristTestFENs[7], // white promotion imminent
	zobristTestFENs[8], // black promotion imminent (black to move)
	zobristTestFENs[9], // sparse late endgame (king and pawn vs king)

	// --- opening theory (early middlegame, white to move) ---
	// Italian Game, Giuoco Piano main line after 5...Bc5 6.d3 (approx).
	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 6",
	// Ruy Lopez, Morphy Defense main line after 3...a6 4.Ba4.
	"r1bqkbnr/1ppp1ppp/p1n5/4p3/B3P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4",
	// Open Sicilian after 3.Nf3 d6.
	"rnbqkb1r/pp2pppp/3p1n2/2p5/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 0 4",

	// --- tactical middlegames (mix of white and black to move) ---
	// Classic mating-net tactic: black to move, ...Qg3! wins (hg3 Bxe1+ etc).
	"1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
	// Closed middlegame with opposite-side attacks brewing, white to move.
	"r1bq1rk1/ppp2ppp/2np1n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQ1RK1 w - - 0 8",

	// --- plain endgames ---
	// King and rook vs king, white to move.
	"8/8/8/8/8/2k5/8/R3K3 w - - 0 1",
	// Bishop vs knight minor-piece endgame, white to move.
	"8/8/4k3/8/3nK3/8/4B3/8 w - - 0 1",
	// King, bishop, and pawn vs king, white to move.
	"8/8/8/3k4/8/3K4/3P4/3B4 w - - 0 1",

	// --- check / near-mate ---
	// Black king in check from a rook on an open file; not mate (Kd8/Kf8/
	// Kd7/Kf7 all escape). Exercises search starting from an in-check root.
	"4k3/8/8/8/8/8/4R3/4K3 b - - 0 1",
	// One move from a back-rank mate: white to move, Ra1-a8# is forced mate
	// in 1 (black king boxed in by its own f7/g7/h7 pawns).
	"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
}

// TestSearchGolden is the acceptance gate for search-internals refactors: it
// searches a fixed set of ~20 positions at a fixed depth (5) with a fixed
// transposition table size and no time cutoff, and asserts that
// (bestmove, score, nodes) for every position exactly match a checked-in
// golden file (testdata/search_golden.json).
//
// Determinism at fixed depth - the precondition this test relies on - is
// established by TestSearchDeterminism_ClearSearchStateResetsAllState in
// determinism_test.go: identical (BestMove, Score, NodesSearched) across
// repeated searches of the same position on the same engine, provided
// ClearSearchState() is called between searches. This test follows exactly
// that pattern (one shared engine, ClearSearchState() before each position,
// a fresh board parsed from FEN each time) so that node counts are stable
// both across repeated runs and across the order in which positions are
// searched.
//
// Any refactor that changes search behavior in an observable way (move
// ordering, pruning conditions, evaluation, etc.) is expected to change
// NodesSearched and/or the chosen move for at least one of these 20
// positions; regenerate the golden file with:
//
//	go test ./internal/search -run TestSearchGolden -update
//
// and manually confirm the new node counts/moves are the intended effect of
// the refactor before committing the updated golden file.
func TestSearchGolden(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(64)

	var golden []goldenRecord
	if !*update {
		golden = loadGoldenFile(t)
		if len(golden) != len(goldenTestFENs) {
			t.Fatalf("golden file has %d records but goldenTestFENs has %d entries; regenerate with -update", len(golden), len(goldenTestFENs))
		}
	}

	results := make([]goldenRecord, len(goldenTestFENs))

	for idx, fen := range goldenTestFENs {
		idx, fen := idx, fen
		t.Run(fmt.Sprintf("fen%d", idx), func(t *testing.T) {
			engine.ClearSearchState()
			b := testutil.MustFromFEN(t, fen)
			player := playerFromSide(b.GetSideToMove())
			result := engine.FindBestMove(context.Background(), b, player, goldenSearchConfig)

			rec := goldenRecord{
				FEN:      fen,
				Depth:    goldenDepth,
				BestMove: moveString(result.BestMove),
				Score:    int(result.Score),
				Nodes:    result.Stats.NodesSearched,
			}
			results[idx] = rec

			if *update {
				return
			}

			want := golden[idx]
			if want.FEN != rec.FEN {
				t.Fatalf("golden record %d is for a different FEN than goldenTestFENs[%d]:\n  golden file: %q\n  test slice:  %q\nregenerate the golden file with -update", idx, idx, want.FEN, rec.FEN)
			}
			if want != rec {
				t.Errorf("fen=%q: golden mismatch\n  golden: depth=%d bestmove=%s score=%d nodes=%d\n  got:    depth=%d bestmove=%s score=%d nodes=%d",
					rec.FEN,
					want.Depth, want.BestMove, want.Score, want.Nodes,
					rec.Depth, rec.BestMove, rec.Score, rec.Nodes,
				)
			}
		})
	}

	if *update {
		writeGoldenFile(t, results)
		t.Logf("wrote %d golden records to %s", len(results), goldenFilePath)
	}
}

// loadGoldenFile reads and decodes the golden JSON file, failing the test if
// it is missing or malformed.
func loadGoldenFile(t *testing.T) []goldenRecord {
	t.Helper()

	data, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("reading golden file %s: %v (run with -update to generate it)", goldenFilePath, err)
	}

	var records []goldenRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("parsing golden file %s: %v", goldenFilePath, err)
	}
	return records
}

// writeGoldenFile encodes records as indented, human-diffable JSON and
// writes them to the golden file path.
func writeGoldenFile(t *testing.T, records []goldenRecord) {
	t.Helper()

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatalf("marshaling golden records: %v", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(goldenFilePath, data, 0o644); err != nil {
		t.Fatalf("writing golden file %s: %v", goldenFilePath, err)
	}
}
