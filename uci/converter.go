// Package uci provides Universal Chess Interface protocol implementation.
package uci

import (
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// MoveConverter handles conversion between internal Move representation and UCI notation.
type MoveConverter struct{}

// NewMoveConverter creates a new move converter.
func NewMoveConverter() *MoveConverter {
	return &MoveConverter{}
}

// ToUCI converts an internal Move to UCI notation (e.g., "e2e4", "e7e8q").
func (mc *MoveConverter) ToUCI(move board.Move) string {
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 ||
		move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return "0000"
	}

	from := squareToUCI(move.From)
	to := squareToUCI(move.To)

	uciMove := from + to
	
	// Convert Chess960 castling notation to standard UCI notation
	// This fixes opening book data that contains Chess960 format moves
	switch uciMove {
	case "e1h1": // White kingside castling
		uciMove = "e1g1"
	case "e1a1": // White queenside castling
		uciMove = "e1c1"
	case "e8h8": // Black kingside castling
		uciMove = "e8g8"
	case "e8a8": // Black queenside castling
		uciMove = "e8c8"
	}

	if move.Promotion != board.Empty {
		promotion := strings.ToLower(string(move.Promotion))
		return uciMove + promotion
	}

	return uciMove
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

func squareToUCI(square board.Square) string {
	if square.File < 0 || square.File > 7 || square.Rank < 0 || square.Rank > 7 {
		return "0000"
	}

	file := rune('a' + square.File)
	rank := rune('1' + square.Rank)
	return string(file) + string(rank)
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

	enPassantTarget := b.GetEnPassantTarget()
	if enPassantTarget != nil &&
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
