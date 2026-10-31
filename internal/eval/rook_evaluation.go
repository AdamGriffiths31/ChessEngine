package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// RookOpenFileBonus awards rooks on completely open files
	RookOpenFileBonus = 20

	// RookSemiOpenFileBonus awards rooks on semi-open files (no friendly pawns)
	RookSemiOpenFileBonus = 10

	// RookOn7thBonus awards rooks on opponent's 7th rank (attacking position)
	RookOn7thBonus = 25

	// RooksConnectedBonus awards rooks protecting each other on same rank/file
	RooksConnectedBonus = 8
)

const seventhRank = 6

func evaluateRooks(b *board.Board, whiteArea, blackArea board.Bitboard) int {
	return evaluateRooksForColor(b, b.GetPieceBitboard(board.WhiteRook), whiteArea, true) -
		evaluateRooksForColor(b, b.GetPieceBitboard(board.BlackRook), blackArea, false)
}

func evaluateRooksForColor(b *board.Board, rooks, area board.Bitboard, forWhite bool) int {
	if rooks == 0 {
		return 0
	}

	score := 0

	if rooks.PopCount() == 2 {
		score += evaluateRookConnection(rooks.MSB(), rooks.LSB())
	}

	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)
	ownPawns := blackPawns
	if forWhite {
		ownPawns = whitePawns
	}

	for rooks != 0 {
		square, remaining := rooks.PopLSB()
		rooks = remaining

		file := square % 8
		rank := square / 8

		score += evaluateFileControl(file, ownPawns, whitePawns|blackPawns)

		if relativeRank(rank, forWhite) == seventhRank {
			score += RookOn7thBonus
		}

		score += rookMobilityScore(square, b.AllPieces, area)
	}

	return score
}

func evaluateFileControl(file int, ownPawns, allPawns board.Bitboard) int {
	if isFileOpen(allPawns, file) {
		return RookOpenFileBonus
	}

	if isFileOpen(ownPawns, file) {
		return RookSemiOpenFileBonus
	}

	return 0
}

func evaluateRookConnection(firstSquare, secondSquare int) int {
	if firstSquare/8 == secondSquare/8 || firstSquare%8 == secondSquare%8 {
		return RooksConnectedBonus
	}

	return 0
}
