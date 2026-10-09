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

	// Tapered positional score: interpolate the middlegame and endgame PST
	// sums by remaining game phase. Material stays at flat values so every
	// downstream consumer (SEE, MVV-LVA, delta pruning) keeps its scale.
	mg := b.GetPSTScore()
	eg := b.GetPSTScoreEndgame()
	if mg == eg {
		return b.GetMaterialScore() + mg
	}
	phase := computeGamePhase(b)
	return b.GetMaterialScore() + (mg*phase+eg*(24-phase))/24
}

// computeGamePhase returns the remaining game phase as 0 (bare endgame) to
// 24 (full starting material): knights and bishops count 1 each, rooks 2,
// queens 4, both sides combined. The starting position sums to exactly 24.
func computeGamePhase(b *board.Board) int {
	phase := b.GetPieceBitboard(board.WhiteKnight).PopCount() +
		b.GetPieceBitboard(board.BlackKnight).PopCount() +
		b.GetPieceBitboard(board.WhiteBishop).PopCount() +
		b.GetPieceBitboard(board.BlackBishop).PopCount() +
		2*(b.GetPieceBitboard(board.WhiteRook).PopCount()+b.GetPieceBitboard(board.BlackRook).PopCount()) +
		4*(b.GetPieceBitboard(board.WhiteQueen).PopCount()+b.GetPieceBitboard(board.BlackQueen).PopCount())
	if phase > 24 {
		phase = 24
	}
	return phase
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
