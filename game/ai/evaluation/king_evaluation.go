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
	KingShelterBonus = 10 // Per pawn in front of castled king
	OpenFileNearKing = -20 // Open file next to king

	// Endgame king activity
	KingCentralizationEndgame = 20 // King in center during endgame

	// Castling status
	HasCastledBonus      = 15  // King has castled
	LostCastlingRights   = -10 // Lost ability to castle
)

// KingSafetyZone provides pre-computed 3x3 zones around each square
// This avoids expensive zone calculation during evaluation
var KingSafetyZone [64]board.Bitboard

func init() {
	// Pre-compute king safety zones for each square
	for square := 0; square < 64; square++ {
		rank := square / 8
		file := square % 8
		zone := board.Bitboard(0)

		// Create 3x3 zone around king position
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
	score := 0

	// Get king positions
	whiteKing := b.GetPieceBitboard(board.WhiteKing)
	blackKing := b.GetPieceBitboard(board.BlackKing)

	// Evaluate white king
	if whiteKing != 0 {
		whiteKingSquare := whiteKing.LSB()
		score += evaluateKingSimple(b, whiteKingSquare, true)
	}

	// Evaluate black king
	if blackKing != 0 {
		blackKingSquare := blackKing.LSB()
		score -= evaluateKingSimple(b, blackKingSquare, false)
	}

	return score
}

// evaluateKingSimple performs streamlined king evaluation for one color
// Adapts evaluation based on game phase (opening/middlegame vs endgame)
//
// Parameters:
//   - b: current board position
//   - kingSquare: square index where the king is located
//   - isWhite: true for white king, false for black king
//
// Returns: evaluation score for the king
func evaluateKingSimple(b *board.Board, kingSquare int, isWhite bool) int {
	score := 0

	// Determine game phase using simple piece count
	whitePieces := b.GetColorBitboard(board.BitboardWhite)
	blackPieces := b.GetColorBitboard(board.BitboardBlack)
	totalPieces := (whitePieces | blackPieces).PopCount()
	isEndgame := totalPieces < 14 // Rough endgame threshold

	if isEndgame {
		// ENDGAME: King should be active and centralized
		score += evaluateKingEndgameActivity(kingSquare)
	} else {
		// OPENING/MIDDLEGAME: King should be safe
		score += evaluateKingSafety(b, kingSquare, isWhite)
	}

	return score
}

// evaluateKingEndgameActivity provides bonus for centralized kings in endgame
func evaluateKingEndgameActivity(kingSquare int) int {
	file := kingSquare % 8
	rank := kingSquare / 8

	// Simple centralization bonus - closer to center is better
	fileDistance := absFloat(float64(file) - 3.5)
	rankDistance := absFloat(float64(rank) - 3.5)
	centerDistance := fileDistance + rankDistance

	// Convert to bonus: closer to center = higher score
	centralizationScore := int((7.0 - centerDistance) * 3)
	return centralizationScore
}

// evaluateKingSafety provides safety evaluation for opening/middlegame
func evaluateKingSafety(b *board.Board, kingSquare int, isWhite bool) int {
	score := 0
	rank := kingSquare / 8

	// 1. Check if king has castled (simple position check)
	if isWhite {
		// White king on g1 (6) or c1 (2) = probably castled
		if kingSquare == 6 || kingSquare == 2 {
			score += HasCastledBonus
			score += evaluatePawnShelter(b, kingSquare, isWhite)
		} else if rank == 0 {
			// Still on back rank but not castled = lost castling opportunity
			score += LostCastlingRights
		}
	} else {
		// Black king on g8 (62) or c8 (58) = probably castled
		if kingSquare == 62 || kingSquare == 58 {
			score += HasCastledBonus
			score += evaluatePawnShelter(b, kingSquare, isWhite)
		} else if rank == 7 {
			// Still on back rank but not castled = lost castling opportunity
			score += LostCastlingRights
		}
	}

	// 2. Check for open files near king (danger indicator)
	score += evaluateOpenFilesNearKing(b, kingSquare)

	return score
}

// evaluatePawnShelter provides simple pawn shelter evaluation for castled kings
func evaluatePawnShelter(b *board.Board, kingSquare int, isWhite bool) int {
	score := 0
	var pawns board.Bitboard

	if isWhite {
		pawns = b.GetPieceBitboard(board.WhitePawn)
	} else {
		pawns = b.GetPieceBitboard(board.BlackPawn)
	}

	// Check specific pawn positions based on castling side
	if isWhite {
		if kingSquare == 6 { // g1 - kingside castling
			if pawns.HasBit(15) { // h2
				score += KingShelterBonus
			}
			if pawns.HasBit(14) { // g2
				score += KingShelterBonus
			}
			if pawns.HasBit(13) { // f2
				score += KingShelterBonus / 2 // Half bonus for f2
			}
		} else if kingSquare == 2 { // c1 - queenside castling
			if pawns.HasBit(9) { // b2
				score += KingShelterBonus
			}
			if pawns.HasBit(10) { // c2
				score += KingShelterBonus
			}
			if pawns.HasBit(11) { // d2
				score += KingShelterBonus / 2 // Half bonus for d2
			}
		}
	} else {
		if kingSquare == 62 { // g8 - kingside castling
			if pawns.HasBit(55) { // h7
				score += KingShelterBonus
			}
			if pawns.HasBit(54) { // g7
				score += KingShelterBonus
			}
			if pawns.HasBit(53) { // f7
				score += KingShelterBonus / 2 // Half bonus for f7
			}
		} else if kingSquare == 58 { // c8 - queenside castling
			if pawns.HasBit(49) { // b7
				score += KingShelterBonus
			}
			if pawns.HasBit(50) { // c7
				score += KingShelterBonus
			}
			if pawns.HasBit(51) { // d7
				score += KingShelterBonus / 2 // Half bonus for d7
			}
		}
	}

	return score
}

// evaluateOpenFilesNearKing penalizes open files near the king
func evaluateOpenFilesNearKing(b *board.Board, kingSquare int) int {
	score := 0
	file := kingSquare % 8

	// Get all pawns to check for open files
	allPawns := b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.BlackPawn)

	// Check files around king (king file and adjacent files)
	for f := max(0, file-1); f <= min(7, file+1); f++ {
		fileMask := board.FileMask(f)
		if (allPawns & fileMask) == 0 {
			// Open file near king is dangerous
			score += OpenFileNearKing
		}
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