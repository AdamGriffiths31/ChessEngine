// Package eval provides chess position evaluation types and interfaces.
package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

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

// Evaluator defines the interface for position evaluation
type Evaluator interface {
	// Evaluate returns the score for the position from White's perspective
	// Positive = good for White, Negative = good for Black
	Evaluate(b *board.Board) EvaluationScore

	// GetName returns the evaluator name
	GetName() string
}
