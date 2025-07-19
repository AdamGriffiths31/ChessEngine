package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveGenerator defines the interface for generating legal chess moves.
// Uses high-performance bitboard operations for optimal speed.
type MoveGenerator interface {
	GenerateAllMoves(b *board.Board, player Player) *MoveList
}

// Generator implements the MoveGenerator interface providing complete chess move generation.
// Uses high-performance bitboard operations exclusively for optimal speed (3-5x faster than array-based).
// Includes specialized handlers for complex moves like castling, en passant, and promotion.
type Generator struct {
	castlingHandler  *CastlingHandler
	enPassantHandler *EnPassantHandler
	promotionHandler *PromotionHandler
	moveExecutor     *MoveExecutor
	attackDetector   *AttackDetector

	// King position cache for performance optimization
	whiteKingPos   *board.Square
	blackKingPos   *board.Square
	kingCacheValid bool

	// Bitboard move generator for high-performance move generation
	bitboardGenerator *BitboardMoveGenerator
}

// NewGenerator creates a new move generator with bitboard-based move generation.
// The generator includes optimizations like king position caching, object pooling,
// and high-performance bitboard operations for efficient move generation during search.
func NewGenerator() *Generator {
	return &Generator{
		castlingHandler:  &CastlingHandler{},
		enPassantHandler: &EnPassantHandler{},
		promotionHandler: &PromotionHandler{},
		moveExecutor:     &MoveExecutor{},
		attackDetector:   &AttackDetector{},

		// Initialize king cache as invalid
		whiteKingPos:   nil,
		blackKingPos:   nil,
		kingCacheValid: false,

		// Always use bitboard generation for optimal performance
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

// IsKingInCheck checks if the king of the given player is in check.
// Uses cached king positions for performance optimization.
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

// findKing finds the king's position for the given player using cache.
// Returns nil if no king is found (which indicates an invalid board state).
func (g *Generator) findKing(b *board.Board, player Player) *board.Square {
	// Initialize cache if not valid
	if !g.kingCacheValid {
		g.initializeKingCache(b)
	}

	// Return cached position
	if player == White {
		return g.whiteKingPos
	}
	return g.blackKingPos
}

// initializeKingCache scans the board once to find and cache both king positions
func (g *Generator) initializeKingCache(b *board.Board) {
	g.whiteKingPos = nil
	g.blackKingPos = nil

	// Scan the board once to find both kings
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			piece := b.GetPiece(rank, file)

			if piece == board.WhiteKing {
				g.whiteKingPos = &board.Square{File: file, Rank: rank}
			} else if piece == board.BlackKing {
				g.blackKingPos = &board.Square{File: file, Rank: rank}
			}

			// Early exit if we found both kings
			if g.whiteKingPos != nil && g.blackKingPos != nil {
				g.kingCacheValid = true
				return
			}
		}
	}

	g.kingCacheValid = true
}

// invalidateKingCache marks the king cache as invalid, forcing a rescan on next access
func (g *Generator) invalidateKingCache() {
	g.kingCacheValid = false
	g.whiteKingPos = nil
	g.blackKingPos = nil
}

// updateKingCache updates the cached king position when a king moves
func (g *Generator) updateKingCache(move board.Move) {
	// Only update if the moving piece is a king
	if move.Piece == board.WhiteKing {
		g.whiteKingPos = &board.Square{File: move.To.File, Rank: move.To.Rank}
	} else if move.Piece == board.BlackKing {
		g.blackKingPos = &board.Square{File: move.To.File, Rank: move.To.Rank}
	}
}

// makeMove is a wrapper that delegates to the MoveExecutor
func (g *Generator) makeMove(b *board.Board, move board.Move) *MoveHistory {
	return g.moveExecutor.MakeMove(b, move, g.updateBoardState)
}

// unmakeMove is a wrapper that delegates to the MoveExecutor
func (g *Generator) unmakeMove(b *board.Board, history *MoveHistory) {
	// Check if the original move involved a king before unmaking
	wasKingMove := history.Move.Piece == board.WhiteKing || history.Move.Piece == board.BlackKing

	g.moveExecutor.UnmakeMove(b, history)

	// Only invalidate cache if a king was moved
	if wasKingMove {
		g.invalidateKingCache()
	}
}

// IsSquareAttacked checks if a square is attacked by the enemy (public method)
func (g *Generator) IsSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	return g.attackDetector.IsSquareAttacked(b, square, player)
}

// Release releases any resources held by the generator
func (g *Generator) Release() {
	if g.bitboardGenerator != nil {
		g.bitboardGenerator.Release()
	}
}

