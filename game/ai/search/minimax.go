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

// MinimaxEngine implements a basic minimax search with opening book support
type MinimaxEngine struct {
	evaluator   ai.Evaluator
	generator   *moves.Generator
	bookService *openings.BookLookupService
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	return &MinimaxEngine{
		evaluator:   evaluation.NewMaterialEvaluator(),
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
				fromSquare := string(rune('a'+bookMove.From.File)) + string(rune('1'+bookMove.From.Rank))
				toSquare := string(rune('a'+bookMove.To.File)) + string(rune('1'+bookMove.To.Rank))
				result.Stats.DebugInfo = append(result.Stats.DebugInfo,
					"✅ Opening book move selected: "+fromSquare+"-"+toSquare)
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

	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - game over
		return result
	}

	bestScore := ai.EvaluationScore(-1000000)
	var bestMove board.Move
	bestMoveFound := false

	// Try each move
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Make the move with undo information
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			panic(err)
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
			bestMoveFound = true
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

	if !bestMoveFound {
		panic("No move found!")
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
			panic("faield to undo nested move")
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
