package search

import (
	"context"
	"fmt"
	"sort"
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
	// MaxKillerDepth defines the maximum depth for killer move storage
	MaxKillerDepth = 64
)

// MinimaxEngine implements a basic minimax search with opening book support
type MinimaxEngine struct {
	evaluator   ai.Evaluator
	generator   *moves.Generator
	bookService *openings.BookLookupService
	// killerTable stores killer moves indexed by [depth][slot] (2 killers per depth)
	killerTable [MaxKillerDepth][2]board.Move
	// previousBestMove stores the PV move from the last completed iteration
	previousBestMove board.Move
	// debugMoveOrdering tracks move ordering for testing (only used in tests)
	debugMoveOrdering bool
	debugMoveOrder    []board.Move
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
	result := ai.SearchResult{
		Stats: ai.SearchStats{},
	}

	// Initialize book service if needed
	if config.UseOpeningBook {
		if err := m.initializeBookService(config); err != nil {
			// Continue without opening book if initialization fails
			if config.DebugMode {
				println("Warning: Failed to initialize opening book:", err.Error())
			}
		}
	}

	if config.UseOpeningBook && m.bookService != nil {
		bookMove, err := m.bookService.FindBookMove(b)
		if err == nil && bookMove != nil {
			result.BestMove = *bookMove
			result.Score = 0
			result.Stats.BookMoveUsed = true
			return result
		}
	}

	startTime := time.Now()

	// Generate all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		result.BestMove = board.Move{From: board.Square{File: -1, Rank: -1}}
		if m.generator.IsKingInCheck(b, player) {
			result.Score = -ai.MateScore
		} else {
			result.Score = ai.DrawScore
		}
		return result
	}

	// Sort moves for better alpha-beta performance
	m.orderMoves(legalMoves, 0)

	// Variables to store the best result from the last completed depth
	lastCompletedBestMove := legalMoves.Moves[0] // Start with first legal move
	lastCompletedScore := ai.EvaluationScore(0)

	// Iterative deepening: search depth 1, 2, 3... until maxDepth or timeout
	for currentDepth := 1; currentDepth <= config.MaxDepth; currentDepth++ {
		// Check for timeout before starting new depth
		select {
		case <-ctx.Done():
			// Return the best move from the last completed depth
			result.BestMove = lastCompletedBestMove
			result.Score = lastCompletedScore
			result.Stats.Time = time.Since(startTime)
			return result
		default:
		}

		// Search at current depth
		depthResult := m.searchAtDepth(ctx, b, player, currentDepth, legalMoves, config)

		// Check if the search was completed or interrupted by timeout
		if depthResult.completed {
			// This depth completed successfully - save the results
			lastCompletedBestMove = depthResult.bestMove
			lastCompletedScore = depthResult.bestScore
			result.Stats.Depth = currentDepth
			result.Stats.NodesSearched = depthResult.nodesSearched

			// Store PV move for next iteration (only if depth > 1 to ensure it's validated)
			if currentDepth > 1 {
				m.previousBestMove = depthResult.bestMove
			}

			// Copy debug info from the last completed depth
			if config.DebugMode {
				result.Stats.DebugInfo = depthResult.debugInfo
			}
		} else {
			// This depth was interrupted by timeout - return last completed results
			result.BestMove = lastCompletedBestMove
			result.Score = lastCompletedScore
			result.Stats.Time = time.Since(startTime)
			return result
		}
	}

	// All depths completed successfully
	result.BestMove = lastCompletedBestMove
	result.Score = lastCompletedScore
	result.Stats.Time = time.Since(startTime)
	return result
}

// depthSearchResult holds results from searching at a specific depth
type depthSearchResult struct {
	bestMove      board.Move
	bestScore     ai.EvaluationScore
	completed     bool
	nodesSearched int64
	debugInfo     []string
}

