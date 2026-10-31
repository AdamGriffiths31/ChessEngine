package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// KnightOutpostBonus rewards a knight on a pawn-defended square in enemy
	// territory.
	KnightOutpostBonus = 30
)

var knightOutpostRanks = [8]bool{false, false, false, true, true, true, false, false}

func evaluateKnights(b *board.Board, whiteArea, blackArea board.Bitboard) int {
	return evaluateKnightsSimple(b, b.GetPieceBitboard(board.WhiteKnight), whiteArea, true) -
		evaluateKnightsSimple(b, b.GetPieceBitboard(board.BlackKnight), blackArea, false)
}

func evaluateKnightsSimple(b *board.Board, knights, area board.Bitboard, isWhite bool) int {
	if knights == 0 {
		return 0
	}

	score := 0

	var friendlyPawns, enemyPawns board.Bitboard
	if isWhite {
		friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
	} else {
		friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
	}

	for knights != 0 {
		square, remaining := knights.PopLSB()
		knights = remaining

		rank := square / 8
		file := square % 8

		score += knightMobilityScore(square, area)

		if knightOutpostRanks[relativeRank(rank, isWhite)] &&
			enemyPawns&adjacentFiles(file)&ranksAhead(rank, isWhite) == 0 &&
			isConnectedPawn(friendlyPawns, square, isWhite) {
			score += KnightOutpostBonus
		}
	}

	return score
}
