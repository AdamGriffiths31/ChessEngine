package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
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

// MaterialEvaluator evaluates positions based only on material balance
type MaterialEvaluator struct{}

// NewMaterialEvaluator creates a new material-only evaluator
func NewMaterialEvaluator() *MaterialEvaluator {
	return &MaterialEvaluator{}
}

// Evaluate returns the material balance from the given player's perspective
func (m *MaterialEvaluator) Evaluate(b *board.Board, player moves.Player) ai.EvaluationScore {
	score := 0

	// Sum up all piece values on the board
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				score += PieceValues[piece]
			}
		}
	}

	// Return score from player's perspective
	if player == moves.Black {
		score = -score
	}

	return ai.EvaluationScore(score)
}

// GetName returns the evaluator name
func (m *MaterialEvaluator) GetName() string {
	return "Material Evaluator"
}