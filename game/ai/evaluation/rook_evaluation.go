package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Rook evaluation - optimized for speed with focus on what matters most
// 
// Design Philosophy:
// 1. Open files are 90% of rook evaluation (prioritize this)
// 2. 7th rank penetration is the second most important factor
// 3. Everything else (mobility, connections) provides marginal value
// 4. Use pre-computed tables and simple checks to avoid expensive calculations
//
// This implementation trades some evaluation accuracy for significant speed gains,
// making it suitable for positions evaluated during lazy evaluation cutoffs.

const (
	// File control bonuses (the primary value of rooks)
	RookOpenFileBonus     = 20 // Rook on completely open file  
	RookSemiOpenFileBonus = 10 // Rook on semi-open file (no friendly pawns)
	
	// Rank penetration bonus
	RookOn7thBonus = 25 // Rook on opponent's 7th rank (attacking position)
	
	// Coordination bonus  
	RooksConnectedBonus = 8 // Rooks protecting each other on same rank/file
	
	// Mobility approximation
	RookMobilityUnit = 2 // Multiplier for pre-calculated mobility values
)

// RookMobilityByRank contains approximate mobility values for each rank
// Values represent typical number of moves available to a rook on that rank
// Back ranks (0,7) have lower mobility due to own pieces, middle ranks are optimal
var RookMobilityByRank = [8]int{
	12, 14, 14, 14, 14, 14, 14, 12, // Ranks 1-8 mobility approximation
}

// evaluateRooks performs fast rook evaluation for both sides
// Returns positive values favoring White, negative favoring Black
func evaluateRooks(b *board.Board) int {
	whiteScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.WhiteRook), true)
	blackScore := evaluateRooksForColor(b, b.GetPieceBitboard(board.BlackRook), false)
	
	return whiteScore - blackScore
}

// evaluateRooksForColor evaluates all rooks of a specific color
// Uses simplified logic focusing on the most important rook features
//
// Parameters:
//   - b: board position
//   - rooks: bitboard containing all rooks of this color
//   - forWhite: true if evaluating white rooks, false for black rooks
func evaluateRooksForColor(b *board.Board, rooks board.Bitboard, forWhite bool) int {
	if rooks == 0 {
		return 0
	}
	
	score := 0
	rookCount := rooks.PopCount()
	
	// Cache pawn bitboards for file evaluation (avoids repeated lookups)
	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)
	
	// Track first rook position for connection analysis
	firstRookSquare := -1
	
	for rooks != 0 {
		square, remaining := rooks.PopLSB()
		rooks = remaining
		
		file := square % 8
		rank := square / 8
		
		// 1. File control evaluation (most important)
		score += evaluateFileControl(file, whitePawns, blackPawns, forWhite)
		
		// 2. Rank penetration bonus (second most important)  
		score += evaluateRankPenetration(rank, forWhite)
		
		// 3. Mobility approximation (minor factor)
		score += RookMobilityByRank[rank] * RookMobilityUnit
		
		// 4. Connection bonus (only for exactly 2 rooks)
		if firstRookSquare != -1 && rookCount == 2 {
			score += evaluateRookConnection(square, firstRookSquare)
		}
		
		firstRookSquare = square
	}
	
	return score
}

// evaluateFileControl determines file control bonus for a rook
// Open files (no pawns) are most valuable, semi-open files (no friendly pawns) are good
func evaluateFileControl(file int, whitePawns, blackPawns board.Bitboard, forWhite bool) int {
	fileMask := board.FileMask(file)
	hasWhitePawns := (whitePawns & fileMask) != 0
	hasBlackPawns := (blackPawns & fileMask) != 0
	
	// Completely open file (no pawns from either side)
	if !hasWhitePawns && !hasBlackPawns {
		return RookOpenFileBonus
	}
	
	// Semi-open file (no friendly pawns blocking the rook's advance)
	if forWhite && !hasWhitePawns {
		return RookSemiOpenFileBonus
	}
	if !forWhite && !hasBlackPawns {
		return RookSemiOpenFileBonus  
	}
	
	// Closed file (blocked by friendly pawns)
	return 0
}

// evaluateRankPenetration awards bonus for rooks on opponent's 7th rank
// 7th rank rooks are extremely strong in chess as they attack enemy pawns/king
func evaluateRankPenetration(rank int, forWhite bool) int {
	// White rooks on 7th rank (index 6), Black rooks on 2nd rank (index 1)
	if (forWhite && rank == 6) || (!forWhite && rank == 1) {
		return RookOn7thBonus
	}
	return 0
}

// evaluateRookConnection awards bonus if rooks are connected (protecting each other)
// Connected rooks are on the same rank or file with no pieces between them
func evaluateRookConnection(currentSquare, previousSquare int) int {
	currentFile := currentSquare % 8
	currentRank := currentSquare / 8
	previousFile := previousSquare % 8 
	previousRank := previousSquare / 8
	
	// Rooks are connected if they're on the same rank or file
	if currentRank == previousRank || currentFile == previousFile {
		return RooksConnectedBonus
	}
	
	return 0
}