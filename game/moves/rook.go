package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateRookMoves generates all legal rook moves for the given player
func (g *Generator) GenerateRookMoves(b *board.Board, player Player) *MoveList {
	// Determine rook piece based on player
	var rookPiece board.Piece
	if player == White {
		rookPiece = board.WhiteRook
	} else {
		rookPiece = board.BlackRook
	}
	
	return g.generateSlidingPieceMoves(b, player, rookPiece, RookDirections)
}

