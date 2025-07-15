package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateKnightMoves generates all legal knight moves for the given player
func (g *Generator) GenerateKnightMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	// Determine knight piece based on player
	var knightPiece board.Piece
	if player == White {
		knightPiece = board.WhiteKnight
	} else {
		knightPiece = board.BlackKnight
	}
	
	// Scan the board for knights of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == knightPiece {
				// Generate moves for this knight
				g.generateKnightMovesFromSquare(b, player, rank, file, moveList)
			}
		}
	}
	
	return moveList
}

// generateKnightMovesFromSquare generates all moves for a knight at a specific square
func (g *Generator) generateKnightMovesFromSquare(b *board.Board, player Player, rank, file int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}
	
	// All possible knight moves: L-shaped moves (2+1 in all directions)
	knightMoves := []struct{ rankDelta, fileDelta int }{
		{2, 1},   // Up 2, Right 1
		{2, -1},  // Up 2, Left 1
		{-2, 1},  // Down 2, Right 1
		{-2, -1}, // Down 2, Left 1
		{1, 2},   // Up 1, Right 2
		{1, -2},  // Up 1, Left 2
		{-1, 2},  // Down 1, Right 2
		{-1, -2}, // Down 1, Left 2
	}
	
	for _, move := range knightMoves {
		newRank := rank + move.rankDelta
		newFile := file + move.fileDelta
		
		// Check if the target square is within board boundaries
		if newRank >= 0 && newRank <= 7 && newFile >= 0 && newFile <= 7 {
			piece := b.GetPiece(newRank, newFile)
			to := board.Square{File: newFile, Rank: newRank}
			
			if piece == board.Empty {
				// Empty square - valid move
				knightMove := board.Move{
					From:      fromSquare,
					To:        to,
					IsCapture: false,
					Promotion: board.Empty,
				}
				moveList.AddMove(knightMove)
			} else if g.isEnemyPiece(piece, player) {
				// Enemy piece - valid capture
				knightMove := board.Move{
					From:      fromSquare,
					To:        to,
					IsCapture: true,
					Captured:  piece,
					Promotion: board.Empty,
				}
				moveList.AddMove(knightMove)
			}
			// Own piece - can't move here, but knight doesn't slide so no need to block further moves
		}
	}
}