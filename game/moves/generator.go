package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveGenerator interface for generating legal moves
type MoveGenerator interface {
	GenerateAllMoves(b *board.Board, player Player) *MoveList
	GeneratePawnMoves(b *board.Board, player Player) *MoveList
	GenerateRookMoves(b *board.Board, player Player) *MoveList
	GenerateBishopMoves(b *board.Board, player Player) *MoveList
	GenerateKnightMoves(b *board.Board, player Player) *MoveList
	GenerateQueenMoves(b *board.Board, player Player) *MoveList
	GenerateKingMoves(b *board.Board, player Player) *MoveList
}

// Generator implements the MoveGenerator interface
type Generator struct{
	castlingHandler  *CastlingHandler
	enPassantHandler *EnPassantHandler
	promotionHandler *PromotionHandler
	moveExecutor     *MoveExecutor
	attackDetector   *AttackDetector
}

// NewGenerator creates a new move generator
func NewGenerator() *Generator {
	return &Generator{
		castlingHandler:  &CastlingHandler{},
		enPassantHandler: &EnPassantHandler{},
		promotionHandler: &PromotionHandler{},
		moveExecutor:     &MoveExecutor{},
		attackDetector:   &AttackDetector{},
	}
}

// GenerateAllMoves generates all legal moves for the given player
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
	pseudoLegalMoves := g.generateAllPseudoLegalMoves(b, player)
	legalMoves := NewMoveList()
	
	// Filter out moves that would leave the king in check
	for _, move := range pseudoLegalMoves.Moves {
		if g.isMoveLegal(b, move, player) {
			legalMoves.AddMove(move)
		}
	}
	
	return legalMoves
}

// generateAllPseudoLegalMoves generates all pseudo-legal moves (without check validation)
func (g *Generator) generateAllPseudoLegalMoves(b *board.Board, player Player) *MoveList {
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

	kingMoves := g.GenerateKingMoves(b, player)
	for _, move := range kingMoves.Moves {
		moveList.AddMove(move)
	}

	return moveList
}

// generateSlidingPieceMoves generates moves for sliding pieces (rook, bishop, queen)
func (g *Generator) generateSlidingPieceMoves(b *board.Board, player Player, pieceType board.Piece, directions []Direction) *MoveList {
	moveList := NewMoveList()
	
	// Scan the board for pieces of the specified type
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pieceType {
				// Generate moves for this piece
				fromSquare := board.Square{File: file, Rank: rank}
				for _, dir := range directions {
					g.generateSlidingMoves(b, player, fromSquare, dir.RankDelta, dir.FileDelta, moveList)
				}
			}
		}
	}
	
	return moveList
}

// generateJumpingPieceMoves generates moves for jumping pieces (knight, king)
func (g *Generator) generateJumpingPieceMoves(b *board.Board, player Player, pieceType board.Piece, directions []Direction) *MoveList {
	moveList := NewMoveList()
	
	// Scan the board for pieces of the specified type
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pieceType {
				// Generate moves for this piece
				fromSquare := board.Square{File: file, Rank: rank}
				for _, dir := range directions {
					g.generateJumpingMove(b, player, fromSquare, dir.RankDelta, dir.FileDelta, moveList)
				}
			}
		}
	}
	
	return moveList
}

// generateJumpingMove generates a single jumping move if valid
func (g *Generator) generateJumpingMove(b *board.Board, player Player, from board.Square, rankDelta, fileDelta int, moveList *MoveList) {
	newRank := from.Rank + rankDelta
	newFile := from.File + fileDelta
	
	// Check if the target square is within board boundaries
	if newRank >= MinRank && newRank <= MaxRank && newFile >= MinFile && newFile <= MaxFile {
		piece := b.GetPiece(newRank, newFile)
		to := board.Square{File: newFile, Rank: newRank}
		
		if piece == board.Empty {
			// Empty square - valid move
			move := g.createMove(b, from, to, false, board.Empty, board.Empty)
			moveList.AddMove(move)
		} else if g.isEnemyPiece(piece, player) {
			// Enemy piece - valid capture
			move := g.createMove(b, from, to, true, piece, board.Empty)
			moveList.AddMove(move)
		}
		// Own piece - can't move here
	}
}

// createMove creates a move with the given parameters
func (g *Generator) createMove(b *board.Board, from, to board.Square, isCapture bool, captured, promotion board.Piece) board.Move {
	piece := b.GetPiece(from.Rank, from.File) // Get the moving piece
	return board.Move{
		From:      from,
		To:        to,
		Piece:     piece,
		IsCapture: isCapture,
		Captured:  captured,
		Promotion: promotion,
	}
}

