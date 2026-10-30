package uci

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// MoveConverter handles conversion between internal Move representation and UCI notation.
type MoveConverter struct{}

// NewMoveConverter creates a new move converter.
func NewMoveConverter() *MoveConverter {
	return &MoveConverter{}
}

// ToUCI converts an internal Move to UCI notation (e.g., "e2e4", "e7e8q").
func (mc *MoveConverter) ToUCI(move board.Move) string {
	return move.UCI()
}

// FromUCI converts UCI notation to internal Move representation.
func (mc *MoveConverter) FromUCI(uciMove string, b *board.Board) (board.Move, error) {
	if b == nil {
		return board.Move{}, fmt.Errorf("board cannot be nil")
	}
	if len(uciMove) < 4 || len(uciMove) > 5 {
		return board.Move{}, fmt.Errorf("invalid UCI move format: %s", uciMove)
	}

	fromSquare, err := parseUCISquare(uciMove[0:2])
	if err != nil {
		return board.Move{}, fmt.Errorf("invalid from square in UCI move %s: %w", uciMove, err)
	}

	toSquare, err := parseUCISquare(uciMove[2:4])
	if err != nil {
		return board.Move{}, fmt.Errorf("invalid to square in UCI move %s: %w", uciMove, err)
	}

	piece := b.GetPiece(fromSquare.Rank, fromSquare.File)
	if piece == board.Empty {
		return board.Move{}, fmt.Errorf("no piece at from square %s", uciMove[0:2])
	}

	captured := b.GetPiece(toSquare.Rank, toSquare.File)

	move := board.Move{
		From:      fromSquare,
		To:        toSquare,
		Piece:     piece,
		Captured:  captured,
		Promotion: board.Empty,
		IsCapture: captured != board.Empty,
	}

	if len(uciMove) == 5 {
		promotion := parsePromotionPiece(uciMove[4], piece)
		move.Promotion = promotion
	}

	move.IsCastling = mc.isCastlingMove(move, piece)
	move.IsEnPassant = mc.isEnPassantMove(move, piece, b)

	return move, nil
}

func parseUCISquare(uciSquare string) (board.Square, error) {
	if len(uciSquare) != 2 {
		return board.Square{}, fmt.Errorf("invalid square notation: %s", uciSquare)
	}

	file := int(uciSquare[0] - 'a')
	rank := int(uciSquare[1] - '1')

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return board.Square{}, fmt.Errorf("square out of bounds: %s", uciSquare)
	}

	return board.Square{File: file, Rank: rank}, nil
}

func parsePromotionPiece(promotionChar byte, originalPiece board.Piece) board.Piece {
	isWhite := originalPiece >= 'A' && originalPiece <= 'Z'

	switch promotionChar {
	case 'q':
		if isWhite {
			return board.WhiteQueen
		}
		return board.BlackQueen
	case 'r':
		if isWhite {
			return board.WhiteRook
		}
		return board.BlackRook
	case 'b':
		if isWhite {
			return board.WhiteBishop
		}
		return board.BlackBishop
	case 'n':
		if isWhite {
			return board.WhiteKnight
		}
		return board.BlackKnight
	default:
		if isWhite {
			return board.WhiteQueen
		}
		return board.BlackQueen
	}
}

func (mc *MoveConverter) isCastlingMove(move board.Move, piece board.Piece) bool {
	if piece == board.WhiteKing || piece == board.BlackKing {
		if abs(move.To.File-move.From.File) == 2 && move.To.Rank == move.From.Rank {
			return true
		}
	}
	return false
}

func (mc *MoveConverter) isEnPassantMove(move board.Move, piece board.Piece, b *board.Board) bool {
	if b == nil {
		return false
	}
	if piece != board.WhitePawn && piece != board.BlackPawn {
		return false
	}
	if abs(move.To.File-move.From.File) != 1 {
		return false
	}
	if move.Captured != board.Empty {
		return false
	}

	enPassantTarget, hasEnPassant := b.GetEnPassantTarget()
	if hasEnPassant &&
		move.To.File == enPassantTarget.File &&
		move.To.Rank == enPassantTarget.Rank {
		return true
	}

	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
