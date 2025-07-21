package search

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// MinimaxEngine implements a basic minimax search
type MinimaxEngine struct {
	evaluator ai.Evaluator
	generator *moves.Generator
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	return &MinimaxEngine{
		evaluator: evaluation.NewMaterialEvaluator(),
		generator: moves.NewGenerator(),
	}
}

// FindBestMove searches for the best move using minimax
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	startTime := time.Now()
	result := ai.SearchResult{
		Stats: ai.SearchStats{},
	}

	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - game over
		return result
	}

	bestScore := ai.EvaluationScore(-1000000)
	var bestMove board.Move

	// Try each move
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Make the move with undo information
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Search deeper
		score := -m.minimaxWithDepthTracking(ctx, b, oppositePlayer(player), config.MaxDepth-1, config.MaxDepth, &result.Stats)

		// Unmake the move
		b.UnmakeMove(undo)

		// Update best move if this is better
		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			result.BestMove = bestMove
			result.Score = bestScore
			result.Stats.Time = time.Since(startTime)
			return result
		default:
		}
	}

	result.BestMove = bestMove
	result.Score = bestScore
	result.Stats.Time = time.Since(startTime)
	// Depth will be set by the minimax function

	return result
}

// minimaxWithDepthTracking is the recursive minimax search with proper depth tracking
func (m *MinimaxEngine) minimaxWithDepthTracking(ctx context.Context, b *board.Board, player moves.Player, depth int, originalMaxDepth int, stats *ai.SearchStats) ai.EvaluationScore {
	stats.NodesSearched++
	
	// Track the maximum depth reached
	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return 0
	default:
	}

	// Terminal node - evaluate position
	if depth == 0 {
		return m.evaluator.Evaluate(b, player)
	}

	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - check for checkmate or stalemate
		if m.generator.IsKingInCheck(b, player) {
			return -ai.MateScore + ai.EvaluationScore(depth) // Checkmate
		}
		return ai.DrawScore // Stalemate
	}

	bestScore := ai.EvaluationScore(-1000000)

	// Try each move
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Make the move with undo information
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Search deeper
		score := -m.minimaxWithDepthTracking(ctx, b, oppositePlayer(player), depth-1, originalMaxDepth, stats)

		// Unmake the move
		b.UnmakeMove(undo)

		// Update best score
		if score > bestScore {
			bestScore = score
		}
	}

	return bestScore
}

// SetEvaluator sets the position evaluator
func (m *MinimaxEngine) SetEvaluator(eval ai.Evaluator) {
	m.evaluator = eval
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}

