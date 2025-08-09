package ai

import (
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// SearchStats tracks statistics during search
type SearchStats struct {
	NodesSearched      int64
	Depth             int
	Time              time.Duration
	PrincipalVariation []board.Move
	BookMoveUsed       bool     // True if move came from opening book
	DebugInfo          []string // Debug messages (when DebugMode is enabled)
	
	// Late Move Reductions (LMR) statistics
	LMRReductions   int64 // Number of moves reduced
	LMRReSearches   int64 // Number of re-searches performed
	LMRNodesSkipped int64 // Estimated nodes saved by LMR
}

// SearchConfig configures the search parameters
type SearchConfig struct {
	MaxDepth     int
	MaxTime      time.Duration
	MaxNodes     int64
	UseAlphaBeta bool
	DebugMode    bool
	
	// Opening book configuration
	UseOpeningBook    bool
	BookFiles         []string
	BookSelectMode    BookSelectionMode
	BookWeightThreshold uint16
	
	// Search enhancement options
	UseNullMove bool
	
	// Late Move Reductions (LMR) configuration
	UseLMR           bool    // Enable/disable LMR
	LMRMinDepth      int     // Minimum depth to apply LMR (default: 3)
	LMRMinMoves      int     // Number of moves to search at full depth (default: 4)
	LMRReductionBase float64 // Base reduction amount (default: 0.75)
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
	// Special scores
	MateScore    EvaluationScore = 10000000
	DrawScore    EvaluationScore = 0
	UnknownScore EvaluationScore = -1000000
)

// SearchResult contains the result of a search
type SearchResult struct {
	BestMove board.Move
	Score    EvaluationScore
	Stats    SearchStats
}