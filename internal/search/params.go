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

	// NullMoveEnabled toggles null-move pruning (see tryNullMove). Defaults to
	// true; used by soundness_test.go to compare pruned vs. unpruned search.
	NullMoveEnabled bool

	// Razoring parameters
	RazoringEnabled  bool
	RazoringMargins  [5]eval.EvaluationScore // Margins for depths 1-4 (index 0 unused)
	RazoringMaxDepth int                     // Maximum depth to apply razoring

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

		// Late Move Reductions: matches the UCI engine defaults.
		LMREnabled:  true,
		LMRMinDepth: 3,
		LMRMinMoves: 4,
	}
}
