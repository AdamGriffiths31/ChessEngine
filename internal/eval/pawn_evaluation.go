package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// IsolatedPawnPenalty applies to a pawn with no friendly pawns on adjacent files.
	IsolatedPawnPenalty = -15
	// DoubledPawnPenalty applies to each extra pawn on the same file.
	DoubledPawnPenalty = -10
	// BackwardPawnPenalty applies to a pawn that cannot advance safely.
	BackwardPawnPenalty = -8
	// ConnectedPawnBonus applies to pawns that protect each other.
	ConnectedPawnBonus = 8
)

// PassedPawnBonus provides exponential bonuses for passed pawns by rank
// Index represents rank (0-7), values increase exponentially toward promotion
var PassedPawnBonus = [8]int{0, 10, 15, 25, 40, 60, 90, 0}

// PawnHashTable provides global caching for pawn evaluations
var PawnHashTable [16384]PawnHashEntry

func evaluatePawnStructure(b *board.Board) int {
	if b == nil {
		return 0
	}
	pawnHash := b.GetPawnHash()
	hashIndex := pawnHash & 16383

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

func evaluatePawnsSimple(whitePawns, blackPawns board.Bitboard) int {
	score := 0

	score += evaluatePawnsByColor(whitePawns, blackPawns, true)

	score -= evaluatePawnsByColor(blackPawns, whitePawns, false)

	return score
}

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

	for file := range 8 {
		if filePawns[file] > 1 {
			score += (filePawns[file] - 1) * DoubledPawnPenalty
		}
	}

	return score
}

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

func isIsolatedPawn(friendlyPawns board.Bitboard, file int) bool {
	if file > 0 && (friendlyPawns&board.FileMask(file-1)) != 0 {
		return false
	}

	if file < 7 && (friendlyPawns&board.FileMask(file+1)) != 0 {
		return false
	}

	return true
}

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

func isBackwardPawn(square int, friendlyPawns, enemyPawns board.Bitboard, isWhite bool) bool {
	file := square % 8
	rank := square / 8

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

func hasPawnAt(pawns board.Bitboard, file, rank int) bool {
	square := rank*8 + file
	return pawns.HasBit(square)
}
