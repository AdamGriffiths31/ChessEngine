package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Pawn evaluation - streamlined for performance with focus on key factors
//
// Design Philosophy:
// 1. Passed pawns are the dominant pawn factor (exponential bonus by rank)
// 2. Pawn structure penalties (isolated, doubled, backward) are secondary
// 3. Simple connected pawns bonus for pawn chains
// 4. Pawn hash table for caching expensive pawn evaluations
// 5. Eliminates complex calculations (pawn storms, candidate passed, weak squares)
//
// This approach focuses on the most impactful pawn features while using caching
// to avoid recalculating identical pawn structures multiple times per search.

// Pawn evaluation constants - only the essentials
const (
	// Structure penalties
	IsolatedPawnPenalty = -15 // Pawn with no friendly pawns on adjacent files
	DoubledPawnPenalty  = -10 // Extra pawns on the same file
	BackwardPawnPenalty = -8  // Pawn that cannot advance safely

	// Connected pawns bonus
	ConnectedPawnBonus = 8 // Bonus for pawns protecting each other
)

// PassedPawnBonus provides exponential bonuses for passed pawns by rank
// Index represents rank (0-7), values increase exponentially toward promotion
var PassedPawnBonus = [8]int{0, 10, 15, 25, 40, 60, 90, 0}

// PawnHashTable provides global caching for pawn evaluations
// 16K entries = ~256KB of memory for significant speed improvement
var PawnHashTable [16384]PawnHashEntry

// evaluatePawnStructure performs cached pawn structure evaluation
// Uses pawn hash table to avoid recalculating identical pawn structures
func evaluatePawnStructure(b *board.Board) int {
	if b == nil {
		return 0
	}
	pawnHash := b.GetPawnHash()
	hashIndex := pawnHash & 16383

	entry := &PawnHashTable[hashIndex]
	if entry.hash == pawnHash {
		return entry.score
	}

	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)
	score := evaluatePawnsSimple(whitePawns, blackPawns)

	entry.hash = pawnHash
	entry.score = score

	return score
}

// evaluatePawnsSimple performs fast pawn evaluation for both colors
// Returns positive values favoring White, negative favoring Black
func evaluatePawnsSimple(whitePawns, blackPawns board.Bitboard) int {
	score := 0

	score += evaluatePawnsByColor(whitePawns, blackPawns, true)

	score -= evaluatePawnsByColor(blackPawns, whitePawns, false)

	return score
}

// evaluatePawnsByColor performs streamlined pawn evaluation for one color
func evaluatePawnsByColor(friendlyPawns, enemyPawns board.Bitboard, isWhite bool) int {
	if friendlyPawns == 0 {
		return 0
	}

	score := 0

	var filePawns [8]int

	tempPawns := friendlyPawns
	for tempPawns != 0 {
		square, remaining := tempPawns.PopLSB()
		tempPawns = remaining

		file := square % 8
		rank := square / 8

		filePawns[file]++

		if isPassedPawn(square, enemyPawns, isWhite) {
			if isWhite {
				score += PassedPawnBonus[rank]
			} else {
				score += PassedPawnBonus[7-rank]
			}
		}

		if isIsolatedPawn(friendlyPawns, file) {
			score += IsolatedPawnPenalty
		}

		if isConnectedPawn(friendlyPawns, square) {
			score += ConnectedPawnBonus
		}
	}

	for file := 0; file < 8; file++ {
		if filePawns[file] > 1 {
			score += (filePawns[file] - 1) * DoubledPawnPenalty
		}
	}

	return score
}

// isPassedPawn checks if a pawn is passed (no enemy pawns can stop it)
func isPassedPawn(square int, enemyPawns board.Bitboard, isWhite bool) bool {
	file := square % 8
	rank := square / 8

	if isWhite {
		for r := rank + 1; r < 8; r++ {
			for f := max(0, file-1); f <= min(7, file+1); f++ {
				if hasPawnAt(enemyPawns, f, r) {
					return false
				}
			}
		}
	} else {
		for r := rank - 1; r >= 0; r-- {
			for f := max(0, file-1); f <= min(7, file+1); f++ {
				if hasPawnAt(enemyPawns, f, r) {
					return false
				}
			}
		}
	}

	return true
}

// isIsolatedPawn checks if a pawn has no friendly pawns on adjacent files
func isIsolatedPawn(friendlyPawns board.Bitboard, file int) bool {
	if file > 0 && (friendlyPawns&board.FileMask(file-1)) != 0 {
		return false
	}

	if file < 7 && (friendlyPawns&board.FileMask(file+1)) != 0 {
		return false
	}

	return true
}

// isConnectedPawn checks if a pawn is protected by another friendly pawn
// Only checks backward diagonals - pawns in front cannot provide protection
func isConnectedPawn(friendlyPawns board.Bitboard, square int) bool {
	file := square % 8
	rank := square / 8

	// Check left diagonal support (backward)
	if file > 0 && rank > 0 {
		supportSquare := (rank-1)*8 + (file - 1)
		if friendlyPawns.HasBit(supportSquare) {
			return true
		}
	}

	// Check right diagonal support (backward)
	if file < 7 && rank > 0 {
		supportSquare := (rank-1)*8 + (file + 1)
		if friendlyPawns.HasBit(supportSquare) {
			return true
		}
	}

	return false
}

// hasPawnAt checks if a pawn exists at the specified file and rank
func hasPawnAt(pawns board.Bitboard, file, rank int) bool {
	square := rank*8 + file
	return pawns.HasBit(square)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
