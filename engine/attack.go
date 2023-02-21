package engine

import "fmt"

//SquareAttacked checks if the square is under attack
func SquareAttacked(square int, side int, pos *Board) bool {
	if !SideValid(side) {
		panic(fmt.Errorf("SquareAttacked: side %v is invalid", side))
	}
	if !SqaureOnBoard(square) {
		panic(fmt.Errorf("SquareAttacked: square %v is invalid", side))
	}

	CheckBoard(pos)

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

//isAttackedByKing checks if the square is being attacked by the King
func isAttackedByKing(sq int, side int, pos *Board) bool {
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[sq+KingDirection[i]]
		if piece != OffBoard && piece != Empty {
			if PieceKing[piece] == True && PieceCol[piece] == side {
				return true
			}
		}
	}
	return false
}

//isAttackedByKnight checks if the square is being attacked by a knight
func isAttackedByKnight(sq int, side int, pos *Board) bool {
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[sq+KnightDirection[i]]
		if piece != OffBoard && piece != Empty {
			if PieceKnight[piece] == True && PieceCol[piece] == side {
				return true
			}
		}
	}
	return false
}

//isAttackedByPawn checks if the square is being attacked by a pawn
func isAttackedByPawn(sq int, side int, pos *Board) bool {
	if side == White {
		if pos.Pieces[sq-11] == WP || pos.Pieces[sq-9] == WP {
			return true
		}
	} else {
		if pos.Pieces[sq+11] == BP || pos.Pieces[sq+9] == BP {
			return true
		}
	}

	return false
}

//isAttackedByRookQueen checks if the square is being attacked by a rook or queen
func isAttackedByRookQueen(sq int, side int, pos *Board) bool {
	for i := 0; i < 4; i++ {
		direction := RookDirection[i]
		tempSq := sq + direction
		piece := pos.Pieces[tempSq]
		for piece != OffBoard {
			if piece != Empty {
				if PieceRookQueen[piece] == True && PieceCol[piece] == side {
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

//isAttackedByBishopQueen checks if the square is being attacked by a bishop or queen
func isAttackedByBishopQueen(sq int, side int, pos *Board) bool {
	for i := 0; i < 4; i++ {
		direction := BishopDirection[i]
		tempSq := sq + direction
		piece := pos.Pieces[tempSq]
		for piece != OffBoard {
			if piece != Empty {
				if PieceBishopQueen[piece] == True && PieceCol[piece] == side {
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
