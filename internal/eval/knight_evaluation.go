package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// KnightOutpostBonus rewards a knight on a pawn-defended square in enemy
	// territory.
	KnightOutpostBonus = 30
)

// WhiteKnightOutpostRanks defines outpost ranks for white knights (ranks 4-6)
var WhiteKnightOutpostRanks = [8]bool{false, false, false, true, true, true, false, false}

// BlackKnightOutpostRanks defines outpost ranks for black knights (ranks 3-5)
var BlackKnightOutpostRanks = [8]bool{false, false, true, true, true, false, false, false}

func evaluateKnights(b *board.Board) int {
	if b == nil {
		return 0
	}
	score := 0

	whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
	blackKnights := b.GetPieceBitboard(board.BlackKnight)

	score += evaluateKnightsSimple(b, whiteKnights, true)
	score -= evaluateKnightsSimple(b, blackKnights, false)

	return score
}

func evaluateKnightsSimple(b *board.Board, knights board.Bitboard, isWhite bool) int {
	if b == nil || knights == 0 {
		return 0
	}

	score := 0
	area := mobilityArea(b, isWhite)

	var enemyPawns board.Bitboard
	var outpostRanks [8]bool
	if isWhite {
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		outpostRanks = WhiteKnightOutpostRanks
	} else {
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		outpostRanks = BlackKnightOutpostRanks
	}

	for knights != 0 {
		square, remaining := knights.PopLSB()
		knights = remaining

		rank := square / 8
		file := square % 8

		score += knightMobilityScore(square, area)

		if outpostRanks[rank] {
			var aheadMask board.Bitboard
			if isWhite {
				aheadMask = board.Bitboard(^uint64(0)) << ((rank + 1) * 8)
			} else {
				aheadMask = board.Bitboard(uint64(1)<<(rank*8) - 1)
			}

			var adjacentFiles board.Bitboard
			if file > 0 {
				adjacentFiles |= board.FileMask(file - 1)
			}
			if file < 7 {
				adjacentFiles |= board.FileMask(file + 1)
			}

			isOutpost := (enemyPawns & adjacentFiles & aheadMask) == 0

			if isOutpost {
				hasSupport := false
				var friendlyPawns board.Bitboard
				if isWhite {
					friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
					if rank > 0 {
						supportRank := rank - 1
						if file > 0 && friendlyPawns.HasBit(supportRank*8+(file-1)) {
							hasSupport = true
						}
						if file < 7 && friendlyPawns.HasBit(supportRank*8+(file+1)) {
							hasSupport = true
						}
					}
				} else {
					friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
					if rank < 7 {
						supportRank := rank + 1
						if file > 0 && friendlyPawns.HasBit(supportRank*8+(file-1)) {
							hasSupport = true
						}
						if file < 7 && friendlyPawns.HasBit(supportRank*8+(file+1)) {
							hasSupport = true
						}
					}
				}

				if hasSupport {
					score += KnightOutpostBonus
				}
			}
		}
	}

	return score
}
