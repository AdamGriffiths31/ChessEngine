package eval

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// Attack-based mobility: a piece scores Base + Unit*(count - Neutral), where
// count is the number of squares it attacks inside its side's mobility area.
const (
	KnightMobilityUnit    = 4
	KnightMobilityNeutral = 4
	KnightMobilityBase    = 24

	BishopMobilityUnit    = 4
	BishopMobilityNeutral = 5
	BishopMobilityBase    = 33

	RookMobilityUnit    = 3
	RookMobilityNeutral = 6
	RookMobilityBase    = 28

	QueenMobilityUnit    = 1
	QueenMobilityNeutral = 9
	QueenMobilityBase    = 32
)

func mobilityArea(b *board.Board, isWhite bool) board.Bitboard {
	var ownPawns, ownKing, enemyPawns board.Bitboard
	if isWhite {
		ownPawns = b.GetPieceBitboard(board.WhitePawn)
		ownKing = b.GetPieceBitboard(board.WhiteKing)
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
	} else {
		ownPawns = b.GetPieceBitboard(board.BlackPawn)
		ownKing = b.GetPieceBitboard(board.BlackKing)
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
	}

	notFileA := ^board.FileMask(0)
	notFileH := ^board.FileMask(7)
	var enemyPawnAttacks board.Bitboard
	if isWhite {
		enemyPawnAttacks = (enemyPawns>>7)&notFileA | (enemyPawns>>9)&notFileH
	} else {
		enemyPawnAttacks = (enemyPawns<<7)&notFileH | (enemyPawns<<9)&notFileA
	}

	return ^(ownPawns | ownKing | enemyPawnAttacks)
}

func mobilityScore(count, base, unit, neutral int) int {
	return base + unit*(count-neutral)
}

func knightMobilityScore(square int, area board.Bitboard) int {
	count := (board.KnightAttacks[square] & area).PopCount()
	return mobilityScore(count, KnightMobilityBase, KnightMobilityUnit, KnightMobilityNeutral)
}

func bishopMobilityScore(square int, occupancy, area board.Bitboard) int {
	count := (board.GetBishopAttacks(square, occupancy) & area).PopCount()
	return mobilityScore(count, BishopMobilityBase, BishopMobilityUnit, BishopMobilityNeutral)
}

func rookMobilityScore(square int, occupancy, area board.Bitboard) int {
	count := (board.GetRookAttacks(square, occupancy) & area).PopCount()
	return mobilityScore(count, RookMobilityBase, RookMobilityUnit, RookMobilityNeutral)
}

func queenMobilityScore(square int, occupancy, area board.Bitboard) int {
	count := (board.GetQueenAttacks(square, occupancy) & area).PopCount()
	return mobilityScore(count, QueenMobilityBase, QueenMobilityUnit, QueenMobilityNeutral)
}
