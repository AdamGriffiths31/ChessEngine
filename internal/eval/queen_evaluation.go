package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// QueenOn7thRank bonus for queen on opponent's 7th rank
	QueenOn7thRank = 20

	// QueenOnOpenFile bonus for queen on open file
	QueenOnOpenFile = 10
)

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
	area := mobilityArea(b, isWhite)

	for queensBitboard != 0 {
		square, remaining := queensBitboard.PopLSB()
		queensBitboard = remaining

		file := square % 8
		rank := square / 8

		score += queenMobilityScore(square, b.AllPieces, area)

		if (isWhite && rank == 6) || (!isWhite && rank == 1) {
			score += QueenOn7thRank
		}

		if isFileOpen(b, file) {
			score += QueenOnOpenFile
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
