package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Rook evaluation constants
const (
	// Open and semi-open file bonuses
	RookOpenFileBonus     = 25
	RookSemiOpenFileBonus = 15

	// 7th rank bonus (when enemy king on 8th)
	RookOnSeventhBonus   = 30
	RookPairSeventhBonus = 50 // Both rooks on 7th

	// Doubled rook bonuses
	DoubledRooksFileBonus = 20
	DoubledRooksRankBonus = 15

	// Mobility
	RookMobilityUnit     = 2
	RookTrappedPenalty   = -50 // Less than 4 moves
	RookPartiallyTrapped = -25 // 4-6 moves

	// Trapped by king
	RookTrappedByKing = -30 // Extra penalty

	// Battery bonuses
	RookQueenBatteryBonus = 15

	// Connected rooks
	ConnectedRooksBonus = 10
)

// evaluateRooks evaluates all rook-specific features
func evaluateRooks(b *board.Board) int {
	score := 0

	// White rooks
	whiteRooks := b.GetPieceBitboard(board.WhiteRook)
	score += evaluateRookFeatures(b, whiteRooks, board.BitboardWhite)

	// Black rooks (subtract since we evaluate from white's perspective)
	blackRooks := b.GetPieceBitboard(board.BlackRook)
	score -= evaluateRookFeatures(b, blackRooks, board.BitboardBlack)

	// Evaluate rook pairs (doubled/connected rooks)
	score += evaluateRookPairs(b, whiteRooks, board.BitboardWhite)
	score -= evaluateRookPairs(b, blackRooks, board.BitboardBlack)

	return score
}

// evaluateRookFeatures evaluates features for rooks of one color
func evaluateRookFeatures(b *board.Board, rooks board.Bitboard, color board.BitboardColor) int {
	if rooks == 0 {
		return 0
	}

	score := 0

	// Get relevant bitboards
	var friendlyPawns, enemyPawns, enemyKing board.Bitboard
	if color == board.BitboardWhite {
		friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyKing = b.GetPieceBitboard(board.BlackKing)
	} else {
		friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
	}

	// Process each rook
	for rooks != 0 {
		square, newRooks := rooks.PopLSB()
		rooks = newRooks

		// Evaluate individual rook features
		score += evaluateOpenFiles(square, friendlyPawns, enemyPawns)
		score += evaluateSeventhRank(square, enemyKing, color)
		score += evaluateRookMobility(b, square, color)
		score += evaluateRookTrappedByKing(b, square, color)
	}

	return score
}

// evaluateOpenFiles checks if rook is on open or semi-open file
func evaluateOpenFiles(rookSquare int, friendlyPawns, enemyPawns board.Bitboard) int {
	file, _ := board.SquareToFileRank(rookSquare)
	fileMask := board.FileMask(file)

	friendlyPawnsOnFile := friendlyPawns & fileMask
	enemyPawnsOnFile := enemyPawns & fileMask

	if friendlyPawnsOnFile == 0 && enemyPawnsOnFile == 0 {
		// Open file
		return RookOpenFileBonus
	} else if friendlyPawnsOnFile == 0 {
		// Semi-open file (only enemy pawns)
		return RookSemiOpenFileBonus
	}

	return 0
}

// evaluateSeventhRank checks if rook is on 7th rank with enemy king on 8th
func evaluateSeventhRank(rookSquare int, enemyKing board.Bitboard, color board.BitboardColor) int {
	_, rank := board.SquareToFileRank(rookSquare)

	// Check if rook is on 7th rank (6 for white, 1 for black)
	var seventhRank int
	var eighthRank int

	if color == board.BitboardWhite {
		seventhRank = 6
		eighthRank = 7
	} else {
		seventhRank = 1
		eighthRank = 0
	}

	if rank != seventhRank {
		return 0
	}

	// Check if enemy king is on 8th rank
	if enemyKing != 0 {
		kingSquare := enemyKing.LSB()
		_, kingRank := board.SquareToFileRank(kingSquare)

		if kingRank == eighthRank {
			return RookOnSeventhBonus
		}
	}

	// Still give small bonus for rook on 7th even without king on 8th
	return 10
}

// evaluateRookPairs evaluates doubled and connected rooks
func evaluateRookPairs(b *board.Board, rooks board.Bitboard, color board.BitboardColor) int {
	if rooks.PopCount() < 2 {
		return 0
	}

	score := 0

	// Get rook positions
	var rookSquares []int
	tempRooks := rooks
	for tempRooks != 0 {
		square, newRooks := tempRooks.PopLSB()
		tempRooks = newRooks
		rookSquares = append(rookSquares, square)
	}

	// Check pairs of rooks
	for i := 0; i < len(rookSquares); i++ {
		for j := i + 1; j < len(rookSquares); j++ {
			score += evaluateRookPair(b, rookSquares[i], rookSquares[j], color)
		}
	}

	return score
}

