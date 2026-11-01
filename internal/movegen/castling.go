package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func generateCastlingMoves(b *board.Board, side *sideInfo, kingSquare int, moveList *MoveList) {
	castlingRights := b.GetCastlingSet()
	if castlingRights == 0 || kingSquare != side.kingStart {
		return
	}

	addCastlingMove(b, side, castlingRights, &side.kingside, moveList)
	addCastlingMove(b, side, castlingRights, &side.queenside, moveList)
}

func addCastlingMove(b *board.Board, side *sideInfo, castlingRights board.CastlingRights, wing *castleWing, moveList *MoveList) {
	if castlingRights&wing.right == 0 {
		return
	}

	if b.AllPieces&board.GetBetween(side.kingStart, wing.rookAt) != 0 {
		return
	}

	if b.IsSquareAttackedByColor(side.kingStart, side.enemy) ||
		b.IsSquareAttackedByColor(wing.rookTo, side.enemy) ||
		b.IsSquareAttackedByColor(wing.kingTo, side.enemy) {
		return
	}

	moveList.AddMove(board.Move{
		From:       squareOf(side.kingStart),
		To:         squareOf(wing.kingTo),
		Piece:      side.king,
		Captured:   board.Empty,
		Promotion:  board.Empty,
		IsCastling: true,
	})
}
