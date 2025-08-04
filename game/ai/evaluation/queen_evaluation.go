package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Queen evaluation constants
const (
	// Early development penalty
	EarlyQueenDevelopmentPenalty  = -30
	EarlyDevelopmentMoveThreshold = 8 // Before move 8

	// Queen safety
	QueenAttackedByMinorPenalty = -25
	QueenAttackedByPawnPenalty  = -40
	QueenTrappedPenalty         = -60

	// Queen mobility
	QueenMobilityUnit         = 1 // Per square
	QueenCentralMobilityBonus = 2 // Extra for central control
	QueenSafeMobilityBonus    = 1 // Extra for safe squares

	// Pin detection
	QueenPinBonus          = 20 // Pinning a piece
	QueenAbsolutePinBonus  = 30 // Pinning to king
	QueenMultiplePinsBonus = 15 // Per additional pin

	// Battery bonuses
	QueenRookBatteryBonus   = 15
	QueenBishopBatteryBonus = 12

	// Queen centralization
	QueenCentralizationBonus = 10 // e4,d4,e5,d5
	QueenExtendedCenterBonus = 5  // c3-f6 rectangle

	// Attack bonuses
	QueenAttackingKingZone = 20
	QueenMultipleAttacks   = 15 // Attacking 2+ pieces
)

// evaluateQueens evaluates all queen-specific features
func evaluateQueens(b *board.Board) int {
	score := 0

	// White queen
	whiteQueen := b.GetPieceBitboard(board.WhiteQueen)
	if whiteQueen != 0 {
		score += evaluateQueenFeatures(b, whiteQueen, board.BitboardWhite)
	}

	// Black queen (subtract since we evaluate from white's perspective)
	blackQueen := b.GetPieceBitboard(board.BlackQueen)
	if blackQueen != 0 {
		score -= evaluateQueenFeatures(b, blackQueen, board.BitboardBlack)
	}

	return score
}

// evaluateQueenFeatures evaluates features for a queen
func evaluateQueenFeatures(b *board.Board, queenBitboard board.Bitboard, color board.BitboardColor) int {
	if queenBitboard == 0 {
		return 0
	}

	score := 0

	// Usually only one queen, but handle promotions
	for queenBitboard != 0 {
		square, newQueens := queenBitboard.PopLSB()
		queenBitboard = newQueens

		// Evaluate individual queen features
		score += evaluateEarlyDevelopment(b, square, color)
		score += evaluateQueenSafety(b, square, color)
		score += evaluateQueenMobility(b, square, color)
		score += evaluateQueenPins(b, square, color)
		score += evaluateQueenBatteries(b, square, color)
		score += evaluateQueenCentralization(square)
		score += evaluateQueenAttacks(b, square, color)
	}

	return score
}

// evaluateEarlyDevelopment penalizes early queen development
func evaluateEarlyDevelopment(b *board.Board, queenSquare int, color board.BitboardColor) int {
	// Check game phase (count moves or developed pieces)
	developedPieces := countDevelopedPieces(b, color)

	// If in opening (few pieces developed)
	if developedPieces < 3 {
		_, rank := board.SquareToFileRank(queenSquare)

		// Check if queen has moved from starting position
		var startingRank int
		if color == board.BitboardWhite {
			startingRank = 0 // rank 1
		} else {
			startingRank = 7 // rank 8
		}

		if rank != startingRank {
			// Queen developed before minor pieces
			return EarlyQueenDevelopmentPenalty
		}
	}

	return 0
}

