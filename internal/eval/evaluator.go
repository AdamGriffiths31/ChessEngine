package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval/values"
)

const (
	lazyEvalMargin      = 1000
	pieceActivityMargin = 500
	maxGamePhase        = 24
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
type StandardEvaluator struct {
	pawnHashTable [pawnHashSize]PawnHashEntry
}

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

	score := evaluateMaterialAndPST(b)

	if abs(score) > lazyEvalMargin {
		return EvaluationScore(score)
	}

	score += e.evaluatePawnStructure(b)

	if abs(score) < pieceActivityMargin {
		score += evaluatePieceActivity(b)
	}

	score += evaluateKings(b)

	return EvaluationScore(score)
}

func evaluateMaterialAndPST(b *board.Board) int {
	mg := b.GetPSTScore()
	eg := b.GetPSTScoreEndgame()
	phase := computeGamePhase(b)
	return b.GetMaterialScore() + (mg*phase+eg*(maxGamePhase-phase))/maxGamePhase
}

func computeGamePhase(b *board.Board) int {
	minors := b.GetPieceBitboard(board.WhiteKnight) | b.GetPieceBitboard(board.BlackKnight) |
		b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.BlackBishop)
	rooks := b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.BlackRook)
	queens := b.GetPieceBitboard(board.WhiteQueen) | b.GetPieceBitboard(board.BlackQueen)

	return min(minors.PopCount()+2*rooks.PopCount()+4*queens.PopCount(), maxGamePhase)
}

func evaluatePieceActivity(b *board.Board) int {
	whiteArea := mobilityArea(b, true)
	blackArea := mobilityArea(b, false)

	return evaluateKnights(b, whiteArea, blackArea) +
		evaluateBishops(b, whiteArea, blackArea) +
		evaluateRooks(b, whiteArea, blackArea) +
		evaluateQueens(b, whiteArea, blackArea)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
