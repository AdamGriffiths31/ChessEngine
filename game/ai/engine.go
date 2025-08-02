package ai

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Evaluator defines the interface for position evaluation
type Evaluator interface {
	// Evaluate returns the score for the position from White's perspective
	// Positive = good for White, Negative = good for Black
	Evaluate(b *board.Board) EvaluationScore

	// GetName returns the evaluator name
	GetName() string
}

// Engine defines the interface for a chess AI engine
type Engine interface {
	// FindBestMove searches for the best move in the given position
	FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config SearchConfig) SearchResult

	// SetEvaluator sets the position evaluator
	SetEvaluator(eval Evaluator)

	// GetName returns the engine name
	GetName() string
}