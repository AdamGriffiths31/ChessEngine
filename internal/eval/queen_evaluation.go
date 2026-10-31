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

func evaluateQueens(b *board.Board, whiteArea, blackArea board.Bitboard) int {
	return evaluateQueensForColor(b, b.GetPieceBitboard(board.WhiteQueen), whiteArea, true) -
		evaluateQueensForColor(b, b.GetPieceBitboard(board.BlackQueen), blackArea, false)
}

func evaluateQueensForColor(b *board.Board, queensBitboard, area board.Bitboard, isWhite bool) int {
	if queensBitboard == 0 {
		return 0
	}

	score := 0
	allPawns := b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.BlackPawn)

	for queensBitboard != 0 {
		square, remaining := queensBitboard.PopLSB()
		queensBitboard = remaining

		file := square % 8
		rank := square / 8

		score += queenMobilityScore(square, b.AllPieces, area)

		if relativeRank(rank, isWhite) == seventhRank {
			score += QueenOn7thRank
		}

		if isFileOpen(allPawns, file) {
			score += QueenOnOpenFile
		}
	}

	return score
}

func isFileOpen(pawns board.Bitboard, file int) bool {
	return pawns&board.FileMask(file) == 0
}
