package search

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

func TestSearchStatsEBF(t *testing.T) {
	// Iteration.Nodes is cumulative; the helper builds iterations from the
	// cumulative totals at the end of each one.
	iterations := func(cumulative ...int64) []Iteration {
		its := make([]Iteration, len(cumulative))
		for i, n := range cumulative {
			its[i] = Iteration{Depth: i + 1, Nodes: n}
		}
		return its
	}

	tests := []struct {
		name string
		its  []Iteration
		want float64
	}{
		{name: "no iterations", want: 0},
		{name: "single iteration - no ratio possible", its: iterations(20), want: 0},
		{
			name: "two iterations, exact ratio",
			// own nodes 20 then 100: 100/20 = 5
			its:  iterations(20, 120),
			want: 5,
		},
		{
			name: "uneven ratios averaged",
			// own nodes 20, 40, 400: (40/20 + 400/40) / 2 = 6
			its:  iterations(20, 60, 460),
			want: 6,
		},
		{
			name: "iteration with no nodes of its own is skipped",
			// own nodes 20, 0, 100: the empty iteration is ignored, 100/20 = 5
			its:  iterations(20, 20, 120),
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{Iterations: tt.its}
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

func TestTopRootMoves(t *testing.T) {
	const unsearched = -eval.MateScore - 1
	moves := &movegen.MoveList{
		Moves: []board.Move{
			{From: sq(0, 1), To: sq(0, 2)},
			{From: sq(1, 1), To: sq(1, 2)},
			{From: sq(2, 1), To: sq(2, 2)},
			{From: sq(3, 1), To: sq(3, 2)},
			{From: sq(4, 1), To: sq(4, 2)},
		},
		Count: 5,
	}
	scores := []eval.EvaluationScore{50, 120, unsearched, 120, 10}
	exact := []bool{true, true, false, true, false}

	got := topRootMoves(moves, scores, exact, 3)

	want := []RootMove{
		{Move: moves.Moves[1], Score: 120},
		{Move: moves.Moves[3], Score: 120}, // tie keeps root order
		{Move: moves.Moves[0], Score: 50},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d moves, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("rank %d = %+v, want %+v", i, got[i], want[i])
		}
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
		`"stop_reason":"","iterations":[],"root_moves":[],` +
		`"lmr_reductions":0,"lmr_re_searches":0,"null_moves":0,"null_cutoffs":0,` +
		`"null_move_zugzwang_skipped":0,"null_move_sparse_skipped":0,` +
		`"q_nodes":0,"tt_cutoffs":0,"first_move_cutoffs":0,"total_cutoffs":0,"delta_pruned":0,` +
		`"rfp_cutoffs":0,` +
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
