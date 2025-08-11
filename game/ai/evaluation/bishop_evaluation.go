package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Bishop evaluation - streamlined for performance with focus on key factors
//
// Design Philosophy:
// 1. Bishop pair bonus is the dominant factor (especially in endgames)
// 2. Bad bishop penalty for pawns on same-color squares
// 3. Pre-computed mobility table for O(1) lookups
// 4. Fianchetto position recognition for positional bonuses
// 5. Eliminates expensive calculations (X-ray attacks, complex diagonal analysis)
//
// This approach trades some evaluation precision for significant speed improvements,
// making it ideal for positions evaluated during lazy evaluation with early cutoffs.

// Bishop evaluation constants - focused on high-impact factors
const (
	// Bishop pair bonus - the most valuable bishop feature
	BishopPairBonus = 50 // Having both light and dark squared bishops

	// Mobility scoring unit
	BishopMobilityUnit = 3 // Multiplier for mobility table values

	// Bad bishop penalty - bishops blocked by own pawns
	BadBishopPenalty = -8 // Penalty per own pawn on same color squares

	// Positional bonus for strong bishop placements
	FianchettoBishopBonus = 10 // Bonus for bishops on fianchetto squares
)

// BishopMobilityTable provides pre-computed mobility approximations for each square
// Values reflect typical bishop mobility: center squares offer more diagonal access
// than edges and corners. This avoids expensive move generation during evaluation.
var BishopMobilityTable = [64]int{
	// Rank 1: Limited mobility near board edges
	7, 7, 7, 7, 7, 7, 7, 7,
	// Rank 2: Slightly better mobility
	7, 9, 9, 9, 9, 9, 9, 7,
	// Rank 3: Good mobility toward center
	7, 9, 11, 11, 11, 11, 9, 7,
	// Rank 4-5: Maximum mobility in center
	7, 9, 11, 13, 13, 11, 9, 7,
	7, 9, 11, 13, 13, 11, 9, 7,
	// Rank 6: Good mobility toward center
	7, 9, 11, 11, 11, 11, 9, 7,
	// Rank 7: Slightly better mobility
	7, 9, 9, 9, 9, 9, 9, 7,
	// Rank 8: Limited mobility near board edges
	7, 7, 7, 7, 7, 7, 7, 7,
}

// evaluateBishops performs fast bishop evaluation for both sides
// Returns positive values favoring White, negative favoring Black
func evaluateBishops(b *board.Board) int {
	score := 0

	// Get bishop bitboards
	whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
	blackBishops := b.GetPieceBitboard(board.BlackBishop)

	// Bishop pair bonus (most important factor)
	score += evaluateBishopPairBonus(whiteBishops, blackBishops)

	// Individual bishop evaluation
	score += evaluateBishopsSimple(b, whiteBishops, true)
	score -= evaluateBishopsSimple(b, blackBishops, false)

	return score
}

// evaluateBishopPairBonus calculates the bishop pair advantage
// The bishop pair is one of the most important positional factors in chess,
// especially valuable in open positions and endgames where both bishops
// can coordinate to control key squares.
func evaluateBishopPairBonus(whiteBishops, blackBishops board.Bitboard) int {
	score := 0

	// Award bonus if White has both light and dark squared bishops
	whiteLightBishops := whiteBishops & board.LightSquares
	whiteDarkBishops := whiteBishops & board.DarkSquares
	if whiteLightBishops != 0 && whiteDarkBishops != 0 {
		score += BishopPairBonus
	}

	// Award bonus if Black has both light and dark squared bishops
	blackLightBishops := blackBishops & board.LightSquares
	blackDarkBishops := blackBishops & board.DarkSquares
	if blackLightBishops != 0 && blackDarkBishops != 0 {
		score -= BishopPairBonus
	}

	return score
}

// evaluateBishopsSimple performs streamlined evaluation of bishops for one color
// Focuses on the most impactful factors while avoiding expensive calculations
//
// Parameters:
//   - b: current board position
//   - bishops: bitboard containing all bishops of this color
//   - isWhite: true for white bishops, false for black bishops
//
// Returns: evaluation score for all bishops of the specified color
func evaluateBishopsSimple(b *board.Board, bishops board.Bitboard, isWhite bool) int {
	if bishops == 0 {
		return 0
	}

	score := 0

	// Cache own pawns bitboard for bad bishop evaluation
	var ownPawns board.Bitboard
	if isWhite {
		ownPawns = b.GetPieceBitboard(board.WhitePawn)
	} else {
		ownPawns = b.GetPieceBitboard(board.BlackPawn)
	}

	// Evaluate each bishop individually
	for bishops != 0 {
		square, remaining := bishops.PopLSB()
		bishops = remaining

		// 1. Mobility approximation using pre-computed table
		score += BishopMobilityTable[square] * BishopMobilityUnit

		// 2. Bad bishop penalty for pawn blockages
		score += evaluateBadBishop(square, ownPawns)

		// 3. Fianchetto positional bonus
		score += evaluateFianchetto(square, isWhite)
	}

	return score
}

// evaluateBadBishop calculates penalty for bishops restricted by own pawns
// A "bad bishop" is one where many own pawns occupy the same colored squares,
// limiting the bishop's effectiveness and scope.
func evaluateBadBishop(bishopSquare int, ownPawns board.Bitboard) int {
	// Determine bishop's square color using coordinate parity
	bishopOnLightSquare := ((bishopSquare/8 + bishopSquare%8) % 2) == 0

	// Count own pawns on the same color squares as the bishop
	var pawnsOnSameColor int
	if bishopOnLightSquare {
		pawnsOnSameColor = (ownPawns & board.LightSquares).PopCount()
	} else {
		pawnsOnSameColor = (ownPawns & board.DarkSquares).PopCount()
	}

	// Apply penalty proportional to the number of blocking pawns
	return pawnsOnSameColor * BadBishopPenalty
}

// evaluateFianchetto provides bonus for bishops in fianchetto positions
// Fianchettoed bishops on long diagonals (a1-h8 or h1-a8) are strategically
// strong, controlling key central squares and providing king safety.
func evaluateFianchetto(square int, isWhite bool) int {
	if isWhite {
		// White fianchetto squares: b2 (9) and g2 (14)
		if square == 9 || square == 14 {
			return FianchettoBishopBonus
		}
	} else {
		// Black fianchetto squares: b7 (49) and g7 (54)
		if square == 49 || square == 54 {
			return FianchettoBishopBonus
		}
	}
	return 0
}
