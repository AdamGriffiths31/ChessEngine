// Package evaluation provides chess position evaluation functions.
package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

const (
	// QueenMobilityUnit multiplier for pre-calculated mobility values
	QueenMobilityUnit = 2

	// QueenOn7thRank bonus for queen on opponent's 7th rank
	QueenOn7thRank = 20

	// QueenOnOpenFile bonus for queen on open file
	QueenOnOpenFile = 10

	// EarlyQueenMovePenalty penalty for premature queen development
	EarlyQueenMovePenalty = -25
)

// QueenMobilityTable contains pre-calculated mobility scores per square.
// This eliminates expensive attack generation during evaluation.
var QueenMobilityTable = [64]int{
	12, 14, 14, 14, 14, 14, 14, 12,
	14, 16, 16, 16, 16, 16, 16, 14,
	14, 16, 18, 18, 18, 18, 16, 14,
	14, 16, 18, 20, 20, 18, 16, 14,
	14, 16, 18, 20, 20, 18, 16, 14,
	14, 16, 18, 18, 18, 18, 16, 14,
	14, 16, 16, 16, 16, 16, 16, 14,
	12, 14, 14, 14, 14, 14, 14, 12,
}

func evaluateQueens(b *board.Board) int {
	if b == nil {
		return 0
	}

	score := 0

	whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
	if whiteQueens != 0 {
		score += evaluateQueensForColor(b, whiteQueens, true)
	}

	blackQueens := b.GetPieceBitboard(board.BlackQueen)
	if blackQueens != 0 {
		score -= evaluateQueensForColor(b, blackQueens, false)
	}

	return score
}

func evaluateQueensForColor(b *board.Board, queensBitboard board.Bitboard, isWhite bool) int {
	if b == nil || queensBitboard == 0 {
		return 0
	}

	score := 0

	for queensBitboard != 0 {
		square, remaining := queensBitboard.PopLSB()
		queensBitboard = remaining

		file := square % 8
		rank := square / 8

		score += QueenMobilityTable[square] * QueenMobilityUnit

		if (isWhite && rank == 6) || (!isWhite && rank == 1) {
			score += QueenOn7thRank
		}

		if isFileOpen(b, file) {
			score += QueenOnOpenFile
		}

		if b.GetFullMoveNumber() <= 5 {
			score += evaluateEarlyDevelopment(b, rank, isWhite)
		}
	}

	return score
}

func isFileOpen(b *board.Board, file int) bool {
	if b == nil {
		return false
	}
	fileMask := board.FileMask(file)
	whitePawns := b.GetPieceBitboard(board.WhitePawn) & fileMask
	blackPawns := b.GetPieceBitboard(board.BlackPawn) & fileMask
	return whitePawns == 0 && blackPawns == 0
}

func evaluateEarlyDevelopment(b *board.Board, queenRank int, isWhite bool) int {
	if b == nil {
		return 0
	}

	startingRank := 0
	if !isWhite {
		startingRank = 7
	}

	if queenRank == startingRank {
		return 0
	}

	developedCount := 0

	if isWhite {
		knights := b.GetPieceBitboard(board.WhiteKnight)
		bishops := b.GetPieceBitboard(board.WhiteBishop)

		if (knights & board.Bitboard(1<<1)) == 0 {
			developedCount++
		}
		if (knights & board.Bitboard(1<<6)) == 0 {
			developedCount++
		}
		if (bishops & board.Bitboard(1<<2)) == 0 {
			developedCount++
		}
		if (bishops & board.Bitboard(1<<5)) == 0 {
			developedCount++
		}
	} else {
		knights := b.GetPieceBitboard(board.BlackKnight)
		bishops := b.GetPieceBitboard(board.BlackBishop)

		if (knights & board.Bitboard(1<<57)) == 0 {
			developedCount++
		}
		if (knights & board.Bitboard(1<<62)) == 0 {
			developedCount++
		}
		if (bishops & board.Bitboard(1<<58)) == 0 {
			developedCount++
		}
		if (bishops & board.Bitboard(1<<61)) == 0 {
			developedCount++
		}
	}

	if developedCount < 2 {
		return EarlyQueenMovePenalty
	}

	return 0
}
