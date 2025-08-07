package evaluation

import (
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Pawn shelter evaluation constants
const (
	// Penalties for missing shelter pawns
	MissingShelterPawnKingFile = 25 // Missing pawn directly in front of king
	MissingShelterPawnAdjFile  = 15 // Missing pawn on adjacent file

	// Penalties for advanced shelter pawns (creating holes)
	AdvancedShelterPawn1 = 10 // Pawn advanced 1 square from starting position
	AdvancedShelterPawn2 = 20 // Pawn advanced 2+ squares from starting position

	// Pawn storm penalty (enemy pawns advancing toward king)
	PawnStormPenalty = 15 // Per advancing enemy pawn near king
)

// Castling rights evaluation constants
const (
	// Castling rights bonuses
	CastlingRightsBonus     = 15 // Bonus for having any castling rights
	BothSidesCastlingBonus  = 25 // Bonus for having both kingside and queenside rights
	KingsideCastlingBonus   = 10 // Specific bonus for kingside rights
	QueensideCastlingBonus  = 8  // Specific bonus for queenside rights (slightly less valuable)
	CastledKingBonus        = 20 // Bonus if king has already castled safely
)

// evaluateKings evaluates both kings and returns the score from White's perspective
func evaluateKings(b *board.Board) int {
	score := 0

	// Get king positions
	whiteKing := b.GetPieceBitboard(board.WhiteKing)
	blackKing := b.GetPieceBitboard(board.BlackKing)

	// Evaluate white king (positive contribution)
	if whiteKing != 0 {
		whiteKingSquare := whiteKing.LSB()
		score += evaluatePawnShelter(b, whiteKingSquare, board.BitboardWhite)
		score += evaluateCastlingRights(b, board.BitboardWhite)
	}

	// Evaluate black king (negative contribution)
	if blackKing != 0 {
		blackKingSquare := blackKing.LSB()
		score -= evaluatePawnShelter(b, blackKingSquare, board.BitboardBlack)
		score -= evaluateCastlingRights(b, board.BitboardBlack)
	}

	return score
}

// evaluatePawnShelter analyzes the pawn shield strength around the king
func evaluatePawnShelter(b *board.Board, kingSquare int, color board.BitboardColor) int {
	score := 0
	kingFile, kingRank := board.SquareToFileRank(kingSquare)

	// Define shelter evaluation parameters based on color
	var friendlyPawns, enemyPawns board.Bitboard
	var shelterRank, enemyDirection int

	if color == board.BitboardWhite {
		friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
		shelterRank = kingRank + 1 // White shelter pawns should be in front
		enemyDirection = -1        // Black pawns advance downward
	} else {
		friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
		shelterRank = kingRank - 1 // Black shelter pawns should be in front
		enemyDirection = 1         // White pawns advance upward
	}

	// Evaluate pawn shelter on three files: king file and both adjacent files
	for fileOffset := -1; fileOffset <= 1; fileOffset++ {
		file := kingFile + fileOffset
		if file < 0 || file > 7 {
			continue // Skip files outside the board
		}

		// Evaluate friendly shelter pawn on this file
		score += evaluatePawnShieldFile(b, file, shelterRank, friendlyPawns, fileOffset == 0)

		// Check for enemy pawn storms
		score -= evaluatePawnStorm(b, file, kingRank, enemyPawns, enemyDirection)
	}

	return score
}

// evaluatePawnShieldFile evaluates a single file of the pawn shield
func evaluatePawnShieldFile(b *board.Board, file int, expectedRank int, friendlyPawns board.Bitboard, isKingFile bool) int {
	score := 0
	fileMask := board.FileMask(file)
	pawnsOnFile := friendlyPawns & fileMask

	if pawnsOnFile == 0 {
		// No pawn on this file - apply penalty
		if isKingFile {
			score -= MissingShelterPawnKingFile
		} else {
			score -= MissingShelterPawnAdjFile
		}
	} else {
		// Find the most advanced pawn on this file
		pawnSquares := pawnsOnFile.BitList()
		for _, square := range pawnSquares {
			_, pawnRank := board.SquareToFileRank(square)

			// Check how far the pawn has advanced from ideal shelter position
			advancement := abs(pawnRank - expectedRank)
			if advancement == 1 {
				score -= AdvancedShelterPawn1
			} else if advancement >= 2 {
				score -= AdvancedShelterPawn2
			}
			// Only evaluate the most advanced pawn
			break
		}
	}

	return score
}

// evaluatePawnStorm checks for enemy pawns advancing toward the king
func evaluatePawnStorm(b *board.Board, file int, kingRank int, enemyPawns board.Bitboard, direction int) int {
	fileMask := board.FileMask(file)
	pawnsOnFile := enemyPawns & fileMask

	if pawnsOnFile == 0 {
		return 0
	}

	// Count all pawns on this file that are close to the king
	penalty := 0
	pawnSquares := pawnsOnFile.BitList()

	for _, square := range pawnSquares {
		_, pawnRank := board.SquareToFileRank(square)

		// Check if enemy pawn is advancing toward king
		distanceToKing := abs(pawnRank - kingRank)
		if distanceToKing <= 3 {
			// Pawn is close to king - potential storm threat
			penalty += PawnStormPenalty
		}
	}

	return penalty
}

// evaluateCastlingRights evaluates castling rights and king safety from castling
func evaluateCastlingRights(b *board.Board, color board.BitboardColor) int {
	score := 0
	castlingRights := b.GetCastlingRights()

	// Check if king has already castled
	if hasKingCastled(b, color) {
		score += CastledKingBonus
		return score // If already castled, don't evaluate remaining rights
	}

	// Check castling rights for the specified color
	var kingsideRight, queensideRight string
	if color == board.BitboardWhite {
		kingsideRight = "K"
		queensideRight = "Q"
	} else {
		kingsideRight = "k"
		queensideRight = "q"
	}

	// Count available castling rights
	hasKingside := strings.Contains(castlingRights, kingsideRight)
	hasQueenside := strings.Contains(castlingRights, queensideRight)

	if hasKingside && hasQueenside {
		// Bonus for having both castling rights available
		score += BothSidesCastlingBonus
	} else if hasKingside {
		// Bonus for having kingside castling available
		score += KingsideCastlingBonus
	} else if hasQueenside {
		// Bonus for having queenside castling available
		score += QueensideCastlingBonus
	}

	// Additional bonus if any castling rights remain
	if hasKingside || hasQueenside {
		score += CastlingRightsBonus
	}

	return score
}

// hasKingCastled checks if the king has already completed castling
func hasKingCastled(b *board.Board, color board.BitboardColor) bool {
	var kingBitboard board.Bitboard
	var expectedCastledSquares []int

	if color == board.BitboardWhite {
		kingBitboard = b.GetPieceBitboard(board.WhiteKing)
		// White castled positions: g1 (kingside) = 6, c1 (queenside) = 2
		expectedCastledSquares = []int{6, 2} // g1, c1
	} else {
		kingBitboard = b.GetPieceBitboard(board.BlackKing)
		// Black castled positions: g8 (kingside) = 62, c8 (queenside) = 58
		expectedCastledSquares = []int{62, 58} // g8, c8
	}

	if kingBitboard == 0 {
		return false
	}

	kingSquare := kingBitboard.LSB()
	
	// Check if king is on a castled square
	for _, castledSquare := range expectedCastledSquares {
		if kingSquare == castledSquare {
			return true
		}
	}

	return false
}


