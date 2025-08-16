package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// King evaluation - streamlined for performance with focus on essential factors
//
// Design Philosophy:
// 1. King safety is paramount in opening/middlegame (castling, shelter)
// 2. King activity becomes important in endgames (centralization)
// 3. Simple position-based checks avoid expensive attack calculations
// 4. Pre-computed safety zones for efficient evaluation
// 5. Eliminates complex king safety calculations (attack weights, pressure zones)
//
// This approach focuses on the most critical king factors while using simple
// position-based heuristics instead of expensive attack generation.

// King evaluation constants - minimal and essential
const (
	// King safety (simplified)
	KingShelterBonus = 15  // Per pawn in front of castled king
	OpenFileNearKing = -20 // Open file next to king

	// Endgame king activity
	KingCentralizationEndgame = 20 // King in center during endgame

	// Castling status
	HasCastledBonus    = 15  // King has castled
	LostCastlingRights = -10 // Lost ability to castle
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

// evaluateKings performs fast king evaluation for both sides
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

// evaluateKingSimple performs streamlined king evaluation for one color
func evaluateKingSimple(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0

	whitePieces := b.GetColorBitboard(board.BitboardWhite)
	blackPieces := b.GetColorBitboard(board.BitboardBlack)
	totalPieces := (whitePieces | blackPieces).PopCount()
	isEndgame := totalPieces < 14

	if isEndgame {
		score += evaluateKingEndgameActivity(kingSquare)
	} else {
		score += evaluateKingSafety(b, kingSquare, isWhite)
	}

	return score
}

// evaluateKingEndgameActivity provides bonus for centralized kings in endgame
func evaluateKingEndgameActivity(kingSquare int) int {
	file := kingSquare % 8
	rank := kingSquare / 8

	fileDistance := absFloat(float64(file) - 3.5)
	rankDistance := absFloat(float64(rank) - 3.5)
	centerDistance := fileDistance + rankDistance

	centralizationScore := int((7.0 - centerDistance) * 3)
	return centralizationScore
}

// evaluateKingSafety provides safety evaluation for opening/middlegame
func evaluateKingSafety(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0
	rank := kingSquare / 8

	if isWhite {
		if kingSquare == 6 || kingSquare == 2 {
			score += HasCastledBonus
			score += evaluatePawnShelter(b, kingSquare, isWhite)
		} else if rank == 0 {
			score += LostCastlingRights
		}
	} else {
		if kingSquare == 62 || kingSquare == 58 {
			score += HasCastledBonus
			score += evaluatePawnShelter(b, kingSquare, isWhite)
		} else if rank == 7 {
			score += LostCastlingRights
		}
	}

	score += evaluateOpenFilesNearKing(b, kingSquare)
	score += evaluateBasicThreats(b, kingSquare, isWhite)

	return score
}

// evaluatePawnShelter provides simple pawn shelter evaluation for castled kings
func evaluatePawnShelter(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0
	var pawns board.Bitboard

	if isWhite {
		pawns = b.GetPieceBitboard(board.WhitePawn)
	} else {
		pawns = b.GetPieceBitboard(board.BlackPawn)
	}

	if isWhite {
		if kingSquare == 6 {
			if pawns.HasBit(15) {
				score += KingShelterBonus
			}
			if pawns.HasBit(14) {
				score += KingShelterBonus
			}
			if pawns.HasBit(13) {
				score += KingShelterBonus / 2
			}
		} else if kingSquare == 2 {
			if pawns.HasBit(9) {
				score += KingShelterBonus
			}
			if pawns.HasBit(10) {
				score += KingShelterBonus
			}
			if pawns.HasBit(11) {
				score += KingShelterBonus / 2
			}
		}
	} else {
		if kingSquare == 62 {
			if pawns.HasBit(55) {
				score += KingShelterBonus
			}
			if pawns.HasBit(54) {
				score += KingShelterBonus
			}
			if pawns.HasBit(53) {
				score += KingShelterBonus / 2
			}
		} else if kingSquare == 58 {
			if pawns.HasBit(49) {
				score += KingShelterBonus
			}
			if pawns.HasBit(50) {
				score += KingShelterBonus
			}
			if pawns.HasBit(51) {
				score += KingShelterBonus / 2
			}
		}
	}

	return score
}

// evaluateOpenFilesNearKing penalizes open files near the king
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

// evaluateBasicThreats provides simple threat detection near the king
// Counts enemy pieces in king zone and applies exponential penalty for multiple threats
func evaluateBasicThreats(b *board.Board, kingSquare int, isWhite bool) int {
	if b == nil {
		return 0
	}
	score := 0
	zone := KingSafetyZone[kingSquare] // Use pre-computed safety zone

	// Get enemy pieces
	var enemyPieces board.Bitboard
	if isWhite {
		enemyPieces = b.GetColorBitboard(board.BitboardBlack)
	} else {
		enemyPieces = b.GetColorBitboard(board.BitboardWhite)
	}

	// Count enemies near king (fast bitboard operation)
	threatsNearKing := (enemyPieces & zone).PopCount()

	// Simple penalty: more threats = exponentially worse
	if threatsNearKing >= 2 {
		score -= threatsNearKing * threatsNearKing * 25
	}

	return score
}

// abs for float calculations
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
