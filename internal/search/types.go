// Package search provides chess search types and interfaces.
package search

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// Config configures the search parameters
type Config struct {
	MaxDepth  int
	MaxTime   time.Duration
	DebugMode bool

	// Opening book configuration
	UseOpeningBook bool
	BookFiles      []string

	// RepetitionHistory is the sequence of position hashes (board.Board.GetHash
	// scheme) for the real game so far, oldest first, ending with the hash of
	// the position being searched (b.GetHash()). Optional: when empty, the
	// search's repetition detection only sees positions reachable within this
	// search's own hypothetical lookahead from the current position, exactly
	// as before this field existed - it cannot recognize that a position has
	// already repeated via real moves played earlier in the actual game. When
	// provided, it seeds that real history so a hypothetical recurrence found
	// during search can be correctly recognized as completing a repetition
	// that started before this search call.
	RepetitionHistory []uint64
}

// Result contains the result of a search
type Result struct {
	BestMove board.Move
	Score    eval.EvaluationScore
	Stats    Stats
}

// Engine defines the interface for a chess AI engine
type Engine interface {
	// FindBestMove searches for the best move in the given position
	FindBestMove(ctx context.Context, b *board.Board, player movegen.Player, config Config) Result

	// SetEvaluator sets the position evaluator
	SetEvaluator(evaluator eval.Evaluator)

	// GetName returns the engine name
	GetName() string
}
