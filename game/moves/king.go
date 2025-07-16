package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Castling constants
const (
	KingStartFile    = 4 // King starts on file e (index 4)
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
	// Determine king piece based on player
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
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
					Piece:      kingPiece,
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
					Piece:      kingPiece,
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
	castlingRights := b.GetCastlingRights()
	
	var kingStartRank int
	var kingPiece board.Piece
	var rookPiece board.Piece
	var kingsideRight, queensideRight rune
	
	if player == White {
		kingStartRank = WhiteKingStartRank
		kingPiece = board.WhiteKing
		rookPiece = board.WhiteRook
		kingsideRight = 'K'
		queensideRight = 'Q'
	} else {
		kingStartRank = BlackKingStartRank
		kingPiece = board.BlackKing
		rookPiece = board.BlackRook
		kingsideRight = 'k'
		queensideRight = 'q'
	}
	
	// Only allow castling if king is on starting square
	if from.Rank != kingStartRank || from.File != KingStartFile {
		return
	}
	
	// Check kingside castling (O-O)
	if g.hasCastlingRight(castlingRights, kingsideRight) && g.canCastleKingside(b, player, kingStartRank, rookPiece) && 
	   g.isCastlingPathSafe(b, player, from, board.Square{File: KingsideTargetFile, Rank: kingStartRank}) {
		castlingMove := board.Move{
			From:       from,
			To:         board.Square{File: KingsideTargetFile, Rank: kingStartRank},
			Piece:      kingPiece,
			IsCapture:  false,
			IsCastling: true,
			Promotion:  board.Empty,
		}
		moveList.AddMove(castlingMove)
	}
	
	// Check queenside castling (O-O-O)
	if g.hasCastlingRight(castlingRights, queensideRight) && g.canCastleQueenside(b, player, kingStartRank, rookPiece) && 
	   g.isCastlingPathSafe(b, player, from, board.Square{File: QueensideTargetFile, Rank: kingStartRank}) {
		castlingMove := board.Move{
			From:       from,
			To:         board.Square{File: QueensideTargetFile, Rank: kingStartRank},
			Piece:      kingPiece,
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

// hasCastlingRight checks if a specific castling right is available
func (g *Generator) hasCastlingRight(castlingRights string, right rune) bool {
	for _, r := range castlingRights {
		if r == right {
			return true
		}
	}
	return false
}

// isCastlingPathSafe checks if the king doesn't pass through or end up in check during castling
func (g *Generator) isCastlingPathSafe(b *board.Board, player Player, from, to board.Square) bool {
	// Check each square the king passes through (including destination)
	startFile := from.File
	endFile := to.File
	
	// Determine direction and squares to check
	var filesToCheck []int
	if endFile > startFile {
		// Kingside castling: check e1, f1, g1
		filesToCheck = []int{startFile, startFile + 1, startFile + 2}
	} else {
		// Queenside castling: check e1, d1, c1
		filesToCheck = []int{startFile, startFile - 1, startFile - 2}
	}
	
	// Test each square for safety
	for _, file := range filesToCheck {
		testSquare := board.Square{File: file, Rank: from.Rank}
		if g.isSquareAttacked(b, testSquare, player) {
			return false
		}
	}
	
	return true
}

// isSquareAttacked checks if a square is attacked by the enemy
func (g *Generator) isSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	enemyPlayer := Black
	if player == Black {
		enemyPlayer = White
	}
	
	// Generate all enemy moves (without castling to avoid recursion)
	enemyMoves := g.generateMovesWithoutCastling(b, enemyPlayer)
	
	for _, move := range enemyMoves.Moves {
		if move.To.File == square.File && move.To.Rank == square.Rank {
			return true
		}
	}
	
	return false
}

// generateMovesWithoutCastling generates all moves except castling (to avoid recursion)
func (g *Generator) generateMovesWithoutCastling(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()

	pawnMoves := g.GeneratePawnMoves(b, player)
	for _, move := range pawnMoves.Moves {
		moveList.AddMove(move)
	}

	rookMoves := g.GenerateRookMoves(b, player)
	for _, move := range rookMoves.Moves {
		moveList.AddMove(move)
	}

	bishopMoves := g.GenerateBishopMoves(b, player)
	for _, move := range bishopMoves.Moves {
		moveList.AddMove(move)
	}

	knightMoves := g.GenerateKnightMoves(b, player)
	for _, move := range knightMoves.Moves {
		moveList.AddMove(move)
	}

	queenMoves := g.GenerateQueenMoves(b, player)
	for _, move := range queenMoves.Moves {
		moveList.AddMove(move)
	}

	// Generate only non-castling king moves
	kingMoves := g.generateKingNonCastlingMoves(b, player)
	for _, move := range kingMoves.Moves {
		moveList.AddMove(move)
	}

	return moveList
}

// generateKingNonCastlingMoves generates only single-square king moves (no castling)
func (g *Generator) generateKingNonCastlingMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()
	
	kingSquare := g.findKingSquare(b, player)
	if kingSquare != nil {
		// Generate only single-square moves for this king
		g.generateKingSingleMoves(b, player, *kingSquare, moveList)
	}
	
	return moveList
}