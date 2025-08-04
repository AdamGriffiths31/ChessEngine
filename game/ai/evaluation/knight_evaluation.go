package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Knight evaluation constants
const (
	// Outpost bonuses
	KnightOutpostBase     = 30
	KnightOutpostAdvanced = 50 // Extra bonus for advanced outposts (rank 5+)

	// Mobility bonuses (per available square)
	KnightMobilityUnit   = 4
	KnightMobilityCenter = 2 // Extra bonus for central moves

	// Penalties
	KnightTrappedPenalty = -50 // Less than 3 moves available
	KnightBadPenalty     = -25 // 3-4 moves available

	// Fork bonuses
	KnightForkThreat = 15 // Potential fork
	KnightForkActive = 25 // Active fork of 2+ pieces
	KnightRoyalFork  = 50 // Fork involving king/queen
)

// evaluateKnights evaluates all knight-specific features
func evaluateKnights(b *board.Board) int {
	score := 0

	// White knights
	whiteKnights := b.GetPieceBitboard(board.WhiteKnight)
	score += evaluateKnightsBitboard(b, whiteKnights, board.BitboardWhite)

	// Black knights (subtract since we evaluate from white's perspective)
	blackKnights := b.GetPieceBitboard(board.BlackKnight)
	score -= evaluateKnightsBitboard(b, blackKnights, board.BitboardBlack)

	return score
}

// evaluateKnightsBitboard evaluates knights for one side
func evaluateKnightsBitboard(b *board.Board, knights board.Bitboard, color board.BitboardColor) int {
	if knights == 0 {
		return 0
	}

	score := 0

	// Get relevant bitboards
	var friendlyPawns, enemyPawns board.Bitboard
	if color == board.BitboardWhite {
		friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
	} else {
		friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
	}

	// Process each knight
	for knights != 0 {
		square, newKnights := knights.PopLSB()
		knights = newKnights

		// Evaluate individual knight features
		score += evaluateKnightOutpost(square, friendlyPawns, enemyPawns, color)
		score += evaluateKnightMobility(b, square, color)
		score += evaluateKnightForks(b, square, color)
	}

	return score
}

// evaluateKnightOutpost checks if knight is on an outpost square
func evaluateKnightOutpost(square int, friendlyPawns, enemyPawns board.Bitboard, color board.BitboardColor) int {
	file, rank := board.SquareToFileRank(square)

	// Check if knight is in enemy territory
	if color == board.BitboardWhite && rank < 3 {
		return 0 // Not advanced enough
	} else if color == board.BitboardBlack && rank > 4 {
		return 0 // Not advanced enough
	}

	// Check if square is supported by friendly pawn
	supportMask := getKnightSupportMask(square, color)
	if (friendlyPawns & supportMask) == 0 {
		return 0 // Not supported
	}

	// Check if square can be attacked by enemy pawns
	if canBeAttackedByPawn(square, enemyPawns, color) {
		return 0 // Not a true outpost
	}

	bonus := KnightOutpostBase

	// Extra bonus for advanced outposts
	if (color == board.BitboardWhite && rank >= 5) ||
		(color == board.BitboardBlack && rank <= 2) {
		bonus = KnightOutpostAdvanced
	}

	// Extra bonus for central outposts
	if file >= 2 && file <= 5 {
		bonus += 10
	}

	return bonus
}

// getKnightSupportMask returns squares that can support the knight with pawns
func getKnightSupportMask(square int, color board.BitboardColor) board.Bitboard {
	file, rank := board.SquareToFileRank(square)
	var mask board.Bitboard

	if color == board.BitboardWhite {
		// Pawns support from behind and diagonally
		if rank > 0 {
			if file > 0 {
				mask = mask.SetBit(board.FileRankToSquare(file-1, rank-1))
			}
			if file < 7 {
				mask = mask.SetBit(board.FileRankToSquare(file+1, rank-1))
			}
		}
	} else {
		// Black pawns support from above
		if rank < 7 {
			if file > 0 {
				mask = mask.SetBit(board.FileRankToSquare(file-1, rank+1))
			}
			if file < 7 {
				mask = mask.SetBit(board.FileRankToSquare(file+1, rank+1))
			}
		}
	}

	return mask
}

