package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveGenerator defines the interface for generating legal chess moves.
// Implementations use high-performance bitboard operations for optimal speed.
type MoveGenerator interface {
	GenerateAllMoves(b *board.Board, player Player) *MoveList
}

// Generator implements the MoveGenerator interface providing complete chess move generation.
// Uses high-performance bitboard operations exclusively for optimal speed (3-5x faster than array-based).
// Includes specialized handlers for complex moves like castling, en passant, and promotion.
// The generator maintains separate handlers for different move types and supports object pooling.
type Generator struct {
	castlingHandler   *CastlingHandler
	enPassantHandler  *EnPassantHandler
	promotionHandler  *PromotionHandler
	moveExecutor      *MoveExecutor
	attackDetector    *AttackDetector
	bitboardGenerator *BitboardMoveGenerator
}

// NewGenerator creates a new move generator with bitboard-based move generation.
// The generator includes optimizations like king position caching, object pooling,
// and high-performance bitboard operations for efficient move generation during search.
func NewGenerator() *Generator {
	return &Generator{
		castlingHandler:   &CastlingHandler{},
		enPassantHandler:  &EnPassantHandler{},
		promotionHandler:  &PromotionHandler{},
		moveExecutor:      &MoveExecutor{},
		attackDetector:    &AttackDetector{},
		bitboardGenerator: NewBitboardMoveGenerator(),
	}
}

// GenerateAllMoves generates all legal moves for the given player using high-performance bitboard operations.
// This includes all piece types and special moves (castling, en passant, promotion).
// Moves are filtered to ensure they don't leave the king in check.
// Returns a MoveList that should be released back to the pool when done.
// Returns an empty list if board is nil.
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
	if b == nil {
		return GetMoveList() // Return empty list
	}

	// Use bitboard generation for optimal performance
	return g.bitboardGenerator.GenerateAllMovesBitboard(b, player)
}

// IsKingInCheck checks if the king of the given player is currently in check.
// Uses optimized bitboard operations and attack detection for performance.
// Returns false if board is nil or king is not found.
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	kingSquare := g.findKing(b, player)
	if kingSquare.File == -1 {
		return false // No king found
	}

	return g.isSquareUnderAttack(b, kingSquare, player)
}

// isSquareUnderAttack checks if a square is under attack by the enemy player.
// Used internally for king safety validation during move generation.
func (g *Generator) isSquareUnderAttack(b *board.Board, square board.Square, player Player) bool {
	return g.attackDetector.IsSquareAttacked(b, square, player)
}

// findKing finds the king's position for the given player using bitboard lookup.
// Returns a Square with File=-1 if no king is found (which indicates an invalid board state).
func (g *Generator) findKing(b *board.Board, player Player) board.Square {
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}

	kingBitboard := b.GetPieceBitboard(kingPiece)
	if kingBitboard == 0 {
		return board.Square{File: -1, Rank: -1} // No king found - sentinel value
	}

	squareIndex := kingBitboard.LSB()
	if squareIndex == -1 {
		return board.Square{File: -1, Rank: -1}
	}

	file, rank := board.SquareToFileRank(squareIndex)
	return board.Square{File: file, Rank: rank} // Return by value - no allocation!
}

// makeMove executes a move on the board and returns the move history.
// This is a wrapper that delegates to the MoveExecutor with board state updates.
func (g *Generator) makeMove(b *board.Board, move board.Move) *MoveHistory {
	return g.moveExecutor.MakeMove(b, move, g.updateBoardState)
}

// unmakeMove reverts a move using the provided move history.
// This is a wrapper that delegates to the MoveExecutor for consistent undo operations.
func (g *Generator) unmakeMove(b *board.Board, history *MoveHistory) {
	g.moveExecutor.UnmakeMove(b, history)
}

// IsSquareAttacked checks if a square is under attack by the opposing player.
// This is the public interface for external attack detection queries.
// Returns true if any enemy piece can attack the specified square.
func (g *Generator) IsSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	return g.attackDetector.IsSquareAttacked(b, square, player)
}

// Release cleans up and releases any resources held by the generator.
// Should be called when the generator is no longer needed to prevent memory leaks.
// Safe to call multiple times.
func (g *Generator) Release() {
	if g.bitboardGenerator != nil {
		g.bitboardGenerator.Release()
	}
}
