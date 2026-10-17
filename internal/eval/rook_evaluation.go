package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// Rook evaluation: open/semi-open file control, seventh-rank penetration,
// attack-based mobility, connected-rooks bonus.

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

// Returns positive values favoring White, negative favoring Black
func evaluateRooks(b *board.Board) int {
	if b == nil {
		return 0
	}
	whiteScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.WhiteRook), true)
	blackScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.BlackRook), false)

	return whiteScore - blackScore
}

func evaluateRooksForColor(b *board.Board, rooks board.Bitboard, forWhite bool) int {
	if b == nil || rooks == 0 {
		return 0
	}

	score := 0
	rookCount := rooks.PopCount()

	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)

	area := mobilityArea(b, forWhite)

	firstRookSquare := -1

	for rooks != 0 {
		square, remaining := rooks.PopLSB()
		rooks = remaining

		file := square % 8
		rank := square / 8

		score += evaluateFileControl(file, whitePawns, blackPawns, forWhite)

		score += evaluateRankPenetration(rank, forWhite)

		score += rookMobilityScore(square, b.AllPieces, area)

		if firstRookSquare != -1 && rookCount == 2 {
			score += evaluateRookConnection(square, firstRookSquare)
		}

		firstRookSquare = square
	}

	return score
}

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

func evaluateRankPenetration(rank int, forWhite bool) int {
	if (forWhite && rank == 6) || (!forWhite && rank == 1) {
		return RookOn7thBonus
	}
	return 0
}

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
