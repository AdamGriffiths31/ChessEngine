package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateBishopMoves generates all legal bishop moves for the given player
func (g *Generator) GenerateBishopMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	// Determine bishop piece based on player
	var bishopPiece board.Piece
	if player == White {
		bishopPiece = board.WhiteBishop
	} else {
		bishopPiece = board.BlackBishop
	}
	
	// Scan the board for bishops of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == bishopPiece {
				// Generate moves for this bishop
				g.generateBishopMovesFromSquare(b, player, rank, file, moveList)
			}
		}
	}
	
	return moveList
}

// generateBishopMovesFromSquare generates all moves for a bishop at a specific square
func (g *Generator) generateBishopMovesFromSquare(b *board.Board, player Player, rank, file int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}
	
	// Generate moves in all four diagonal directions
	directions := []struct{ rankDelta, fileDelta int }{
		{1, 1},   // Up-right
		{1, -1},  // Up-left
		{-1, 1},  // Down-right
		{-1, -1}, // Down-left
	}
	
	for _, dir := range directions {
		g.generateSlidingMoves(b, player, fromSquare, dir.rankDelta, dir.fileDelta, moveList)
	}
}