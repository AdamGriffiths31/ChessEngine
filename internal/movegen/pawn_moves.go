package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func generatePawnMoves(b *board.Board, side *sideInfo, moveList *MoveList) {
	pawns := b.GetPieceBitboard(side.pawn)
	if pawns == 0 {
		return
	}

	emptySquares := ^b.AllPieces

	singlePushes := side.pushPawns(pawns) & emptySquares
	startingPawnPushes := side.pushPawns(pawns&board.RankMask(side.startRank)) & emptySquares
	doublePushes := side.pushPawns(startingPawnPushes) & emptySquares

	for singlePushes != 0 {
		toSquare, newBitboard := singlePushes.PopLSB()
		singlePushes = newBitboard

		fromSquare := toSquare - side.push

		if toSquare/8 == side.promotionRank {
			addPromotionMoves(moveList, fromSquare, toSquare, side, board.Empty)
		} else {
			addPawnPush(moveList, fromSquare, toSquare, side.pawn)
		}
	}

	for doublePushes != 0 {
		toSquare, newBitboard := doublePushes.PopLSB()
		doublePushes = newBitboard

		addPawnPush(moveList, toSquare-2*side.push, toSquare, side.pawn)
	}

	generatePawnCaptures(b, side, pawns, b.GetColorBitboard(side.enemy), moveList)
	generateEnPassantCaptures(b, side, pawns, moveList)
}

func addPawnPush(moveList *MoveList, fromSquare, toSquare int, pawnPiece board.Piece) {
	moveList.add(squareOf(fromSquare), squareOf(toSquare), pawnPiece, board.Empty, board.Empty, false)
}

func generatePawnCaptures(b *board.Board, side *sideInfo, pawns, enemyPieces board.Bitboard, moveList *MoveList) {
	capturingPawns := pawns & side.pawnsAttacking(enemyPieces)

	for capturingPawns != 0 {
		fromSquare, newBitboard := capturingPawns.PopLSB()
		capturingPawns = newBitboard

		captures := board.GetPawnAttacks(fromSquare, side.color) & enemyPieces

		for captures != 0 {
			toSquare, newCaptures := captures.PopLSB()
			captures = newCaptures

			capturedPiece := b.GetPieceOnSquare(toSquare)

			if toSquare/8 == side.promotionRank {
				addPromotionMoves(moveList, fromSquare, toSquare, side, capturedPiece)
			} else {
				moveList.add(squareOf(fromSquare), squareOf(toSquare), side.pawn, capturedPiece, board.Empty, true)
			}
		}
	}
}

func generateEnPassantCaptures(b *board.Board, side *sideInfo, pawns board.Bitboard, moveList *MoveList) {
	enPassantTarget, hasEnPassant := b.GetEnPassantTarget()
	if !hasEnPassant {
		return
	}

	targetSquare := enPassantTarget.Index()

	attackers := pawns & board.RankMask(side.enPassantRank) & board.GetPawnAttacks(targetSquare, side.enemy)

	for attackers != 0 {
		fromSquare, newBitboard := attackers.PopLSB()
		attackers = newBitboard

		moveList.AddMove(board.Move{
			From:        squareOf(fromSquare),
			To:          squareOf(targetSquare),
			Piece:       side.pawn,
			Captured:    sides[side.enemy].pawn,
			Promotion:   board.Empty,
			IsCapture:   true,
			IsEnPassant: true,
		})
	}
}

func addPromotionMoves(moveList *MoveList, fromSquare, toSquare int, side *sideInfo, capturedPiece board.Piece) {
	for _, promotion := range &side.promotions {
		moveList.add(squareOf(fromSquare), squareOf(toSquare), side.pawn, capturedPiece, promotion, capturedPiece != board.Empty)
	}
}
