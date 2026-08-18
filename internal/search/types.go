// Package search provides chess search types and interfaces.
package search

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// SearchConfig configures the search parameters
type SearchConfig struct {
	MaxDepth  int
	MaxTime   time.Duration
	DebugMode bool

	// Opening book configuration
	UseOpeningBook bool
	BookFiles      []string
}

// SearchResult contains the result of a search
type SearchResult struct {
	BestMove board.Move
	Score    eval.EvaluationScore
	Stats    SearchStats
}

// Engine defines the interface for a chess AI engine
type Engine interface {
	// FindBestMove searches for the best move in the given position
	FindBestMove(ctx context.Context, b *board.Board, player movegen.Player, config SearchConfig) SearchResult

	// SetEvaluator sets the position evaluator
	SetEvaluator(evaluator eval.Evaluator)

	// GetName returns the engine name
	GetName() string
}
