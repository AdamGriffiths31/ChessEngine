package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Knight evaluation - streamlined for performance with focus on key factors
//
// Design Philosophy:
// 1. Outpost bonus is the dominant knight factor (defended squares in enemy territory)
// 2. Pre-computed mobility table for O(1) lookups
// 3. Eliminates expensive calculations (fork detection, complex mobility analysis, knight pair penalties)
//
// This approach trades some tactical awareness for significant speed improvements,
// making it ideal for positions evaluated during lazy evaluation with early cutoffs.

// Knight evaluation constants - only the essentials
const (
	// Outpost bonus (most important)
	KnightOutpostBonus = 30 // Knight on defended square in enemy territory

	// Mobility penalty (knights hate being on edges)
	KnightMobilityUnit = 4 // Per available square
)

// KnightMobilityTable provides pre-computed mobility approximations for each square
// Knights in center have 8 moves, corner knights have only 2. This avoids expensive
// move generation during evaluation.
var KnightMobilityTable = [64]int{
	2, 3, 4, 4, 4, 4, 3, 2,
	3, 4, 6, 6, 6, 6, 4, 3,
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	3, 4, 6, 6, 6, 6, 4, 3,
	2, 3, 4, 4, 4, 4, 3, 2,
}

// WhiteKnightOutpostRanks defines outpost ranks for white knights (ranks 4-6)
var WhiteKnightOutpostRanks = [8]bool{false, false, false, true, true, true, false, false}

// BlackKnightOutpostRanks defines outpost ranks for black knights (ranks 3-5)
var BlackKnightOutpostRanks = [8]bool{false, false, true, true, true, false, false, false}

// evaluateKnights performs fast knight evaluation for both sides
// Returns positive values favoring White, negative favoring Black
func evaluateKnights(b *board.Board) int {
	if b == nil {
		return 0
	}
	score := 0

	whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
	blackKnights := b.GetPieceBitboard(board.BlackKnight)

	score += evaluateKnightsSimple(b, whiteKnights, true)
	score -= evaluateKnightsSimple(b, blackKnights, false)

	return score
}

// evaluateKnightsSimple performs streamlined evaluation of knights for one color
func evaluateKnightsSimple(b *board.Board, knights board.Bitboard, isWhite bool) int {
	if b == nil || knights == 0 {
		return 0
	}

	score := 0

	var enemyPawns board.Bitboard
	var outpostRanks [8]bool
	if isWhite {
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		outpostRanks = WhiteKnightOutpostRanks
	} else {
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		outpostRanks = BlackKnightOutpostRanks
	}

	for knights != 0 {
		square, remaining := knights.PopLSB()
		knights = remaining

		rank := square / 8
		file := square % 8

		score += KnightMobilityTable[square] * KnightMobilityUnit

		if outpostRanks[rank] {
			isOutpost := true

			if file > 0 {
				leftFileMask := board.FileMask(file - 1)
				if (enemyPawns & leftFileMask) != 0 {
					isOutpost = false
				}
			}

			if file < 7 && isOutpost {
				rightFileMask := board.FileMask(file + 1)
				if (enemyPawns & rightFileMask) != 0 {
					isOutpost = false
				}
			}

			if isOutpost {
				score += KnightOutpostBonus
			}
		}
	}

	return score
}
