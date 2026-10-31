package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// BishopPairBonus rewards owning both the light- and dark-squared bishop.
	BishopPairBonus = 50

	// BadBishopPenalty applies per own pawn on the bishop's square colour.
	BadBishopPenalty = -8

	// FianchettoBishopBonus rewards a bishop on its fianchetto square.
	FianchettoBishopBonus = 10
)

func evaluateBishops(b *board.Board, whiteArea, blackArea board.Bitboard) int {
	whiteBishops := b.GetPieceBitboard(board.WhiteBishop)
	blackBishops := b.GetPieceBitboard(board.BlackBishop)

	return evaluateBishopPairBonus(whiteBishops, blackBishops) +
		evaluateBishopsSimple(b, whiteBishops, whiteArea, true) -
		evaluateBishopsSimple(b, blackBishops, blackArea, false)
}

func evaluateBishopPairBonus(whiteBishops, blackBishops board.Bitboard) int {
	score := 0

	if hasBishopPair(whiteBishops) {
		score += BishopPairBonus
	}

	if hasBishopPair(blackBishops) {
		score -= BishopPairBonus
	}

	return score
}

func hasBishopPair(bishops board.Bitboard) bool {
	return bishops&board.LightSquares != 0 && bishops&board.DarkSquares != 0
}

func evaluateBishopsSimple(b *board.Board, bishops, area board.Bitboard, isWhite bool) int {
	if bishops == 0 {
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

		score += bishopMobilityScore(square, b.AllPieces, area)

		score += evaluateBadBishop(square, ownPawns)

		score += evaluateFianchetto(square, isWhite)
	}

	return score
}

func evaluateBadBishop(bishopSquare int, ownPawns board.Bitboard) int {
	sameColorSquares := board.DarkSquares
	if board.LightSquares.HasBit(bishopSquare) {
		sameColorSquares = board.LightSquares
	}

	return (ownPawns & sameColorSquares).PopCount() * BadBishopPenalty
}

func evaluateFianchetto(square int, isWhite bool) int {
	if isWhite {
		if square == board.B2 || square == board.G2 {
			return FianchettoBishopBonus
		}
	} else {
		if square == board.B7 || square == board.G7 {
			return FianchettoBishopBonus
		}
	}
	return 0
}
