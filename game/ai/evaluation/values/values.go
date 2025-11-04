// Package values provides piece values and position evaluation tables for chess evaluation.
package values

// Piece type for use in this package (maps to board.Piece numeric values)
type Piece int

// Piece constants matching board.Piece values
const (
	Empty       Piece = 0
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
	case WhiteKing:
		return 0
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
	case BlackKing:
		return 0
	default:
		return 0
	}
}

// GetPositionalBonus returns the PST bonus for a piece at a given position
func GetPositionalBonus(piece Piece, rank, file int) int {
	switch piece {
	case WhiteKnight:
		return KnightTable[rank*8+file]
	case BlackKnight:
		flippedRank := 7 - rank
		return -KnightTable[flippedRank*8+file]
	case WhiteBishop:
		return BishopTable[rank*8+file]
	case BlackBishop:
		flippedRank := 7 - rank
		return -BishopTable[flippedRank*8+file]
	case WhiteRook:
		return RookTable[rank*8+file]
	case BlackRook:
		flippedRank := 7 - rank
		return -RookTable[flippedRank*8+file]
	case WhitePawn:
		return PawnTable[rank*8+file]
	case BlackPawn:
		flippedRank := 7 - rank
		return -PawnTable[flippedRank*8+file]
	case WhiteQueen:
		return QueenTable[rank*8+file]
	case BlackQueen:
		flippedRank := 7 - rank
		return -QueenTable[flippedRank*8+file]
	case WhiteKing:
		return KingTable[rank*8+file]
	case BlackKing:
		flippedRank := 7 - rank
		return -KingTable[flippedRank*8+file]
	default:
		return 0
	}
}

// KnightTable contains positional bonuses/penalties for knights
var KnightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
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
var PawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, -20, -20, 10, 10, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}