// generateSlidingMoves generates moves in a straight line until blocked or edge reached
func (g *Generator) generateSlidingMoves(b *board.Board, player Player, from board.Square, rankDelta, fileDelta int, moveList *MoveList) {
	currentRank := from.Rank + rankDelta
	currentFile := from.File + fileDelta
	
	// Continue sliding in the direction until we hit the board edge or a piece
	for currentRank >= MinRank && currentRank <= MaxRank && currentFile >= MinFile && currentFile <= MaxFile {
		piece := b.GetPiece(currentRank, currentFile)
		to := board.Square{File: currentFile, Rank: currentRank}
		
		if piece == board.Empty {
			// Empty square - valid move
			move := g.createMove(b, from, to, false, board.Empty, board.Empty)
			moveList.AddMove(move)
		} else if g.isEnemyPiece(piece, player) {
			// Enemy piece - valid capture, but can't continue sliding
			move := g.createMove(b, from, to, true, piece, board.Empty)
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

// IsKingInCheck checks if the king of the given player is in check
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	return g.IsKingInCheckFast(b, player)
}

// IsKingInCheckFast optimized version that avoids repeated king searches
func (g *Generator) IsKingInCheckFast(b *board.Board, player Player) bool {
	kingSquare := g.findKing(b, player)
	if kingSquare == nil {
		return false // No king found
	}
	
	return g.isSquareUnderAttack(b, *kingSquare, player)
}

// isSquareUnderAttack checks if a square is attacked by the enemy player
func (g *Generator) isSquareUnderAttack(b *board.Board, square board.Square, player Player) bool {
	return g.attackDetector.IsSquareAttacked(b, square, player)
}

// findKing finds the king's position for the given player
func (g *Generator) findKing(b *board.Board, player Player) *board.Square {
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			if b.GetPiece(rank, file) == kingPiece {
				return &board.Square{File: file, Rank: rank}
			}
		}
	}
	return nil
}

// makeMove is a wrapper that delegates to the MoveExecutor
func (g *Generator) makeMove(b *board.Board, move board.Move) *MoveHistory {
	return g.moveExecutor.MakeMove(b, move, g.updateBoardState)
}

// unmakeMove is a wrapper that delegates to the MoveExecutor
func (g *Generator) unmakeMove(b *board.Board, history *MoveHistory) {
	g.moveExecutor.UnmakeMove(b, history)
}

// IsSquareAttacked checks if a square is attacked by the enemy (public method)
func (g *Generator) IsSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	return g.attackDetector.IsSquareAttacked(b, square, player)
}

// isEnemyPiece checks if a piece belongs to the enemy
func (g *Generator) isEnemyPiece(piece board.Piece, player Player) bool {
	if player == White {
		// White player - enemy pieces are lowercase (black)
		return piece >= 'a' && piece <= 'z'
	} else {
		// Black player - enemy pieces are uppercase (white)
		return piece >= 'A' && piece <= 'Z'
	}
}

// isMoveLegal checks if a move is legal (doesn't leave king in check)
func (g *Generator) isMoveLegal(b *board.Board, move board.Move, player Player) bool {
	// First, perform basic move validation
	if !g.isValidBasicMove(b, move, player) {
		return false
	}
	
	// For castling moves, use special validation
	if move.IsCastling {
		// Check that king is not currently in check
		if g.IsKingInCheck(b, player) {
			return false
		}
		
		// Create attack detection function for castling handler
		isSquareAttacked := func(square board.Square) bool {
			return g.attackDetector.IsSquareAttacked(b, square, player)
		}
		
		return g.castlingHandler.IsLegal(b, move, player, isSquareAttacked)
	}
	
	// For en passant moves, use special validation
	if move.IsEnPassant {
		// Create helper functions for en passant handler
		makeMove := func(m board.Move) *MoveHistory {
			return g.makeMove(b, m)
		}
		unmakeMove := func(h *MoveHistory) {
			g.unmakeMove(b, h)
		}
		isKingInCheck := func() bool {
			return g.IsKingInCheck(b, player)
		}
		
		return g.enPassantHandler.IsLegal(b, move, player, makeMove, unmakeMove, isKingInCheck)
	}
	
	// For all other moves, test by making the move and checking if king is in check
	history := g.makeMove(b, move)
	legal := !g.IsKingInCheck(b, player)
	g.unmakeMove(b, history)
	
	return legal
}

// isValidBasicMove performs comprehensive basic move validation
func (g *Generator) isValidBasicMove(b *board.Board, move board.Move, player Player) bool {
	// Check bounds
	if !g.isSquareValid(move.From) || !g.isSquareValid(move.To) {
		return false
	}
	
	// Check piece exists at from square
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == board.Empty {
		return false
	}
	
	// Check piece belongs to current player
	if !g.isPieceOwnedByPlayer(piece, player) {
		return false
	}
	
	// Check destination square
	destPiece := b.GetPiece(move.To.Rank, move.To.File)
	if destPiece != board.Empty {
		// If destination has a piece, it must be enemy piece (capture)
		if !g.isEnemyPiece(destPiece, player) {
			return false // Can't capture own piece
		}
	}
	
	return true
}

// isSquareValid checks if a square is within board boundaries
func (g *Generator) isSquareValid(square board.Square) bool {
	return square.Rank >= MinRank && square.Rank <= MaxRank && 
		   square.File >= MinFile && square.File <= MaxFile
}

// isPieceOwnedByPlayer checks if a piece belongs to the current player
func (g *Generator) isPieceOwnedByPlayer(piece board.Piece, player Player) bool {
	if player == White {
		return piece >= 'A' && piece <= 'Z'
	} else {
		return piece >= 'a' && piece <= 'z'
	}
}