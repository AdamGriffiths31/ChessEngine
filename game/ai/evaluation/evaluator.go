package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

// PieceValues defines the standard piece values in centipawns
var PieceValues = map[board.Piece]int{
	board.WhitePawn:   100,
	board.WhiteKnight: 320,
	board.WhiteBishop: 330,
	board.WhiteRook:   500,
	board.WhiteQueen:  900,
	board.WhiteKing:   0, // King has no material value

	board.BlackPawn:   -100,
	board.BlackKnight: -320,
	board.BlackBishop: -330,
	board.BlackRook:   -500,
	board.BlackQueen:  -900,
	board.BlackKing:   0,
}

// Evaluator evaluates positions based on material balance and piece-square tables
type Evaluator struct{}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate returns the evaluation from White's perspective
// combining material value and positional bonuses
// Positive = good for White, Negative = good for Black
func (e *Evaluator) Evaluate(b *board.Board) ai.EvaluationScore {
	score := 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				score += PieceValues[piece]
				score += getPositionalBonus(piece, rank, file)
			}
		}
	}

	// Add pawn structure evaluation
	score += e.evaluatePawnStructure(b)

	return ai.EvaluationScore(score)
}

// getPositionalBonus returns the positional bonus for a piece at the given position
func getPositionalBonus(piece board.Piece, rank, file int) int {
	switch piece {
	case board.WhiteKnight:
		return KnightTable[rank*8+file]
	case board.BlackKnight:
		flippedRank := 7 - rank
		return -KnightTable[flippedRank*8+file]
	case board.WhiteBishop:
		return BishopTable[rank*8+file]
	case board.BlackBishop:
		flippedRank := 7 - rank
		return -BishopTable[flippedRank*8+file]
	case board.WhiteRook:
		return RookTable[rank*8+file]
	case board.BlackRook:
		flippedRank := 7 - rank
		return -RookTable[flippedRank*8+file]
	case board.WhitePawn:
		return PawnTable[rank*8+file]
	case board.BlackPawn:
		flippedRank := 7 - rank
		return -PawnTable[flippedRank*8+file]
	case board.WhiteQueen:
		return QueenTable[rank*8+file]
	case board.BlackQueen:
		flippedRank := 7 - rank
		return -QueenTable[flippedRank*8+file]
	case board.WhiteKing:
		return KingTable[rank*8+file]
	case board.BlackKing:
		flippedRank := 7 - rank
		return -KingTable[flippedRank*8+file]
	default:
		return 0
	}
}

// evaluatePawnStructure evaluates pawn structure and returns the score from White's perspective
func (e *Evaluator) evaluatePawnStructure(b *board.Board) int {
	score := 0
	
	// Get pawn bitboards
	whitePawns := b.GetPieceBitboard(board.WhitePawn)
	blackPawns := b.GetPieceBitboard(board.BlackPawn)
	
	// Penalize isolated pawns (pawns with no friendly pawns on adjacent files)
	const isolatedPawnPenalty = 20
	score -= countIsolatedPawns(whitePawns) * isolatedPawnPenalty  // White penalty
	score += countIsolatedPawns(blackPawns) * isolatedPawnPenalty  // Black penalty
	
	// Penalize doubled pawns (multiple pawns on same file)
	const doubledPawnPenalty = 15
	score -= countDoubledPawns(whitePawns) * doubledPawnPenalty  // White penalty
	score += countDoubledPawns(blackPawns) * doubledPawnPenalty  // Black penalty
	
	// Bonus for passed pawns (pawns with clear path to promotion)
	score += evaluatePassedPawns(whitePawns, blackPawns, board.BitboardWhite)  // White bonus
	score -= evaluatePassedPawns(blackPawns, whitePawns, board.BitboardBlack)  // Black bonus (subtracted since from white perspective)
	
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

// GetName returns the evaluator name
func (e *Evaluator) GetName() string {
	return "Evaluator"
}
