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
	// Rank 1: Edge squares have limited knight mobility
	2, 3, 4, 4, 4, 4, 3, 2,
	// Rank 2: Better mobility moving toward center
	3, 4, 6, 6, 6, 6, 4, 3,
	// Rank 3-6: Maximum mobility in center ranks
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	4, 6, 8, 8, 8, 8, 6, 4,
	// Rank 7: Better mobility moving toward center
	3, 4, 6, 6, 6, 6, 4, 3,
	// Rank 8: Edge squares have limited knight mobility
	2, 3, 4, 4, 4, 4, 3, 2,
}

// Outpost rank definitions - key squares where knights want to be
var WhiteKnightOutpostRanks = [8]bool{false, false, false, true, true, true, false, false} // Ranks 4-6
var BlackKnightOutpostRanks = [8]bool{false, false, true, true, true, false, false, false} // Ranks 3-5

// evaluateKnights performs fast knight evaluation for both sides
// Returns positive values favoring White, negative favoring Black
func evaluateKnights(b *board.Board) int {
	score := 0

	// Get knight bitboards
	whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
	blackKnights := b.GetPieceBitboard(board.BlackKnight)

	// Individual knight evaluation
	score += evaluateKnightsSimple(b, whiteKnights, true)
	score -= evaluateKnightsSimple(b, blackKnights, false)

	return score
}

// evaluateKnightsSimple performs streamlined evaluation of knights for one color
// Focuses on the most impactful factors while avoiding expensive calculations
//
// Parameters:
//   - b: current board position
//   - knights: bitboard containing all knights of this color
//   - isWhite: true for white knights, false for black knights
//
// Returns: evaluation score for all knights of the specified color
func evaluateKnightsSimple(b *board.Board, knights board.Bitboard, isWhite bool) int {
	if knights == 0 {
		return 0
	}

	score := 0

	// Get enemy pawns for outpost detection
	var enemyPawns board.Bitboard
	var outpostRanks [8]bool
	if isWhite {
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		outpostRanks = WhiteKnightOutpostRanks
	} else {
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		outpostRanks = BlackKnightOutpostRanks
	}

	// Evaluate each knight individually
	for knights != 0 {
		square, remaining := knights.PopLSB()
		knights = remaining

		rank := square / 8
		file := square % 8

		// 1. Mobility approximation using pre-computed table
		score += KnightMobilityTable[square] * KnightMobilityUnit

		// 2. Outpost bonus (simplified detection)
		if outpostRanks[rank] {
			// Check if square can't be attacked by enemy pawns
			// Simple check: no enemy pawns on adjacent files that can advance to attack
			isOutpost := true

			// Check left file
			if file > 0 {
				leftFileMask := board.FileMask(file - 1)
				if (enemyPawns & leftFileMask) != 0 {
					isOutpost = false
				}
			}

			// Check right file
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