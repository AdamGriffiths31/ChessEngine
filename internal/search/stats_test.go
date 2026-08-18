package search

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
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
			s := &SearchStats{NodesByDepth: tt.nodes}
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

func TestSearchStatsTTHitRate(t *testing.T) {
	tests := []struct {
		name     string
		ttProbes int64
		ttHits   int64
		want     float64
	}{
		{name: "zero probes", ttProbes: 0, ttHits: 0, want: 0},
		{name: "zero probes with nonzero hits (degenerate input)", ttProbes: 0, ttHits: 5, want: 0},
		{name: "all misses", ttProbes: 100, ttHits: 0, want: 0},
		{name: "all hits", ttProbes: 100, ttHits: 100, want: 1},
		{name: "half hits", ttProbes: 200, ttHits: 50, want: 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SearchStats{TTProbes: tt.ttProbes, TTHits: tt.ttHits}
			got := s.TTHitRate()
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("TTHitRate() = %v, want a finite number", got)
			}
			if got != tt.want {
				t.Errorf("TTHitRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchStatsOrderingQuality(t *testing.T) {
	tests := []struct {
		name             string
		firstMoveCutoffs int64
		totalCutoffs     int64
		want             float64
	}{
		{name: "zero cutoffs", firstMoveCutoffs: 0, totalCutoffs: 0, want: 0},
		{name: "zero total with nonzero first-move (degenerate input)", firstMoveCutoffs: 3, totalCutoffs: 0, want: 0},
		{name: "perfect ordering", firstMoveCutoffs: 40, totalCutoffs: 40, want: 1},
		{name: "quarter first-move", firstMoveCutoffs: 10, totalCutoffs: 40, want: 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SearchStats{FirstMoveCutoffs: tt.firstMoveCutoffs, TotalCutoffs: tt.totalCutoffs}
			got := s.OrderingQuality()
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("OrderingQuality() = %v, want a finite number", got)
			}
			if got != tt.want {
				t.Errorf("OrderingQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchStatsNullMoveEfficiency(t *testing.T) {
	tests := []struct {
		name        string
		nullMoves   int64
		nullCutoffs int64
		want        float64
	}{
		{name: "zero attempts", nullMoves: 0, nullCutoffs: 0, want: 0},
		{name: "zero attempts with nonzero cutoffs (degenerate input)", nullMoves: 0, nullCutoffs: 2, want: 0},
		{name: "all cutoffs", nullMoves: 10, nullCutoffs: 10, want: 1},
		{name: "no cutoffs", nullMoves: 10, nullCutoffs: 0, want: 0},
		{name: "partial", nullMoves: 8, nullCutoffs: 2, want: 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SearchStats{NullMoves: tt.nullMoves, NullCutoffs: tt.nullCutoffs}
			got := s.NullMoveEfficiency()
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("NullMoveEfficiency() = %v, want a finite number", got)
			}
			if got != tt.want {
				t.Errorf("NullMoveEfficiency() = %v, want %v", got, tt.want)
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

func TestSearchStatsMarshalJSON(t *testing.T) {
	s := SearchStats{
		NodesSearched: 12345,
		Depth:         6,
		Time:          2500 * time.Millisecond,
		PrincipalVariation: []board.Move{
			{From: board.Square{File: 4, Rank: 1}, To: board.Square{File: 4, Rank: 3}, Promotion: board.Empty},
			{From: board.Square{File: 4, Rank: 6}, To: board.Square{File: 4, Rank: 7}, Promotion: 'Q'},
		},
		BookMoveUsed:     false,
		LMRReductions:    10,
		LMRReSearches:    2,
		LMRNodesSkipped:  500,
		NullMoves:        8,
		NullCutoffs:      2,
		QNodes:           300,
		TTCutoffs:        15,
		FirstMoveCutoffs: 10,
		TotalCutoffs:     40,
		DeltaPruned:      5,
		RazoringAttempts: 4,
		RazoringCutoffs:  3,
		RazoringFailed:   1,
		TTProbes:         200,
		TTHits:           50,
		PVNodes:          100,
		CutNodes:         40,
		AllNodes:         60,
	}
	s.CutoffsByMoveIndex[0] = 10
	s.CutoffsByMoveIndex[1] = 20
	s.NodesByDepth[0] = 20
	s.NodesByDepth[1] = 100

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	// Raw counters round-trip.
	if decoded["nodes_searched"] != float64(12345) {
		t.Errorf("nodes_searched = %v, want 12345", decoded["nodes_searched"])
	}
	if decoded["depth"] != float64(6) {
		t.Errorf("depth = %v, want 6", decoded["depth"])
	}
	if decoded["time_ms"] != float64(2500) {
		t.Errorf("time_ms = %v, want 2500", decoded["time_ms"])
	}
	if decoded["tt_probes"] != float64(200) {
		t.Errorf("tt_probes = %v, want 200", decoded["tt_probes"])
	}
	if decoded["tt_hits"] != float64(50) {
		t.Errorf("tt_hits = %v, want 50", decoded["tt_hits"])
	}

	// PrincipalVariation as UCI strings.
	pv, ok := decoded["principal_variation"].([]interface{})
	if !ok || len(pv) != 2 {
		t.Fatalf("principal_variation = %v, want a 2-element array", decoded["principal_variation"])
	}
	if pv[0] != "e2e4" {
		t.Errorf("principal_variation[0] = %v, want e2e4", pv[0])
	}
	if pv[1] != "e7e8q" {
		t.Errorf("principal_variation[1] = %v, want e7e8q", pv[1])
	}

	// Non-zero-prefix truncation of the histogram arrays.
	cutoffs, ok := decoded["cutoffs_by_move_index"].([]interface{})
	if !ok || len(cutoffs) != 2 {
		t.Fatalf("cutoffs_by_move_index = %v, want length 2", decoded["cutoffs_by_move_index"])
	}
	nodesByDepth, ok := decoded["nodes_by_depth"].([]interface{})
	if !ok || len(nodesByDepth) != 2 {
		t.Fatalf("nodes_by_depth = %v, want length 2", decoded["nodes_by_depth"])
	}

	// Derived fields present and correct.
	wantEBF := s.EBF()
	if decoded["ebf"] != wantEBF {
		t.Errorf("ebf = %v, want %v", decoded["ebf"], wantEBF)
	}
	wantTTHitRate := s.TTHitRate()
	if decoded["tt_hit_rate"] != wantTTHitRate {
		t.Errorf("tt_hit_rate = %v, want %v", decoded["tt_hit_rate"], wantTTHitRate)
	}
	wantOrderingQuality := s.OrderingQuality()
	if decoded["ordering_quality"] != wantOrderingQuality {
		t.Errorf("ordering_quality = %v, want %v", decoded["ordering_quality"], wantOrderingQuality)
	}
	wantNullMoveEfficiency := s.NullMoveEfficiency()
	if decoded["null_move_efficiency"] != wantNullMoveEfficiency {
		t.Errorf("null_move_efficiency = %v, want %v", decoded["null_move_efficiency"], wantNullMoveEfficiency)
	}
}

func TestSearchStatsMarshalJSONZeroValue(t *testing.T) {
	var s SearchStats

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded["ebf"] != float64(0) {
		t.Errorf("ebf = %v, want 0", decoded["ebf"])
	}
	if decoded["tt_hit_rate"] != float64(0) {
		t.Errorf("tt_hit_rate = %v, want 0", decoded["tt_hit_rate"])
	}
	if decoded["ordering_quality"] != float64(0) {
		t.Errorf("ordering_quality = %v, want 0", decoded["ordering_quality"])
	}
	if decoded["null_move_efficiency"] != float64(0) {
		t.Errorf("null_move_efficiency = %v, want 0", decoded["null_move_efficiency"])
	}
	pv, ok := decoded["principal_variation"].([]interface{})
	if !ok || len(pv) != 0 {
		t.Errorf("principal_variation = %v, want empty array", decoded["principal_variation"])
	}
	cutoffs, ok := decoded["cutoffs_by_move_index"].([]interface{})
	if !ok || len(cutoffs) != 0 {
		t.Errorf("cutoffs_by_move_index = %v, want empty array", decoded["cutoffs_by_move_index"])
	}
	nodesByDepth, ok := decoded["nodes_by_depth"].([]interface{})
	if !ok || len(nodesByDepth) != 0 {
		t.Errorf("nodes_by_depth = %v, want empty array", decoded["nodes_by_depth"])
	}
}

// TestSearchStatsMarshalJSONFieldOrder locks in the stable field order of the
// marshaled JSON so an accidental switch back to map-based marshaling (which
// would randomize key order across runs) is caught by a plain byte-for-byte
// diff.
func TestSearchStatsMarshalJSONFieldOrder(t *testing.T) {
	var s SearchStats
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	want := `{"nodes_searched":0,"depth":0,"time_ms":0,"principal_variation":[],"book_move_used":false,` +
		`"lmr_reductions":0,"lmr_re_searches":0,"lmr_nodes_skipped":0,"null_moves":0,"null_cutoffs":0,` +
		`"q_nodes":0,"tt_cutoffs":0,"first_move_cutoffs":0,"total_cutoffs":0,"delta_pruned":0,` +
		`"razoring_attempts":0,"razoring_cutoffs":0,"razoring_failed":0,"cutoffs_by_move_index":[],` +
		`"tt_probes":0,"tt_hits":0,"nodes_by_depth":[],"pv_nodes":0,"cut_nodes":0,"all_nodes":0,` +
		`"ebf":0,"tt_hit_rate":0,"ordering_quality":0,"null_move_efficiency":0}`

	if string(data) != want {
		t.Errorf("MarshalJSON field order/content changed:\n got: %s\nwant: %s", string(data), want)
	}
}

// TestSearchStatsMarshalJSONValueVsPointer verifies MarshalJSON fires
// correctly whether SearchStats is marshaled by value or via a pointer, and
// when embedded (by value) inside another struct passed to json.Marshal by
// value -- the case that motivated the value receiver (see MarshalJSON's
// doc comment).
func TestSearchStatsMarshalJSONValueVsPointer(t *testing.T) {
	s := SearchStats{NodesSearched: 7}

	byValue, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal(value) returned error: %v", err)
	}
	byPointer, err := json.Marshal(&s)
	if err != nil {
		t.Fatalf("json.Marshal(pointer) returned error: %v", err)
	}
	if string(byValue) != string(byPointer) {
		t.Errorf("value and pointer marshal differ:\nvalue:   %s\npointer: %s", byValue, byPointer)
	}

	type wrapper struct {
		Stats SearchStats
	}
	w := wrapper{Stats: s}
	wrapped, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("json.Marshal(wrapper value) returned error: %v", err)
	}
	if !containsCustomMarshal(wrapped) {
		t.Errorf("wrapper.Stats did not use custom MarshalJSON, got: %s", wrapped)
	}
}

// containsCustomMarshal checks for a derived-metric key that only appears if
// SearchStats.MarshalJSON actually ran (as opposed to falling back to
// encoding/json's default struct reflection, which would use Go field names
// and include unexported fields differently).
func containsCustomMarshal(data []byte) bool {
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return false
	}
	stats, ok := decoded["Stats"].(map[string]interface{})
	if !ok {
		return false
	}
	_, hasEBF := stats["ebf"]
	return hasEBF
}
