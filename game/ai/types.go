package ai

import (
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// SearchStats tracks statistics during search
type SearchStats struct {
	NodesSearched      int64
	Depth              int
	Time               time.Duration
	PrincipalVariation []board.Move
	BookMoveUsed       bool     // True if move came from opening book
	DebugInfo          []string // Debug messages (when DebugMode is enabled)

	// Late Move Reductions (LMR) statistics
	LMRReductions   int64 // Number of moves reduced
	LMRReSearches   int64 // Number of re-searches performed
	LMRNodesSkipped int64 // Estimated nodes saved by LMR

	// Null move pruning statistics
	NullMoves   int64 // Number of null move attempts
	NullCutoffs int64 // Number of successful null move cutoffs

	// Additional search statistics
	QNodes           int64 // Quiescence search nodes
	TTCutoffs        int64 // Beta cutoffs from transposition table
	FirstMoveCutoffs int64 // Beta cutoffs on first move tried
	TotalCutoffs     int64 // Total beta cutoffs (for move ordering calculation)
	DeltaPruned      int64 // Captures skipped by delta pruning

	// Razoring statistics
	RazoringAttempts int64 // Number of razoring attempts
	RazoringCutoffs  int64 // Successful razoring cutoffs
	RazoringFailed   int64 // Razoring attempts that failed verification
}

// SearchConfig configures the search parameters
type SearchConfig struct {
	MaxDepth  int
	MaxTime   time.Duration
	MaxNodes  int64
	DebugMode bool

	// Opening book configuration
	UseOpeningBook      bool
	BookFiles           []string
	BookSelectMode      BookSelectionMode
	BookWeightThreshold uint16

	// Late Move Reductions (LMR) configuration
	LMRMinDepth int // Minimum depth to apply LMR (default: 3)
	LMRMinMoves int // Number of moves to search at full depth (default: 4)

	// Null move pruning configuration
	DisableNullMove bool // If true, disables null move pruning for comparison testing

	// Razoring configuration
	DisableRazoring     bool    // If true, disables razoring for comparison testing
	RazoringMarginScale float64 // Scale factor for margins (default 1.0)

	// Parallel search configuration
	NumThreads int // Number of threads to use for parallel search (default: 1 for sequential)
}

// BookSelectionMode defines how to select moves from opening books
type BookSelectionMode int

const (
	// BookSelectBest always chooses the highest-weighted move
	BookSelectBest BookSelectionMode = iota

	// BookSelectRandom chooses randomly (equal probability)
	BookSelectRandom

	// BookSelectWeightedRandom uses weighted random selection based on move weights
	BookSelectWeightedRandom
)

// EvaluationScore represents the score of a position
type EvaluationScore int32

const (
	// MateScore represents a checkmate position value
	MateScore EvaluationScore = 10000000
	// DrawScore represents a drawn position value
	DrawScore EvaluationScore = 0
	// UnknownScore represents an unknown or invalid position value
	UnknownScore EvaluationScore = -1000000
)

// SearchResult contains the result of a search
type SearchResult struct {
	BestMove board.Move
	Score    EvaluationScore
	Stats    SearchStats
}
