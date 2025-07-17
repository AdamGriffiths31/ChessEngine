package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Castling constants
const (
	KingsideTargetFile = 6 // King moves to file g (index 6) for kingside castling
	QueensideTargetFile = 2 // King moves to file c (index 2) for queenside castling
	WhiteKingStartRank = 0 // White king starts on rank 1 (index 0)
	BlackKingStartRank = 7 // Black king starts on rank 8 (index 7)
)

// findKingSquare finds the king's position for the given player
func (g *Generator) findKingSquare(b *board.Board, player Player) *board.Square {
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == kingPiece {
				return &board.Square{File: file, Rank: rank}
			}
		}
	}
	return nil
}

// GenerateKingMoves generates all legal king moves for the given player
func (g *Generator) GenerateKingMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	kingSquare := g.findKingSquare(b, player)
	if kingSquare != nil {
		// Generate moves for this king
		g.generateKingMovesFromSquare(b, player, kingSquare.Rank, kingSquare.File, moveList)
	}
	
	return moveList
}

// generateKingMovesFromSquare generates all moves for a king at a specific square
func (g *Generator) generateKingMovesFromSquare(b *board.Board, player Player, rank, file int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}
	
	// Generate single-square moves in all 8 directions
	g.generateKingSingleMoves(b, player, fromSquare, moveList)
	
	// Generate castling moves
	g.generateCastlingMoves(b, player, fromSquare, moveList)
}

// generateKingSingleMoves generates single-square moves in all 8 directions
func (g *Generator) generateKingSingleMoves(b *board.Board, player Player, from board.Square, moveList *MoveList) {
	// All possible king moves: one square in any direction
	kingMoves := []struct{ rankDelta, fileDelta int }{
		{1, 0},   // Up
		{-1, 0},  // Down
		{0, 1},   // Right
		{0, -1},  // Left
		{1, 1},   // Up-right
		{1, -1},  // Up-left
		{-1, 1},  // Down-right
		{-1, -1}, // Down-left
	}
	
	for _, move := range kingMoves {
		newRank := from.Rank + move.rankDelta
		newFile := from.File + move.fileDelta
		
		// Check if the target square is within board boundaries
		if newRank >= 0 && newRank <= 7 && newFile >= 0 && newFile <= 7 {
			piece := b.GetPiece(newRank, newFile)
			to := board.Square{File: newFile, Rank: newRank}
			
			if piece == board.Empty {
				// Empty square - valid move
				kingMove := g.createMove(b, from, to, false, board.Empty, board.Empty)
				kingMove.IsCastling = false
				moveList.AddMove(kingMove)
			} else if g.isEnemyPiece(piece, player) {
				// Enemy piece - valid capture
				kingMove := g.createMove(b, from, to, true, piece, board.Empty)
				kingMove.IsCastling = false
				moveList.AddMove(kingMove)
			}
			// Own piece - can't move here
		}
	}
}

// generateCastlingMoves generates castling moves if conditions are met
func (g *Generator) generateCastlingMoves(b *board.Board, player Player, from board.Square, moveList *MoveList) {
	castlingRights := b.GetCastlingRights()
	
	var kingStartRank int
	var rookPiece board.Piece
	var kingsideRight, queensideRight rune
	
	if player == White {
		kingStartRank = WhiteKingStartRank
		rookPiece = board.WhiteRook
		kingsideRight = 'K'
		queensideRight = 'Q'
	} else {
		kingStartRank = BlackKingStartRank
		rookPiece = board.BlackRook
		kingsideRight = 'k'
		queensideRight = 'q'
	}
	
	// Only allow castling if king is on starting square
	if from.Rank != kingStartRank || from.File != KingStartFile {
		return
	}
	
	// Check kingside castling (O-O)
	if g.hasCastlingRight(castlingRights, kingsideRight) && g.canCastleKingside(b, player, kingStartRank, rookPiece) {
		castlingMove := g.createMove(b, from, board.Square{File: KingsideTargetFile, Rank: kingStartRank}, false, board.Empty, board.Empty)
		castlingMove.IsCastling = true
		// Use castling handler for validation
		isSquareAttacked := func(square board.Square) bool {
			return g.attackDetector.IsSquareAttacked(b, square, player)
		}
		if g.castlingHandler.IsLegal(b, castlingMove, player, isSquareAttacked) {
			moveList.AddMove(castlingMove)
		}
	}
	
	// Check queenside castling (O-O-O)
	if g.hasCastlingRight(castlingRights, queensideRight) && g.canCastleQueenside(b, player, kingStartRank, rookPiece) {
		castlingMove := g.createMove(b, from, board.Square{File: QueensideTargetFile, Rank: kingStartRank}, false, board.Empty, board.Empty)
		castlingMove.IsCastling = true
		// Use castling handler for validation
		isSquareAttacked := func(square board.Square) bool {
			return g.attackDetector.IsSquareAttacked(b, square, player)
		}
		if g.castlingHandler.IsLegal(b, castlingMove, player, isSquareAttacked) {
			moveList.AddMove(castlingMove)
		}
	}
}

// canCastleKingside checks if kingside castling is possible
func (g *Generator) canCastleKingside(b *board.Board, player Player, rank int, rookPiece board.Piece) bool {
	// Check if rook is in correct position
	if b.GetPiece(rank, 7) != rookPiece {
		return false
	}
	
	// Check if squares between king and rook are empty
	if b.GetPiece(rank, 5) != board.Empty || b.GetPiece(rank, 6) != board.Empty {
		return false
	}
	
	return true
}

// canCastleQueenside checks if queenside castling is possible
func (g *Generator) canCastleQueenside(b *board.Board, player Player, rank int, rookPiece board.Piece) bool {
	// Check if rook is in correct position
	if b.GetPiece(rank, 0) != rookPiece {
		return false
	}
	
	// Check if squares between king and rook are empty
	if b.GetPiece(rank, 1) != board.Empty || b.GetPiece(rank, 2) != board.Empty || b.GetPiece(rank, 3) != board.Empty {
		return false
	}
	
	return true
}

// hasCastlingRight checks if a specific castling right is available
func (g *Generator) hasCastlingRight(castlingRights string, right rune) bool {
	for _, r := range castlingRights {
		if r == right {
			return true
		}
	}
	return false
}

