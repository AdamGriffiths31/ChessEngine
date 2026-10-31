// Package values provides piece values and position evaluation tables for chess evaluation.
package values

// Piece type for use in this package (maps to board.Piece numeric values)
type Piece int

// Piece constants matching board.Piece values
const (
	WhitePawn   Piece = 'P'
	WhiteRook   Piece = 'R'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
	BlackPawn   Piece = 'p'
	BlackRook   Piece = 'r'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
)

// GetPieceValue returns the value of a piece in centipawns
func GetPieceValue(piece Piece) int {
	switch piece {
	case WhitePawn:
		return 100
	case WhiteKnight:
		return 320
	case WhiteBishop:
		return 330
	case WhiteRook:
		return 500
	case WhiteQueen:
		return 900
	case BlackPawn:
		return -100
	case BlackKnight:
		return -320
	case BlackBishop:
		return -330
	case BlackRook:
		return -500
	case BlackQueen:
		return -900
	default:
		return 0
	}
}

// GetPositionalBonus returns the middlegame PST bonus for a piece at a given position
func GetPositionalBonus(piece Piece, rank, file int) int {
	return positionalBonus(piece, rank, file, false)
}

// GetPositionalBonusEndgame returns the endgame PST bonus for a piece at a given position.
// For pieces whose tables do not differ by phase this matches the middlegame value.
func GetPositionalBonusEndgame(piece Piece, rank, file int) int {
	return positionalBonus(piece, rank, file, true)
}

func positionalBonus(piece Piece, rank, file int, endgame bool) int {
	square := rank*8 + file
	flipped := square ^ 56

	switch piece {
	case WhiteKnight:
		return KnightTable[square]
	case BlackKnight:
		return -KnightTable[flipped]
	case WhiteBishop:
		return BishopTable[square]
	case BlackBishop:
		return -BishopTable[flipped]
	case WhiteRook:
		return RookTable[square]
	case BlackRook:
		return -RookTable[flipped]
	case WhiteQueen:
		return QueenTable[square]
	case BlackQueen:
		return -QueenTable[flipped]
	case WhitePawn:
		if endgame {
			return PawnTableEndgame[square]
		}
		return PawnTable[square]
	case BlackPawn:
		if endgame {
			return -PawnTableEndgame[flipped]
		}
		return -PawnTable[flipped]
	case WhiteKing:
		if endgame {
			return KingTableEndgame[square]
		}
		return KingTable[square]
	case BlackKing:
		if endgame {
			return -KingTableEndgame[flipped]
		}
		return -KingTable[flipped]
	default:
		return 0
	}
}

// KnightTable contains positional bonuses/penalties for knights
// Oriented a1-first: index 0 = a1, index 63 = h8 (White's perspective)
var KnightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

// BishopTable contains positional bonuses/penalties for bishops
var BishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

// RookTable contains positional bonuses/penalties for rooks
var RookTable = [64]int{
	0, 0, 0, 5, 5, 0, 0, 0,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// QueenTable contains positional bonuses/penalties for queens
var QueenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-10, 5, 5, 5, 5, 5, 0, -10,
	0, 0, 5, 5, 5, 5, 0, -5,
	-5, 0, 5, 5, 5, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

// KingTable contains positional bonuses/penalties for kings (middle game)
var KingTable = [64]int{
	20, 30, 10, 0, 0, 10, 30, 20,
	20, 20, 0, 0, 0, 0, 20, 20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
}

// PawnTable contains positional bonuses/penalties for pawns
// Oriented a1-first: index 0 = a1, index 63 = h8 (White's perspective)
var PawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, -20, -20, 10, 10, 5,
	5, -5, -10, 0, 0, -10, -5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// PawnTableEndgame contains endgame pawn bonuses.
// Oriented a1-first like PawnTable.
var PawnTableEndgame = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	10, 10, 10, 10, 10, 10, 10, 10,
	10, 10, 10, 10, 10, 10, 10, 10,
	20, 20, 20, 20, 20, 20, 20, 20,
	30, 30, 30, 30, 30, 30, 30, 30,
	50, 50, 50, 50, 50, 50, 50, 50,
	80, 80, 80, 80, 80, 80, 80, 80,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// KingTableEndgame contains endgame king bonuses.
// Oriented a1-first like KingTable.
var KingTableEndgame = [64]int{
	-50, -40, -30, -20, -20, -30, -40, -50,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-20, -10, 20, 30, 30, 20, -10, -20,
	-10, 0, 30, 40, 40, 30, 0, -10,
	-10, 0, 30, 40, 40, 30, 0, -10,
	-20, -10, 20, 30, 30, 20, -10, -20,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-50, -40, -30, -20, -20, -30, -40, -50,
}
