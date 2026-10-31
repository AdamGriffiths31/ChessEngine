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
	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)

	if isWhite {
		enemyPawnAttacks := blackPawns.ShiftSouthEast() | blackPawns.ShiftSouthWest()
		return ^(whitePawns | b.GetPieceBitboard(board.WhiteKing) | enemyPawnAttacks)
	}
	enemyPawnAttacks := whitePawns.ShiftNorthEast() | whitePawns.ShiftNorthWest()
	return ^(blackPawns | b.GetPieceBitboard(board.BlackKing) | enemyPawnAttacks)
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
