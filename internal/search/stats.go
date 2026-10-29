package search

import (
	"encoding/json"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

// Reasons a search stopped, reported in Stats.StopReason.
const (
	// StopReasonDepth means every iteration up to Config.MaxDepth completed.
	StopReasonDepth = "depth"
	// StopReasonTime means the time budget (Config.MaxTime or the context
	// deadline) ran out.
	StopReasonTime = "time"
	// StopReasonMate means a mate in one was found, so deeper search was skipped.
	StopReasonMate = "mate"
	// StopReasonCancelled means the context was cancelled (e.g. a UCI "stop").
	StopReasonCancelled = "cancelled"
	// StopReasonNoLegalMoves means the root position has no legal moves.
	StopReasonNoLegalMoves = "no_legal_moves"
	// StopReasonBook means the move came from the opening book, with no search.
	StopReasonBook = "book"
)

// Iteration records one completed iterative-deepening pass.
type Iteration struct {
	Depth    int
	Score    eval.EvaluationScore // from the side to move's point of view
	BestMove board.Move
	Nodes    int64 // cumulative nodes searched when the iteration finished
	Elapsed  time.Duration

	// Times the aspiration window failed low or high before this iteration's
	// score landed inside it. Both are zero when the first window held.
	AspirationFailLows  int
	AspirationFailHighs int
}

// MaxLoggedRootMoves is how many root moves Stats.RootMoves keeps.
const MaxLoggedRootMoves = 5

// RootMove is a root move with an exact score from the last completed
// iteration. Only moves searched with a full window have one: the rest failed
// low on a null-window scout, which reports the current alpha whatever the
// move's real value, so they carry no information and are not recorded.
type RootMove struct {
	Move  board.Move
	Score eval.EvaluationScore // from the side to move's point of view
}

// Stats tracks statistics during search
type Stats struct {
	NodesSearched      int64
	Depth              int
	Time               time.Duration
	PrincipalVariation []board.Move
	BookMoveUsed       bool // True if move came from opening book

	// Iterations lists each completed iteration in depth order; StopReason says
	// why the search ended (one of the StopReason constants).
	Iterations []Iteration
	StopReason string

	// RootMoves lists the root moves with an exact score in the last completed
	// iteration, best first, at most MaxLoggedRootMoves. The chosen move is
	// always first; the others raised alpha before a better move was found.
	RootMoves []RootMove

	// Late Move Reductions (LMR) statistics
	LMRReductions int64 // Number of moves reduced
	LMRReSearches int64 // Number of re-searches performed

	// Null move pruning statistics
	NullMoves               int64 // Number of null move attempts
	NullCutoffs             int64 // Number of successful null move cutoffs
	NullMoveZugzwangSkipped int64 // Null move declined: side to move has no non-pawn material
	NullMoveSparseSkipped   int64 // Null move declined: board has NullMoveMaxPieces pieces or fewer

	QNodes           int64 // Quiescence search nodes
	TTCutoffs        int64 // Beta cutoffs from transposition table
	FirstMoveCutoffs int64 // Beta cutoffs on first move tried
	TotalCutoffs     int64 // Total beta cutoffs (for move ordering calculation)
	DeltaPruned      int64 // Captures skipped by delta pruning

	// Reverse futility pruning statistics
	RFPCutoffs int64 // Nodes pruned as fail-high by static eval minus margin

	// Razoring statistics
	RazoringAttempts int64 // Number of razoring attempts
	RazoringCutoffs  int64 // Successful razoring cutoffs
	RazoringFailed   int64 // Razoring attempts that failed verification

	// Quiet-move pruning statistics
	FutilityPrunes int64 // Quiets skipped because eval + margin cannot reach alpha
	LMPPrunes      int64 // Late quiets skipped outright (late move pruning)

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

// adoptCounters replaces every search counter in s with the value in counters
// and keeps the result fields s has accumulated itself (Depth, Time,
// PrincipalVariation, BookMoveUsed, Iterations, StopReason, RootMoves). The
// counters are whatever Stats holds beyond those, so a new counter is picked
// up without touching this function; a new result field must be added to the
// list below.
func (s *Stats) adoptCounters(counters Stats) {
	counters.Depth, counters.Time = s.Depth, s.Time
	counters.PrincipalVariation, counters.BookMoveUsed = s.PrincipalVariation, s.BookMoveUsed
	counters.Iterations, counters.StopReason, counters.RootMoves = s.Iterations, s.StopReason, s.RootMoves
	*s = counters
}

// EBF returns the effective branching factor: the mean ratio of nodes searched
// by one iterative-deepening iteration to nodes searched by the one before it.
// Iteration.Nodes is cumulative, so each iteration's own cost is the difference
// from its predecessor. Iterations with no nodes of their own are skipped
// rather than treated as zero, and it returns 0 when fewer than two iterations
// qualify, never NaN or Inf.
//
// NodesByDepth is deliberately not used: it counts nodes by distance from the
// root within the whole search, so the shrinking tail beyond the nominal depth
// drags a per-ply ratio well below the true branching factor.
func (s *Stats) EBF() float64 {
	var sum float64
	var count int
	var prevCumulative, prevOwn int64
	for _, it := range s.Iterations {
		own := it.Nodes - prevCumulative
		prevCumulative = it.Nodes
		if own <= 0 {
			continue
		}
		if prevOwn > 0 {
			sum += float64(own) / float64(prevOwn)
			count++
		}
		prevOwn = own
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// TTHitRate returns the fraction of transposition table probes that were
// hits (TTHits/TTProbes). Returns 0 if there were no probes.
func (s *Stats) TTHitRate() float64 {
	if s.TTProbes == 0 {
		return 0
	}
	return float64(s.TTHits) / float64(s.TTProbes)
}

// OrderingQuality returns the fraction of beta cutoffs that occurred on the
// first move tried (FirstMoveCutoffs/TotalCutoffs), a standard proxy for
// move-ordering effectiveness. Returns 0 if there were no cutoffs.
func (s *Stats) OrderingQuality() float64 {
	if s.TotalCutoffs == 0 {
		return 0
	}
	return float64(s.FirstMoveCutoffs) / float64(s.TotalCutoffs)
}

// NullMoveEfficiency returns the fraction of null move attempts that
// produced a cutoff (NullCutoffs/NullMoves). Returns 0 if no null moves were
// attempted.
func (s *Stats) NullMoveEfficiency() float64 {
	if s.NullMoves == 0 {
		return 0
	}
	return float64(s.NullCutoffs) / float64(s.NullMoves)
}

// searchStatsJSON mirrors Stats for marshaling: it fixes the field
// order (struct field order, not map iteration order, so json.Marshal output
// is deterministic and diff-stable) and adds the derived metrics alongside
// the raw counters. Field names are snake_case per the task brief.
type searchStatsJSON struct {
	NodesSearched      int64    `json:"nodes_searched"`
	Depth              int      `json:"depth"`
	TimeMS             int64    `json:"time_ms"`
	PrincipalVariation []string `json:"principal_variation"`
	BookMoveUsed       bool     `json:"book_move_used"`

	StopReason string          `json:"stop_reason"`
	Iterations []iterationJSON `json:"iterations"`
	RootMoves  []rootMoveJSON  `json:"root_moves"`

	LMRReductions int64 `json:"lmr_reductions"`
	LMRReSearches int64 `json:"lmr_re_searches"`

	NullMoves               int64 `json:"null_moves"`
	NullCutoffs             int64 `json:"null_cutoffs"`
	NullMoveZugzwangSkipped int64 `json:"null_move_zugzwang_skipped"`
	NullMoveSparseSkipped   int64 `json:"null_move_sparse_skipped"`

	QNodes           int64 `json:"q_nodes"`
	TTCutoffs        int64 `json:"tt_cutoffs"`
	FirstMoveCutoffs int64 `json:"first_move_cutoffs"`
	TotalCutoffs     int64 `json:"total_cutoffs"`
	DeltaPruned      int64 `json:"delta_pruned"`

	RFPCutoffs int64 `json:"rfp_cutoffs"`

	RazoringAttempts int64 `json:"razoring_attempts"`
	RazoringCutoffs  int64 `json:"razoring_cutoffs"`
	RazoringFailed   int64 `json:"razoring_failed"`

	FutilityPrunes int64 `json:"futility_prunes"`
	LMPPrunes      int64 `json:"lmp_prunes"`

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

// iterationJSON is the marshaled form of Iteration: the score is in centipawns
// from the side to move's point of view, the best move in UCI notation.
type iterationJSON struct {
	Depth     int    `json:"depth"`
	ScoreCP   int    `json:"score_cp"`
	BestMove  string `json:"best_move"`
	Nodes     int64  `json:"nodes"`
	ElapsedMS int64  `json:"elapsed_ms"`

	AspirationFailLows  int `json:"aspiration_fail_lows"`
	AspirationFailHighs int `json:"aspiration_fail_highs"`
}

// rootMoveJSON is the marshaled form of RootMove.
type rootMoveJSON struct {
	Move    string `json:"move"`
	ScoreCP int    `json:"score_cp"`
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
// invokes it even when Stats is embedded by value in another struct
// (e.g. Result.Stats) and the top-level value passed to json.Marshal
// is not itself a pointer — encoding/json only auto-takes the address of
// addressable values, and a struct field of a non-pointer argument to
// json.Marshal is not addressable.
func (s Stats) MarshalJSON() ([]byte, error) {
	pv := make([]string, len(s.PrincipalVariation))
	for i, m := range s.PrincipalVariation {
		pv[i] = moveToUCI(m)
	}

	iterations := make([]iterationJSON, len(s.Iterations))
	for i, it := range s.Iterations {
		iterations[i] = iterationJSON{
			Depth:     it.Depth,
			ScoreCP:   int(it.Score),
			BestMove:  moveToUCI(it.BestMove),
			Nodes:     it.Nodes,
			ElapsedMS: it.Elapsed.Milliseconds(),

			AspirationFailLows:  it.AspirationFailLows,
			AspirationFailHighs: it.AspirationFailHighs,
		}
	}

	rootMoves := make([]rootMoveJSON, len(s.RootMoves))
	for i, rm := range s.RootMoves {
		rootMoves[i] = rootMoveJSON{Move: moveToUCI(rm.Move), ScoreCP: int(rm.Score)}
	}

	snapshot := searchStatsJSON{
		NodesSearched:      s.NodesSearched,
		Depth:              s.Depth,
		TimeMS:             s.Time.Milliseconds(),
		PrincipalVariation: pv,
		BookMoveUsed:       s.BookMoveUsed,

		StopReason: s.StopReason,
		Iterations: iterations,
		RootMoves:  rootMoves,

		RFPCutoffs: s.RFPCutoffs,

		LMRReductions: s.LMRReductions,
		LMRReSearches: s.LMRReSearches,

		NullMoves:               s.NullMoves,
		NullCutoffs:             s.NullCutoffs,
		NullMoveZugzwangSkipped: s.NullMoveZugzwangSkipped,
		NullMoveSparseSkipped:   s.NullMoveSparseSkipped,

		QNodes:           s.QNodes,
		TTCutoffs:        s.TTCutoffs,
		FirstMoveCutoffs: s.FirstMoveCutoffs,
		TotalCutoffs:     s.TotalCutoffs,
		DeltaPruned:      s.DeltaPruned,

		RazoringAttempts: s.RazoringAttempts,
		RazoringCutoffs:  s.RazoringCutoffs,
		RazoringFailed:   s.RazoringFailed,

		FutilityPrunes: s.FutilityPrunes,
		LMPPrunes:      s.LMPPrunes,

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
