package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateBishopMoves generates all legal bishop moves for the given player
func (g *Generator) GenerateBishopMoves(b *board.Board, player Player) *MoveList {
	// Determine bishop piece based on player
	var bishopPiece board.Piece
	if player == White {
		bishopPiece = board.WhiteBishop
	} else {
		bishopPiece = board.BlackBishop
	}
	
	return g.generateSlidingPieceMoves(b, player, bishopPiece, BishopDirections)
}