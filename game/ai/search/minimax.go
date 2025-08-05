package search

import (
	"context"
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
	// debugMoveOrdering tracks move ordering for testing (only used in tests)
	debugMoveOrdering bool
	debugMoveOrder    []board.Move
	// transposition table for caching search results
	transpositionTable *TranspositionTable
	// zobrist hash for position keys
	zobrist *openings.ZobristHash
	// useTranspositions controls whether transposition table is used
	useTranspositions bool
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	return &MinimaxEngine{
		evaluator:          evaluation.NewEvaluator(),
		generator:          moves.NewGenerator(),
		bookService:        nil, // Will be initialized when needed based on config
		transpositionTable: nil, // Will be initialized when configured
		zobrist:            openings.GetPolyglotHash(),
		useTranspositions:  false, // Disabled by default
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

	// Increment transposition table age at start of new search
	if m.useTranspositions && m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

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

	// Check transposition table for best move to help with initial move ordering
	var rootTTMove board.Move
	if m.useTranspositions && m.transpositionTable != nil {
		hash := m.zobrist.HashPosition(b)
		if entry, found := m.transpositionTable.Probe(hash); found {
			rootTTMove = entry.BestMove
			// Debug logging for tests
			if m.debugMoveOrdering {
				// This will be visible in test logs when debug is enabled
			}
		}
	}

	// Sort moves for better alpha-beta performance
	m.orderMoves(legalMoves, 0, rootTTMove)

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

		// Search at current depth using root-level alpha-beta
		bestScore := ai.EvaluationScore(-ai.MateScore - 1)
		bestMove := legalMoves.Moves[0]
		alpha := ai.EvaluationScore(-ai.MateScore - 1)
		beta := ai.EvaluationScore(ai.MateScore + 1)
		var searchStats ai.SearchStats
		completed := false

		// Try each move at this depth
		for i := 0; i < legalMoves.Count; i++ {
			move := legalMoves.Moves[i]

			// Check for timeout before each move
			select {
			case <-ctx.Done():
				completed = false
				goto searchComplete
			default:
			}

			// Make the move
			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				continue
			}

			// Search with negamax
			var moveStats ai.SearchStats
			score := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, &moveStats)

			// Unmake the move
			b.UnmakeMove(undo)

			// Add to total stats
			searchStats.NodesSearched += moveStats.NodesSearched

			// Check for timeout after search
			select {
			case <-ctx.Done():
				completed = false
				goto searchComplete
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
		}
		completed = true

	searchComplete:
		if completed {
			// This depth completed successfully - save the results
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			result.Stats.Depth = currentDepth
			result.Stats.NodesSearched = searchStats.NodesSearched

			// Store root position result in transposition table
			if m.useTranspositions && m.transpositionTable != nil {
				hash := m.zobrist.HashPosition(b)
				m.transpositionTable.Store(hash, currentDepth, bestScore, EntryExact, bestMove)
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

	// Transposition table lookup for quiescence
	hash := m.zobrist.HashPosition(b)
	if m.useTranspositions && m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			// For quiescence, we're less strict about depth requirements
			if entry.Depth >= 0 { // Accept any non-negative depth
				switch entry.Type {
				case EntryExact:
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						return entry.Score
					}
					if entry.Score > alpha {
						alpha = entry.Score
					}
				case EntryUpperBound:
					if entry.Score <= alpha {
						return entry.Score
					}
					if entry.Score < beta {
						beta = entry.Score
					}
				}
			}
		}
	}

	// Get in-check status
	inCheck := m.generator.IsKingInCheck(b, player)

	// Store original alpha for TT entry type determination
	originalAlpha := alpha

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
			// Checkmate - current player loses, return negative mate score
			// In quiescence, we're at depth 0, so mate distance is 1
			return -ai.MateScore + ai.EvaluationScore(1)
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

	// Store result in transposition table
	if m.useTranspositions && m.transpositionTable != nil {
		entryType := EntryExact
		if alpha <= originalAlpha {
			entryType = EntryUpperBound
		} else if alpha >= beta {
			entryType = EntryLowerBound
		}
		m.transpositionTable.Store(hash, 0, alpha, entryType, board.Move{})
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

	// Transposition table lookup
	var ttMove board.Move
	hash := m.zobrist.HashPosition(b)

	if m.useTranspositions && m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			// Use TT move for move ordering even if we can't use the score
			ttMove = entry.BestMove

			// Check if we can use the stored score
			if entry.Depth >= depth {
				switch entry.Type {
				case EntryExact:
					// Exact score - we can return immediately
					return entry.Score
				case EntryLowerBound:
					// Fail high (beta cutoff was found)
					if entry.Score >= beta {
						return entry.Score
					}
					if entry.Score > alpha {
						alpha = entry.Score
					}
				case EntryUpperBound:
					// Fail low
					if entry.Score <= alpha {
						return entry.Score
					}
					if entry.Score < beta {
						beta = entry.Score
					}
				}
			}
		}
	}

	// Terminal node - call quiescence search with proper bounds
	if depth == 0 {
		// Enter quiescence search to resolve captures
		score := m.quiescence(ctx, b, player, alpha, beta, stats)

		// Store in transposition table
		if m.useTranspositions && m.transpositionTable != nil {
			m.transpositionTable.Store(hash, 0, score, EntryExact, board.Move{})
		}

		return score
	}

	// Get all legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// No legal moves - check for checkmate or stalemate
		if m.generator.IsKingInCheck(b, player) {
			// Checkmate - current player loses, return negative mate score
			score := -ai.MateScore + ai.EvaluationScore(originalMaxDepth-depth)
			if m.useTranspositions && m.transpositionTable != nil {
				m.transpositionTable.Store(hash, depth, score, EntryExact, board.Move{})
			}
			return score
		}
		// Stalemate
		if m.useTranspositions && m.transpositionTable != nil {
			m.transpositionTable.Store(hash, depth, ai.DrawScore, EntryExact, board.Move{})
		}
		return ai.DrawScore
	}

	// Sort moves to improve alpha-beta efficiency (captures first, with TT move prioritized)
	m.orderMoves(legalMoves, currentDepth, ttMove)

	// Search moves
	bestScore := ai.EvaluationScore(-ai.MateScore - 1)
	bestMove := board.Move{}
	entryType := EntryUpperBound // Assume fail-low initially

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

		// Update best score
		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		// Update alpha (best score for current player)
		if score > alpha {
			alpha = score
			entryType = EntryExact // We have an exact score

			// Beta cutoff - opponent won't allow this line
			if alpha >= beta {
				// Store killer move if it's not a capture
				if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
					m.storeKiller(move, currentDepth)
				}

				// Store in transposition table
				if m.useTranspositions && m.transpositionTable != nil {
					m.transpositionTable.Store(hash, depth, beta, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	// Store result in transposition table
	if m.useTranspositions && m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
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

// orderMoves sorts moves to improve alpha-beta efficiency using TT move, MVV-LVA and killer moves
func (m *MinimaxEngine) orderMoves(moveList *moves.MoveList, depth int, ttMove board.Move) {
	// Create slice of move indices with their scores
	type moveScore struct {
		index int
		score int
	}

	scores := make([]moveScore, moveList.Count)
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		var score int

		// TT move gets absolute highest priority
		if ttMove.From.File != -1 && ttMove.From.Rank != -1 &&
			move.From == ttMove.From && move.To == ttMove.To {
			score = 3000000 // Highest priority
		} else if move.IsCapture {
			// Captures get second priority with MVV-LVA scoring
			score = 1000000 + m.getMVVLVAScore(move)
		} else if m.isKillerMove(move, depth) {
			// Killer moves get third priority
			if depth >= 0 && depth < MaxKillerDepth &&
				move.From == m.killerTable[depth][0].From && move.To == m.killerTable[depth][0].To {
				score = 500000 // First killer
			} else {
				score = 490000 // Second killer
			}
		} else {
			// Non-captures get lowest priority
			score = 0
		}
		scores[i] = moveScore{index: i, score: score}
	}

	// Sort by score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Reorder moves in-place using the sorted indices
	for i := 0; i < moveList.Count; i++ {
		// Find where the i-th best move currently is
		targetIndex := scores[i].index
		if targetIndex != i {
			// Swap moves and update indices
			moveList.Moves[i], moveList.Moves[targetIndex] = moveList.Moves[targetIndex], moveList.Moves[i]

			// Update the index tracking for the swapped move
			for j := i + 1; j < moveList.Count; j++ {
				if scores[j].index == i {
					scores[j].index = targetIndex
					break
				}
			}
		}
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

// SetTranspositionTableSize initializes the transposition table with the given size in MB
func (m *MinimaxEngine) SetTranspositionTableSize(sizeMB int) {
	if sizeMB <= 0 {
		m.transpositionTable = nil
		m.useTranspositions = false
		return
	}
	m.transpositionTable = NewTranspositionTable(sizeMB)
	m.useTranspositions = true
}

// SetTranspositionTableEnabled enables or disables transposition table usage
func (m *MinimaxEngine) SetTranspositionTableEnabled(enabled bool) {
	if m.transpositionTable != nil {
		m.useTranspositions = enabled
	}
}

// GetTranspositionTableStats returns transposition table statistics if available
func (m *MinimaxEngine) GetTranspositionTableStats() (hits, misses, collisions uint64, hitRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetStats()
	}
	return 0, 0, 0, 0
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	// Clear killer moves
	for i := 0; i < MaxKillerDepth; i++ {
		m.killerTable[i][0] = board.Move{}
		m.killerTable[i][1] = board.Move{}
	}
	// Clear transposition table if enabled
	if m.useTranspositions && m.transpositionTable != nil {
		m.transpositionTable.Clear()
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

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}