// countDevelopedPieces counts developed minor pieces
func countDevelopedPieces(b *board.Board, color board.BitboardColor) int {
	count := 0

	if color == board.BitboardWhite {
		// Check white knights
		knights := b.GetPieceBitboard(board.WhiteKnight)
		if (knights & board.Bitboard(1<<board.B1)) == 0 { // b1 knight moved
			count++
		}
		if (knights & board.Bitboard(1<<board.G1)) == 0 { // g1 knight moved
			count++
		}

		// Check white bishops
		bishops := b.GetPieceBitboard(board.WhiteBishop)
		if (bishops & board.Bitboard(1<<board.C1)) == 0 { // c1 bishop moved
			count++
		}
		if (bishops & board.Bitboard(1<<board.F1)) == 0 { // f1 bishop moved
			count++
		}
	} else {
		// Similar for black pieces
		knights := b.GetPieceBitboard(board.BlackKnight)
		if (knights & board.Bitboard(1<<board.B8)) == 0 {
			count++
		}
		if (knights & board.Bitboard(1<<board.G8)) == 0 {
			count++
		}

		bishops := b.GetPieceBitboard(board.BlackBishop)
		if (bishops & board.Bitboard(1<<board.C8)) == 0 {
			count++
		}
		if (bishops & board.Bitboard(1<<board.F8)) == 0 {
			count++
		}
	}

	return count
}

// evaluateQueenSafety checks if queen is in danger
func evaluateQueenSafety(b *board.Board, queenSquare int, color board.BitboardColor) int {
	penalty := 0

	// Get enemy attackers
	enemyColor := board.OppositeBitboardColor(color)
	attackers := b.GetAttackersToSquare(queenSquare, enemyColor)

	if attackers == 0 {
		return 0 // Queen is safe
	}

	// Check what's attacking the queen
	var enemyPawns, enemyMinors board.Bitboard
	if enemyColor == board.BitboardWhite {
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyMinors = b.GetPieceBitboard(board.WhiteKnight) |
			b.GetPieceBitboard(board.WhiteBishop)
	} else {
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyMinors = b.GetPieceBitboard(board.BlackKnight) |
			b.GetPieceBitboard(board.BlackBishop)
	}

	// Penalty for pawn attacks (very bad)
	if (attackers & enemyPawns) != 0 {
		penalty += QueenAttackedByPawnPenalty
	}

	// Penalty for minor piece attacks
	if (attackers & enemyMinors) != 0 {
		penalty += QueenAttackedByMinorPenalty
	}

	// Check if queen is trapped (few escape squares)
	escapeSquares := evaluateQueenEscapeSquares(b, queenSquare, color)
	if escapeSquares < 3 {
		penalty += QueenTrappedPenalty
	}

	return penalty
}

// evaluateQueenEscapeSquares counts safe escape squares
func evaluateQueenEscapeSquares(b *board.Board, queenSquare int, color board.BitboardColor) int {
	// Get queen moves
	queenMoves := board.GetQueenAttacks(queenSquare, b.AllPieces)

	// Remove friendly pieces
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := queenMoves &^ friendlyPieces

	// Count safe squares (not attacked by enemy)
	safeSquares := 0
	enemyColor := board.OppositeBitboardColor(color)

	for validMoves != 0 {
		toSquare, newMoves := validMoves.PopLSB()
		validMoves = newMoves

		// Check if square is safe
		if !b.IsSquareAttackedByColor(toSquare, enemyColor) {
			safeSquares++
		}
	}

	return safeSquares
}

// evaluateQueenMobility evaluates queen mobility with safety considerations
func evaluateQueenMobility(b *board.Board, queenSquare int, color board.BitboardColor) int {
	// Get queen attacks
	attacks := board.GetQueenAttacks(queenSquare, b.AllPieces)

	// Remove friendly pieces
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := attacks &^ friendlyPieces

	// Basic mobility
	mobility := validMoves.PopCount()
	score := mobility * QueenMobilityUnit

	// Bonus for central control
	centralMask := getCentralMask()
	centralControl := (validMoves & centralMask).PopCount()
	score += centralControl * QueenCentralMobilityBonus

	// Bonus for safe mobility
	enemyColor := board.OppositeBitboardColor(color)
	enemyPawnAttacks := getEnemyPawnAttacks(b, enemyColor)
	safeMoves := validMoves &^ enemyPawnAttacks
	safeMobility := safeMoves.PopCount()
	score += safeMobility * QueenSafeMobilityBonus

	// Penalty if mobility is too low
	if mobility < 5 {
		score -= 20
	}

	return score
}

