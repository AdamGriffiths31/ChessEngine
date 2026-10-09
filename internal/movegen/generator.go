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
	MoveExecutor      *MoveExecutor
	attackDetector    *AttackDetector
	bitboardGenerator *BitboardMoveGenerator
}

// NewGenerator creates a new move generator with bitboard-based move generation.
func NewGenerator() *Generator {
	return &Generator{
		MoveExecutor:      &MoveExecutor{},
		attackDetector:    &AttackDetector{},
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

// findKing finds the king's position for the given player.
// Returns a Square with File=-1 if no king is found.
func (g *Generator) findKing(b *board.Board, player Player) board.Square {
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}

	kingBitboard := b.GetPieceBitboard(kingPiece)
	if kingBitboard == 0 {
		return board.Square{File: -1, Rank: -1}
	}

	squareIndex := kingBitboard.LSB()
	if squareIndex == -1 {
		return board.Square{File: -1, Rank: -1}
	}

	file, rank := board.SquareToFileRank(squareIndex)
	return board.Square{File: file, Rank: rank}
}

// updateBoardState updates castling rights, en passant, and move counters
//
//nolint:gocyclo // refactored in Phase 4
func (g *Generator) updateBoardState(b *board.Board, move board.Move) {
	piece := b.GetPiece(move.To.Rank, move.To.File)

	// Castling-rights bookkeeping is shared with board.Board.MakeMove via
	// UpdateCastlingRights rather than reimplemented here - see that
	// method's doc comment for why (this file used to have its own
	// separate copy, and the two diverged).
	b.UpdateCastlingRights(move, piece)

	if piece == board.WhitePawn || piece == board.BlackPawn {
		if abs(move.To.Rank-move.From.Rank) == 2 {
			targetRank := (move.From.Rank + move.To.Rank) / 2
			enPassantTarget := board.Square{File: move.From.File, Rank: targetRank}
			b.SetEnPassantTarget(enPassantTarget, true)
		} else {
			b.SetEnPassantTarget(board.Square{}, false)
		}
	} else {
		b.SetEnPassantTarget(board.Square{}, false)
	}

	halfMoveClock := b.GetHalfMoveClock()
	if move.IsCapture || piece == board.WhitePawn || piece == board.BlackPawn {
		halfMoveClock = 0
	} else {
		halfMoveClock++
	}
	b.SetHalfMoveClock(halfMoveClock)

	if b.GetSideToMove() == "b" {
		b.SetFullMoveNumber(b.GetFullMoveNumber() + 1)
	}
}
