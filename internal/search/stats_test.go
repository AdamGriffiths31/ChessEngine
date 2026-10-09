package search

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

func TestSearchStatsEBF(t *testing.T) {
	tests := []struct {
		name  string
		nodes [100]int64
		want  float64
	}{
		{
			name: "zero value - all depths empty",
			want: 0,
		},
		{
			name: "single depth recorded - no ratio possible",
			nodes: [100]int64{
				0: 20,
			},
			want: 0,
		},
		{
			name: "two depths, exact ratio",
			// depth1/depth0 = 100/20 = 5
			nodes: [100]int64{
				0: 20,
				1: 100,
			},
			want: 5,
		},
		{
			name: "three depths, mean of two ratios",
			// depth1/depth0 = 100/20 = 5
			// depth2/depth1 = 500/100 = 5
			// mean = 5
			nodes: [100]int64{
				0: 20,
				1: 100,
				2: 500,
			},
			want: 5,
		},
		{
			name: "uneven ratios averaged",
			// depth1/depth0 = 40/20 = 2
			// depth2/depth1 = 400/40 = 10
			// mean = (2+10)/2 = 6
			nodes: [100]int64{
				0: 20,
				1: 40,
				2: 400,
			},
			want: 6,
		},
		{
			name: "gap in the middle is skipped, not treated as zero",
			// depth0=20, depth1=0 (missing), depth2=200
			// only eligible pair would need both prev and curr nonzero;
			// (1,2) has prev=0 skipped, (0,1) has curr=0 skipped.
			// So no eligible pairs at all -> 0.
			nodes: [100]int64{
				0: 20,
				2: 200,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{NodesByDepth: tt.nodes}
			got := s.EBF()
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("EBF() = %v, want a finite number", got)
			}
			if got != tt.want {
				t.Errorf("EBF() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The derived ratios must return 0, never NaN/Inf, when the denominator is
// zero (including degenerate inputs where the numerator is nonzero).
func TestSearchStatsRatios(t *testing.T) {
	tests := []struct {
		name string
		s    Stats
		fn   func(*Stats) float64
		want float64
	}{
		{"tt_hit_rate zero probes", Stats{}, (*Stats).TTHitRate, 0},
		{"tt_hit_rate zero probes nonzero hits", Stats{TTHits: 5}, (*Stats).TTHitRate, 0},
		{"tt_hit_rate quarter", Stats{TTProbes: 200, TTHits: 50}, (*Stats).TTHitRate, 0.25},
		{"ordering_quality zero cutoffs", Stats{}, (*Stats).OrderingQuality, 0},
		{"ordering_quality zero total nonzero first", Stats{FirstMoveCutoffs: 3}, (*Stats).OrderingQuality, 0},
		{"ordering_quality perfect", Stats{FirstMoveCutoffs: 40, TotalCutoffs: 40}, (*Stats).OrderingQuality, 1},
		{"ordering_quality quarter", Stats{FirstMoveCutoffs: 10, TotalCutoffs: 40}, (*Stats).OrderingQuality, 0.25},
		{"null_move_efficiency zero attempts", Stats{}, (*Stats).NullMoveEfficiency, 0},
		{"null_move_efficiency zero attempts nonzero cutoffs", Stats{NullCutoffs: 2}, (*Stats).NullMoveEfficiency, 0},
		{"null_move_efficiency quarter", Stats{NullMoves: 8, NullCutoffs: 2}, (*Stats).NullMoveEfficiency, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(&tt.s)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("got %v, want a finite number", got)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNonZeroPrefix(t *testing.T) {
	tests := []struct {
		name string
		in   []int64
		want []int64
	}{
		{name: "empty slice", in: []int64{}, want: []int64{}},
		{name: "all zero", in: []int64{0, 0, 0}, want: []int64{}},
		{name: "no trailing zero", in: []int64{1, 2, 3}, want: []int64{1, 2, 3}},
		{name: "trailing zeros truncated", in: []int64{1, 2, 0, 0}, want: []int64{1, 2}},
		{name: "leading zero kept if followed by nonzero", in: []int64{0, 5, 0, 3, 0, 0}, want: []int64{0, 5, 0, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonZeroPrefix(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("nonZeroPrefix(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("nonZeroPrefix(%v)[%d] = %v, want %v", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestSearchStatsMarshalJSONFieldOrder locks in the stable field order of the
// marshaled JSON so an accidental switch back to map-based marshaling (which
// would randomize key order across runs) is caught by a plain byte-for-byte
// diff.
func TestSearchStatsMarshalJSONFieldOrder(t *testing.T) {
	var s Stats
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	want := `{"nodes_searched":0,"depth":0,"time_ms":0,"principal_variation":[],"book_move_used":false,` +
		`"lmr_reductions":0,"lmr_re_searches":0,"lmr_nodes_skipped":0,"null_moves":0,"null_cutoffs":0,` +
		`"null_move_zugzwang_skipped":0,` +
		`"q_nodes":0,"tt_cutoffs":0,"first_move_cutoffs":0,"total_cutoffs":0,"delta_pruned":0,` +
		`"razoring_attempts":0,"razoring_cutoffs":0,"razoring_failed":0,` +
		`"futility_prunes":0,"lmp_prunes":0,"cutoffs_by_move_index":[],` +
		`"tt_probes":0,"tt_hits":0,"nodes_by_depth":[],"pv_nodes":0,"cut_nodes":0,"all_nodes":0,` +
		`"ebf":0,"tt_hit_rate":0,"ordering_quality":0,"null_move_efficiency":0}`

	if string(data) != want {
		t.Errorf("MarshalJSON field order/content changed:\n got: %s\nwant: %s", string(data), want)
	}
}

// Stats must describe only the latest FindBestMove call. The UCI adapter
// reuses one engine across a game's moves without ClearSearchState, and the
// counters used to accumulate across calls.
//
// Two different positions are searched: repeating one would let the warm TT
// legitimately shrink the second node count. A 10% tolerance absorbs the history
// table carrying over; the bug would add the first search's whole node count.
func TestFindBestMove_StatsAreNotCumulativeAcrossCalls(t *testing.T) {
	cfg := Config{MaxDepth: 5, UseOpeningBook: false}
	search := func(engine *MinimaxEngine, fen string) Result {
		b := testutil.MustFromFEN(t, fen)
		return engine.FindBestMove(context.Background(), b, playerFromSide(b.GetSideToMove()), cfg)
	}
	newEngine := func() *MinimaxEngine {
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(64)
		return engine
	}

	first, second := zobristTestFENs[0], zobristTestFENs[1]

	reused := newEngine()
	firstResult := search(reused, first)
	if firstResult.Stats.NodesSearched <= 0 {
		t.Fatalf("first search reported %d nodes, want a positive count", firstResult.Stats.NodesSearched)
	}
	reusedNodes := search(reused, second).Stats.NodesSearched
	freshNodes := search(newEngine(), second).Stats.NodesSearched

	diff := reusedNodes - freshNodes
	if diff < 0 {
		diff = -diff
	}
	if tolerance := freshNodes / 10; diff > tolerance {
		t.Fatalf("reused engine reported %d nodes, a fresh engine %d (tolerance %d); the first search used %d, so stats are bleeding across calls",
			reusedNodes, freshNodes, tolerance, firstResult.Stats.NodesSearched)
	}
}
