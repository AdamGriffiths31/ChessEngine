package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// Bishop evaluation: bishop-pair bonus, bad-bishop penalty for own pawns
// on the bishop's square color, static mobility and fianchetto tables.

const (
	// Bishop pair bonus - the most valuable bishop feature
	BishopPairBonus = 50 // Having both light and dark squared bishops

	// Mobility scoring unit
	BishopMobilityUnit = 3 // Multiplier for mobility table values

	// Bad bishop penalty - bishops blocked by own pawns
	BadBishopPenalty = -8 // Penalty per own pawn on same color squares

	// Positional bonus for strong bishop placements
	FianchettoBishopBonus = 10
)

// BishopMobilityTable provides pre-computed mobility approximations for each square
// Values reflect typical bishop mobility: center squares offer more diagonal access
// than edges and corners. This avoids expensive move generation during evaluation.
var BishopMobilityTable = [64]int{
	7, 7, 7, 7, 7, 7, 7, 7,
	7, 9, 9, 9, 9, 9, 9, 7,
	7, 9, 11, 11, 11, 11, 9, 7,
	7, 9, 11, 13, 13, 11, 9, 7,
	7, 9, 11, 13, 13, 11, 9, 7,
	7, 9, 11, 11, 11, 11, 9, 7,
	7, 9, 9, 9, 9, 9, 9, 7,
	7, 7, 7, 7, 7, 7, 7, 7,
}

// Returns positive values favoring White, negative favoring Black
func evaluateBishops(b *board.Board) int {
	if b == nil {
		return 0
	}
	score := 0

	whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
	blackBishops := b.GetPieceBitboard(board.BlackBishop)

	score += evaluateBishopPairBonus(whiteBishops, blackBishops)

	score += evaluateBishopsSimple(b, whiteBishops, true)
	score -= evaluateBishopsSimple(b, blackBishops, false)

	return score
}

func evaluateBishopPairBonus(whiteBishops, blackBishops board.Bitboard) int {
	score := 0

	whiteLightBishops := whiteBishops & board.LightSquares
	whiteDarkBishops := whiteBishops & board.DarkSquares
	if whiteLightBishops != 0 && whiteDarkBishops != 0 {
		score += BishopPairBonus
	}

	blackLightBishops := blackBishops & board.LightSquares
	blackDarkBishops := blackBishops & board.DarkSquares
	if blackLightBishops != 0 && blackDarkBishops != 0 {
		score -= BishopPairBonus
	}

	return score
}

func evaluateBishopsSimple(b *board.Board, bishops board.Bitboard, isWhite bool) int {
	if b == nil || bishops == 0 {
		return 0
	}

	score := 0

	var ownPawns board.Bitboard
	if isWhite {
		ownPawns = b.GetPieceBitboard(board.WhitePawn)
	} else {
		ownPawns = b.GetPieceBitboard(board.BlackPawn)
	}

	for bishops != 0 {
		square, remaining := bishops.PopLSB()
		bishops = remaining

		score += BishopMobilityTable[square] * BishopMobilityUnit

		score += evaluateBadBishop(square, ownPawns)

		score += evaluateFianchetto(square, isWhite)
	}

	return score
}

// evaluateBadBishop calculates penalty for bishops restricted by own pawns
func evaluateBadBishop(bishopSquare int, ownPawns board.Bitboard) int {
	bishopOnLightSquare := board.LightSquares.HasBit(bishopSquare)

	var pawnsOnSameColor int
	if bishopOnLightSquare {
		pawnsOnSameColor = (ownPawns & board.LightSquares).PopCount()
	} else {
		pawnsOnSameColor = (ownPawns & board.DarkSquares).PopCount()
	}

	return pawnsOnSameColor * BadBishopPenalty
}

func evaluateFianchetto(square int, isWhite bool) int {
	if isWhite {
		if square == 9 || square == 14 {
			return FianchettoBishopBonus
		}
	} else {
		if square == 49 || square == 54 {
			return FianchettoBishopBonus
		}
	}
	return 0
}
