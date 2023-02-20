package engine

import "fmt"

//TODO Refactor
func SquareAttacked(square int, side int, pos *Board) bool {
	if !SideValid(side) {
		panic(fmt.Errorf("SquareAttacked: side %v is invalid", side))
	}
	if !SqaureOnBoard(square) {
		panic(fmt.Errorf("SquareAttacked: square %v is invalid", side))
	}
	CheckBoard(pos)
	//Pawns
	if side == White {
		if pos.Pieces[square-11] == WP || pos.Pieces[square-9] == WP {
			return true
		}
	} else {
		if pos.Pieces[square+11] == BP || pos.Pieces[square+9] == BP {
			return true
		}
	}
	//Knights
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[square+KnightDirection[i]]
		if piece != OffBoard && piece != Empty {
			if PieceKnight[piece] == True && PieceCol[piece] == side {
				return true
			}
		}
	}
	//Rooks & Queens
	for i := 0; i < 4; i++ {
		direction := RookDirection[i]
		tempSq := square + direction
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
	//Bishops & Queens
	for i := 0; i < 4; i++ {
		direction := BishopDirection[i]
		tempSq := square + direction
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
	//Kings
	for i := 0; i < 8; i++ {
		piece := pos.Pieces[square+KingDirection[i]]
		if piece != OffBoard && piece != Empty {
			if PieceKing[piece] == True && PieceCol[piece] == side {
				return true
			}
		}
	}

	return false
}