// searchAtDepth searches at a specific depth and returns whether it completed
func (m *MinimaxEngine) searchAtDepth(ctx context.Context, b *board.Board, player moves.Player, depth int, legalMoves *moves.MoveList, config ai.SearchConfig) depthSearchResult {
	bestScore := ai.EvaluationScore(-ai.MateScore - 1)
	bestMove := legalMoves.Moves[0] // Default to first move
	var debugInfo []string
	var nodesSearched int64

	// Initial alpha-beta bounds
	alpha := ai.EvaluationScore(-ai.MateScore - 1)
	beta := ai.EvaluationScore(ai.MateScore + 1)

	// Try each move at this depth
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		// Check for timeout before each move
		select {
		case <-ctx.Done():
			// Timeout occurred - return incomplete result
			return depthSearchResult{
				bestMove:      bestMove,
				bestScore:     bestScore,
				completed:     false,
				nodesSearched: nodesSearched,
				debugInfo:     debugInfo,
			}
		default:
		}

		// Make the move
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Create a stats tracker for this search
		var moveStats ai.SearchStats

		// Search with negamax
		var score ai.EvaluationScore
		if depth > 0 {
			// Use negamax with negated bounds
			score = -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player),
				depth-1, -beta, -alpha, depth, &moveStats)
		} else {
			// Depth 0 - just evaluate
			eval := m.evaluator.Evaluate(b)
			if player == moves.Black {
				eval = -eval
			}
			score = eval
			moveStats.NodesSearched = 1
		}

		// Unmake the move
		b.UnmakeMove(undo)

		// Add nodes from this move's search
		nodesSearched += moveStats.NodesSearched

		// Check for timeout after search
		select {
		case <-ctx.Done():
			// Timeout occurred during search - return incomplete result
			return depthSearchResult{
				bestMove:      bestMove,
				bestScore:     bestScore,
				completed:     false,
				nodesSearched: nodesSearched,
				debugInfo:     debugInfo,
			}
		default:
		}

		// Update best move if this is better
		if score > bestScore {
			bestScore = score
			bestMove = move

			// Update alpha for next iteration
			if score > alpha {
				alpha = score
			}
		}

		// Add debug info if enabled
		if config.DebugMode && i < 10 {
			debugInfo = append(debugInfo, fmt.Sprintf("Move %s%s: score=%d",
				move.From.String(), move.To.String(), score))
		}
	}

	// All moves at this depth completed successfully
	return depthSearchResult{
		bestMove:      bestMove,
		bestScore:     bestScore,
		completed:     true,
		nodesSearched: nodesSearched,
		debugInfo:     debugInfo,
	}
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

	// Check for terminal positions (checkmate or stalemate)
	if legalMoves.Count == 0 {
		if inCheck {
			// Checkmate - shorter mates should have worse scores
			return -ai.MateScore - ai.EvaluationScore(1)
		} else {
			// Stalemate - return draw score
			return ai.DrawScore
		}
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
			// Shorter mates should have worse scores (closer to -MateScore)
			return -ai.MateScore - ai.EvaluationScore(currentDepth)
		}
		return ai.DrawScore // Stalemate
	}

	// Sort moves to improve alpha-beta efficiency (captures first)
	m.orderMoves(legalMoves, currentDepth)

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
			// Store killer move if it's not a capture
			if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
				m.storeKiller(move, currentDepth)
			}
			return beta
		}
	}

	return alpha
}

// getMVVLVAScore calculates the MVV-LVA score for a capture move
// Higher scores indicate more valuable captures (better moves to try first)
func (m *MinimaxEngine) getMVVLVAScore(move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0 // Non-captures get score 0
	}

	// Get absolute piece values (remove sign for black pieces)
	victimValue := evaluation.PieceValues[move.Captured]
	if victimValue < 0 {
		victimValue = -victimValue
	}

	attackerValue := evaluation.PieceValues[move.Piece]
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	// MVV-LVA: Most valuable victim first, then least valuable attacker
	// Multiply victim by 10 to prioritize victim value over attacker value
	return victimValue*10 - attackerValue
}

