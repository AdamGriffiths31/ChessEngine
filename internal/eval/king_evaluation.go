package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// OpenFileNearKing is the penalty per open file next to the king.
	OpenFileNearKing = -20

	kingThreatWeight = 25
)

// KingSafetyZone provides pre-computed 3x3 zones around each square
var KingSafetyZone [64]board.Bitboard

func init() {
	for square := range 64 {
		KingSafetyZone[square] = board.KingAttacks[square].SetBit(square)
	}
}

func evaluateKings(b *board.Board) int {
	score := 0

	whiteKing := b.GetPieceBitboard(board.WhiteKing)
	blackKing := b.GetPieceBitboard(board.BlackKing)

	if whiteKing != 0 {
		whiteKingSquare := whiteKing.LSB()
		score += evaluateKingSafety(b, whiteKingSquare, true)
	}

	if blackKing != 0 {
		blackKingSquare := blackKing.LSB()
		score -= evaluateKingSafety(b, blackKingSquare, false)
	}

	return score
}

func evaluateKingSafety(b *board.Board, kingSquare int, isWhite bool) int {
	return evaluateOpenFilesNearKing(b, kingSquare) + evaluateBasicThreats(b, kingSquare, isWhite)
}

func evaluateOpenFilesNearKing(b *board.Board, kingSquare int) int {
	score := 0
	file := kingSquare % 8

	allPawns := b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.BlackPawn)

	for f := max(0, file-1); f <= min(7, file+1); f++ {
		if isFileOpen(allPawns, f) {
			score += OpenFileNearKing
		}
	}

	return score
}

func evaluateBasicThreats(b *board.Board, kingSquare int, isWhite bool) int {
	var enemyPieces board.Bitboard
	if isWhite {
		enemyPieces = b.BlackPieces &^ (b.GetPieceBitboard(board.BlackPawn) | b.GetPieceBitboard(board.BlackKing))
	} else {
		enemyPieces = b.WhitePieces &^ (b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.WhiteKing))
	}

	threatsNearKing := (enemyPieces & KingSafetyZone[kingSquare]).PopCount()

	if threatsNearKing >= 2 {
		return -threatsNearKing * threatsNearKing * kingThreatWeight
	}

	return 0
}
