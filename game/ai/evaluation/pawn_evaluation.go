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

	// Entries store hash XOR score so a torn read (hash and score from
	// different positions) fails validation instead of returning garbage.
	entry := &PawnHashTable[hashIndex]
	cachedScore := entry.score
	if entry.hash^uint64(int64(cachedScore)) == pawnHash {
		return cachedScore
	}

	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)
	score := evaluatePawnsSimple(whitePawns, blackPawns)

	entry.score = score
	entry.hash = pawnHash ^ uint64(int64(score))

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
		} else if isBackwardPawn(square, friendlyPawns, enemyPawns, isWhite) {
			score += BackwardPawnPenalty
		}

		if isConnectedPawn(friendlyPawns, square, isWhite) {
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
// Support comes from behind relative to the pawn's direction of travel:
// rank-1 for White, rank+1 for Black
func isConnectedPawn(friendlyPawns board.Bitboard, square int, isWhite bool) bool {
	file := square % 8
	rank := square / 8

	supportRank := rank - 1
	if !isWhite {
		supportRank = rank + 1
	}
	if supportRank < 0 || supportRank > 7 {
		return false
	}

	if file > 0 && friendlyPawns.HasBit(supportRank*8+(file-1)) {
		return true
	}

	if file < 7 && friendlyPawns.HasBit(supportRank*8+(file+1)) {
		return true
	}

	return false
}

// isBackwardPawn checks if a pawn has fallen behind its neighbors and cannot
// advance safely: no friendly pawn on an adjacent file is level with it or
// behind it, and its stop square is covered by an enemy pawn
func isBackwardPawn(square int, friendlyPawns, enemyPawns board.Bitboard, isWhite bool) bool {
	file := square % 8
	rank := square / 8

	// A pawn that can still be supported from behind is not backward
	for _, f := range [2]int{file - 1, file + 1} {
		if f < 0 || f > 7 {
			continue
		}
		adjacentPawns := friendlyPawns & board.FileMask(f)
		for adjacentPawns != 0 {
			adjSquare, remaining := adjacentPawns.PopLSB()
			adjacentPawns = remaining
			adjRank := adjSquare / 8
			if (isWhite && adjRank <= rank) || (!isWhite && adjRank >= rank) {
				return false
			}
		}
	}

	// Backward only if the stop square is covered by an enemy pawn
	var stopRank, attackerRank int
	if isWhite {
		stopRank = rank + 1
		attackerRank = stopRank + 1
	} else {
		stopRank = rank - 1
		attackerRank = stopRank - 1
	}
	if stopRank < 0 || stopRank > 7 || attackerRank < 0 || attackerRank > 7 {
		return false
	}

	if file > 0 && enemyPawns.HasBit(attackerRank*8+(file-1)) {
		return true
	}
	if file < 7 && enemyPawns.HasBit(attackerRank*8+(file+1)) {
		return true
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
