package search

import (
	"encoding/json"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// SearchStats tracks statistics during search
type SearchStats struct {
	NodesSearched      int64
	Depth              int
	Time               time.Duration
	PrincipalVariation []board.Move
	BookMoveUsed       bool // True if move came from opening book

	// Late Move Reductions (LMR) statistics
	LMRReductions   int64 // Number of moves reduced
	LMRReSearches   int64 // Number of re-searches performed
	LMRNodesSkipped int64 // Estimated nodes saved by LMR

	// Null move pruning statistics
	NullMoves   int64 // Number of null move attempts
	NullCutoffs int64 // Number of successful null move cutoffs

	QNodes           int64 // Quiescence search nodes
	TTCutoffs        int64 // Beta cutoffs from transposition table
	FirstMoveCutoffs int64 // Beta cutoffs on first move tried
	TotalCutoffs     int64 // Total beta cutoffs (for move ordering calculation)
	DeltaPruned      int64 // Captures skipped by delta pruning

	// Razoring statistics
	RazoringAttempts int64 // Number of razoring attempts
	RazoringCutoffs  int64 // Successful razoring cutoffs
	RazoringFailed   int64 // Razoring attempts that failed verification

	// Move ordering effectiveness
	CutoffsByMoveIndex [64]int64 // Histogram of which move caused beta cutoff

	// Transposition table effectiveness
	TTProbes int64 // Total TT lookups attempted
	TTHits   int64 // Successful TT hits

	// Effective branching factor calculation
	NodesByDepth [100]int64 // Nodes searched at each depth from root

	// Node type distribution
	PVNodes  int64 // Principal variation nodes (best move found)
	CutNodes int64 // Nodes that caused beta cutoff
	AllNodes int64 // Nodes where all moves were searched
}

// EBF returns the effective branching factor, computed as the mean of the
// per-depth node-count ratios NodesByDepth[d]/NodesByDepth[d-1] over every
// depth d where both NodesByDepth[d] and NodesByDepth[d-1] are nonzero. This
// is the standard "average ratio of successive depths" approximation to the
// branching factor (as opposed to the alternative geometric-mean-over-total
// definition, b such that NodesSearched = b^Depth). Depths with a zero
// numerator or denominator are skipped rather than treated as zero, since a
// missing sample (search didn't reach that depth, or reported no nodes)
// should not be interpreted as "no branching" and drag the average down.
// Returns 0 if there are no eligible depth pairs (e.g. fewer than two
// non-zero depths were recorded), never NaN or Inf.
func (s *SearchStats) EBF() float64 {
	var sum float64
	var count int
	for d := 1; d < len(s.NodesByDepth); d++ {
		prev := s.NodesByDepth[d-1]
		curr := s.NodesByDepth[d]
		if prev == 0 || curr == 0 {
			continue
		}
		sum += float64(curr) / float64(prev)
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// TTHitRate returns the fraction of transposition table probes that were
// hits (TTHits/TTProbes). Returns 0 if there were no probes.
func (s *SearchStats) TTHitRate() float64 {
	if s.TTProbes == 0 {
		return 0
	}
	return float64(s.TTHits) / float64(s.TTProbes)
}

// OrderingQuality returns the fraction of beta cutoffs that occurred on the
// first move tried (FirstMoveCutoffs/TotalCutoffs), a standard proxy for
// move-ordering effectiveness. Returns 0 if there were no cutoffs.
func (s *SearchStats) OrderingQuality() float64 {
	if s.TotalCutoffs == 0 {
		return 0
	}
	return float64(s.FirstMoveCutoffs) / float64(s.TotalCutoffs)
}

// NullMoveEfficiency returns the fraction of null move attempts that
// produced a cutoff (NullCutoffs/NullMoves). Returns 0 if no null moves were
// attempted.
func (s *SearchStats) NullMoveEfficiency() float64 {
	if s.NullMoves == 0 {
		return 0
	}
	return float64(s.NullCutoffs) / float64(s.NullMoves)
}

// searchStatsJSON mirrors SearchStats for marshaling: it fixes the field
// order (struct field order, not map iteration order, so json.Marshal output
// is deterministic and diff-stable) and adds the derived metrics alongside
// the raw counters. Field names are snake_case per the task brief.
type searchStatsJSON struct {
	NodesSearched      int64    `json:"nodes_searched"`
	Depth              int      `json:"depth"`
	TimeMS             int64    `json:"time_ms"`
	PrincipalVariation []string `json:"principal_variation"`
	BookMoveUsed       bool     `json:"book_move_used"`

	LMRReductions   int64 `json:"lmr_reductions"`
	LMRReSearches   int64 `json:"lmr_re_searches"`
	LMRNodesSkipped int64 `json:"lmr_nodes_skipped"`

	NullMoves   int64 `json:"null_moves"`
	NullCutoffs int64 `json:"null_cutoffs"`

	QNodes           int64 `json:"q_nodes"`
	TTCutoffs        int64 `json:"tt_cutoffs"`
	FirstMoveCutoffs int64 `json:"first_move_cutoffs"`
	TotalCutoffs     int64 `json:"total_cutoffs"`
	DeltaPruned      int64 `json:"delta_pruned"`

	RazoringAttempts int64 `json:"razoring_attempts"`
	RazoringCutoffs  int64 `json:"razoring_cutoffs"`
	RazoringFailed   int64 `json:"razoring_failed"`

	CutoffsByMoveIndex []int64 `json:"cutoffs_by_move_index"`

	TTProbes int64 `json:"tt_probes"`
	TTHits   int64 `json:"tt_hits"`

	NodesByDepth []int64 `json:"nodes_by_depth"`

	PVNodes  int64 `json:"pv_nodes"`
	CutNodes int64 `json:"cut_nodes"`
	AllNodes int64 `json:"all_nodes"`

	EBF                float64 `json:"ebf"`
	TTHitRate          float64 `json:"tt_hit_rate"`
	OrderingQuality    float64 `json:"ordering_quality"`
	NullMoveEfficiency float64 `json:"null_move_efficiency"`
}

// nonZeroPrefix returns arr truncated after the last non-zero element (empty
// if arr is entirely zero). NodesByDepth and CutoffsByMoveIndex are fixed-size
// arrays sized for the deepest/widest search the engine could ever run, so
// most entries are trailing zeros for any given search; truncating to the
// non-zero prefix keeps the JSON output compact and its length meaningful
// (e.g. len(nodes_by_depth) approximates the depth reached) instead of always
// emitting 100 or 64 entries.
func nonZeroPrefix(arr []int64) []int64 {
	last := -1
	for i, v := range arr {
		if v != 0 {
			last = i
		}
	}
	if last < 0 {
		return []int64{}
	}
	out := make([]int64, last+1)
	copy(out, arr[:last+1])
	return out
}

// MarshalJSON produces a stable snapshot of s: the raw counters in a fixed
// field order plus the derived metrics (EBF, TTHitRate, OrderingQuality,
// NullMoveEfficiency). NodesByDepth and CutoffsByMoveIndex are truncated to
// their non-zero prefix (see nonZeroPrefix); PrincipalVariation is rendered
// as UCI move strings (e.g. "e2e4", "e7e8q").
//
// This uses a value receiver (rather than pointer) so that json.Marshal
// invokes it even when SearchStats is embedded by value in another struct
// (e.g. SearchResult.Stats) and the top-level value passed to json.Marshal
// is not itself a pointer — encoding/json only auto-takes the address of
// addressable values, and a struct field of a non-pointer argument to
// json.Marshal is not addressable.
func (s SearchStats) MarshalJSON() ([]byte, error) {
	pv := make([]string, len(s.PrincipalVariation))
	for i, m := range s.PrincipalVariation {
		pv[i] = moveToUCI(m)
	}

	snapshot := searchStatsJSON{
		NodesSearched:      s.NodesSearched,
		Depth:              s.Depth,
		TimeMS:             s.Time.Milliseconds(),
		PrincipalVariation: pv,
		BookMoveUsed:       s.BookMoveUsed,

		LMRReductions:   s.LMRReductions,
		LMRReSearches:   s.LMRReSearches,
		LMRNodesSkipped: s.LMRNodesSkipped,

		NullMoves:   s.NullMoves,
		NullCutoffs: s.NullCutoffs,

		QNodes:           s.QNodes,
		TTCutoffs:        s.TTCutoffs,
		FirstMoveCutoffs: s.FirstMoveCutoffs,
		TotalCutoffs:     s.TotalCutoffs,
		DeltaPruned:      s.DeltaPruned,

		RazoringAttempts: s.RazoringAttempts,
		RazoringCutoffs:  s.RazoringCutoffs,
		RazoringFailed:   s.RazoringFailed,

		CutoffsByMoveIndex: nonZeroPrefix(s.CutoffsByMoveIndex[:]),

		TTProbes: s.TTProbes,
		TTHits:   s.TTHits,

		NodesByDepth: nonZeroPrefix(s.NodesByDepth[:]),

		PVNodes:  s.PVNodes,
		CutNodes: s.CutNodes,
		AllNodes: s.AllNodes,

		EBF:                s.EBF(),
		TTHitRate:          s.TTHitRate(),
		OrderingQuality:    s.OrderingQuality(),
		NullMoveEfficiency: s.NullMoveEfficiency(),
	}

	return json.Marshal(snapshot)
}

// moveToUCI renders m in UCI long algebraic notation (e.g. "e2e4", "e7e8q"),
// matching the format used by internal/uci.MoveConverter.ToUCI. It is
// reimplemented locally (rather than imported) because internal/uci imports
// internal/search, so importing internal/uci here would create a cycle.
func moveToUCI(m board.Move) string {
	uciMove := m.From.String() + m.To.String()
	if m.Promotion != board.Empty {
		uciMove += string(toLowerRune(rune(m.Promotion)))
	}
	return uciMove
}

// toLowerRune lowercases a single ASCII letter rune. Promotion pieces are
// stored uppercase (e.g. 'Q'); UCI notation requires lowercase (e.g. 'q').
func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
