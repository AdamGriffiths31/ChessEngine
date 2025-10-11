// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"math"

	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

const (
	// MinEval represents the minimum possible evaluation score
	MinEval = ai.EvaluationScore(-32000)
	// MaxKillerDepth is the maximum depth for killer move tables
	MaxKillerDepth = 128
	// MateDistanceThreshold is the threshold for detecting mate distances
	MateDistanceThreshold = 1000
	// MaxGamePly is the maximum number of plies to track for repetition detection
	MaxGamePly = 1024
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
	LMRDivisor           float64
	NullMoveReduction    int
	HistoryHighThreshold int32
	HistoryMedThreshold  int32
	HistoryLowThreshold  int32

	// Razoring parameters
	RazoringEnabled  bool
	RazoringMargins  [5]ai.EvaluationScore // Margins for depths 1-4 (index 0 unused)
	RazoringMaxDepth int                   // Maximum depth to apply razoring

	// Futility pruning parameters
	FutilityMargins [5]ai.EvaluationScore // Futility margins for depths 1-4 (index 0 unused)

	// Extension thresholds
	CheckExtensionThreshold int                // Whether to extend single checks (0=no, 1=yes)
	SingularExtensionMargin ai.EvaluationScore // Margin for singular extensions
}

// getParams returns well-tuned search parameters
func getParams() Params {
	return Params{
		LMRDivisor:           1.8,  // Standard LMR divisor
		NullMoveReduction:    2,    // Conservative null move
		HistoryHighThreshold: 2000, // Well-tested history values
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,

		// Stockfish-aligned razoring margins - targeting 10-15% attempt rate
		RazoringEnabled:  true,
		RazoringMargins:  [5]ai.EvaluationScore{0, 100, 150, 200, 250},
		RazoringMaxDepth: 3,

		// Standard futility pruning
		FutilityMargins: [5]ai.EvaluationScore{0, 100, 200, 300, 400},

		// Standard extensions
		CheckExtensionThreshold: 1,
		SingularExtensionMargin: 100,
	}
}