// orderMoves sorts moves to improve alpha-beta efficiency using PV, MVV-LVA and killer moves
func (m *MinimaxEngine) orderMoves(moveList *moves.MoveList, depth int) {
	// Create slice of move indices with their scores
	type moveScore struct {
		index int
		score int
	}

	scores := make([]moveScore, moveList.Count)
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		var score int
		if m.isPVMove(move) {
			// PV move gets highest priority
			score = 2000000
			// Debug: log when PV move is detected (only in debug mode)
		} else if move.IsCapture {
			// Use MVV-LVA score for captures (second highest priority)
			score = 1000000 + m.getMVVLVAScore(move)
		} else if m.isKillerMove(move, depth) {
			// Killer moves get high priority (between captures and other moves)
			// Differentiate between first and second killer
			if depth >= 0 && depth < MaxKillerDepth &&
				move.From == m.killerTable[depth][0].From && move.To == m.killerTable[depth][0].To {
				score = 500000 // First killer
			} else {
				score = 490000 // Second killer
			}
		} else {
			// Non-captures get lower priority
			score = 0
		}
		scores[i] = moveScore{index: i, score: score}
	}

	// Sort by score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Reorder moves based on sorted scores
	originalMoves := make([]board.Move, moveList.Count)
	copy(originalMoves, moveList.Moves[:moveList.Count])

	for i := 0; i < moveList.Count; i++ {
		moveList.Moves[i] = originalMoves[scores[i].index]
	}

	// Debug: capture move ordering if enabled (only for the root level, depth 0)
	if m.debugMoveOrdering && depth == 0 {
		m.debugMoveOrder = make([]board.Move, moveList.Count)
		copy(m.debugMoveOrder, moveList.Moves[:moveList.Count])
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

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	// Clear PV move from previous search
	m.previousBestMove = board.Move{}
	// Clear killer moves
	for i := 0; i < MaxKillerDepth; i++ {
		m.killerTable[i][0] = board.Move{}
		m.killerTable[i][1] = board.Move{}
	}
}

// SetDebugMoveOrdering enables/disables move ordering debug tracking (for tests only)
func (m *MinimaxEngine) SetDebugMoveOrdering(enabled bool) {
	m.debugMoveOrdering = enabled
	if !enabled {
		m.debugMoveOrder = nil
	}
}

// GetLastMoveOrder returns the move order from the last orderMoves call (for tests only)
func (m *MinimaxEngine) GetLastMoveOrder() []board.Move {
	return m.debugMoveOrder
}

// isKillerMove checks if a move is a killer move at the given depth
func (m *MinimaxEngine) isKillerMove(move board.Move, depth int) bool {
	if depth < 0 || depth >= MaxKillerDepth {
		return false
	}

	// Check both killer slots at this depth
	return (move.From == m.killerTable[depth][0].From && move.To == m.killerTable[depth][0].To) ||
		(move.From == m.killerTable[depth][1].From && move.To == m.killerTable[depth][1].To)
}

// storeKiller stores a move as a killer move, shifting existing killers
func (m *MinimaxEngine) storeKiller(move board.Move, depth int) {
	if depth < 0 || depth >= MaxKillerDepth {
		return
	}

	// Don't store the same killer move twice
	if m.isKillerMove(move, depth) {
		return
	}

	// Shift existing killers and store new one in first slot
	m.killerTable[depth][1] = m.killerTable[depth][0]
	m.killerTable[depth][0] = move
}

// isPVMove checks if a move is the principal variation move from the previous iteration
func (m *MinimaxEngine) isPVMove(move board.Move) bool {
	// Check for invalid move markers (used to indicate no valid PV move)
	if m.previousBestMove.From.File == -1 || m.previousBestMove.From.Rank == -1 {
		return false // Invalid PV move stored
	}

	// Check if this move matches the stored PV move (From and To squares)
	return move.From == m.previousBestMove.From && move.To == m.previousBestMove.To
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}
