package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

// PieceValues defines the standard piece values in centipawns
var PieceValues = map[board.Piece]int{
	board.WhitePawn:   100,
	board.WhiteKnight: 320,
	board.WhiteBishop: 330,
	board.WhiteRook:   500,
	board.WhiteQueen:  900,
	board.WhiteKing:   0, // King has no material value

	board.BlackPawn:   -100,
	board.BlackKnight: -320,
	board.BlackBishop: -330,
	board.BlackRook:   -500,
	board.BlackQueen:  -900,
	board.BlackKing:   0,
}

// Evaluator evaluates positions based on material balance and piece-square tables
type Evaluator struct{}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate returns the evaluation from White's perspective
// combining material value and positional bonuses
// Positive = good for White, Negative = good for Black
func (e *Evaluator) Evaluate(b *board.Board) ai.EvaluationScore {
	score := 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				score += PieceValues[piece]
				score += getPositionalBonus(piece, rank, file)
			}
		}
	}

	// Add pawn-specific evaluation
	score += evaluatePawnStructure(b)

	// Add knight-specific evaluation
	score += evaluateKnights(b)

	// Add bishop-specific evaluation
	score += evaluateBishops(b)

	// Add rook-specific evaluation
	score += evaluateRooks(b)

	// Add queen-specific evaluation
	score += evaluateQueens(b)

	// Add king-specific evaluation
	score += evaluateKings(b)

	return ai.EvaluationScore(score)
}

// getPositionalBonus returns the positional bonus for a piece at the given position
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

// GetName returns the evaluator name
func (e *Evaluator) GetName() string {
	return "Evaluator"
}
