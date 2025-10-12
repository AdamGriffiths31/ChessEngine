// Package evaluation provides chess position evaluation functions.
package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

// GetPieceValue returns the value of a piece in centipawns.
func GetPieceValue(piece board.Piece) int {
	switch piece {
	case board.WhitePawn:
		return 100
	case board.WhiteKnight:
		return 320
	case board.WhiteBishop:
		return 330
	case board.WhiteRook:
		return 500
	case board.WhiteQueen:
		return 900
	case board.WhiteKing:
		return 0
	case board.BlackPawn:
		return -100
	case board.BlackKnight:
		return -320
	case board.BlackBishop:
		return -330
	case board.BlackRook:
		return -500
	case board.BlackQueen:
		return -900
	case board.BlackKing:
		return 0
	default:
		return 0
	}
}

// PieceValues map kept for backward compatibility but marked for deprecation.
// Use GetPieceValue() for better performance.
var PieceValues = map[board.Piece]int{
	board.WhitePawn:   100,
	board.WhiteKnight: 320,
	board.WhiteBishop: 330,
	board.WhiteRook:   500,
	board.WhiteQueen:  900,
	board.WhiteKing:   0,

	board.BlackPawn:   -100,
	board.BlackKnight: -320,
	board.BlackBishop: -330,
	board.BlackRook:   -500,
	board.BlackQueen:  -900,
	board.BlackKing:   0,
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

	score := 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				score += GetPieceValue(piece)
				score += getPositionalBonus(piece, rank, file)
			}
		}
	}

	return score
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

func getPositionalBonus(piece board.Piece, rank, file int) int {
	switch piece {
	case board.WhiteKnight:
		return KnightTable[rank*8+file]
	case board.BlackKnight:
		flippedRank := 7 - rank
		return -KnightTable[flippedRank*8+file]
	case board.WhiteBishop:
		return BishopTable[rank*8+file]
	case board.BlackBishop:
		flippedRank := 7 - rank
		return -BishopTable[flippedRank*8+file]
	case board.WhiteRook:
		return RookTable[rank*8+file]
	case board.BlackRook:
		flippedRank := 7 - rank
		return -RookTable[flippedRank*8+file]
	case board.WhitePawn:
		return PawnTable[rank*8+file]
	case board.BlackPawn:
		flippedRank := 7 - rank
		return -PawnTable[flippedRank*8+file]
	case board.WhiteQueen:
		return QueenTable[rank*8+file]
	case board.BlackQueen:
		flippedRank := 7 - rank
		return -QueenTable[flippedRank*8+file]
	case board.WhiteKing:
		return KingTable[rank*8+file]
	case board.BlackKing:
		flippedRank := 7 - rank
		return -KingTable[flippedRank*8+file]
	default:
		return 0
	}
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