// getEnemyPawnAttacks returns all squares attacked by enemy pawns
func getEnemyPawnAttacks(b *board.Board, enemyColor board.BitboardColor) board.Bitboard {
	var pawnAttacks board.Bitboard

	if enemyColor == board.BitboardWhite {
		whitePawns := b.GetPieceBitboard(board.WhitePawn)
		// White pawns attack diagonally upward
		pawnAttacks = whitePawns.ShiftNorthEast() | whitePawns.ShiftNorthWest()
	} else {
		blackPawns := b.GetPieceBitboard(board.BlackPawn)
		// Black pawns attack diagonally downward
		pawnAttacks = blackPawns.ShiftSouthEast() | blackPawns.ShiftSouthWest()
	}

	return pawnAttacks
}

// evaluateQueenPins detects pins created by the queen
func evaluateQueenPins(b *board.Board, queenSquare int, color board.BitboardColor) int {
	score := 0
	pinCount := 0

	// Get enemy king position
	enemyColor := board.OppositeBitboardColor(color)
	var enemyKing board.Bitboard
	if enemyColor == board.BitboardWhite {
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
	} else {
		enemyKing = b.GetPieceBitboard(board.BlackKing)
	}

	// Check for absolute pins (to king)
	if enemyKing != 0 {
		kingSquare := enemyKing.LSB()
		pins := detectPinsFromQueen(b, queenSquare, kingSquare, enemyColor)

		if pins != 0 {
			score += QueenAbsolutePinBonus
			pinCount += pins.PopCount()
		}
	}

	// Check for relative pins (to valuable pieces)
	valuablePieces := getValuableEnemyPieces(b, enemyColor)

	for valuablePieces != 0 {
		targetSquare, newPieces := valuablePieces.PopLSB()
		valuablePieces = newPieces

		pins := detectPinsFromQueen(b, queenSquare, targetSquare, enemyColor)
		if pins != 0 {
			score += QueenPinBonus
			pinCount += pins.PopCount()
		}
	}

	// Bonus for multiple pins
	if pinCount > 1 {
		score += (pinCount - 1) * QueenMultiplePinsBonus
	}

	return score
}

// detectPinsFromQueen detects pieces pinned by queen
func detectPinsFromQueen(b *board.Board, queenSquare, targetSquare int, targetColor board.BitboardColor) board.Bitboard {
	// Check if squares are on same line
	if !areOnSameLine(queenSquare, targetSquare) {
		return 0
	}

	// Get pieces between queen and target
	between := board.GetBetween(queenSquare, targetSquare)
	blockers := between & b.AllPieces

	// Exactly one piece between = pin
	if blockers.PopCount() == 1 {
		blockerSquare := blockers.LSB()
		piece := b.GetPieceOnSquare(blockerSquare)

		// Check if blocker is same color as target
		if getPieceColor(piece) == targetColor {
			return blockers
		}
	}

	return 0
}

// areOnSameLine checks if two squares are on same rank, file, or diagonal
func areOnSameLine(sq1, sq2 int) bool {
	file1, rank1 := board.SquareToFileRank(sq1)
	file2, rank2 := board.SquareToFileRank(sq2)

	// Same file or rank
	if file1 == file2 || rank1 == rank2 {
		return true
	}

	// Same diagonal
	fileDiff := abs(file2 - file1)
	rankDiff := abs(rank2 - rank1)
	return fileDiff == rankDiff
}

// getValuableEnemyPieces returns enemy rooks and minor pieces
func getValuableEnemyPieces(b *board.Board, enemyColor board.BitboardColor) board.Bitboard {
	if enemyColor == board.BitboardWhite {
		return b.GetPieceBitboard(board.WhiteRook) |
			b.GetPieceBitboard(board.WhiteKnight) |
			b.GetPieceBitboard(board.WhiteBishop)
	}
	return b.GetPieceBitboard(board.BlackRook) |
		b.GetPieceBitboard(board.BlackKnight) |
		b.GetPieceBitboard(board.BlackBishop)
}

