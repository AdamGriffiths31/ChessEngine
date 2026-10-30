package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func (bmg *BitboardMoveGenerator) generateBishopMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var bishopPiece board.Piece
	var bitboardColor, enemyColor board.BitboardColor

	if player == White {
		bishopPiece = board.WhiteBishop
		bitboardColor = board.BitboardWhite
		enemyColor = board.BitboardBlack
	} else {
		bishopPiece = board.BlackBishop
		bitboardColor = board.BitboardBlack
		enemyColor = board.BitboardWhite
	}

	bishops := b.GetPieceBitboard(bishopPiece)
	if bishops == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	enemyPieces := b.GetColorBitboard(enemyColor)
	occupancy := b.AllPieces

	for bishops != 0 {
		fromSquare, newBitboard := bishops.PopLSB()
		bishops = newBitboard

		attacks := board.GetBishopAttacks(fromSquare, occupancy)
		validMoves := attacks &^ friendlyPieces

		bmg.addSlidingPieceMoves(b, moveList, fromSquare, validMoves, enemyPieces, bishopPiece)
	}
}

func (bmg *BitboardMoveGenerator) generateRookMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var rookPiece board.Piece
	var bitboardColor, enemyColor board.BitboardColor

	if player == White {
		rookPiece = board.WhiteRook
		bitboardColor = board.BitboardWhite
		enemyColor = board.BitboardBlack
	} else {
		rookPiece = board.BlackRook
		bitboardColor = board.BitboardBlack
		enemyColor = board.BitboardWhite
	}

	rooks := b.GetPieceBitboard(rookPiece)
	if rooks == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	enemyPieces := b.GetColorBitboard(enemyColor)
	occupancy := b.AllPieces

	for rooks != 0 {
		fromSquare, newBitboard := rooks.PopLSB()
		rooks = newBitboard

		attacks := board.GetRookAttacks(fromSquare, occupancy)
		validMoves := attacks &^ friendlyPieces

		bmg.addSlidingPieceMoves(b, moveList, fromSquare, validMoves, enemyPieces, rookPiece)
	}
}

func (bmg *BitboardMoveGenerator) generateQueenMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var queenPiece board.Piece
	var bitboardColor, enemyColor board.BitboardColor

	if player == White {
		queenPiece = board.WhiteQueen
		bitboardColor = board.BitboardWhite
		enemyColor = board.BitboardBlack
	} else {
		queenPiece = board.BlackQueen
		bitboardColor = board.BitboardBlack
		enemyColor = board.BitboardWhite
	}

	queens := b.GetPieceBitboard(queenPiece)
	if queens == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	enemyPieces := b.GetColorBitboard(enemyColor)
	occupancy := b.AllPieces

	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard

		rookAttacks := board.GetRookAttacks(fromSquare, occupancy)
		rookValidMoves := rookAttacks &^ friendlyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, rookValidMoves, enemyPieces, queenPiece)

		bishopAttacks := board.GetBishopAttacks(fromSquare, occupancy)
		bishopValidMoves := bishopAttacks &^ friendlyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, bishopValidMoves, enemyPieces, queenPiece)
	}
}

func (bmg *BitboardMoveGenerator) addSlidingPieceMoves(b *board.Board, moveList *MoveList, fromSquare int, validMoves, enemyPieces board.Bitboard, piece board.Piece) {
	fromFile, fromRank := board.SquareToFileRank(fromSquare)

	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		isCapture := enemyPieces.HasBit(toSquare)
		var capturedPiece board.Piece
		if isCapture {
			capturedPiece = b.GetPieceOnSquare(toSquare)
		}

		toFile, toRank := board.SquareToFileRank(toSquare)

		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       piece,
			Captured:    capturedPiece,
			Promotion:   board.Empty,
			IsCapture:   isCapture,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}
}