// canBeAttackedByPawn checks if square can be attacked by enemy pawns
func canBeAttackedByPawn(square int, enemyPawns board.Bitboard, friendlyColor board.BitboardColor) bool {
	file, rank := board.SquareToFileRank(square)

	// Check files where enemy pawns could attack from
	for f := file - 1; f <= file+1; f += 2 {
		if f < 0 || f > 7 {
			continue
		}

		// Check all ranks ahead where enemy pawns could be
		if friendlyColor == board.BitboardWhite {
			// Check black pawns ahead
			for r := rank + 1; r < 8; r++ {
				if enemyPawns.HasBit(board.FileRankToSquare(f, r)) {
					return true
				}
			}
		} else {
			// Check white pawns ahead
			for r := rank - 1; r >= 0; r-- {
				if enemyPawns.HasBit(board.FileRankToSquare(f, r)) {
					return true
				}
			}
		}
	}

	return false
}

// evaluateKnightMobility evaluates knight mobility
func evaluateKnightMobility(b *board.Board, square int, color board.BitboardColor) int {
	// Get all possible knight moves
	attacks := board.GetKnightAttacks(square)

	// Remove squares occupied by friendly pieces
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := attacks &^ friendlyPieces

	// Count mobility
	mobility := validMoves.PopCount()

	// Check if knight is trapped or has bad mobility
	if mobility < 3 {
		return KnightTrappedPenalty
	} else if mobility < 5 {
		return KnightBadPenalty
	}

	// Base mobility bonus
	score := mobility * KnightMobilityUnit

	// Bonus for central mobility
	centralMask := getCentralMask()
	centralMoves := validMoves & centralMask
	score += centralMoves.PopCount() * KnightMobilityCenter

	return score
}

// getCentralMask returns a bitboard of central squares
func getCentralMask() board.Bitboard {
	var mask board.Bitboard
	// Define central squares (c3-f6 rectangle)
	for rank := 2; rank <= 5; rank++ {
		for file := 2; file <= 5; file++ {
			mask = mask.SetBit(board.FileRankToSquare(file, rank))
		}
	}
	return mask
}

// evaluateKnightForks detects knight forks
func evaluateKnightForks(b *board.Board, square int, color board.BitboardColor) int {
	attacks := board.GetKnightAttacks(square)

	// Get enemy pieces
	enemyColor := board.OppositeBitboardColor(color)
	enemyPieces := b.GetColorBitboard(enemyColor)

	// Find attacked enemy pieces
	attackedPieces := attacks & enemyPieces
	attackedCount := attackedPieces.PopCount()

	if attackedCount < 2 {
		// Check for potential forks (one move away)
		return evaluatePotentialForks(b, square, color)
	}

	// Active fork detected
	score := KnightForkActive

	// Check if it's a royal fork (involves king or queen)
	var enemyKing, enemyQueen board.Bitboard
	if color == board.BitboardWhite {
		enemyKing = b.GetPieceBitboard(board.BlackKing)
		enemyQueen = b.GetPieceBitboard(board.BlackQueen)
	} else {
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
		enemyQueen = b.GetPieceBitboard(board.WhiteQueen)
	}

	if (attackedPieces & (enemyKing | enemyQueen)) != 0 {
		score = KnightRoyalFork
	}

	return score
}

// evaluatePotentialForks checks for fork threats
func evaluatePotentialForks(b *board.Board, knightSquare int, color board.BitboardColor) int {
	// Get squares the knight can move to
	knightMoves := board.GetKnightAttacks(knightSquare)
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := knightMoves &^ friendlyPieces

	maxForkValue := 0

	// Check each potential knight destination
	for validMoves != 0 {
		toSquare, newMoves := validMoves.PopLSB()
		validMoves = newMoves

		// Check if this square would create a fork
		potentialAttacks := board.GetKnightAttacks(toSquare)
		enemyColor := board.OppositeBitboardColor(color)
		enemyPieces := b.GetColorBitboard(enemyColor)
		attackedPieces := potentialAttacks & enemyPieces

		if attackedPieces.PopCount() >= 2 {
			// This move would create a fork
			value := KnightForkThreat

			// Check for royal fork potential
			var enemyKing, enemyQueen board.Bitboard
			if color == board.BitboardWhite {
				enemyKing = b.GetPieceBitboard(board.BlackKing)
				enemyQueen = b.GetPieceBitboard(board.BlackQueen)
			} else {
				enemyKing = b.GetPieceBitboard(board.WhiteKing)
				enemyQueen = b.GetPieceBitboard(board.WhiteQueen)
			}

			if (attackedPieces & (enemyKing | enemyQueen)) != 0 {
				value = KnightForkThreat * 2
			}

			if value > maxForkValue {
				maxForkValue = value
			}
		}
	}

	return maxForkValue
}
