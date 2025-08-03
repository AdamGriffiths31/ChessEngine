package search

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

const (
	// MinEval represents the worst possible evaluation score
	MinEval = ai.EvaluationScore(-1000000)
)

// MinimaxEngine implements a basic minimax search with opening book support
type MinimaxEngine struct {
	evaluator   ai.Evaluator
	generator   *moves.Generator
	bookService *openings.BookLookupService
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	return &MinimaxEngine{
		evaluator:   evaluation.NewEvaluator(),
		generator:   moves.NewGenerator(),
		bookService: nil, // Will be initialized when needed based on config
	}
}

// initializeBookService initializes the opening book service based on configuration
func (m *MinimaxEngine) initializeBookService(config ai.SearchConfig) error {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		m.bookService = nil
		return nil
	}

	// Convert AI selection mode to openings selection mode
	var selectionMode openings.SelectionMode
	switch config.BookSelectMode {
	case ai.BookSelectBest:
		selectionMode = openings.SelectBest
	case ai.BookSelectRandom:
		selectionMode = openings.SelectRandom
	case ai.BookSelectWeightedRandom:
		selectionMode = openings.SelectWeightedRandom
	default:
		selectionMode = openings.SelectWeightedRandom
	}

	bookConfig := openings.BookConfig{
		Enabled:         true,
		BookFiles:       config.BookFiles,
		SelectionMode:   selectionMode,
		WeightThreshold: config.BookWeightThreshold,
	}

	service := openings.NewBookLookupService(bookConfig)
	err := service.LoadBooks()
	if err != nil {
		return err
	}

	m.bookService = service

	// Debug: Log successful book loading
	if config.DebugMode {
		loadedBooks := service.GetLoadedBooks()
		for _, info := range loadedBooks {
			println("📚 Loaded opening book:", info.Filename, "with", info.EntryCount, "entries")
		}
	}

	return nil
}

// FindBestMove searches for the best move using minimax with optional opening book
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	startTime := time.Now()
	result := ai.SearchResult{
		Stats: ai.SearchStats{
			DebugInfo: make([]string, 0),
		},
	}

	// Initialize book service if needed
	if err := m.initializeBookService(config); err != nil {
		// Log error but continue with regular search
		if config.DebugMode {
			result.Stats.DebugInfo = append(result.Stats.DebugInfo,
				"Opening book initialization failed: "+err.Error())
		}
	}

	// Try opening book first
	if m.bookService != nil && m.bookService.IsEnabled() {
		// Debug: Show position hash
		if config.DebugMode {
			hash := openings.GetPolyglotHash().HashPosition(b)
			result.Stats.DebugInfo = append(result.Stats.DebugInfo,
				fmt.Sprintf("🔍 Position hash: %016X", hash))
		}

		bookMove, err := m.bookService.FindBookMove(b)
		if err == nil && bookMove != nil {
			// Found a book move - return it immediately
			result.BestMove = *bookMove
			result.Score = 0 // Book moves don't have evaluation scores
			result.Stats.Time = time.Since(startTime)
			result.Stats.Depth = 0         // Book lookup doesn't count as search depth
			result.Stats.NodesSearched = 0 // No nodes searched for book moves
			result.Stats.BookMoveUsed = true

			if config.DebugMode {
				moveStr := bookMove.From.String() + bookMove.To.String()
				result.Stats.DebugInfo = append(result.Stats.DebugInfo,
					"✅ Opening book move selected: "+moveStr)
			}
			return result
		}

		// Book lookup failed or found no moves
		if config.DebugMode {
			if err != nil {
				result.Stats.DebugInfo = append(result.Stats.DebugInfo,
					"❌ Opening book lookup failed: "+err.Error()+", falling back to search")
			} else {
				result.Stats.DebugInfo = append(result.Stats.DebugInfo,
					"ℹ️  No moves found in opening book, falling back to search")
			}
		}
	}

	// Always use iterative deepening - it provides the same deterministic results
	// for tests (no timeout) while handling time constraints gracefully in real games
	return m.findBestMoveIterative(ctx, b, player, config, startTime, result)
}

