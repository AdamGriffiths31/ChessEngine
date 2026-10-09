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
	case WhiteQueen:
		return QueenTable[rank*8+file]
	case BlackQueen:
		flippedRank := 7 - rank
		return -QueenTable[flippedRank*8+file]
	case WhitePawn:
		if endgame {
			return PawnTableEndgame[rank*8+file]
		}
		return PawnTable[rank*8+file]
	case BlackPawn:
		flippedRank := 7 - rank
		if endgame {
			return -PawnTableEndgame[flippedRank*8+file]
		}
		return -PawnTable[flippedRank*8+file]
	case WhiteKing:
		if endgame {
			return KingTableEndgame[rank*8+file]
		}
		return KingTable[rank*8+file]
	case BlackKing:
		flippedRank := 7 - rank
		if endgame {
			return -KingTableEndgame[flippedRank*8+file]
		}
		return -KingTable[flippedRank*8+file]
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

// PawnTableEndgame contains endgame pawn bonuses: advancement dominates over
// shape, so a passed pawn's march is worth far more than its file.
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

// KingTableEndgame contains endgame king bonuses: centralization. In the
// middlegame the king scores from KingTable (safety corners); blending this
// table in by game phase is what replaces the former hard <14-pieces switch
// that scored endgame kings with the middlegame safety table.
// Oriented a1-first like KingTable (symmetric under left-right flip).
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
