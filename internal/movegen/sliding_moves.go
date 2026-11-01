package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func generateBishopMoves(b *board.Board, side *sideInfo, targets board.Bitboard, moveList *MoveList) {
	bishops := b.GetPieceBitboard(side.bishop)
	enemyPieces := b.GetColorBitboard(side.enemy)

	for bishops != 0 {
		fromSquare, newBitboard := bishops.PopLSB()
		bishops = newBitboard

		attacks := board.GetBishopAttacks(fromSquare, b.AllPieces)
		addSlidingPieceMoves(b, moveList, fromSquare, attacks&targets, enemyPieces, side.bishop)
	}
}

func generateRookMoves(b *board.Board, side *sideInfo, targets board.Bitboard, moveList *MoveList) {
	rooks := b.GetPieceBitboard(side.rook)
	enemyPieces := b.GetColorBitboard(side.enemy)

	for rooks != 0 {
		fromSquare, newBitboard := rooks.PopLSB()
		rooks = newBitboard

		attacks := board.GetRookAttacks(fromSquare, b.AllPieces)
		addSlidingPieceMoves(b, moveList, fromSquare, attacks&targets, enemyPieces, side.rook)
	}
}

func generateQueenMoves(b *board.Board, side *sideInfo, targets board.Bitboard, moveList *MoveList) {
	queens := b.GetPieceBitboard(side.queen)
	enemyPieces := b.GetColorBitboard(side.enemy)

	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard

		rookAttacks := board.GetRookAttacks(fromSquare, b.AllPieces)
		addSlidingPieceMoves(b, moveList, fromSquare, rookAttacks&targets, enemyPieces, side.queen)

		bishopAttacks := board.GetBishopAttacks(fromSquare, b.AllPieces)
		addSlidingPieceMoves(b, moveList, fromSquare, bishopAttacks&targets, enemyPieces, side.queen)
	}
}

func addSlidingPieceMoves(b *board.Board, moveList *MoveList, fromSquare int, validMoves, enemyPieces board.Bitboard, piece board.Piece) {
	from := squareOf(fromSquare)

	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		isCapture := enemyPieces.HasBit(toSquare)
		var capturedPiece board.Piece
		if isCapture {
			capturedPiece = b.GetPieceOnSquare(toSquare)
		}

		moveList.add(from, squareOf(toSquare), piece, capturedPiece, board.Empty, isCapture)
	}
}
