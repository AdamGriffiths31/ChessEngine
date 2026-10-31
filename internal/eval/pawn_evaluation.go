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

const (
	pawnHashBits       = 14
	pawnHashSize       = 1 << pawnHashBits
	pawnHashMultiplier = 0x9E3779B97F4A7C15
)

// PassedPawnBonus provides exponential bonuses for passed pawns by rank
// Index represents rank (0-7), values increase exponentially toward promotion
var PassedPawnBonus = [8]int{0, 10, 15, 25, 40, 60, 90, 0}

func (e *StandardEvaluator) evaluatePawnStructure(b *board.Board) int {
	pawnHash := b.GetPawnHash()
	entry := &e.pawnHashTable[pawnHashIndex(pawnHash)]
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

func pawnHashIndex(pawnHash uint64) uint64 {
	return (pawnHash * pawnHashMultiplier) >> (64 - pawnHashBits)
}

func evaluatePawnsSimple(whitePawns, blackPawns board.Bitboard) int {
	return evaluatePawnsByColor(whitePawns, blackPawns, true) -
		evaluatePawnsByColor(blackPawns, whitePawns, false)
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
			score += PassedPawnBonus[relativeRank(rank, isWhite)]
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

	return enemyPawns&(board.FileMask(file)|adjacentFiles(file))&ranksAhead(rank, isWhite) == 0
}

func isIsolatedPawn(friendlyPawns board.Bitboard, file int) bool {
	return friendlyPawns&adjacentFiles(file) == 0
}

func isConnectedPawn(friendlyPawns board.Bitboard, square int, isWhite bool) bool {
	if isWhite {
		return friendlyPawns&board.BlackPawnAttacks[square] != 0
	}
	return friendlyPawns&board.WhitePawnAttacks[square] != 0
}

func isBackwardPawn(square int, friendlyPawns, enemyPawns board.Bitboard, isWhite bool) bool {
	file := square % 8
	rank := square / 8

	if friendlyPawns&adjacentFiles(file)&^ranksAhead(rank, isWhite) != 0 {
		return false
	}

	if isWhite {
		return enemyPawns&board.GetPawnAttacks(square+8, board.BitboardWhite) != 0
	}
	return enemyPawns&board.GetPawnAttacks(square-8, board.BitboardBlack) != 0
}

func adjacentFiles(file int) board.Bitboard {
	mask := board.FileMask(file)
	return mask.ShiftEast() | mask.ShiftWest()
}

func ranksAhead(rank int, isWhite bool) board.Bitboard {
	if isWhite {
		return board.Bitboard(^uint64(0)) << ((rank + 1) * 8)
	}
	return board.Bitboard(uint64(1)<<(rank*8) - 1)
}

func relativeRank(rank int, isWhite bool) int {
	if isWhite {
		return rank
	}
	return 7 - rank
}
