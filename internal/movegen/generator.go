package movegen

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// MoveGenerator defines the interface for generating legal chess moves.
type MoveGenerator interface {
	GenerateAllMoves(b *board.Board, player Player) *MoveList
	GeneratePseudoLegalMoves(b *board.Board, player Player) *MoveList
}

// Generator implements the MoveGenerator interface for complete chess move generation.
// Uses high-performance bitboard operations for all move types.
type Generator struct {
	bitboardGenerator *BitboardMoveGenerator
}

// NewGenerator creates a new move generator with bitboard-based move generation.
func NewGenerator() *Generator {
	return &Generator{
		bitboardGenerator: NewBitboardMoveGenerator(),
	}
}

// GenerateAllMoves generates all legal moves for the given player.
// Returns an empty list if board is nil.
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
	if b == nil {
		return GetMoveList()
	}

	return g.bitboardGenerator.GenerateAllMovesBitboard(b, player)
}

// GeneratePseudoLegalMoves generates pseudo-legal moves without king safety validation.
// Callers must verify moves don't leave their own king in check before execution.
// This method is optimized for search algorithms that use lazy move validation.
// Returns an empty list if board is nil.
func (g *Generator) GeneratePseudoLegalMoves(b *board.Board, player Player) *MoveList {
	if b == nil {
		return GetMoveList()
	}

	return g.bitboardGenerator.GeneratePseudoLegalMoves(b, player)
}

// GenerateCaptures generates pseudo-legal captures and promotions only.
// Intended for quiescence search; quiet non-promotion moves are not generated.
// Returns an empty list if board is nil.
func (g *Generator) GenerateCaptures(b *board.Board, player Player) *MoveList {
	if b == nil {
		return GetMoveList()
	}

	return g.bitboardGenerator.GenerateCaptures(b, player)
}

// IsKingInCheck checks if the king of the given player is currently in check.
// Returns false if board is nil or king is not found.
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	if b == nil {
		return false
	}

	var color board.BitboardColor
	if player == White {
		color = board.BitboardWhite
	} else {
		color = board.BitboardBlack
	}

	return b.IsInCheck(color)
}
