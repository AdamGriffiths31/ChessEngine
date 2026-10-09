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

// update regenerates testdata/search_golden.json from the engine's current
// output: go test ./internal/search -run TestSearchGolden -update
var update = flag.Bool("update", false, "regenerate the golden file (testdata/search_golden.json) from current engine output")

const goldenFilePath = "testdata/search_golden.json"

// goldenDepth is deep enough to exercise LMR and aspiration windows, and
// shallow enough to finish in seconds.
const goldenDepth = 5

// goldenSearchConfig is depth-bounded only (no time cutoff, no opening book).
var goldenSearchConfig = Config{
	MaxDepth:       goldenDepth,
	MaxTime:        0,
	UseOpeningBook: false,
}

// goldenRecord is one golden search result; field order fixes the JSON key order.
type goldenRecord struct {
	FEN      string `json:"fen"`
	Depth    int    `json:"depth"`
	BestMove string `json:"bestmove"`
	Score    int    `json:"score"`
	Nodes    int64  `json:"nodes"`
}

// goldenTestFENs is a fixed, ordered set of 20 positions spanning opening,
// middlegame, tactical, endgame, in-check and near-mate. Never reorder: the
// test compares index-for-index. Append new positions and regenerate instead.
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

// TestSearchGolden is the acceptance gate for search and eval changes. It
// searches every golden position at depth 5 and requires an exact match of
// (bestmove, score, nodes) with the golden file, so ANY change to search or
// eval behaviour fails it. After confirming the new values are the intended
// effect, regenerate with -update and commit the golden file with the change.
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

	if err := os.WriteFile(goldenFilePath, data, 0o600); err != nil {
		t.Fatalf("writing golden file %s: %v", goldenFilePath, err)
	}
}
