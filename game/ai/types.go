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
	BookMoveUsed       bool // True if move came from opening book

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

// SearchConfig configures the search parameters
type SearchConfig struct {
	MaxDepth  int
	MaxTime   time.Duration
	DebugMode bool

	// Opening book configuration
	UseOpeningBook bool
	BookFiles      []string

	// Late Move Reductions (LMR) configuration
	LMRMinDepth int // Minimum depth to apply LMR (default: 3)
	LMRMinMoves int // Number of moves to search at full depth (default: 4)
}

// EvaluationScore represents the score of a position
type EvaluationScore int16

const (
	// MateScore represents a checkmate position value
	MateScore EvaluationScore = 30000
	// DrawScore represents a drawn position value
	DrawScore EvaluationScore = 0
	// UnknownScore represents an unknown or invalid position value
	UnknownScore EvaluationScore = -32000
)

// SearchResult contains the result of a search
type SearchResult struct {
	BestMove board.Move
	Score    EvaluationScore
	Stats    SearchStats
}
