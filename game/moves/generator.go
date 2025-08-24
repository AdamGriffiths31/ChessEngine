package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

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

// IsKingInCheck checks if the king of the given player is currently in check.
// Returns false if board is nil or king is not found.
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	kingSquare := g.findKing(b, player)
	if kingSquare.File == -1 {
		return false
	}

	return g.attackDetector.IsSquareAttacked(b, kingSquare, player)
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
func (g *Generator) updateBoardState(b *board.Board, move board.Move) {
	castlingRights := b.GetCastlingRights()
	piece := b.GetPiece(move.To.Rank, move.To.File)

	if piece == board.WhiteKing {
		castlingRights = g.removeCastlingRights(castlingRights, "KQ")
	} else if piece == board.BlackKing {
		castlingRights = g.removeCastlingRights(castlingRights, "kq")
	}

	if piece == board.WhiteRook {
		if move.From.File == QueensideRookFromFile && move.From.Rank == 0 {
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.From.File == KingsideRookFromFile && move.From.Rank == 0 {
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		}
	} else if piece == board.BlackRook {
		if move.From.File == QueensideRookFromFile && move.From.Rank == 7 {
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.From.File == KingsideRookFromFile && move.From.Rank == 7 {
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}

	if move.IsCapture {
		if move.To.File == QueensideRookFromFile && move.To.Rank == 0 {
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.To.File == KingsideRookFromFile && move.To.Rank == 0 {
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		} else if move.To.File == QueensideRookFromFile && move.To.Rank == 7 {
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.To.File == KingsideRookFromFile && move.To.Rank == 7 {
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}

	b.SetCastlingRights(castlingRights)

	if piece == board.WhitePawn || piece == board.BlackPawn {
		if abs(move.To.Rank-move.From.Rank) == 2 {
			targetRank := (move.From.Rank + move.To.Rank) / 2
			enPassantTarget := &board.Square{File: move.From.File, Rank: targetRank}
			b.SetEnPassantTarget(enPassantTarget)
		} else {
			b.SetEnPassantTarget(nil)
		}
	} else {
		b.SetEnPassantTarget(nil)
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

// removeCastlingRights removes specific castling rights from the string
func (g *Generator) removeCastlingRights(rights, toRemove string) string {
	result := ""
	for _, r := range rights {
		remove := false
		for _, removeR := range toRemove {
			if r == removeR {
				remove = true
				break
			}
		}
		if !remove {
			result += string(r)
		}
	}
	if result == "" {
		return "-"
	}
	return result
}
