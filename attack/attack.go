package attack

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

// SquareAttacked checks if the square is under attack
func SquareAttacked(square int, side int, pos *data.Board) bool {
	if !validate.SideValid(side) {
		panic(fmt.Errorf("SquareAttacked: side %v is invalid", side))
	}
	if !validate.SquareOnBoard(square) {
		panic(fmt.Errorf("SquareAttacked: square %v is invalid", side))
	}

	board.CheckBoard(pos)

	if isAttackedByPawn(square, side, pos) {
		return true
	}

	if isAttackedByKnight(square, side, pos) {
		return true
	}

	if isAttackedByRookQueen(square, side, pos) {
		return true
	}

	if isAttackedByBishopQueen(square, side, pos) {
		return true
	}

	if isAttackedByKing(square, side, pos) {
		return true
	}

	return false
}

// isAttackedByKing checks if the square is being attacked by the King
func isAttackedByKing(sq int, side int, pos *data.Board) bool {
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[sq+data.KingDirection[i]]
		if piece != data.OffBoard && piece != data.Empty {
			if data.PieceKing[piece] == data.True && data.PieceCol[piece] == side {
				return true
			}
		}
	}
	return false
}

// isAttackedByKnight checks if the square is being attacked by a knight
func isAttackedByKnight(sq int, side int, pos *data.Board) bool {
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[sq+data.KnightDirection[i]]
		if piece != data.OffBoard && piece != data.Empty {
			if data.PieceKnight[piece] == data.True && data.PieceCol[piece] == side {
				return true
			}
		}
	}
	return false
}

// isAttackedByPawn checks if the square is being attacked by a pawn
func isAttackedByPawn(sq int, side int, pos *data.Board) bool {
	if side == data.White {
		if pos.Pieces[sq-11] == data.WP || pos.Pieces[sq-9] == data.WP {
			return true
		}
	} else {
		if pos.Pieces[sq+11] == data.BP || pos.Pieces[sq+9] == data.BP {
			return true
		}
	}

	return false
}

// isAttackedByRookQueen checks if the square is being attacked by a rook or queen
func isAttackedByRookQueen(sq int, side int, pos *data.Board) bool {
	for i := 0; i < 4; i++ {
		direction := data.RookDirection[i]
		tempSq := sq + direction
		piece := pos.Pieces[tempSq]
		for piece != data.OffBoard {
			if piece != data.Empty {
				if data.PieceRookQueen[piece] == data.True && data.PieceCol[piece] == side {
					return true
				}
				break
			}
			tempSq += direction
			piece = pos.Pieces[tempSq]
		}
	}
	return false
}

// isAttackedByBishopQueen checks if the square is being attacked by a bishop or queen
func isAttackedByBishopQueen(sq int, side int, pos *data.Board) bool {
	for i := 0; i < 4; i++ {
		direction := data.BishopDirection[i]
		tempSq := sq + direction
		piece := pos.Pieces[tempSq]
		for piece != data.OffBoard {
			if piece != data.Empty {
				if data.PieceBishopQueen[piece] == data.True && data.PieceCol[piece] == side {
					return true
				}
				break
			}
			tempSq += direction
			piece = pos.Pieces[tempSq]
		}
	}
	return false
}
