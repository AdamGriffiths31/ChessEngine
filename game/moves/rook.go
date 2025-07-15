package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateRookMoves generates all legal rook moves for the given player
func (g *Generator) GenerateRookMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	// Determine rook piece based on player
	var rookPiece board.Piece
	if player == White {
		rookPiece = board.WhiteRook
	} else {
		rookPiece = board.BlackRook
	}
	
	// Scan the board for rooks of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == rookPiece {
				// Generate moves for this rook
				g.generateRookMovesFromSquare(b, player, rank, file, moveList)
			}
		}
	}
	
	return moveList
}

// generateRookMovesFromSquare generates all moves for a rook at a specific square
func (g *Generator) generateRookMovesFromSquare(b *board.Board, player Player, rank, file int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}
	
	// Generate moves in all four directions: up, down, left, right
	directions := []struct{ rankDelta, fileDelta int }{
		{1, 0},  // Up (increasing rank)
		{-1, 0}, // Down (decreasing rank)
		{0, 1},  // Right (increasing file)
		{0, -1}, // Left (decreasing file)
	}
	
	for _, dir := range directions {
		g.generateSlidingMoves(b, player, fromSquare, dir.rankDelta, dir.fileDelta, moveList)
	}
}

// generateSlidingMoves generates moves in a straight line until blocked or edge reached
func (g *Generator) generateSlidingMoves(b *board.Board, player Player, from board.Square, rankDelta, fileDelta int, moveList *MoveList) {
	currentRank := from.Rank + rankDelta
	currentFile := from.File + fileDelta
	
	// Continue sliding in the direction until we hit the board edge or a piece
	for currentRank >= 0 && currentRank <= 7 && currentFile >= 0 && currentFile <= 7 {
		piece := b.GetPiece(currentRank, currentFile)
		to := board.Square{File: currentFile, Rank: currentRank}
		
		if piece == board.Empty {
			// Empty square - valid move
			move := board.Move{
				From:      from,
				To:        to,
				IsCapture: false,
				Promotion: board.Empty,
			}
			moveList.AddMove(move)
		} else if g.isEnemyPiece(piece, player) {
			// Enemy piece - valid capture, but can't continue sliding
			move := board.Move{
				From:      from,
				To:        to,
				IsCapture: true,
				Captured:  piece,
				Promotion: board.Empty,
			}
			moveList.AddMove(move)
			break // Stop sliding in this direction
		} else {
			// Own piece - can't move here and can't continue sliding
			break
		}
		
		// Move to next square in this direction
		currentRank += rankDelta
		currentFile += fileDelta
	}
}