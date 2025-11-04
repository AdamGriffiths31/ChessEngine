// Package evaluation provides chess position evaluation functions.
package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation/values"
)

// GetPieceValue returns the value of a piece in centipawns.
func GetPieceValue(piece board.Piece) int {
	return values.GetPieceValue(values.Piece(piece))
}

// PawnHashEntry represents a cached pawn structure evaluation.
type PawnHashEntry struct {
	hash  uint64
	score int
}

// Evaluator evaluates positions based on material balance and piece-square tables.
type Evaluator struct{}

// NewEvaluator creates a new evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate returns the evaluation from White's perspective using lazy evaluation with early cutoffs.
// Positive values favor White, negative values favor Black.
func (e *Evaluator) Evaluate(b *board.Board) ai.EvaluationScore {
	if b == nil {
		return ai.EvaluationScore(0)
	}

	score := 0

	score = e.evaluateMaterialAndPST(b)

	if abs(score) > 1000 {
		return ai.EvaluationScore(score)
	}

	pawnScore := evaluatePawnStructure(b)
	score += pawnScore

	if abs(score) < 500 {
		score += e.evaluatePieceActivity(b)
	}

	score += evaluateKings(b)

	return ai.EvaluationScore(score)
}

func (e *Evaluator) evaluateMaterialAndPST(b *board.Board) int {
	if b == nil {
		return 0
	}

	// Use incrementally maintained scores
	return b.GetMaterialScore() + b.GetPSTScore()
}

func (e *Evaluator) evaluatePieceActivity(b *board.Board) int {
	if b == nil {
		return 0
	}

	score := 0

	score += evaluateKnights(b)
	score += evaluateBishops(b)
	score += evaluateRooks(b)
	score += evaluateQueens(b)

	return score
}

// GetName returns the evaluator name.
func (e *Evaluator) GetName() string {
	return "Evaluator"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
