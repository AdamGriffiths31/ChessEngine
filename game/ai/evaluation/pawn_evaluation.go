package evaluation

import "github.com/AdamGriffiths31/ChessEngine/board"

// evaluatePawnStructure evaluates pawn structure and returns the score from White's perspective
func evaluatePawnStructure(b *board.Board) int {
	score := 0

	// Get pawn bitboards
	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)

	// Penalize isolated pawns (pawns with no friendly pawns on adjacent files)
	const isolatedPawnPenalty = 20
	score -= countIsolatedPawns(whitePawns) * isolatedPawnPenalty // White penalty
	score += countIsolatedPawns(blackPawns) * isolatedPawnPenalty // Black penalty

	// Penalize doubled pawns (multiple pawns on same file)
	const doubledPawnPenalty = 15
	score -= countDoubledPawns(whitePawns) * doubledPawnPenalty // White penalty
	score += countDoubledPawns(blackPawns) * doubledPawnPenalty // Black penalty

	// Bonus for passed pawns (pawns with clear path to promotion)
	score += evaluatePassedPawns(whitePawns, blackPawns, board.BitboardWhite) // White bonus
	score -= evaluatePassedPawns(blackPawns, whitePawns, board.BitboardBlack) // Black bonus (subtracted since from white perspective)

	return score
}

// countIsolatedPawns counts pawns that have no friendly pawns on adjacent files
func countIsolatedPawns(pawns board.Bitboard) int {
	isolated := 0
	for file := 0; file < 8; file++ {
		fileMask := board.FileMask(file)
		if pawns&fileMask != 0 {
			// Check adjacent files for supporting pawns
			leftFile := fileMask.ShiftWest()
			rightFile := fileMask.ShiftEast()
			if pawns&(leftFile|rightFile) == 0 {
				isolated += (pawns & fileMask).PopCount()
			}
		}
	}
	return isolated
}

// countDoubledPawns counts pawns that have multiple pawns on the same file
func countDoubledPawns(pawns board.Bitboard) int {
	doubled := 0
	for file := 0; file < 8; file++ {
		fileMask := board.FileMask(file)
		pawnsOnFile := (pawns & fileMask).PopCount()
		if pawnsOnFile > 1 {
			doubled += pawnsOnFile - 1 // Only count the extra pawns as penalties
		}
	}
	return doubled
}

// evaluatePassedPawns calculates bonus points for passed pawns
func evaluatePassedPawns(friendlyPawns, enemyPawns board.Bitboard, color board.BitboardColor) int {
	passedPawns := board.GetPassedPawns(friendlyPawns, enemyPawns, color)
	if passedPawns == 0 {
		return 0
	}

	score := 0
	pawnList := passedPawns.BitList()

	for _, square := range pawnList {
		rank := square / 8

		// Calculate bonus based on how advanced the passed pawn is
		var advancement int
		if color == board.BitboardWhite {
			advancement = rank // For white, higher rank = more advanced
		} else {
			advancement = 7 - rank // For black, lower rank = more advanced
		}

		// Progressive bonus: more advanced pawns get exponentially higher bonuses
		// Base bonus of 20, with increasing rewards for advancement
		const basePassedPawnBonus = 20
		bonus := basePassedPawnBonus + (advancement * advancement * 5)
		score += bonus
	}

	return score
}