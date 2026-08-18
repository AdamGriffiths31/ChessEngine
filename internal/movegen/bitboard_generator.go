// Package movegen provides high-performance chess move generation using bitboard operations.
package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// BitboardMoveGenerator provides high-performance move generation using bitboard operations
// This implementation aims for 3-5x performance improvement over array-based generation
type BitboardMoveGenerator struct {
	// No fields needed - all operations use pooled MoveList objects
}

// NewBitboardMoveGenerator creates a new bitboard-based move generator
func NewBitboardMoveGenerator() *BitboardMoveGenerator {
	return &BitboardMoveGenerator{}
}

// GenerateAllMovesBitboard generates all legal moves using bitboard operations
// This is the main entry point that coordinates all piece-specific generators
func (bmg *BitboardMoveGenerator) GenerateAllMovesBitboard(b *board.Board, player Player) *MoveList {

	moveList := GetMoveList()

	bmg.generatePawnMovesBitboard(b, player, moveList)
	bmg.GenerateKnightMovesBitboard(b, player, moveList)
	bmg.generateBishopMovesBitboard(b, player, moveList)
	bmg.generateRookMovesBitboard(b, player, moveList)
	bmg.generateQueenMovesBitboard(b, player, moveList)
	bmg.generateKingMovesBitboard(b, player, moveList)

	bmg.filterLegalMovesInPlace(b, player, moveList)

	return moveList
}

// GeneratePseudoLegalMoves generates pseudo-legal moves without king safety validation.
// Callers must verify moves don't leave their own king in check before execution.
// This method is optimized for search algorithms that use lazy move validation.
func (bmg *BitboardMoveGenerator) GeneratePseudoLegalMoves(b *board.Board, player Player) *MoveList {

	moveList := GetMoveList()

	bmg.generatePawnMovesBitboard(b, player, moveList)
	bmg.GenerateKnightMovesBitboard(b, player, moveList)
	bmg.generateBishopMovesBitboard(b, player, moveList)
	bmg.generateRookMovesBitboard(b, player, moveList)
	bmg.generateQueenMovesBitboard(b, player, moveList)
	bmg.generateKingMovesBitboard(b, player, moveList)

	return moveList
}

// GenerateCaptures generates pseudo-legal captures and promotions only.
// This is the move set quiescence search examines; quiet non-promotion moves
// are never generated, avoiding the cost of building and discarding them.
func (bmg *BitboardMoveGenerator) GenerateCaptures(b *board.Board, player Player) *MoveList {
	moveList := GetMoveList()

	var pawnPiece, knightPiece, bishopPiece, rookPiece, queenPiece, kingPiece board.Piece
	var enemyColor board.BitboardColor
	var promotionRank int

	if player == White {
		pawnPiece = board.WhitePawn
		knightPiece = board.WhiteKnight
		bishopPiece = board.WhiteBishop
		rookPiece = board.WhiteRook
		queenPiece = board.WhiteQueen
		kingPiece = board.WhiteKing
		enemyColor = board.BitboardBlack
		promotionRank = 7
	} else {
		pawnPiece = board.BlackPawn
		knightPiece = board.BlackKnight
		bishopPiece = board.BlackBishop
		rookPiece = board.BlackRook
		queenPiece = board.BlackQueen
		kingPiece = board.BlackKing
		enemyColor = board.BitboardWhite
		promotionRank = 0
	}

	enemyPieces := b.GetColorBitboard(enemyColor)
	occupancy := b.AllPieces

	pawns := b.GetPieceBitboard(pawnPiece)
	if pawns != 0 {
		bmg.generatePawnCapturesBitboard(b, player, pawns, enemyPieces, moveList, promotionRank)
		bmg.generateEnPassantCapturesBitboard(b, player, pawns, moveList)

		// Quiet promotions: single pushes onto the promotion rank
		emptySquares := ^occupancy
		var promotionPushes board.Bitboard
		if player == White {
			promotionPushes = pawns.ShiftNorth() & emptySquares & board.RankMask(promotionRank)
		} else {
			promotionPushes = pawns.ShiftSouth() & emptySquares & board.RankMask(promotionRank)
		}
		for promotionPushes != 0 {
			toSquare, newBitboard := promotionPushes.PopLSB()
			promotionPushes = newBitboard

			var fromSquare int
			if player == White {
				fromSquare = toSquare - 8
			} else {
				fromSquare = toSquare + 8
			}
			bmg.addPromotionMoves(moveList, fromSquare, toSquare, pawnPiece, board.Empty, player)
		}
	}

	knights := b.GetPieceBitboard(knightPiece)
	for knights != 0 {
		fromSquare, newBitboard := knights.PopLSB()
		knights = newBitboard

		captures := board.GetKnightAttacks(fromSquare) & enemyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, knightPiece)
	}

	bishops := b.GetPieceBitboard(bishopPiece)
	for bishops != 0 {
		fromSquare, newBitboard := bishops.PopLSB()
		bishops = newBitboard

		captures := board.GetBishopAttacks(fromSquare, occupancy) & enemyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, bishopPiece)
	}

	rooks := b.GetPieceBitboard(rookPiece)
	for rooks != 0 {
		fromSquare, newBitboard := rooks.PopLSB()
		rooks = newBitboard

		captures := board.GetRookAttacks(fromSquare, occupancy) & enemyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, rookPiece)
	}

	queens := b.GetPieceBitboard(queenPiece)
	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard

		captures := (board.GetRookAttacks(fromSquare, occupancy) | board.GetBishopAttacks(fromSquare, occupancy)) & enemyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, queenPiece)
	}

	// King captures (castling is never a capture)
	kings := b.GetPieceBitboard(kingPiece)
	if kings != 0 {
		fromSquare := kings.LSB()
		captures := board.GetKingAttacks(fromSquare) & enemyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, kingPiece)
	}

	return moveList
}
