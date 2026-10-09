// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"math"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

const (
	// MinEval represents the minimum possible evaluation score
	MinEval = eval.EvaluationScore(-32000)
	// MaxKillerDepth is the maximum depth for killer move tables
	MaxKillerDepth = 128
	// MateDistanceThreshold is the threshold for detecting mate distances
	MateDistanceThreshold = 1000
	// MaxGamePly is the maximum number of plies to track for repetition detection
	MaxGamePly = 1024
	// AspirationWindow is the half-width of the aspiration search window around
	// the previous iteration's score.
	AspirationWindow = 50
	// PVArrayMargin is the extra headroom added to the deepest search depth when
	// sizing the principal-variation array, to accommodate check extensions.
	PVArrayMargin = 20

	// NullMoveMaxPieces is the piece count above which null-move pruning is
	// trusted. At or below it, null-move cutoffs are skipped: sparse positions
	// are where zugzwang lives, and even one immobile minor piece (e.g. a bishop
	// jailed by its own pawns) can lose to a forced tempo, which
	// hasNonPawnMaterial alone does not catch.
	NullMoveMaxPieces = 8
)

// LMRTable is a pre-calculated reduction table for Late Move Reductions
// Indexed by [depth][moveCount] to get reduction amount
var LMRTable [16][64]int

func init() {
	for depth := 1; depth < 16; depth++ {
		for moveCount := 1; moveCount < 64; moveCount++ {
			LMRTable[depth][moveCount] = int(math.Log(float64(depth)) * math.Log(float64(moveCount)) / 1.8)
		}
	}
}

// Params holds search parameters
type Params struct {
	NullMoveReduction    int
	HistoryHighThreshold int32
	HistoryMedThreshold  int32
	HistoryLowThreshold  int32

	// NullMoveEnabled toggles null-move pruning (see tryNullMove). Defaults to true.
	NullMoveEnabled bool

	// Razoring parameters
	RazoringEnabled  bool
	RazoringMargins  [5]eval.EvaluationScore // Margins for depths 1-4 (index 0 unused)
	RazoringMaxDepth int                     // Maximum depth to apply razoring

	// Internal futility pruning: at shallow depth, skip quiet moves when the
	// static eval plus a depth-scaled margin still cannot reach alpha. The
	// move cannot improve on alpha by enough to matter, so searching it is
	// wasted work. Giving-check moves are exempt (they are free tactics).
	FutilityEnabled  bool
	FutilityMaxDepth int
	FutilityMargins  [6]eval.EvaluationScore // Indexed by depth (index 0 unused)

	// Late Move Pruning (LMP): in quiet nodes, once enough quiet moves have
	// been searched without improving alpha, skip remaining quiets outright.
	LMPEnabled  bool
	LMPMaxDepth int

	// Late Move Reductions (LMR) configuration
	LMREnabled  bool // Whether LMR is applied at all (see calculateLMRReduction). Defaults to true.
	LMRMinDepth int  // Minimum depth to apply LMR (default: 3)
	LMRMinMoves int  // Number of moves to search at full depth (default: 4)
}

// getParams returns well-tuned search parameters
func getParams() Params {
	return Params{
		NullMoveReduction:    2,    // Conservative null move
		HistoryHighThreshold: 2000, // Well-tested history values
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,

		NullMoveEnabled: true,

		// Razoring: Like Stockfish, only apply at depth 1
		// Applying at deeper depths is too risky in tactical positions
		// Margin of 125cp is conservative enough to avoid pruning tactics
		RazoringEnabled:  true,
		RazoringMargins:  [5]eval.EvaluationScore{0, 125, 175, 225, 275},
		RazoringMaxDepth: 1,

		// Futility pruning: implemented but default-off. Probing showed its
		// tactical safety is chaotic in this engine - identical margins both
		// found and lost a known mate depending on exact values - because
		// qsearch generates no quiet checking moves, making main-search
		// quiets the only carrier of those tactics. Revisit alongside
		// qsearch check generation; LMP alone delivers most of the node win.
		FutilityEnabled:  false,
		FutilityMaxDepth: 4,
		FutilityMargins:  [6]eval.EvaluationScore{0, 150, 250, 350, 450},

		// LMP: implemented but default-OFF. Empirically catastrophic in this
		// engine despite textbook formulation: SPRT measured roughly -150 to
		// -230 Elo at LMPMaxDepth 5, 2, and with the all-pruned-terminal bug
		// fixed, while fixed-depth tactical probes stayed clean. Root cause:
		// LMP presupposes late quiets are genuinely worse than early ones,
		// which requires well-ranked quiet ordering - see the history-scale
		// mismatch in move_ordering.go (history caps at 10k while tactical
		// bonuses reach 150k+, leaving plain quiets nearly arbitrarily
		// ordered). Revisit only after the history heuristic is reworked.
		LMPEnabled:  false,
		LMPMaxDepth: 5,

		// Late Move Reductions: matches the UCI engine defaults.
		LMREnabled:  true,
		LMRMinDepth: 3,
		LMRMinMoves: 4,
	}
}
