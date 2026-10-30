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
	// trusted. At or below it, null-move cutoffs are skipped.
	NullMoveMaxPieces = 8

	noStaticEval = eval.EvaluationScore(-32001)

	// FiftyMoveHalfMoves is the halfmove-clock value at which the fifty-move
	// rule makes a position drawn (50 full moves = 100 plies without a pawn
	// move or capture).
	FiftyMoveHalfMoves = 100

	// NullMoveMinDepth is the shallowest remaining depth at which null-move
	// pruning is tried.
	NullMoveMinDepth = 3
	// NullMoveDeepDepth is the depth at which the null-move reduction grows by one.
	NullMoveDeepDepth = 6
	// NullMoveMaxReduction caps the null-move reduction.
	NullMoveMaxReduction = 3

	// QDeltaMargin is added to the stand-pat eval and the victim's value. A
	// capture is skipped when the total still falls below alpha.
	QDeltaMargin = 200
	// QDeltaPromoBonus is added to that total for promotions.
	QDeltaPromoBonus = 800
	// QSEEPruneThreshold is the SEE below which quiescence skips a capture.
	QSEEPruneThreshold = -100
)

const (
	lmrTableDepths = 16
	lmrTableMoves  = 64
)

// LMRTable is a pre-calculated reduction table for Late Move Reductions
// Indexed by [depth][moveCount] to get reduction amount
var LMRTable [lmrTableDepths][lmrTableMoves]int

func init() {
	for depth := 1; depth < lmrTableDepths; depth++ {
		for moveCount := 1; moveCount < lmrTableMoves; moveCount++ {
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

	NullMoveEnabled bool

	// Reverse futility pruning (static null move): in non-PV nodes at shallow
	// depth, if the static eval is already above beta by a depth-scaled margin,
	// assume the node fails high without searching it.
	RFPEnabled  bool
	RFPMaxDepth int
	RFPMargin   eval.EvaluationScore // Per-ply margin: eval - depth*RFPMargin >= beta

	RazoringEnabled  bool
	RazoringMargins  [5]eval.EvaluationScore // Margins for depths 1-4 (index 0 unused)
	RazoringMaxDepth int

	// Internal futility pruning: at shallow depth, skip quiet moves when the
	// static eval plus a depth-scaled margin still cannot reach alpha.
	// Giving-check moves are exempt.
	FutilityEnabled  bool
	FutilityMaxDepth int
	FutilityMargins  [6]eval.EvaluationScore // Indexed by depth (index 0 unused)

	// Late Move Pruning (LMP): in quiet nodes, once enough quiet moves have
	// been searched without improving alpha, skip remaining quiets outright.
	LMPEnabled  bool
	LMPMaxDepth int

	LMREnabled  bool
	LMRMinDepth int
	LMRMinMoves int
}

func getParams() Params {
	return Params{
		NullMoveReduction:    2,
		HistoryHighThreshold: 2000,
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,

		NullMoveEnabled: true,

		RFPEnabled:  true,
		RFPMaxDepth: 6,
		RFPMargin:   90,

		RazoringEnabled:  true,
		RazoringMargins:  [5]eval.EvaluationScore{0, 125, 175, 225, 275},
		RazoringMaxDepth: 1,

		FutilityEnabled:  true,
		FutilityMaxDepth: 4,
		FutilityMargins:  [6]eval.EvaluationScore{0, 150, 250, 350, 450},

		LMPEnabled:  true,
		LMPMaxDepth: 5,

		LMREnabled:  true,
		LMRMinDepth: 3,
		LMRMinMoves: 4,
	}
}
