package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Queen evaluation constants - optimized for performance over accuracy
// Focus: Fast evaluation suitable for lazy evaluation with early cutoffs
const (
	// Mobility scoring
	QueenMobilityUnit = 2  // Multiplier for pre-calculated mobility values
	
	// Positional bonuses
	QueenOn7thRank  = 20 // Queen on opponent's 7th rank (strong attacking position)
	QueenOnOpenFile = 10 // Queen on open file (increased mobility and threats)
	
	// Opening penalties
	EarlyQueenMovePenalty = -25 // Penalty for premature queen development
)

// QueenMobilityTable contains pre-calculated mobility scores per square
// Values represent typical queen mobility from each square (accounting for common piece placement)
// Range: 12-20 (corners have lowest mobility, center squares have highest)
// This eliminates expensive attack generation during evaluation
var QueenMobilityTable = [64]int{
	// Rank 1 (back rank) - limited mobility due to own pieces
	12, 14, 14, 14, 14, 14, 14, 12,
	// Rank 2 - slightly better, fewer blocking pieces
	14, 16, 16, 16, 16, 16, 16, 14,
	// Rank 3-6 - progressively better mobility toward center
	14, 16, 18, 18, 18, 18, 16, 14,
	14, 16, 18, 20, 20, 18, 16, 14, // Central ranks peak
	14, 16, 18, 20, 20, 18, 16, 14,
	14, 16, 18, 18, 18, 18, 16, 14,
	// Rank 7 - good mobility for attacking
	14, 16, 16, 16, 16, 16, 16, 14,
	// Rank 8 - limited by opponent pieces
	12, 14, 14, 14, 14, 14, 14, 12,
}

// evaluateQueens performs fast queen evaluation for both sides
// Uses table lookups and simple checks instead of complex calculations
// Suitable for positions evaluated during lazy evaluation cutoffs
func evaluateQueens(b *board.Board) int {
	score := 0

	// Evaluate white queens (usually 1, but could be more after promotion)
	whiteQueens := b.GetPieceBitboard(board.WhiteQueen)
	if whiteQueens != 0 {
		score += evaluateQueensForColor(b, whiteQueens, true)
	}

	// Evaluate black queens (subtract for white's perspective)
	blackQueens := b.GetPieceBitboard(board.BlackQueen)
	if blackQueens != 0 {
		score -= evaluateQueensForColor(b, blackQueens, false)
	}

	return score
}

// evaluateQueensForColor evaluates all queens of a specific color
// Parameters:
//   - b: board position
//   - queensBitboard: bitboard containing all queens of this color
//   - isWhite: true for white queens, false for black queens
func evaluateQueensForColor(b *board.Board, queensBitboard board.Bitboard, isWhite bool) int {
	score := 0

	// Process each queen (typically just one, but handle promotions)
	for queensBitboard != 0 {
		square, remaining := queensBitboard.PopLSB()
		queensBitboard = remaining

		file := square % 8
		rank := square / 8

		// 1. Pre-calculated mobility bonus (O(1) table lookup)
		score += QueenMobilityTable[square] * QueenMobilityUnit

		// 2. 7th rank bonus (queens on opponent's 7th rank are very strong)
		if (isWhite && rank == 6) || (!isWhite && rank == 1) {
			score += QueenOn7thRank
		}

		// 3. Open file bonus (queens on open files have increased scope)
		if isFileOpen(b, file) {
			score += QueenOnOpenFile
		}

		// 4. Early development penalty (only check in opening)
		if b.GetFullMoveNumber() <= 5 {
			score += evaluateEarlyDevelopment(b, rank, isWhite)
		}
	}

	return score
}

// isFileOpen checks if a file has no pawns (O(1) bitboard operation)
func isFileOpen(b *board.Board, file int) bool {
	fileMask := board.FileMask(file)
	whitePawns := b.GetPieceBitboard(board.WhitePawn) & fileMask
	blackPawns := b.GetPieceBitboard(board.BlackPawn) & fileMask
	return whitePawns == 0 && blackPawns == 0
}

// evaluateEarlyDevelopment penalizes premature queen development
// Only applies in the opening when the queen moves before minor pieces are developed
func evaluateEarlyDevelopment(b *board.Board, queenRank int, isWhite bool) int {
	// Quick check: is queen still on starting rank?
	startingRank := 0 // White queen starts on rank 1 (index 0)
	if !isWhite {
		startingRank = 7 // Black queen starts on rank 8 (index 7)
	}

	if queenRank == startingRank {
		return 0 // Queen hasn't moved yet, no penalty
	}

	// Count minor pieces that have moved from starting squares
	developedCount := 0
	
	if isWhite {
		knights := b.GetPieceBitboard(board.WhiteKnight)
		bishops := b.GetPieceBitboard(board.WhiteBishop)
		
		// Check if knights have moved from b1 and g1
		if (knights & board.Bitboard(1<<1)) == 0 { // b1 knight moved
			developedCount++
		}
		if (knights & board.Bitboard(1<<6)) == 0 { // g1 knight moved
			developedCount++
		}
		
		// Check if bishops have moved from c1 and f1
		if (bishops & board.Bitboard(1<<2)) == 0 { // c1 bishop moved
			developedCount++
		}
		if (bishops & board.Bitboard(1<<5)) == 0 { // f1 bishop moved
			developedCount++
		}
	} else {
		knights := b.GetPieceBitboard(board.BlackKnight)
		bishops := b.GetPieceBitboard(board.BlackBishop)
		
		// Check if knights have moved from b8 and g8
		if (knights & board.Bitboard(1<<57)) == 0 { // b8 knight moved
			developedCount++
		}
		if (knights & board.Bitboard(1<<62)) == 0 { // g8 knight moved
			developedCount++
		}
		
		// Check if bishops have moved from c8 and f8
		if (bishops & board.Bitboard(1<<58)) == 0 { // c8 bishop moved
			developedCount++
		}
		if (bishops & board.Bitboard(1<<61)) == 0 { // f8 bishop moved
			developedCount++
		}
	}

	// Penalty if queen moved early with few pieces developed
	if developedCount < 2 {
		return EarlyQueenMovePenalty
	}

	return 0
}