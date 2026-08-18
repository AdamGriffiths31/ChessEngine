package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval/values"
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

// StandardEvaluator evaluates positions based on material balance and piece-square tables.
type StandardEvaluator struct{}

// NewEvaluator creates a new evaluator.
func NewEvaluator() *StandardEvaluator {
	return &StandardEvaluator{}
}

// Evaluate returns the evaluation from White's perspective using lazy evaluation with early cutoffs.
// Positive values favor White, negative values favor Black.
func (e *StandardEvaluator) Evaluate(b *board.Board) EvaluationScore {
	if b == nil {
		return EvaluationScore(0)
	}

	score := 0

	score = e.evaluateMaterialAndPST(b)

	if abs(score) > 1000 {
		return EvaluationScore(score)
	}

	pawnScore := evaluatePawnStructure(b)
	score += pawnScore

	if abs(score) < 500 {
		score += e.evaluatePieceActivity(b)
	}

	score += evaluateKings(b)

	return EvaluationScore(score)
}

func (e *StandardEvaluator) evaluateMaterialAndPST(b *board.Board) int {
	if b == nil {
		return 0
	}

	// Use incrementally maintained scores
	return b.GetMaterialScore() + b.GetPSTScore()
}

func (e *StandardEvaluator) evaluatePieceActivity(b *board.Board) int {
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
func (e *StandardEvaluator) GetName() string {
	return "Evaluator"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
