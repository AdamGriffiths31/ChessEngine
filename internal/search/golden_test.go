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

var update = flag.Bool("update", false, "regenerate the golden file (testdata/search_golden.json) from current engine output")

const goldenFilePath = "testdata/search_golden.json"

const goldenDepth = 5

var goldenSearchConfig = Config{
	MaxDepth:       goldenDepth,
	MaxTime:        0,
	UseOpeningBook: false,
}

type goldenRecord struct {
	FEN      string `json:"fen"`
	Depth    int    `json:"depth"`
	BestMove string `json:"bestmove"`
	Score    int    `json:"score"`
	Nodes    int64  `json:"nodes"`
}

var goldenTestFENs = []string{
	zobristTestFENs[0],
	zobristTestFENs[1],
	zobristTestFENs[2],
	zobristTestFENs[3],
	zobristTestFENs[4],
	zobristTestFENs[5],
	zobristTestFENs[6],
	zobristTestFENs[7],
	zobristTestFENs[8],
	zobristTestFENs[9],

	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 6",
	"r1bqkbnr/1ppp1ppp/p1n5/4p3/B3P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4",
	"rnbqkb1r/pp2pppp/3p1n2/2p5/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 0 4",

	"1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
	"r1bq1rk1/ppp2ppp/2np1n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQ1RK1 w - - 0 8",

	"8/8/8/8/8/2k5/8/R3K3 w - - 0 1",
	"8/8/4k3/8/3nK3/8/4B3/8 w - - 0 1",
	"8/8/8/3k4/8/3K4/3P4/3B4 w - - 0 1",

	"4k3/8/8/8/8/8/4R3/4K3 b - - 0 1",
	"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
}

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
