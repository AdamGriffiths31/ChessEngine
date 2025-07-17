package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// AttackDetector handles all attack detection logic
type AttackDetector struct{}

// IsSquareAttacked checks if a square is attacked by the enemy
func (ad *AttackDetector) IsSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	enemyPlayer := Black
	if player == Black {
		enemyPlayer = White
	}
	
	// Check for enemy pawn attacks
	if ad.isSquareAttackedByPawns(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy knight attacks
	if ad.isSquareAttackedByKnights(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy sliding piece attacks (rooks, bishops, queens)
	if ad.isSquareAttackedBySlidingPieces(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy king attacks (single square moves only, no castling)
	if ad.isSquareAttackedByKing(b, square, enemyPlayer) {
		return true
	}
	
	return false
}

// isSquareAttackedByPawns checks if pawns attack the square
func (ad *AttackDetector) isSquareAttackedByPawns(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var pawnPiece board.Piece
	var pawnDirection int
	
	if enemyPlayer == White {
		pawnPiece = board.WhitePawn
		pawnDirection = 1 // White pawns attack upward
	} else {
		pawnPiece = board.BlackPawn
		pawnDirection = -1 // Black pawns attack downward
	}
	
	// Check diagonally backwards (where pawns would attack from)
	pawnRank := square.Rank - pawnDirection
	if pawnRank >= MinRank && pawnRank <= MaxRank {
		// Check left diagonal
		if square.File > MinFile && b.GetPiece(pawnRank, square.File-1) == pawnPiece {
			return true
		}
		// Check right diagonal
		if square.File < MaxFile && b.GetPiece(pawnRank, square.File+1) == pawnPiece {
			return true
		}
	}
	
	return false
}

// isSquareAttackedByKnights checks if knights attack the square
func (ad *AttackDetector) isSquareAttackedByKnights(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var knightPiece board.Piece
	if enemyPlayer == White {
		knightPiece = board.WhiteKnight
	} else {
		knightPiece = board.BlackKnight
	}
	
	// Check all knight move patterns
	for _, dir := range KnightDirections {
		knightRank := square.Rank - dir.RankDelta
		knightFile := square.File - dir.FileDelta
		
		if knightRank >= MinRank && knightRank <= MaxRank && knightFile >= MinFile && knightFile <= MaxFile {
			if b.GetPiece(knightRank, knightFile) == knightPiece {
				return true
			}
		}
	}
	
	return false
}

// isSquareAttackedBySlidingPieces checks if sliding pieces attack the square
func (ad *AttackDetector) isSquareAttackedBySlidingPieces(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var rookPiece, bishopPiece, queenPiece board.Piece
	
	if enemyPlayer == White {
		rookPiece = board.WhiteRook
		bishopPiece = board.WhiteBishop
		queenPiece = board.WhiteQueen
	} else {
		rookPiece = board.BlackRook
		bishopPiece = board.BlackBishop
		queenPiece = board.BlackQueen
	}
	
	// Check rook/queen attacks (straight lines)
	for _, dir := range RookDirections {
		if ad.isSquareAttackedFromDirection(b, square, dir, rookPiece, queenPiece) {
			return true
		}
	}
	
	// Check bishop/queen attacks (diagonals)
	for _, dir := range BishopDirections {
		if ad.isSquareAttackedFromDirection(b, square, dir, bishopPiece, queenPiece) {
			return true
		}
	}
	
	return false
}

// isSquareAttackedFromDirection checks if a square is attacked from a specific direction
func (ad *AttackDetector) isSquareAttackedFromDirection(b *board.Board, square board.Square, dir Direction, piece1, piece2 board.Piece) bool {
	currentRank := square.Rank + dir.RankDelta
	currentFile := square.File + dir.FileDelta
	
	// Slide in the direction until we hit a piece or board edge
	for currentRank >= MinRank && currentRank <= MaxRank && currentFile >= MinFile && currentFile <= MaxFile {
		pieceAtSquare := b.GetPiece(currentRank, currentFile)
		
		if pieceAtSquare != board.Empty {
			// Found a piece - check if it's an attacking piece
			return pieceAtSquare == piece1 || pieceAtSquare == piece2
		}
		
		currentRank += dir.RankDelta
		currentFile += dir.FileDelta
	}
	
	return false
}

// isSquareAttackedByKing checks king attacks
func (ad *AttackDetector) isSquareAttackedByKing(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var kingPiece board.Piece
	if enemyPlayer == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	// Check all 8 directions around the square (single square moves only)
	for _, dir := range QueenDirections {
		kingRank := square.Rank - dir.RankDelta
		kingFile := square.File - dir.FileDelta
		
		if kingRank >= MinRank && kingRank <= MaxRank && kingFile >= MinFile && kingFile <= MaxFile {
			if b.GetPiece(kingRank, kingFile) == kingPiece {
				return true
			}
		}
	}
	
	return false
}