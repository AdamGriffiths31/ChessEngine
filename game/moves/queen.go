package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateQueenMoves generates all legal queen moves for the given player
func (g *Generator) GenerateQueenMoves(b *board.Board, player Player) *MoveList {
	// Determine queen piece based on player
	var queenPiece board.Piece
	if player == White {
		queenPiece = board.WhiteQueen
	} else {
		queenPiece = board.BlackQueen
	}
	
	return g.generateSlidingPieceMoves(b, player, queenPiece, QueenDirections)
}