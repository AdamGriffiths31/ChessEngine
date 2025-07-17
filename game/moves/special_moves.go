package moves

import (
	"strings"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// CastlingHandler handles all castling-related logic
type CastlingHandler struct{}

// IsLegal checks if a castling move is legal
func (ch *CastlingHandler) IsLegal(b *board.Board, move board.Move, player Player, isSquareAttacked func(board.Square) bool) bool {
	// Basic castling checks
	if !ch.isBasicValid(b, move, player) {
		return false
	}
	
	// Check that king doesn't pass through check
	return ch.isPathSafe(b, move, player, isSquareAttacked)
}

// isBasicValid performs basic castling validation
func (ch *CastlingHandler) isBasicValid(b *board.Board, move board.Move, player Player) bool {
	// Check castling rights
	castlingRights := b.GetCastlingRights()
	
	var kingside bool
	var hasRight bool
	
	if move.To.File == KingsideFile { // Kingside
		kingside = true
		if player == White {
			hasRight = strings.Contains(castlingRights, "K")
		} else {
			hasRight = strings.Contains(castlingRights, "k")
		}
	} else if move.To.File == QueensideFile { // Queenside
		kingside = false
		if player == White {
			hasRight = strings.Contains(castlingRights, "Q")
		} else {
			hasRight = strings.Contains(castlingRights, "q")
		}
	} else {
		return false // Invalid castling destination
	}
	
	if !hasRight {
		return false
	}
	
	// Check that path is clear
	return ch.isPathClear(b, move, player, kingside)
}

// isPathClear checks if the castling path is clear of pieces
func (ch *CastlingHandler) isPathClear(b *board.Board, move board.Move, player Player, kingside bool) bool {
	rank := move.From.Rank
	
	if kingside {
		// Check f and g squares are empty
		return b.GetPiece(rank, KingsideRookToFile) == board.Empty && b.GetPiece(rank, KingsideFile) == board.Empty
	} else {
		// Check b, c, and d squares are empty
		return b.GetPiece(rank, 1) == board.Empty && b.GetPiece(rank, QueensideFile) == board.Empty && b.GetPiece(rank, QueensideRookToFile) == board.Empty
	}
}

// isPathSafe checks castling path safety
func (ch *CastlingHandler) isPathSafe(b *board.Board, move board.Move, player Player, isSquareAttacked func(board.Square) bool) bool {
	// Test each square the king passes through
	startFile := move.From.File
	endFile := move.To.File
	
	var filesToCheck []int
	if endFile > startFile {
		// Kingside castling: check e, f, g
		filesToCheck = []int{startFile, startFile + 1, startFile + 2}
	} else {
		// Queenside castling: check e, d, c
		filesToCheck = []int{startFile, startFile - 1, startFile - 2}
	}
	
	for _, file := range filesToCheck {
		testSquare := board.Square{File: file, Rank: move.From.Rank}
		if isSquareAttacked(testSquare) {
			return false
		}
	}
	
	return true
}

// EnPassantHandler handles all en passant-related logic
type EnPassantHandler struct{}

// IsLegal checks if an en passant move is legal
func (eh *EnPassantHandler) IsLegal(b *board.Board, move board.Move, player Player, makeMove func(board.Move) *MoveHistory, unmakeMove func(*MoveHistory), isKingInCheck func() bool) bool {
	// Basic en passant validation is already done in generation
	// Additional check: make sure the move doesn't leave king in check
	history := makeMove(move)
	legal := !isKingInCheck()
	unmakeMove(history)
	
	return legal
}

// PromotionHandler handles all promotion-related logic
type PromotionHandler struct{}

// AddPromotionMoves adds all four promotion moves (Q, R, B, N)
func (ph *PromotionHandler) AddPromotionMoves(b *board.Board, from, to board.Square, player Player, moveList *MoveList, createMove func(board.Square, board.Square, bool, board.Piece, board.Piece) board.Move) {
	var promotionPieces []board.Piece
	
	if player == White {
		promotionPieces = []board.Piece{
			board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight,
		}
	} else {
		promotionPieces = []board.Piece{
			board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight,
		}
	}

	// Check if this is a capture promotion
	destPiece := b.GetPiece(to.Rank, to.File)
	isCapture := destPiece != board.Empty

	for _, piece := range promotionPieces {
		move := createMove(from, to, isCapture, destPiece, piece)
		moveList.AddMove(move)
	}
}