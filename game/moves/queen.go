package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateQueenMoves generates all legal queen moves for the given player
func (g *Generator) GenerateQueenMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	// Determine queen piece based on player
	var queenPiece board.Piece
	if player == White {
		queenPiece = board.WhiteQueen
	} else {
		queenPiece = board.BlackQueen
	}
	
	// Scan the board for queens of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == queenPiece {
				// Generate moves for this queen
				g.generateQueenMovesFromSquare(b, player, rank, file, moveList)
			}
		}
	}
	
	return moveList
}

// generateQueenMovesFromSquare generates all moves for a queen at a specific square
func (g *Generator) generateQueenMovesFromSquare(b *board.Board, player Player, rank, file int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}
	
	// Queen combines rook and bishop movement patterns
	// Generate moves in all eight directions: 4 straight + 4 diagonal
	directions := []struct{ rankDelta, fileDelta int }{
		// Rook-like moves (straight lines)
		{1, 0},   // Up (increasing rank)
		{-1, 0},  // Down (decreasing rank)
		{0, 1},   // Right (increasing file)
		{0, -1},  // Left (decreasing file)
		// Bishop-like moves (diagonals)
		{1, 1},   // Up-right
		{1, -1},  // Up-left
		{-1, 1},  // Down-right
		{-1, -1}, // Down-left
	}
	
	for _, dir := range directions {
		g.generateSlidingMoves(b, player, fromSquare, dir.rankDelta, dir.fileDelta, moveList)
	}
}