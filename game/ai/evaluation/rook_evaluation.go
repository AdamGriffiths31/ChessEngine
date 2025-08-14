package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Rook evaluation - optimized for speed with focus on what matters most
//
// Design Philosophy:
// 1. Open files are 90% of rook evaluation (prioritize this)
// 2. 7th rank penetration is the second most important factor
// 3. Everything else (mobility, connections) provides marginal value
// 4. Use pre-computed tables and simple checks to avoid expensive calculations
//
// This implementation trades some evaluation accuracy for significant speed gains,
// making it suitable for positions evaluated during lazy evaluation cutoffs.

const (
	// RookOpenFileBonus awards rooks on completely open files
	RookOpenFileBonus = 20

	// RookSemiOpenFileBonus awards rooks on semi-open files (no friendly pawns)
	RookSemiOpenFileBonus = 10

	// RookOn7thBonus awards rooks on opponent's 7th rank (attacking position)
	RookOn7thBonus = 25

	// RooksConnectedBonus awards rooks protecting each other on same rank/file
	RooksConnectedBonus = 8

	// RookMobilityUnit multiplier for pre-calculated mobility values
	RookMobilityUnit = 2
)

// RookMobilityByRank contains approximate mobility values for each rank
// Values represent typical number of moves available to a rook on that rank
// Back ranks (0,7) have lower mobility due to own pieces, middle ranks are optimal
var RookMobilityByRank = [8]int{
	12, 14, 14, 14, 14, 14, 14, 12,
}

// evaluateRooks performs fast rook evaluation for both sides
// Returns positive values favoring White, negative favoring Black
func evaluateRooks(b *board.Board) int {
	if b == nil {
		return 0
	}
	whiteScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.WhiteRook), true)
	blackScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.BlackRook), false)

	return whiteScore - blackScore
}

// evaluateRooksForColor evaluates all rooks of a specific color
func evaluateRooksForColor(b *board.Board, rooks board.Bitboard, forWhite bool) int {
	if b == nil || rooks == 0 {
		return 0
	}

	score := 0
	rookCount := rooks.PopCount()

	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)

	firstRookSquare := -1

	for rooks != 0 {
		square, remaining := rooks.PopLSB()
		rooks = remaining

		file := square % 8
		rank := square / 8

		score += evaluateFileControl(file, whitePawns, blackPawns, forWhite)

		score += evaluateRankPenetration(rank, forWhite)

		score += RookMobilityByRank[rank] * RookMobilityUnit

		if firstRookSquare != -1 && rookCount == 2 {
			score += evaluateRookConnection(square, firstRookSquare)
		}

		firstRookSquare = square
	}

	return score
}

// evaluateFileControl determines file control bonus for a rook
func evaluateFileControl(file int, whitePawns, blackPawns board.Bitboard, forWhite bool) int {
	fileMask := board.FileMask(file)
	hasWhitePawns := (whitePawns & fileMask) != 0
	hasBlackPawns := (blackPawns & fileMask) != 0

	if !hasWhitePawns && !hasBlackPawns {
		return RookOpenFileBonus
	}

	if forWhite && !hasWhitePawns {
		return RookSemiOpenFileBonus
	}
	if !forWhite && !hasBlackPawns {
		return RookSemiOpenFileBonus
	}

	return 0
}

// evaluateRankPenetration awards bonus for rooks on opponent's 7th rank
func evaluateRankPenetration(rank int, forWhite bool) int {
	if (forWhite && rank == 6) || (!forWhite && rank == 1) {
		return RookOn7thBonus
	}
	return 0
}

// evaluateRookConnection awards bonus if rooks are connected
func evaluateRookConnection(currentSquare, previousSquare int) int {
	currentFile := currentSquare % 8
	currentRank := currentSquare / 8
	previousFile := previousSquare % 8
	previousRank := previousSquare / 8

	if currentRank == previousRank || currentFile == previousFile {
		return RooksConnectedBonus
	}

	return 0
}
