package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// King evaluation: open-file and nearby-threat safety. Endgame king
// centralization comes from the phase-blended KingTableEndgame PST (see
// evaluator.go), replacing the former hard <14-pieces switch that scored
// endgame kings with the middlegame safety table. A castling/pawn-shelter
// bonus was tried and removed - SPRT showed it cost ~80 Elo, since keying
// it to the king's exact square created a large score cliff the instant a
// castled king took one more step (see sprt-mobility-eval-fix session notes).

const (
	// King safety (simplified)
	OpenFileNearKing = -20
)

// KingSafetyZone provides pre-computed 3x3 zones around each square
// This avoids expensive zone calculation during evaluation
var KingSafetyZone [64]board.Bitboard

func init() {
	for square := 0; square < 64; square++ {
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

// Returns positive values favoring White, negative favoring Black
func evaluateKings(b *board.Board) int {
	if b == nil {
		return 0
	}
	score := 0

	whiteKing := b.GetPieceBitboard(board.WhiteKing)
	blackKing := b.GetPieceBitboard(board.BlackKing)

	if whiteKing != 0 {
		whiteKingSquare := whiteKing.LSB()
		score += evaluateKingSimple(b, whiteKingSquare, true)
	}

	if blackKing != 0 {
		blackKingSquare := blackKing.LSB()
		score -= evaluateKingSimple(b, blackKingSquare, false)
	}

	return score
}

func evaluateKingSimple(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}

	// Safety terms run in all phases: shelter only fires at castled squares
	// (which endgame kings have left), while open files and threats near the
	// king stay meaningful. Positional centralization is handled by the
	// tapered PST blend.
	return evaluateKingSafety(b, kingSquare, isWhite)
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

// Counts enemy pieces in king zone and applies exponential penalty for multiple threats
func evaluateBasicThreats(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0
	zone := KingSafetyZone[kingSquare]

	// Only attacking pieces count as threats - enemy pawns and the enemy king
	// near our king (e.g. locked pawn chains) are not an attack
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
