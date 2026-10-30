package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// OpenFileNearKing is the penalty per open file next to the king.
const OpenFileNearKing = -20

// KingSafetyZone provides pre-computed 3x3 zones around each square
var KingSafetyZone [64]board.Bitboard

func init() {
	for square := range 64 {
		rank := square / 8
		file := square % 8
		zone := board.Bitboard(0)

		for dr := -1; dr <= 1; dr++ {
			for df := -1; df <= 1; df++ {
				newRank := rank + dr
				newFile := file + df

				if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
					zone |= board.Bitboard(1) << (newRank*8 + newFile)
				}
			}
		}

		KingSafetyZone[square] = zone
	}
}

func evaluateKings(b *board.Board) int {
	if b == nil {
		return 0
	}
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
	if b == nil {
		return 0
	}
	score := 0

	score += evaluateOpenFilesNearKing(b, kingSquare)
	score += evaluateBasicThreats(b, kingSquare, isWhite)

	return score
}

func evaluateOpenFilesNearKing(b *board.Board, kingSquare int) int {
	if b == nil {
		return 0
	}
	score := 0
	file := kingSquare % 8

	allPawns := b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.BlackPawn)

	for f := max(0, file-1); f <= min(7, file+1); f++ {
		fileMask := board.FileMask(f)
		if (allPawns & fileMask) == 0 {
			score += OpenFileNearKing
		}
	}

	return score
}

func evaluateBasicThreats(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0
	zone := KingSafetyZone[kingSquare]

	var enemyPieces board.Bitboard
	if isWhite {
		enemyPieces = b.GetPieceBitboard(board.BlackKnight) |
			b.GetPieceBitboard(board.BlackBishop) |
			b.GetPieceBitboard(board.BlackRook) |
			b.GetPieceBitboard(board.BlackQueen)
	} else {
		enemyPieces = b.GetPieceBitboard(board.WhiteKnight) |
			b.GetPieceBitboard(board.WhiteBishop) |
			b.GetPieceBitboard(board.WhiteRook) |
			b.GetPieceBitboard(board.WhiteQueen)
	}

	threatsNearKing := (enemyPieces & zone).PopCount()

	if threatsNearKing >= 2 {
		score -= threatsNearKing * threatsNearKing * 25
	}

	return score
}
