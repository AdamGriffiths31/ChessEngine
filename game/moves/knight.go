package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateKnightMoves generates all legal knight moves for the given player
func (g *Generator) GenerateKnightMoves(b *board.Board, player Player) *MoveList {
	// Determine knight piece based on player
	var knightPiece board.Piece
	if player == White {
		knightPiece = board.WhiteKnight
	} else {
		knightPiece = board.BlackKnight
	}
	
	return g.generateJumpingPieceMoves(b, player, knightPiece, KnightDirections)
}