// findBestMoveIterative implements iterative deepening search
// This is now the only search algorithm - works for both timed and untimed searches
func (m *MinimaxEngine) findBestMoveIterative(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig, startTime time.Time, result ai.SearchResult) ai.SearchResult {
	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - game over
		result.Stats.Time = time.Since(startTime)
		return result
	}

	var bestMove board.Move
	bestScore := MinEval
	bestMoveFound := false

	// Iterative deepening: search 1, 2, 3, ..., MaxDepth
	for currentDepth := 1; currentDepth <= config.MaxDepth; currentDepth++ {
		// Check for timeout before starting new depth
		select {
		case <-ctx.Done():
			if !bestMoveFound {
				bestMove = legalMoves.Moves[0]
				bestScore = MinEval
			}
			result.BestMove = bestMove
			result.Score = bestScore
			result.Stats.Time = time.Since(startTime)
			return result
		default:
		}

		// Search all moves at current depth with full alpha-beta window
		currentBestScore := MinEval
		var currentBestMove board.Move
		currentBestFound := false

		for i := 0; i < legalMoves.Count; i++ {
			move := legalMoves.Moves[i]

			// Make the move
			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				continue
			}

			// Search at current depth with full window
			// Note: We use -negamax, so we negate alpha and beta
			score := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player), currentDepth-1, MinEval, -currentBestScore, currentDepth, &result.Stats)

			// Unmake the move
			b.UnmakeMove(undo)

			// Update best move for this depth
			if score > currentBestScore {
				currentBestScore = score
				currentBestMove = move
				currentBestFound = true
			}
		}

		// Update overall best move if we found one at this depth
		if currentBestFound {
			bestScore = currentBestScore
			bestMove = currentBestMove
			bestMoveFound = true
		}

		// DEBUG: Log iterative deepening progress
		if config.DebugMode && bestMoveFound {
			result.Stats.DebugInfo = append(result.Stats.DebugInfo,
				fmt.Sprintf("Depth %d: best=%s score=%d",
					currentDepth, bestMove.From.String()+bestMove.To.String(), bestScore))
		}
	}

	if !bestMoveFound {
		// Fallback to first legal move
		bestMove = legalMoves.Moves[0]
		bestScore = 0
	}

	result.BestMove = bestMove
	result.Score = bestScore
	result.Stats.Time = time.Since(startTime)
	return result
}

func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, stats *ai.SearchStats) ai.EvaluationScore {
	stats.NodesSearched++

	// Check for cancellation
	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	// Get in-check status
	inCheck := m.generator.IsKingInCheck(b, player)

	// Stand-pat evaluation - only if not in check
	if !inCheck {
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}

		// Beta cutoff
		if eval >= beta {
			return beta
		}

		// Update alpha
		if eval > alpha {
			alpha = eval
		}
	}

	// Generate all moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	// If in check and no moves, it's checkmate
	if inCheck && legalMoves.Count == 0 {
		return -ai.MateScore + ai.EvaluationScore(1)
	}

	// Try captures (and all moves if in check)
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Only search captures in quiescence (unless in check)
		if !inCheck && !move.IsCapture {
			continue
		}

		// Make the move
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Recursive quiescence search
		score := -m.quiescence(ctx, b, oppositePlayer(player), -beta, -alpha, stats)

		// Unmake the move
		b.UnmakeMove(undo)

		// Update alpha
		if score > alpha {
			alpha = score

			// Beta cutoff
			if alpha >= beta {
				return beta
			}
		}
	}

	return alpha
}

func (m *MinimaxEngine) negamaxWithAlphaBeta(ctx context.Context, b *board.Board, player moves.Player, depth int, alpha, beta ai.EvaluationScore, originalMaxDepth int, stats *ai.SearchStats) ai.EvaluationScore {
	stats.NodesSearched++

	// Track the maximum depth reached
	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	// Terminal node - call quiescence search with proper bounds
	if depth == 0 {
		// Enter quiescence search to resolve captures
		return m.quiescence(ctx, b, player, alpha, beta, stats)
	}

	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - check for checkmate or stalemate
		if m.generator.IsKingInCheck(b, player) {
			// Checkmate - return worst score for current player
			return -ai.MateScore + ai.EvaluationScore(currentDepth)
		}
		return ai.DrawScore // Stalemate
	}

	// Sort moves to improve alpha-beta efficiency (captures first)
	m.orderMoves(legalMoves)

	// Try each move
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Make the move with undo information
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Search deeper with negamax and alpha-beta
		score := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player), depth-1, -beta, -alpha, originalMaxDepth, stats)

		// Unmake the move
		b.UnmakeMove(undo)

		// Update alpha (best score for current player)
		if score > alpha {
			alpha = score
		}

		// Beta cutoff - opponent won't allow this line
		if alpha >= beta {
			return beta
		}
	}

	return alpha
}

// orderMoves sorts moves to improve alpha-beta efficiency
func (m *MinimaxEngine) orderMoves(moveList *moves.MoveList) {
	// Simple move ordering: captures first, then quiet moves
	// Better ordering would use MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
	captureCount := 0
	
	// Partition captures to the front
	for i := 0; i < moveList.Count; i++ {
		if moveList.Moves[i].IsCapture {
			if i != captureCount {
				// Swap capture to front
				moveList.Moves[i], moveList.Moves[captureCount] = moveList.Moves[captureCount], moveList.Moves[i]
			}
			captureCount++
		}
	}
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
