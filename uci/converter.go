package uci

import (
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// MoveConverter handles conversion between internal Move representation and UCI notation
type MoveConverter struct{}

// NewMoveConverter creates a new move converter
func NewMoveConverter() *MoveConverter {
	return &MoveConverter{}
}

// ToUCI converts an internal Move to UCI notation (e.g., "e2e4", "e7e8q")
func (mc *MoveConverter) ToUCI(move board.Move) string {
	from := squareToUCI(move.From)
	to := squareToUCI(move.To)
	
	// Handle promotion
	if move.Promotion != board.Empty {
		promotion := strings.ToLower(string(move.Promotion))
		return from + to + promotion
	}
	
	return from + to
}

// FromUCI converts UCI notation to internal Move representation
func (mc *MoveConverter) FromUCI(uciMove string, b *board.Board) (board.Move, error) {
	if len(uciMove) < 4 || len(uciMove) > 5 {
		return board.Move{}, fmt.Errorf("invalid UCI move format: %s", uciMove)
	}
	
	// Parse from square
	fromSquare, err := parseUCISquare(uciMove[0:2])
	if err != nil {
		return board.Move{}, fmt.Errorf("invalid from square in UCI move %s: %v", uciMove, err)
	}
	
	// Parse to square
	toSquare, err := parseUCISquare(uciMove[2:4])
	if err != nil {
		return board.Move{}, fmt.Errorf("invalid to square in UCI move %s: %v", uciMove, err)
	}
	
	// Get the piece being moved
	piece := b.GetPiece(fromSquare.Rank, fromSquare.File)
	if piece == board.Empty {
		return board.Move{}, fmt.Errorf("no piece at from square %s", uciMove[0:2])
	}
	
	// Get captured piece (if any)
	captured := b.GetPiece(toSquare.Rank, toSquare.File)
	
	move := board.Move{
		From:      fromSquare,
		To:        toSquare,
		Piece:     piece,
		Captured:  captured,
		IsCapture: captured != board.Empty,
	}
	
	// Handle promotion
	if len(uciMove) == 5 {
		promotion := parsePromotionPiece(uciMove[4], piece)
		move.Promotion = promotion
	}
	
	// Detect special moves
	move.IsCastling = mc.isCastlingMove(move, piece)
	move.IsEnPassant = mc.isEnPassantMove(move, piece, b)
	
	return move, nil
}

// squareToUCI converts a Square to UCI notation (e.g., Square{0,0} -> "a1")
func squareToUCI(square board.Square) string {
	file := rune('a' + square.File)
	rank := rune('1' + square.Rank)
	return string(file) + string(rank)
}

// parseUCISquare converts UCI square notation to Square struct
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

// parsePromotionPiece converts UCI promotion character to piece
func parsePromotionPiece(promotionChar byte, originalPiece board.Piece) board.Piece {
	// Determine if the piece is white or black
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
		// Default to queen if invalid promotion piece
		if isWhite {
			return board.WhiteQueen
		}
		return board.BlackQueen
	}
}

// isCastlingMove detects if a move is a castling move
func (mc *MoveConverter) isCastlingMove(move board.Move, piece board.Piece) bool {
	// King moves
	if piece == board.WhiteKing || piece == board.BlackKing {
		// Check if it's a 2-square horizontal move (castling)
		if abs(move.To.File-move.From.File) == 2 && move.To.Rank == move.From.Rank {
			return true
		}
	}
	return false
}

// isEnPassantMove detects if a move is an en passant capture
func (mc *MoveConverter) isEnPassantMove(move board.Move, piece board.Piece, b *board.Board) bool {
	// Must be a pawn move
	if piece != board.WhitePawn && piece != board.BlackPawn {
		return false
	}
	
	// Must be a diagonal move
	if abs(move.To.File-move.From.File) != 1 {
		return false
	}
	
	// Target square must be empty (captured piece is not on target square in en passant)
	if move.Captured != board.Empty {
		return false
	}
	
	// Check if target square matches en passant target
	enPassantTarget := b.GetEnPassantTarget()
	if enPassantTarget != nil && 
		move.To.File == enPassantTarget.File && 
		move.To.Rank == enPassantTarget.Rank {
		return true
	}
	
	return false
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}