// evaluateRookPair evaluates a pair of rooks
func evaluateRookPair(b *board.Board, rook1, rook2 int, color board.BitboardColor) int {
	file1, rank1 := board.SquareToFileRank(rook1)
	file2, rank2 := board.SquareToFileRank(rook2)

	score := 0

	// Doubled on same file
	if file1 == file2 {
		score += DoubledRooksFileBonus

		// Extra bonus if on open file
		fileMask := board.FileMask(file1)
		allPawns := b.GetPieceBitboard(board.WhitePawn) | b.GetPieceBitboard(board.BlackPawn)
		if (fileMask & allPawns) == 0 {
			score += 10 // Extra for doubled on open file
		}
	}

	// Doubled on same rank
	if rank1 == rank2 {
		score += DoubledRooksRankBonus

		// Check if both on 7th rank
		var seventhRank int
		if color == board.BitboardWhite {
			seventhRank = 6
		} else {
			seventhRank = 1
		}

		if rank1 == seventhRank {
			score += RookPairSeventhBonus
		}
	}

	// Connected rooks (protecting each other)
	if areRooksConnected(b, rook1, rook2) {
		score += ConnectedRooksBonus
	}

	return score
}

// areRooksConnected checks if rooks protect each other
func areRooksConnected(b *board.Board, rook1, rook2 int) bool {
	// Check if rook1 attacks rook2's square
	rook1Attacks := board.GetRookAttacks(rook1, b.AllPieces)
	if rook1Attacks.HasBit(rook2) {
		// Verify no pieces between them
		between := board.GetBetween(rook1, rook2)
		if (between & b.AllPieces) == 0 {
			return true
		}
	}
	return false
}

// evaluateRookMobility evaluates rook mobility
func evaluateRookMobility(b *board.Board, square int, color board.BitboardColor) int {
	// Get rook attacks
	attacks := board.GetRookAttacks(square, b.AllPieces)

	// Remove squares occupied by friendly pieces
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := attacks &^ friendlyPieces

	// Count mobility
	mobility := validMoves.PopCount()

	// Check if rook is trapped
	if mobility < 4 {
		return RookTrappedPenalty
	} else if mobility < 7 {
		return RookPartiallyTrapped
	}

	// Base mobility bonus
	score := mobility * RookMobilityUnit

	// Bonus for horizontal mobility (rank control)
	file, rank := board.SquareToFileRank(square)
	rankMask := board.RankMask(rank)
	horizontalMobility := (validMoves & rankMask).PopCount()
	score += horizontalMobility // Extra point per horizontal move

	// Bonus for vertical mobility (file control)
	fileMask := board.FileMask(file)
	verticalMobility := (validMoves & fileMask).PopCount()
	score += verticalMobility // Extra point per vertical move

	return score
}

// evaluateRookTrappedByKing checks if rook is trapped by own king
func evaluateRookTrappedByKing(b *board.Board, rookSquare int, color board.BitboardColor) int {
	// Get king position
	var kingBitboard board.Bitboard
	if color == board.BitboardWhite {
		kingBitboard = b.GetPieceBitboard(board.WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(board.BlackKing)
	}

	if kingBitboard == 0 {
		return 0
	}

	kingSquare := kingBitboard.LSB()
	kingFile, kingRank := board.SquareToFileRank(kingSquare)
	rookFile, rookRank := board.SquareToFileRank(rookSquare)

	// Check common trapped rook patterns
	if color == board.BitboardWhite {
		// Kingside castling trap (king on g1, rook on h1/f1)
		if kingFile == 6 && kingRank == 0 { // g1
			if rookFile == 7 && rookRank == 0 { // h1
				return RookTrappedByKing
			}
			if rookFile == 5 && rookRank == 0 { // f1
				return RookTrappedByKing
			}
		}
		// Queenside castling trap
		if kingFile == 2 && kingRank == 0 { // c1
			if rookFile == 0 && rookRank == 0 { // a1
				return RookTrappedByKing / 2 // Less severe
			}
		}
	} else {
		// Similar patterns for black
		if kingFile == 6 && kingRank == 7 { // g8
			if rookFile == 7 && rookRank == 7 { // h8
				return RookTrappedByKing
			}
			if rookFile == 5 && rookRank == 7 { // f8
				return RookTrappedByKing
			}
		}
		if kingFile == 2 && kingRank == 7 { // c8
			if rookFile == 0 && rookRank == 7 { // a8
				return RookTrappedByKing / 2
			}
		}
	}

	return 0
}
