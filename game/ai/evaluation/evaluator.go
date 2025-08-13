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

// PawnHashEntry represents a cached pawn structure evaluation
type PawnHashEntry struct {
	hash  uint64
	score int
}

// Evaluator evaluates positions based on material balance and piece-square tables
type Evaluator struct {}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate returns the evaluation from White's perspective using lazy evaluation with early cutoffs
// Positive = good for White, Negative = good for Black
func (e *Evaluator) Evaluate(b *board.Board) ai.EvaluationScore {
	// Start with cheap material+PST evaluation
	score := 0

	// Phase 1: Material + PST (very fast, always needed)
	score = e.evaluateMaterialAndPST(b)

	// Early return if position is overwhelmingly one-sided
	if abs(score) > 1000 { // Material advantage > Queen
		return ai.EvaluationScore(score)
	}

	// Phase 2: Pawn structure (uses global pawn hash table for caching)
	pawnScore := evaluatePawnStructure(b)
	score += pawnScore

	// Phase 3: Expensive piece evaluations only if needed
	if abs(score) < 500 { // Position is relatively balanced
		score += e.evaluatePieceActivity(b)
	}

	// Always evaluate kings (important for tactical safety)
	score += evaluateKings(b)

	return ai.EvaluationScore(score)
}

// evaluateMaterialAndPST computes just material and piece-square table values
// This is the fastest, most essential evaluation component
func (e *Evaluator) evaluateMaterialAndPST(b *board.Board) int {
	score := 0

	// Scan board for material and PST values only
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				score += PieceValues[piece]
				score += getPositionalBonus(piece, rank, file)
			}
		}
	}

	return score
}

// evaluatePieceActivity computes expensive piece-specific evaluations
// Only called for balanced positions where these features matter
func (e *Evaluator) evaluatePieceActivity(b *board.Board) int {
	score := 0

	score += evaluateKnights(b)
	score += evaluateBishops(b)
	score += evaluateRooks(b)
	score += evaluateQueens(b)

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

// abs returns the absolute value of x
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
