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

	// Sum up all piece values and positional bonuses on the board
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				// Add material value
				score += PieceValues[piece]

				// Add positional bonus
				score += getPositionalBonus(piece, rank, file)
			}
		}
	}

	// Always return from White's perspective
	return ai.EvaluationScore(score)
}

// getPositionalBonus returns the positional bonus for a piece at the given position
func getPositionalBonus(piece board.Piece, rank, file int) int {
	switch piece {
	case board.WhiteKnight:
		// White knights use table as-is
		return KnightTable[rank*8+file]
	case board.BlackKnight:
		// Black knights use flipped table (flip rank to black's perspective)
		flippedRank := 7 - rank
		return -KnightTable[flippedRank*8+file]
	case board.WhiteBishop:
		// White bishops use table as-is
		return BishopTable[rank*8+file]
	case board.BlackBishop:
		// Black bishops use flipped table with negated values
		flippedRank := 7 - rank
		return -BishopTable[flippedRank*8+file]
	case board.WhiteRook:
		// White rooks use table as-is
		return RookTable[rank*8+file]
	case board.BlackRook:
		// Black rooks use flipped table with negated values
		flippedRank := 7 - rank
		return -RookTable[flippedRank*8+file]
	case board.WhitePawn:
		// White pawns use table as-is
		return PawnTable[rank*8+file]
	case board.BlackPawn:
		// Black pawns use flipped table with negated values
		flippedRank := 7 - rank
		return -PawnTable[flippedRank*8+file]
	case board.WhiteQueen:
		// White queens use table as-is
		return QueenTable[rank*8+file]
	case board.BlackQueen:
		// Black queens use flipped table with negated values
		flippedRank := 7 - rank
		return -QueenTable[flippedRank*8+file]
	case board.WhiteKing:
		// White kings use table as-is
		return KingTable[rank*8+file]
	case board.BlackKing:
		// Black kings use flipped table with negated values
		flippedRank := 7 - rank
		return -KingTable[flippedRank*8+file]
	default:
		// Other pieces have no positional bonus yet
		return 0
	}
}

// GetName returns the evaluator name
func (e *Evaluator) GetName() string {
	return "Evaluator"
}

