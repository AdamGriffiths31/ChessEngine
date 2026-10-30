package search

import (
	"encoding/json"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

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
	BookMoveUsed       bool

	// Iterations lists each completed iteration in depth order; StopReason says
	// why the search ended (one of the StopReason constants).
	Iterations []Iteration
	StopReason string

	// RootMoves lists the root moves with an exact score in the last completed
	// iteration, best first, at most MaxLoggedRootMoves. The chosen move is
	// always first; the others raised alpha before a better move was found.
	RootMoves []RootMove

	LMRReductions int64
	LMRReSearches int64

	NullMoves               int64
	NullCutoffs             int64
	NullMoveZugzwangSkipped int64 // Null move declined: side to move has no non-pawn material
	NullMoveSparseSkipped   int64 // Null move declined: board has NullMoveMaxPieces pieces or fewer

	QNodes           int64
	TTCutoffs        int64
	FirstMoveCutoffs int64
	TotalCutoffs     int64
	DeltaPruned      int64

	RFPCutoffs int64

	RazoringAttempts int64
	RazoringCutoffs  int64
	RazoringFailed   int64

	FutilityPrunes int64
	LMPPrunes      int64

	CutoffsByMoveIndex [64]int64

	TTProbes int64
	TTHits   int64

	NodesByDepth [100]int64 // Nodes searched at each depth from root

	PVNodes  int64
	CutNodes int64
	AllNodes int64
}

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

type iterationJSON struct {
	Depth     int    `json:"depth"`
	ScoreCP   int    `json:"score_cp"`
	BestMove  string `json:"best_move"`
	Nodes     int64  `json:"nodes"`
	ElapsedMS int64  `json:"elapsed_ms"`

	AspirationFailLows  int `json:"aspiration_fail_lows"`
	AspirationFailHighs int `json:"aspiration_fail_highs"`
}

type rootMoveJSON struct {
	Move    string `json:"move"`
	ScoreCP int    `json:"score_cp"`
}

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

func moveToUCI(m board.Move) string {
	uciMove := m.From.String() + m.To.String()
	if m.Promotion != board.Empty {
		uciMove += string(toLowerRune(rune(m.Promotion)))
	}
	return uciMove
}

func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
