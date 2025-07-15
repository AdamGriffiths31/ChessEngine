package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GenerateKingMoves generates all legal king moves for the given player
func (g *Generator) GenerateKingMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	// Determine king piece based on player
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	// Scan the board for the king of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == kingPiece {
				// Generate moves for this king
				g.generateKingMovesFromSquare(b, player, rank, file, moveList)
				return moveList // There should only be one king
			}
		}
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
				kingMove := board.Move{
					From:       from,
					To:         to,
					IsCapture:  false,
					IsCastling: false,
					Promotion:  board.Empty,
				}
				moveList.AddMove(kingMove)
			} else if g.isEnemyPiece(piece, player) {
				// Enemy piece - valid capture
				kingMove := board.Move{
					From:       from,
					To:         to,
					IsCapture:  true,
					Captured:   piece,
					IsCastling: false,
					Promotion:  board.Empty,
				}
				moveList.AddMove(kingMove)
			}
			// Own piece - can't move here
		}
	}
}

// generateCastlingMoves generates castling moves if conditions are met
func (g *Generator) generateCastlingMoves(b *board.Board, player Player, from board.Square, moveList *MoveList) {
	// Basic castling conditions:
	// 1. King must be on starting square
	// 2. No pieces between king and rook
	// 3. King and rook must not have moved (we'll assume they haven't for now)
	// 4. King must not be in check (we'll skip this check for now)
	// 5. King must not pass through or land on a square that is attacked (we'll skip this for now)
	
	var kingStartRank int
	var rookPiece board.Piece
	
	if player == White {
		kingStartRank = 0 // White king starts on rank 1 (index 0)
		rookPiece = board.WhiteRook
	} else {
		kingStartRank = 7 // Black king starts on rank 8 (index 7)
		rookPiece = board.BlackRook
	}
	
	// Only allow castling if king is on starting square
	if from.Rank != kingStartRank || from.File != 4 {
		return
	}
	
	// Check kingside castling (O-O)
	if g.canCastleKingside(b, player, kingStartRank, rookPiece) {
		castlingMove := board.Move{
			From:       from,
			To:         board.Square{File: 6, Rank: kingStartRank},
			IsCapture:  false,
			IsCastling: true,
			Promotion:  board.Empty,
		}
		moveList.AddMove(castlingMove)
	}
	
	// Check queenside castling (O-O-O)
	if g.canCastleQueenside(b, player, kingStartRank, rookPiece) {
		castlingMove := board.Move{
			From:       from,
			To:         board.Square{File: 2, Rank: kingStartRank},
			IsCapture:  false,
			IsCastling: true,
			Promotion:  board.Empty,
		}
		moveList.AddMove(castlingMove)
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