// evaluateQueenBatteries detects queen batteries with rooks/bishops
func evaluateQueenBatteries(b *board.Board, queenSquare int, color board.BitboardColor) int {
	score := 0

	// Get friendly rooks and bishops
	var rooks, bishops board.Bitboard
	if color == board.BitboardWhite {
		rooks = b.GetPieceBitboard(board.WhiteRook)
		bishops = b.GetPieceBitboard(board.WhiteBishop)
	} else {
		rooks = b.GetPieceBitboard(board.BlackRook)
		bishops = b.GetPieceBitboard(board.BlackBishop)
	}

	// Check rook-queen batteries
	for rooks != 0 {
		rookSquare, newRooks := rooks.PopLSB()
		rooks = newRooks

		if formsBattery(b, queenSquare, rookSquare) {
			score += QueenRookBatteryBonus

			// Extra bonus if battery attacks enemy king
			if batteryAttacksKing(b, queenSquare, rookSquare, color) {
				score += 10
			}
		}
	}

	// Check bishop-queen batteries
	for bishops != 0 {
		bishopSquare, newBishops := bishops.PopLSB()
		bishops = newBishops

		if formsBattery(b, queenSquare, bishopSquare) {
			score += QueenBishopBatteryBonus

			// Extra bonus if battery attacks enemy king
			if batteryAttacksKing(b, queenSquare, bishopSquare, color) {
				score += 10
			}
		}
	}

	return score
}

// formsBattery checks if two pieces form a battery
func formsBattery(b *board.Board, piece1, piece2 int) bool {
	// Check if on same line
	if !areOnSameLine(piece1, piece2) {
		return false
	}

	// Check if no pieces between them
	between := board.GetBetween(piece1, piece2)
	return (between & b.AllPieces) == 0
}

// batteryAttacksKing checks if battery points toward enemy king
func batteryAttacksKing(b *board.Board, piece1, piece2 int, color board.BitboardColor) bool {
	// Get enemy king
	enemyColor := board.OppositeBitboardColor(color)
	var enemyKing board.Bitboard
	if enemyColor == board.BitboardWhite {
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
	} else {
		enemyKing = b.GetPieceBitboard(board.BlackKing)
	}

	if enemyKing == 0 {
		return false
	}

	kingSquare := enemyKing.LSB()

	// Check if king is on the battery's line
	return areOnSameLine(piece1, kingSquare) && areOnSameLine(piece2, kingSquare)
}

// evaluateQueenCentralization rewards central queen placement
func evaluateQueenCentralization(queenSquare int) int {
	file, rank := board.SquareToFileRank(queenSquare)

	// Central squares (d4, e4, d5, e5)
	if (file == 3 || file == 4) && (rank == 3 || rank == 4) {
		return QueenCentralizationBonus
	}

	// Extended center (c3-f6 rectangle)
	if file >= 2 && file <= 5 && rank >= 2 && rank <= 5 {
		return QueenExtendedCenterBonus
	}

	return 0
}

// evaluateQueenAttacks evaluates queen attacking enemy pieces
func evaluateQueenAttacks(b *board.Board, queenSquare int, color board.BitboardColor) int {
	score := 0

	// Get queen attacks
	attacks := board.GetQueenAttacks(queenSquare, b.AllPieces)

	// Get enemy pieces
	enemyColor := board.OppositeBitboardColor(color)
	enemyPieces := b.GetColorBitboard(enemyColor)

	// Count attacked enemy pieces
	attackedPieces := attacks & enemyPieces
	attackCount := attackedPieces.PopCount()

	if attackCount >= 2 {
		score += QueenMultipleAttacks
	}

	// Check if attacking king zone
	var enemyKing board.Bitboard
	if enemyColor == board.BitboardWhite {
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
	} else {
		enemyKing = b.GetPieceBitboard(board.BlackKing)
	}

	if enemyKing != 0 {
		kingSquare := enemyKing.LSB()
		kingZone := getKingZone(kingSquare)

		if (attacks & kingZone) != 0 {
			score += QueenAttackingKingZone
		}
	}

	return score
}

// getKingZone returns squares around the king
func getKingZone(kingSquare int) board.Bitboard {
	// King attacks plus the king square itself
	zone := board.GetKingAttacks(kingSquare)
	zone = zone.SetBit(kingSquare)
	return zone
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
