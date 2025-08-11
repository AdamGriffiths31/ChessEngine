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

	// Optimized: Use bitboard iteration instead of double board scanning
	// This eliminates the expensive nested loops and GetPiece() calls
	
	// White pieces (positive contribution)
	score += e.evaluatePiecesBitboard(b, board.WhitePawn, PawnTable, false)
	score += e.evaluatePiecesBitboard(b, board.WhiteKnight, KnightTable, false)
	score += e.evaluatePiecesBitboard(b, board.WhiteBishop, BishopTable, false)
	score += e.evaluatePiecesBitboard(b, board.WhiteRook, RookTable, false)
	score += e.evaluatePiecesBitboard(b, board.WhiteQueen, QueenTable, false)
	score += e.evaluatePiecesBitboard(b, board.WhiteKing, KingTable, false)
	
	// Black pieces (negative contribution)
	score -= e.evaluatePiecesBitboard(b, board.BlackPawn, PawnTable, true)
	score -= e.evaluatePiecesBitboard(b, board.BlackKnight, KnightTable, true)
	score -= e.evaluatePiecesBitboard(b, board.BlackBishop, BishopTable, true)
	score -= e.evaluatePiecesBitboard(b, board.BlackRook, RookTable, true)
	score -= e.evaluatePiecesBitboard(b, board.BlackQueen, QueenTable, true)
	score -= e.evaluatePiecesBitboard(b, board.BlackKing, KingTable, true)

	// Add pawn-specific evaluation
	score += evaluatePawnStructure(b)

	// Add knight-specific evaluation
	score += evaluateKnights(b)

	// Add bishop-specific evaluation
	score += evaluateBishops(b)

	// Add rook-specific evaluation
	score += evaluateRooks(b)

	// Add queen-specific evaluation
	score += evaluateQueens(b)

	// Add king-specific evaluation
	score += evaluateKings(b)

	return ai.EvaluationScore(score)
}

// evaluatePiecesBitboard efficiently evaluates material and positional value for a piece type
// using bitboard iteration instead of scanning the entire board
func (e *Evaluator) evaluatePiecesBitboard(b *board.Board, pieceType board.Piece, psTable [64]int, isBlack bool) int {
	pieces := b.GetPieceBitboard(pieceType)
	if pieces == 0 {
		return 0 // Early exit if no pieces of this type
	}
	
	score := 0
	materialValue := PieceValues[pieceType]
	
	// Iterate over all pieces of this type using bitboard iteration
	for pieces != 0 {
		square, newPieces := pieces.PopLSB()
		pieces = newPieces
		
		// Add material value
		score += materialValue
		
		// Add positional bonus from piece-square table
		if isBlack {
			// Flip rank for black pieces (same logic as getPositionalBonus)
			rank := square / 8
			file := square % 8
			flippedRank := 7 - rank
			flippedSquare := flippedRank*8 + file
			score += psTable[flippedSquare]
		} else {
			score += psTable[square]
		}
	}
	
	return score
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

// GetName returns the evaluator name
func (e *Evaluator) GetName() string {
	return "Evaluator"
}
