package movegen

import (
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func generateCastlingMoves(b *board.Board, side *sideInfo, kingSquare int, moveList *MoveList) {
	castlingRights := b.GetCastlingRights()
	if castlingRights == "-" || kingSquare != side.kingStart {
		return
	}

	addCastlingMove(b, side, castlingRights, &side.kingside, moveList)
	addCastlingMove(b, side, castlingRights, &side.queenside, moveList)
}

func addCastlingMove(b *board.Board, side *sideInfo, castlingRights string, wing *castleWing, moveList *MoveList) {
	if !strings.ContainsRune(castlingRights, wing.right) {
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
