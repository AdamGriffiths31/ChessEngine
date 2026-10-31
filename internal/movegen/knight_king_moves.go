package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func generateKnightMoves(b *board.Board, side *sideInfo, targets board.Bitboard, moveList *MoveList) {
	knights := b.GetPieceBitboard(side.knight)

	for knights != 0 {
		fromSquare, newBitboard := knights.PopLSB()
		knights = newBitboard

		addStepMoves(b, moveList, fromSquare, board.GetKnightAttacks(fromSquare)&targets, side.knight)
	}
}

func generateKingMoves(b *board.Board, side *sideInfo, moveList *MoveList) {
	kings := b.GetPieceBitboard(side.king)
	if kings == 0 {
		return
	}

	kingSquare := kings.LSB()

	addStepMoves(b, moveList, kingSquare, board.GetKingAttacks(kingSquare)&^b.GetColorBitboard(side.color), side.king)

	generateCastlingMoves(b, side, kingSquare, moveList)
}

func addStepMoves(b *board.Board, moveList *MoveList, fromSquare int, validMoves board.Bitboard, piece board.Piece) {
	from := squareOf(fromSquare)

	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		capturedPiece := b.GetPieceOnSquare(toSquare)

		moveList.AddMove(board.Move{
			From:      from,
			To:        squareOf(toSquare),
			Piece:     piece,
			Captured:  capturedPiece,
			Promotion: board.Empty,
			IsCapture: capturedPiece != board.Empty,
		})
	}
}
