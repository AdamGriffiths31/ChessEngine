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
}

// SearchConfig configures the search parameters
type SearchConfig struct {
	MaxDepth     int
	MaxTime      time.Duration
	MaxNodes     int64
	UseAlphaBeta bool
	DebugMode    bool
}

// EvaluationScore represents the score of a position
type EvaluationScore int32

const (
	// Special scores
	MateScore    EvaluationScore = 100000
	DrawScore    EvaluationScore = 0
	UnknownScore EvaluationScore = -1000000
)

// SearchResult contains the result of a search
type SearchResult struct {
	BestMove board.Move
	Score    EvaluationScore
	Stats    SearchStats
}