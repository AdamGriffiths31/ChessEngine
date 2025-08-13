package search

import (
	"context"
	"fmt"
	"math"
	"sync"
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
	// MateDistanceThreshold defines the distance from mate scores where special handling is needed
	MateDistanceThreshold = 1000
)

// LMRTable is a pre-calculated reduction table for Late Move Reductions
// Indexed by [depth][moveCount] to get reduction amount
var LMRTable [16][64]int

func init() {
	for depth := 1; depth < 16; depth++ {
		for moveCount := 1; moveCount < 64; moveCount++ {
			// More aggressive LMR reduction - changed from /2.0 to /1.8
			// This provides stronger pruning while maintaining search quality
			LMRTable[depth][moveCount] = int(math.Log(float64(depth)) * math.Log(float64(moveCount)) / 1.8)
		}
	}
}

// ThreadSearchParams holds thread-specific search parameters for Lazy SMP diversity
type ThreadSearchParams struct {
	LMRDivisor          float64 // LMR reduction divisor (1.6 to 2.0)
	NullMoveReduction   int     // Null move reduction R value (2-4)
	HistoryHighThreshold int32  // High history score threshold
	HistoryMedThreshold int32   // Medium history score threshold  
	HistoryLowThreshold int32   // Low history score threshold
}

// getThreadSearchParams returns thread-specific search parameters based on threadID
func getThreadSearchParams(threadID int) ThreadSearchParams {
	// Use very subtle variations to maintain search quality while providing minimal diversity
	// The key insight: threads should cooperate, not compete with completely different strategies
	
	baseParams := ThreadSearchParams{
		LMRDivisor:          1.8,
		NullMoveReduction:   2,
		HistoryHighThreshold: 2000,
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,
	}
	
	// Thread 0 always uses baseline parameters
	if threadID == 0 || threadID == -1 {
		return baseParams
	}
	
	// Other threads use more significant variations for better diversity
	// Following Stockfish's approach: threads should explore different paths
	switch threadID % 4 {
	case 1:
		return ThreadSearchParams{
			LMRDivisor:          1.6,   // More aggressive LMR: -11%
			NullMoveReduction:   3,     // More aggressive null move
			HistoryHighThreshold: 1500, // Lower thresholds: -25%
			HistoryMedThreshold:  300,  // Lower thresholds: -40%
			HistoryLowThreshold:  -700, // Lower thresholds: +40%
		}
	case 2:
		return ThreadSearchParams{
			LMRDivisor:          2.2,   // More conservative LMR: +22%
			NullMoveReduction:   2,     // Standard null move
			HistoryHighThreshold: 2800, // Higher thresholds: +40%
			HistoryMedThreshold:  800,  // Higher thresholds: +60%
			HistoryLowThreshold:  -300, // Higher thresholds: -40%
		}
	case 3:
		return ThreadSearchParams{
			LMRDivisor:          1.4,   // Very aggressive LMR: -22%
			NullMoveReduction:   4,     // Very aggressive null move
			HistoryHighThreshold: 1200, // Very low thresholds: -40%
			HistoryMedThreshold:  200,  // Very low thresholds: -60%
			HistoryLowThreshold:  -900, // Very low thresholds: +80%
		}
	default:
		return baseParams
	}
}

// ThreadLocalState contains state for single-threaded search
// Kept for backwards compatibility and single-threaded mode
type ThreadLocalState struct {
	// Killer move table
	killerTable [MaxKillerDepth][2]board.Move

	// Move ordering buffer
	moveOrderBuffer []moveScore

	// Reusable move buffer to avoid allocations during ordering
	reorderBuffer []board.Move

	// Search statistics
	searchStats ai.SearchStats

	// Thread-specific search parameters
	searchParams ThreadSearchParams

	// Debug information for move ordering (if enabled)
	debugMoveOrder []board.Move

	// Thread-local history table for move ordering diversity
	// This is CRITICAL for Lazy SMP thread diversity
	historyTable *HistoryTable

	// Thread ID for debugging and diversity
	threadID int
}


// Cleanup properly shuts down the engine and its resources
func (m *MinimaxEngine) Cleanup() {
	// Cleanup resources
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, SEE-based move ordering, and opening book support
// Now thread-safe with per-thread state management
type MinimaxEngine struct {
	// Shared read-only or synchronized resources
	evaluator          ai.Evaluator
	generator          *moves.Generator
	bookService        *openings.BookLookupService
	transpositionTable *TranspositionTable
	zobrist            *openings.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *evaluation.SEECalculator

	// Global debug flag (read-only during search)
	debugMoveOrdering bool

	// Thread-local state management
	threadStates sync.Map // map[int64]*ThreadLocalState - thread ID to state

	// Statistics aggregation across all threads
	globalStats ai.SearchStats
	statsMutex  sync.Mutex // For aggregating thread-local stats

}

// moveScore holds move index and score for ordering
type moveScore struct {
	index int
	score int
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	engine := &MinimaxEngine{
		evaluator:          evaluation.NewEvaluator(),
		generator:          moves.NewGenerator(),
		bookService:        nil,
		transpositionTable: nil,
		zobrist:            openings.GetPolyglotHash(),
		historyTable:       NewHistoryTable(),
		seeCalculator:      evaluation.NewSEECalculator(),
	}

	// Pre-initialize thread states for up to 16 worker threads
	// This eliminates the expensive on-demand thread state creation during search
	engine.initializeThreadStates(16)

	return engine
}

// initializeThreadStates pre-allocates thread-local state for parallel search
// This prevents expensive on-demand allocation during search and reduces contention
func (m *MinimaxEngine) initializeThreadStates(maxThreads int) {
	for threadID := 0; threadID < maxThreads; threadID++ {
		key := generateThreadKey(threadID)
		
		// Create pre-allocated thread state
		threadState := &ThreadLocalState{
			// Pre-allocate killer move table
			killerTable: [MaxKillerDepth][2]board.Move{},
			
			// Pre-allocate move ordering buffer - increased size to avoid reallocation
			// Maximum legal moves in chess is around 218, so 512 provides good headroom
			moveOrderBuffer: make([]moveScore, 512),
			
			// Pre-allocate reorder buffer
			reorderBuffer: make([]board.Move, 512),
			
			
			// Initialize empty search stats
			searchStats: ai.SearchStats{},
			
			// Set thread-specific search parameters
			searchParams: getThreadSearchParams(threadID),
			
			// Pre-allocate debug move order slice
			debugMoveOrder: make([]board.Move, 0, 512),
			
			// CRITICAL: Each thread gets its own history table for move ordering diversity
			historyTable: NewHistoryTable(),
			
			// Store thread ID for debugging
			threadID: threadID,
		}
		
		// Store in thread states map
		m.threadStates.Store(key, threadState)
	}
	
	// Also pre-initialize the main thread state (threadID -1)
	mainKey := "thread_main"
	mainState := &ThreadLocalState{
		killerTable:     [MaxKillerDepth][2]board.Move{},
		moveOrderBuffer: make([]moveScore, 512),
		reorderBuffer:   make([]board.Move, 512),
		searchStats:     ai.SearchStats{},
		searchParams:    getThreadSearchParams(-1), // Use baseline parameters for main thread
		historyTable:    NewHistoryTable(), // Main thread also gets its own history table
		threadID:        -1, // Main thread ID
		debugMoveOrder:  make([]board.Move, 0, 512),
	}
	m.threadStates.Store(mainKey, mainState)
}

// fastCopyBoard creates an optimized deep copy of a board for thread isolation
// This is much faster than the previous piece-by-piece approach
func (m *MinimaxEngine) fastCopyBoard(original *board.Board) *board.Board {
	if original == nil {
		return nil
	}

	// Create new board instance
	newBoard := &board.Board{}

	// Copy simple value fields
	newBoard.SetCastlingRights(original.GetCastlingRights())
	newBoard.SetHalfMoveClock(original.GetHalfMoveClock())
	newBoard.SetFullMoveNumber(original.GetFullMoveNumber())
	newBoard.SetSideToMove(original.GetSideToMove())

	// Copy en passant target (pointer field)
	if epTarget := original.GetEnPassantTarget(); epTarget != nil {
		newTarget := &board.Square{
			File: epTarget.File,
			Rank: epTarget.Rank,
		}
		newBoard.SetEnPassantTarget(newTarget)
	}

	// Bulk copy bitboards (much faster than piece-by-piece)
	for i := 0; i < 12; i++ {
		newBoard.PieceBitboards[i] = original.PieceBitboards[i]
	}

	// Copy color bitboards
	newBoard.WhitePieces = original.WhitePieces
	newBoard.BlackPieces = original.BlackPieces
	newBoard.AllPieces = original.AllPieces

	// Copy mailbox representation
	for i := 0; i < 64; i++ {
		newBoard.Mailbox[i] = original.Mailbox[i]
	}

	// Copy hash state
	newBoard.SetHash(original.GetHash())
	newBoard.SetHashUpdater(m)

	return newBoard
}

// getThreadLocalState returns state for single-threaded search
// For multi-threaded search, workers will have their own WorkerState
func (m *MinimaxEngine) getThreadLocalState() *ThreadLocalState {
	const singleThreadKey = "single_thread"

	// Try to load existing state
	if state, exists := m.threadStates.Load(singleThreadKey); exists {
		return state.(*ThreadLocalState)
	}

	// Create new thread-local state for single-threaded mode
	newState := &ThreadLocalState{
		moveOrderBuffer: make([]moveScore, 256), // Pre-allocate buffer
		debugMoveOrder:  make([]board.Move, 0),  // Debug move order tracking
	}

	// Store and return the new state
	m.threadStates.Store(singleThreadKey, newState)
	return newState
}

// generateThreadKey creates a consistent key for thread-specific state storage
func generateThreadKey(threadID int) string {
	return fmt.Sprintf("worker_%d", threadID)
}

// getThreadSpecificState returns state for a specific thread ID
// This ensures each parallel search thread has isolated state
func (m *MinimaxEngine) getThreadSpecificState(threadID int) *ThreadLocalState {
	// Use consistent worker key format
	key := generateThreadKey(threadID)

	// Try to load existing state
	if state, exists := m.threadStates.Load(key); exists {
		return state.(*ThreadLocalState)
	}

	// This shouldn't happen if workers are pre-initialized properly
	// But create as fallback for safety
	fmt.Printf("WARNING: Creating thread state on-demand for worker %d\n", threadID)
	newState := &ThreadLocalState{
		moveOrderBuffer: make([]moveScore, 256),
		searchParams:    getThreadSearchParams(threadID), // Set thread-specific parameters
		debugMoveOrder:  make([]board.Move, 0),
	}

	// Store and return the new state
	m.threadStates.Store(key, newState)
	return newState
}

// getThreadLocalStateFromContext returns thread state based on context
// If context has threadID, returns thread-specific state; otherwise returns single-thread state
func (m *MinimaxEngine) getThreadLocalStateFromContext(ctx context.Context) *ThreadLocalState {
	// Check if this is a parallel search with thread ID in context
	if threadID, ok := ctx.Value("threadID").(int); ok {
		return m.getThreadSpecificState(threadID)
	}

	// Fall back to single-thread state for normal (non-parallel) search
	return m.getThreadLocalState()
}

// AggregateThreadStats collects statistics from all thread-local states
// Should be called after search completion for accurate reporting
func (m *MinimaxEngine) AggregateThreadStats() ai.SearchStats {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// Reset global stats
	m.globalStats = ai.SearchStats{}

	// Aggregate from all threads and reset thread-local stats
	m.threadStates.Range(func(key, value interface{}) bool {
		threadState := value.(*ThreadLocalState)
		stats := threadState.searchStats

		// Add thread statistics to global totals
		m.globalStats.NodesSearched += stats.NodesSearched
		m.globalStats.LMRReductions += stats.LMRReductions
		m.globalStats.LMRReSearches += stats.LMRReSearches
		m.globalStats.LMRNodesSkipped += stats.LMRNodesSkipped

		// Reset thread-local stats to prevent accumulation across searches
		threadState.searchStats = ai.SearchStats{}

		return true // Continue iteration
	})

	return m.globalStats
}

// resetThreadLocalStats clears statistics from all thread-local states
// Should be called after aggregating stats for the next search
func (m *MinimaxEngine) resetThreadLocalStats() {
	m.threadStates.Range(func(key, value interface{}) bool {
		threadState := value.(*ThreadLocalState)
		threadState.searchStats = ai.SearchStats{}
		return true
	})
}

// GetHashDelta implements the board.HashUpdater interface
// Calculates the zobrist hash delta for a move from old state to new state
func (m *MinimaxEngine) GetHashDelta(b *board.Board, move board.Move, oldState board.BoardState) uint64 {
	var hashDelta uint64

	// Always flip side to move
	hashDelta ^= m.zobrist.GetSideKey()

	// Handle piece movement (remove from source, add to destination)
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// Remove piece from source square
	if move.Piece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(move.Piece)
		hashDelta ^= m.zobrist.GetPieceKey(fromSquare, pieceIndex)
	}

	// Add piece to destination square (or promotion piece)
	var destPiece board.Piece
	if move.Promotion != board.Empty && move.Promotion != 0 {
		destPiece = move.Promotion
	} else {
		destPiece = move.Piece
	}
	if destPiece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(destPiece)
		hashDelta ^= m.zobrist.GetPieceKey(toSquare, pieceIndex)
	}

	// Handle captured piece
	if move.IsCapture && move.Captured != board.Empty {
		capturedPieceIndex := m.zobrist.GetPieceIndex(move.Captured)
		if move.IsEnPassant {
			// En passant capture is on a different square
			var captureRank int
			if move.Piece == board.WhitePawn {
				captureRank = 4 // Black pawn on rank 5
			} else {
				captureRank = 3 // White pawn on rank 4
			}
			captureSquare := captureRank*8 + move.To.File
			hashDelta ^= m.zobrist.GetPieceKey(captureSquare, capturedPieceIndex)
		} else {
			hashDelta ^= m.zobrist.GetPieceKey(toSquare, capturedPieceIndex)
		}
	}

	// Handle castling rook movement
	if move.IsCastling {
		var rookFrom, rookTo int
		switch move.To.File {
		case 6: // King-side castling
			rookFrom = move.From.Rank*8 + 7
			rookTo = move.From.Rank*8 + 5
		case 2: // Queen-side castling
			rookFrom = move.From.Rank*8 + 0
			rookTo = move.From.Rank*8 + 3
		}

		// Remove rook from original square, add to new square
		var rook board.Piece
		if move.From.Rank == 0 {
			rook = board.BlackRook
		} else {
			rook = board.WhiteRook
		}
		rookIndex := m.zobrist.GetPieceIndex(rook)
		hashDelta ^= m.zobrist.GetPieceKey(rookFrom, rookIndex)
		hashDelta ^= m.zobrist.GetPieceKey(rookTo, rookIndex)
	}

	// Handle castling rights changes
	if oldState.CastlingRights != b.GetCastlingRights() {
		oldRights := m.zobrist.GetCastlingKey(oldState.CastlingRights)
		newRights := m.zobrist.GetCastlingKey(b.GetCastlingRights())
		hashDelta ^= oldRights ^ newRights
	}

	// Handle en passant target changes
	// Only include en passant in hash if there are adjacent pawns that can capture
	if (oldState.EnPassantTarget == nil) != (b.GetEnPassantTarget() == nil) ||
		(oldState.EnPassantTarget != nil && b.GetEnPassantTarget() != nil &&
			oldState.EnPassantTarget.File != b.GetEnPassantTarget().File) {

		if oldState.EnPassantTarget != nil && m.hasAdjacentCapturingPawn(b, oldState.EnPassantTarget, oldState.SideToMove) {
			hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantTarget.File)
		}
		if b.GetEnPassantTarget() != nil && m.hasAdjacentCapturingPawn(b, b.GetEnPassantTarget(), b.GetSideToMove()) {
			hashDelta ^= m.zobrist.GetEnPassantKey(b.GetEnPassantTarget().File)
		}
	}

	return hashDelta
}

// GetNullMoveDelta returns the hash delta for a null move (flip side to move)
func (m *MinimaxEngine) GetNullMoveDelta() uint64 {
	return m.zobrist.GetSideKey()
}

// hasAdjacentCapturingPawn checks if there's a pawn adjacent to the en passant target that can capture
// This implements the same logic as the full HashPosition function
func (m *MinimaxEngine) hasAdjacentCapturingPawn(b *board.Board, epTarget *board.Square, sideToMove string) bool {
	// Determine the rank where the capturing pawn should be and what piece to look for
	var pawnRank int
	var pawnPiece board.Piece

	if sideToMove == "b" { // Black to move, so white pawn moved and black can capture
		pawnRank = 4 // Black pawn should be on rank 5 (0-indexed rank 4)
		pawnPiece = board.BlackPawn
	} else { // White to move, so black pawn moved and white can capture
		pawnRank = 3 // White pawn should be on rank 4 (0-indexed rank 3)
		pawnPiece = board.WhitePawn
	}

	epFile := epTarget.File

	// Check adjacent files for capturing pawn
	for _, df := range []int{-1, 1} {
		adjFile := epFile + df
		if adjFile >= 0 && adjFile < 8 {
			if b.GetPiece(pawnRank, adjFile) == pawnPiece {
				return true
			}
		}
	}

	return false
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

	// Get thread state for single-threaded or main thread
	threadState := m.getThreadLocalStateFromContext(ctx)


	// Opening book lookup
	if config.UseOpeningBook {
		m.initializeBookService(config)
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

	// Initialize incremental hashing for the board
	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	startTime := time.Now()

	// Age transposition table and history table
	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

	if m.historyTable != nil {
		m.historyTable.Age()
	}

	// Route to appropriate search implementation
	if config.NumThreads > 1 {
		// Use Lazy SMP for multi-threaded search
		return m.lazySMPSearch(ctx, b, player, config)
	}
	
	// Single-threaded search using runIterativeDeepening
	return m.runIterativeDeepening(ctx, b, player, config, threadState, startTime)
}

// lazySMPSearch implements Lazy SMP (Symmetric MultiProcessing) parallel search
//
// How Lazy SMP Works:
// Each thread searches the ENTIRE tree from the root position independently.
// This seems redundant but actually provides excellent work distribution because:
//
// 1. Transposition Table creates diversity:
//   - Thread A searches move X first, stores result in TT
//   - Thread B hits this TT entry, changes move ordering, searches Y first
//   - Different move orders → different subtrees → natural work split
//
// 2. Timing variations compound:
//   - Threads start at slightly different times
//   - Small differences in TT hits cascade into large search differences
//   - Each thread naturally explores different parts of the search tree
//
// 3. No explicit coordination needed:
//   - No work queues or move assignments
//   - No communication overhead between threads
//   - Load balancing happens automatically via TT
//
// 4. Proven effectiveness:
//   - Used by Stockfish and other top engines
//   - Typically achieves 5-7x speedup with 8 threads
//   - Much simpler than work-splitting approaches
//
// The "lazy" refers to lazy coordination - we don't explicitly distribute
// work, instead letting probability and the shared TT handle it naturally.
func (m *MinimaxEngine) lazySMPSearch(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	// Input validation
	if b == nil {
		return ai.SearchResult{
			BestMove: board.Move{},
			Score:    ai.EvaluationScore(-ai.MateScore),
			Stats:    ai.SearchStats{},
		}
	}

	// Handle opening book
	if config.UseOpeningBook {
		m.initializeBookService(config)
		if m.bookService != nil {
			bookMove, err := m.bookService.FindBookMove(b)
			if err == nil && bookMove != nil {
				return ai.SearchResult{
					BestMove: *bookMove,
					Score:    0,
					Stats:    ai.SearchStats{BookMoveUsed: true},
				}
			}
		}
	}

	// Initialize hashing
	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	// Age caches
	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}
	if m.historyTable != nil {
		m.historyTable.Age()
	}

	// Generate legal moves once
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		// PANIC: This should not happen in our test case - main thread finding no legal moves
		isCheck := m.generator.IsKingInCheck(b, player)
		panic(fmt.Sprintf("THREADING BUG: Main thread found no legal moves! Player=%v, InCheck=%v, FEN=%s", 
			player, isCheck, b.ToFEN()))
	}

	// Launch independent search threads
	return m.launchLazySMPWorkers(ctx, b, player, config)
}

// lazySMPWorkerResult represents the result from a single Lazy SMP worker thread
type lazySMPWorkerResult struct {
	bestMove board.Move
	score    ai.EvaluationScore
	depth    int
	stats    ai.SearchStats
	workerID int
}

// launchLazySMPWorkers launches independent worker threads for Lazy SMP search
func (m *MinimaxEngine) launchLazySMPWorkers(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {

	resultChan := make(chan lazySMPWorkerResult, config.NumThreads)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Launch worker threads - each runs complete iterative deepening independently
	for workerID := 0; workerID < config.NumThreads; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					fmt.Printf("[PANIC-worker-%d] Worker thread panicked: %v\n", id, r)
					panic(r) // Re-panic to maintain original behavior
				}
			}()

			// Each worker gets its own board copy and thread state
			workerBoard := m.fastCopyBoard(b)
			if workerBoard == nil {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - fastCopyBoard returned nil!", id))
			}
			
			threadState := m.getThreadSpecificState(id)
			if threadState == nil {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - getThreadSpecificState returned nil!", id))
			}

			// Verify board copy integrity
			
			// Verify FEN matches original
			originalFEN := b.ToFEN()
			workerFEN := workerBoard.ToFEN()
			if originalFEN != workerFEN {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - Board copy mismatch! Original: %s, Worker: %s", 
					id, originalFEN, workerFEN))
			}


			// Age thread-local history table for this worker
			if threadState.historyTable != nil {
				threadState.historyTable.Age()
			}

			// Run single-threaded search with this worker's state
			// This is the key - each thread runs the SAME algorithm but gets
			// different results due to TT interactions and timing
			result := m.runIterativeDeepening(ctx, workerBoard, player, config, threadState, startTime, id)

			// Validate result before proceeding
			if result.Stats.Depth == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - runIterativeDeepening returned depth 0! Nodes=%d, Score=%d", 
					id, result.Stats.NodesSearched, result.Score))
			}
			
			if result.Stats.NodesSearched == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - runIterativeDeepening searched 0 nodes! Depth=%d, Score=%d", 
					id, result.Stats.Depth, result.Score))
			}


			// Report result
			workerResult := lazySMPWorkerResult{
				bestMove: result.BestMove,
				score:    result.Score,
				depth:    result.Stats.Depth,
				stats:    result.Stats,
				workerID: id,
			}
			
			// Validate worker result before sending
			if workerResult.depth == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - About to send result with depth 0!", id))
			}
			
			// Always send result - worker has completed the work
			// Don't let context cancellation prevent result reporting
			resultChan <- workerResult
		}(workerID)
	}

	// Close results when all workers complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect and select best result
	return m.selectBestLazySMPResult(resultChan, startTime)
}

// runIterativeDeepening runs the core iterative deepening search
// Used for both single-threaded search and individual worker threads in Lazy SMP
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig, threadState *ThreadLocalState, startTime time.Time, workerID ...int) ai.SearchResult {
	threadID := "main"
	if len(workerID) > 0 {
		threadID = fmt.Sprintf("worker-%d", workerID[0])
	}
	
	// Generate legal moves
	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)


	if legalMoves.Count == 0 {
		// PANIC: This should not happen in our test case - if it does, we need to see exactly why
		isCheck := m.generator.IsKingInCheck(b, player)
		panic(fmt.Sprintf("THREADING BUG: No legal moves found in thread %s! Player=%v, InCheck=%v, FEN=%s", 
			threadID, player, isCheck, b.ToFEN()))
	}

	// Get TT move for initial ordering
	var rootTTMove board.Move
	if m.transpositionTable != nil {
		hash := b.GetHash()
		if entry, found := m.transpositionTable.Probe(hash); found {
			rootTTMove = entry.BestMove
		}
	}

	// Use thread-specific root move ordering for diversity
	m.orderRootMovesWithDiversity(b, legalMoves, rootTTMove, threadState)
	

	// Track best move across depths
	lastCompletedBestMove := legalMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)
	var finalStats ai.SearchStats

	// THREAD DIVERSITY: Different starting depths based on thread ID
	// This is a key technique used by Stockfish and other engines
	startingDepth := 1
	if threadState != nil && threadState.threadID >= 0 {
		// Even threads start at depth 1, odd threads start at depth 2
		if threadState.threadID % 2 == 1 {
			startingDepth = 2
		}
	}

	// Iterative deepening loop
	for currentDepth := startingDepth; currentDepth <= config.MaxDepth; currentDepth++ {
		select {
		case <-ctx.Done():
			// Time expired, return last completed result
			finalStats.Time = time.Since(startTime)
			return ai.SearchResult{
				BestMove: lastCompletedBestMove,
				Score:    lastCompletedScore,
				Stats:    finalStats,
			}
		default:
		}

		// Check global time limit
		if time.Since(startTime) >= config.MaxTime {
			break
		}

		// Search at this depth using the same logic as single-threaded
		bestScore := ai.EvaluationScore(-ai.MateScore - 1)
		bestMove := legalMoves.Moves[0]

		// Use aspiration windows for depths > 1
		alpha := ai.EvaluationScore(-ai.MateScore - 1)
		beta := ai.EvaluationScore(ai.MateScore + 1)

		if currentDepth > 1 {
			// Aspiration window around last result
			window := ai.EvaluationScore(50)
			alpha = lastCompletedScore - window
			beta = lastCompletedScore + window
		}

		// Search all moves at this depth
		moveIndex := 0

		for _, move := range legalMoves.Moves[:legalMoves.Count] {
			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				continue
			}

			var score ai.EvaluationScore
			var moveStats ai.SearchStats

			if moveIndex == 0 {
				// First move - use full window
				score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, threadState, &moveStats)
			} else {
				// Try null window search first
				score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -alpha-1, -alpha, currentDepth, config, threadState, &moveStats)

				if score > alpha && score < beta {
					// Research with full window
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, threadState, &moveStats)
				}
			}

			b.UnmakeMove(undo)

			if score > bestScore {
				bestScore = score
				bestMove = move

				if score > alpha {
					alpha = score
				}
			}

			moveIndex++

			// Check for time expiry during search
			if time.Since(startTime) >= config.MaxTime {
				break
			}
		}

		// Update final stats
		finalStats.NodesSearched = threadState.searchStats.NodesSearched
		finalStats.Depth = currentDepth
		finalStats.LMRReductions = threadState.searchStats.LMRReductions
		finalStats.LMRReSearches = threadState.searchStats.LMRReSearches
		finalStats.LMRNodesSkipped = threadState.searchStats.LMRNodesSkipped

		// Check if we completed this depth
		if time.Since(startTime) < config.MaxTime {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
		} else {
		}

		// Early termination on mate
		if bestScore >= ai.MateScore-1000 || bestScore <= -ai.MateScore+1000 {
			break
		}
	}

	finalStats.Time = time.Since(startTime)
	
	// Final validation before returning
	if finalStats.Depth == 0 {
		panic(fmt.Sprintf("THREADING BUG: %s - About to return result with depth 0! Nodes=%d, Score=%d, BestMove=%s", 
			threadID, finalStats.NodesSearched, lastCompletedScore, moveToDebugString(lastCompletedBestMove)))
	}
	
	if finalStats.NodesSearched == 0 {
		panic(fmt.Sprintf("THREADING BUG: %s - About to return result with 0 nodes! Depth=%d, Score=%d, BestMove=%s", 
			threadID, finalStats.Depth, lastCompletedScore, moveToDebugString(lastCompletedBestMove)))
	}
	
	
	return ai.SearchResult{
		BestMove: lastCompletedBestMove,
		Score:    lastCompletedScore,
		Stats:    finalStats,
	}
}

// selectBestLazySMPResult selects the best result from all Lazy SMP workers using thread voting
func (m *MinimaxEngine) selectBestLazySMPResult(resultChan <-chan lazySMPWorkerResult, startTime time.Time) ai.SearchResult {
	var aggregatedStats ai.SearchStats
	resultsReceived := 0

	// Thread voting mechanism - each move gets votes weighted by depth and score
	moveVotes := make(map[board.Move]float64)
	moveResults := make(map[board.Move]lazySMPWorkerResult)
	var allResults []lazySMPWorkerResult


	// Collect all results and calculate voting weights
	var minScore, maxScore ai.EvaluationScore
	var maxDepth int
	firstResult := true

	for result := range resultChan {
		resultsReceived++
		allResults = append(allResults, result)

		// Validate received result
		if result.depth == 0 {
			panic(fmt.Sprintf("THREADING BUG: Main thread received result with depth 0 from worker-%d!", result.workerID))
		}
		
		if result.stats.NodesSearched == 0 {
			panic(fmt.Sprintf("THREADING BUG: Main thread received result with 0 nodes from worker-%d!", result.workerID))
		}

		// Track score range and max depth for normalization
		if firstResult {
			minScore = result.score
			maxScore = result.score
			firstResult = false
		} else {
			if result.score < minScore {
				minScore = result.score
			}
			if result.score > maxScore {
				maxScore = result.score
			}
		}
		
		if result.depth > maxDepth {
			maxDepth = result.depth
		}

		// Aggregate stats from all workers
		aggregatedStats.NodesSearched += result.stats.NodesSearched
		aggregatedStats.LMRReductions += result.stats.LMRReductions
		aggregatedStats.LMRReSearches += result.stats.LMRReSearches
		aggregatedStats.LMRNodesSkipped += result.stats.LMRNodesSkipped

		// Store result for this move (prefer deeper/better scored result)
		if existing, exists := moveResults[result.bestMove]; !exists ||
			result.depth > existing.depth ||
			(result.depth == existing.depth && result.score > existing.score) {
			moveResults[result.bestMove] = result
		}
	}

	// Calculate votes for each move
	scoreRange := maxScore - minScore
	if scoreRange == 0 {
		scoreRange = 1 // Avoid division by zero
	}


	for _, result := range allResults {
		// Vote weight based on:
		// 1. Depth achieved (higher is better)
		// 2. Score relative to worst score (better scores get more weight)
		// 3. Small bonus for nodes searched (tie breaker)
		
		depthWeight := float64(result.depth) / float64(maxDepth)
		
		// Normalize score to 0-1 range
		scoreWeight := float64(result.score-minScore) / float64(scoreRange)
		
		// Node weight (very small, just for tie breaking)
		nodeWeight := math.Log10(float64(result.stats.NodesSearched)+1) / 100.0
		
		// Combined vote weight
		voteWeight := depthWeight*0.6 + scoreWeight*0.35 + nodeWeight*0.05
		
		moveVotes[result.bestMove] += voteWeight
		
	}

	// Select move with highest vote count
	var bestMove board.Move
	var bestVotes float64
	for move, votes := range moveVotes {
		if votes > bestVotes {
			bestVotes = votes
			bestMove = move
		}
	}


	// Get the best result for the winning move
	bestResult := moveResults[bestMove]

	// Set aggregated depth to max achieved
	aggregatedStats.Depth = maxDepth

	aggregatedStats.Time = time.Since(startTime)

	// Fallback if no valid results were received
	if resultsReceived == 0 {
		panic("THREADING BUG: No worker results received from any worker threads!")
	}
	
	if bestResult.depth == 0 {
		panic(fmt.Sprintf("THREADING BUG: Best result has depth 0! Received %d results, best worker was %d", 
			resultsReceived, bestResult.workerID))
	}

	return ai.SearchResult{
		BestMove: bestResult.bestMove,
		Score:    bestResult.score,
		Stats:    aggregatedStats,
	}
}

// negamax performs negamax search with alpha-beta pruning and optimizations
func (m *MinimaxEngine) negamax(ctx context.Context, b *board.Board, player moves.Player, depth int, alpha, beta ai.EvaluationScore, originalMaxDepth int, config ai.SearchConfig, threadState *ThreadLocalState, stats *ai.SearchStats) ai.EvaluationScore {
	// Count nodes in provided thread state
	threadState.searchStats.NodesSearched++

	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	// Check extension - extend search when in check
	inCheck := m.generator.IsKingInCheck(b, player)
	if inCheck && depth < originalMaxDepth {
		depth++
	}

	var ttMove board.Move
	hash := b.GetHash()

	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			ttMove = entry.BestMove
			
			// Validate TT move is legal (Stockfish-style validation)
			if ttMove.From.File >= 0 && ttMove.From.File <= 7 && 
			   ttMove.To.File >= 0 && ttMove.To.File <= 7 {
				// Additional validation: check if move makes basic sense
				if ttMove.From != ttMove.To {
					// TT move is basically valid, can use it
				} else {
					// Invalid TT move, clear it
					ttMove = board.Move{}
				}
			} else {
				// Invalid coordinates, clear TT move
				ttMove = board.Move{}
			}

			// Only use TT score if depth is sufficient AND in multi-threaded mode,
			// require higher depth to reduce interference
			minDepthRequired := depth
			if config.NumThreads > 1 {
				// In multi-threaded mode, require deeper entries to reduce interference
				minDepthRequired = depth + 1
			}

			if entry.GetDepth() >= minDepthRequired {
				switch entry.GetType() {
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

	// Null move pruning with static evaluation check
	staticEval := m.evaluator.Evaluate(b)
	if !config.DisableNullMove && depth >= 3 &&
		staticEval >= beta &&
		beta < ai.MateScore-MateDistanceThreshold &&
		beta > -ai.MateScore+MateDistanceThreshold {
		if !inCheck {
			threadState.searchStats.NullMoves++

			// Use thread-specific null move reduction
			nullReduction := threadState.searchParams.NullMoveReduction
			if depth >= 6 && nullReduction < 3 {
				nullReduction++ // Increase reduction for deep searches, but keep it conservative
			}

			nullUndo := b.MakeNullMove()

			nullScore := -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-nullReduction, -beta, -beta+1, originalMaxDepth, config, threadState, stats)

			b.UnmakeNullMove(nullUndo)

			// If null move score >= beta, position is too good for opponent
			if nullScore >= beta {
				if nullScore < ai.MateScore-MateDistanceThreshold {
					threadState.searchStats.NullCutoffs++
					return beta
				}
			}
		}
	}

	// Terminal node - call quiescence search with thread state
	// Also handle negative depths that can result from aggressive reductions
	if depth <= 0 {
		score := m.quiescence(ctx, b, player, alpha, beta, originalMaxDepth-depth, threadState, stats)

		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, 0, score, EntryExact, board.Move{})
		}

		return score
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		if m.generator.IsKingInCheck(b, player) {
			score := -ai.MateScore + ai.EvaluationScore(originalMaxDepth-depth)
			if m.transpositionTable != nil {
				m.transpositionTable.Store(hash, depth, score, EntryExact, board.Move{})
			}
			return score
		}
		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, depth, ai.DrawScore, EntryExact, board.Move{})
		}
		return ai.DrawScore
	}

	// Sort moves using thread state
	m.orderMoves(b, legalMoves, currentDepth, ttMove, threadState)

	// Search moves
	bestScore := ai.EvaluationScore(-ai.MateScore - 1)
	bestMove := board.Move{}
	entryType := EntryUpperBound // Assume fail-low initially
	moveCount := 0

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		moveCount++
		var score ai.EvaluationScore

		// Calculate LMR reduction if applicable
		reduction := 0

		// Only apply LMR if enabled and conditions are met
		if depth >= config.LMRMinDepth &&
			moveCount > config.LMRMinMoves &&
			!inCheck &&
			!move.IsCapture &&
			move.Promotion == board.Empty &&
			!m.isKillerMove(move, currentDepth, threadState) {

			// Check if move gives check (prevents reducing tactical moves)
			givesCheck := m.moveGivesCheck(b, move)

			if !givesCheck {
				// Safe to reduce this quiet move - use thread-specific LMR divisor
				reduction = int(math.Log(float64(min(depth, 15))) * math.Log(float64(min(moveCount, 63))) / threadState.searchParams.LMRDivisor)

				// Adjust reduction based on history score with thread-specific thresholds
				historyScore := m.getHistoryScore(move, threadState)
				if historyScore > threadState.searchParams.HistoryHighThreshold {
					reduction = 0 // Don't reduce high history moves
				} else if historyScore > threadState.searchParams.HistoryMedThreshold && reduction > 0 {
					// Reduce less - ensure we don't underflow
					newReduction := reduction * 2 / 3
					if newReduction >= 0 {
						reduction = newReduction
					}
				} else if historyScore < threadState.searchParams.HistoryLowThreshold && reduction >= 0 {
					// Reduce more - ensure we don't overflow relative to depth
					newReduction := reduction * 4 / 3
					if newReduction < depth {
						reduction = newReduction
					} else {
						reduction = depth - 1
					}
				}

				// Final bounds checking
				if reduction >= depth {
					reduction = depth - 1
				}
				if reduction < 0 {
					reduction = 0
				}
			}
		}

		if reduction > 0 {
			// Track LMR statistics in thread state
			threadState.searchStats.LMRReductions++

			// Reduced-depth search with null window
			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-reduction, -alpha-1, -alpha, originalMaxDepth, config, threadState, stats)

			// Re-search at full depth if reduced search failed high (score > alpha)
			if score > alpha {
				threadState.searchStats.LMRReSearches++

				// Use full window for re-search
				score = -m.negamax(ctx, b, oppositePlayer(player),
					depth-1, -beta, -alpha, originalMaxDepth, config, threadState, stats)
			}
		} else {
			// Search at full depth with full window
			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1, -beta, -alpha, originalMaxDepth, config, threadState, stats)
		}

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
			entryType = EntryExact

			// Beta cutoff - opponent won't allow this line
			if alpha >= beta {
				if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
					m.storeKiller(move, currentDepth, threadState)
				}

				if !move.IsCapture {
					// Update thread-local history table for move ordering diversity
					if threadState != nil && threadState.historyTable != nil {
						threadState.historyTable.UpdateHistory(move, depth)
					} else {
						m.historyTable.UpdateHistory(move, depth)
					}
				}

				if m.transpositionTable != nil {
					m.transpositionTable.Store(hash, depth, beta, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
}

// quiescence performs quiescence search using provided thread state
func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, depthFromRoot int, threadState *ThreadLocalState, stats *ai.SearchStats) ai.EvaluationScore {
	// Count nodes in provided thread state
	threadState.searchStats.NodesSearched++

	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	hash := b.GetHash()
	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			if entry.GetDepth() >= 0 {
				switch entry.GetType() {
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

	inCheck := m.generator.IsKingInCheck(b, player)
	originalAlpha := alpha

	// Always calculate eval for delta pruning
	eval := m.evaluator.Evaluate(b)
	if player == moves.Black {
		eval = -eval
	}

	// Stand pat - if not in check, consider current position
	if !inCheck {
		if eval >= beta {
			return beta
		}
		if eval > alpha {
			alpha = eval
		}
	}

	// Generate all moves and filter for captures and promotions
	allMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(allMoves)

	// Create capture list from pool
	captureList := moves.GetMoveList()
	defer moves.ReleaseMoveList(captureList)

	// Filter for captures and promotions
	for i := 0; i < allMoves.Count; i++ {
		move := allMoves.Moves[i]
		if move.IsCapture || move.Promotion != board.Empty {
			captureList.AddMove(move)
		}
	}

	if captureList.Count == 0 {
		if inCheck {
			// No captures in check means checkmate
			return -ai.MateScore + ai.EvaluationScore(depthFromRoot)
		}
		// No captures available, return stand-pat evaluation
		return eval
	}

	// Order captures by SEE (Static Exchange Evaluation)
	m.orderCapturesWithThreadState(b, captureList, threadState)

	bestScore := eval
	for i := 0; i < captureList.Count; i++ {
		move := captureList.Moves[i]

		// Delta pruning - skip obviously bad captures
		if !inCheck {
			// Get approximate value of captured piece
			captureValue := ai.EvaluationScore(0)
			switch move.Captured {
			case board.WhitePawn, board.BlackPawn:
				captureValue = 100
			case board.WhiteKnight, board.BlackKnight, board.WhiteBishop, board.BlackBishop:
				captureValue = 300
			case board.WhiteRook, board.BlackRook:
				captureValue = 500
			case board.WhiteQueen, board.BlackQueen:
				captureValue = 900
			}

			// Add promotion value
			if move.Promotion != board.Empty {
				captureValue += 800 // Approximate value of promotion
			}

			// Delta pruning margin
			margin := ai.EvaluationScore(200)
			if eval+captureValue+margin < alpha {
				continue
			}
		}

		// SEE pruning - skip obviously losing captures
		if !inCheck && move.IsCapture {
			if seeScore := m.seeWithThreadState(b, move, threadState); seeScore < 0 {
				continue
			}
		}

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		score := -m.quiescence(ctx, b, oppositePlayer(player), -beta, -alpha, depthFromRoot+1, threadState, stats)
		b.UnmakeMove(undo)

		if score > bestScore {
			bestScore = score
		}

		if score > alpha {
			alpha = score
			if alpha >= beta {
				break
			}
		}
	}

	// Store result in transposition table
	if m.transpositionTable != nil {
		var entryType EntryType
		if bestScore <= originalAlpha {
			entryType = EntryUpperBound
		} else if bestScore >= beta {
			entryType = EntryLowerBound
		} else {
			entryType = EntryExact
		}
		m.transpositionTable.Store(hash, 0, bestScore, entryType, board.Move{})
	}

	return bestScore
}

// orderMoves orders moves using the provided thread state
func (m *MinimaxEngine) orderMoves(b *board.Board, moveList *moves.MoveList, depth int, ttMove board.Move, threadState *ThreadLocalState) {
	if moveList.Count <= 1 {
		return
	}

	// Reuse buffer if possible, only grow if needed to avoid excessive allocations
	if cap(threadState.moveOrderBuffer) < moveList.Count {
		threadState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		threadState.moveOrderBuffer = threadState.moveOrderBuffer[:moveList.Count]
	}

	// Score each move
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := 0

		// Transposition table move gets highest priority
		if move.From == ttMove.From && move.To == ttMove.To && move.Promotion == ttMove.Promotion {
			score = 3000000
		} else {
			// Score captures using the comprehensive getCaptureScore system
			if move.IsCapture {
				score = m.getCaptureScore(b, move, threadState)
			}

			// Score promotions
			if move.Promotion != board.Empty {
				switch move.Promotion {
				case board.WhiteQueen, board.BlackQueen:
					score += 9000
				case board.WhiteRook, board.BlackRook:
					score += 5000
				case board.WhiteBishop, board.BlackBishop, board.WhiteKnight, board.BlackKnight:
					score += 3000
				}
			}

			// Killer moves
			if !move.IsCapture && m.isKillerMove(move, depth, threadState) {
				score = 500000
			}

			// History heuristic for quiet moves
			if !move.IsCapture && move.Promotion == board.Empty {
				score += int(m.getHistoryScore(move, threadState))
			}
		}

		threadState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Sort moves by score (highest first)
	for i := 0; i < moveList.Count-1; i++ {
		for j := i + 1; j < moveList.Count; j++ {
			if threadState.moveOrderBuffer[j].score > threadState.moveOrderBuffer[i].score {
				// Swap buffer entries
				threadState.moveOrderBuffer[i], threadState.moveOrderBuffer[j] = threadState.moveOrderBuffer[j], threadState.moveOrderBuffer[i]
			}
		}
	}

	// Reuse reorder buffer if possible, only grow if needed
	if cap(threadState.reorderBuffer) < moveList.Count {
		threadState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		threadState.reorderBuffer = threadState.reorderBuffer[:moveList.Count]
	}

	// Reorder the actual moves based on sorted buffer
	for i := 0; i < moveList.Count; i++ {
		origIndex := threadState.moveOrderBuffer[i].index
		threadState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	// Copy back to original move list
	copy(moveList.Moves[:moveList.Count], threadState.reorderBuffer)
	
	// Populate debug move order if enabled
	if m.debugMoveOrdering {
		if cap(threadState.debugMoveOrder) < moveList.Count {
			threadState.debugMoveOrder = make([]board.Move, moveList.Count)
		} else {
			threadState.debugMoveOrder = threadState.debugMoveOrder[:moveList.Count]
		}
		copy(threadState.debugMoveOrder, moveList.Moves[:moveList.Count])
	}
}

// orderRootMovesWithDiversity applies thread-specific move ordering at the root position
// This is CRITICAL for Lazy SMP thread diversity - different threads explore different move orders
func (m *MinimaxEngine) orderRootMovesWithDiversity(b *board.Board, moveList *moves.MoveList, ttMove board.Move, threadState *ThreadLocalState) {
	// First, do standard move ordering
	m.orderMoves(b, moveList, 0, ttMove, threadState)
	
	// Now apply thread-specific variations for diversity
	if threadState == nil || threadState.threadID < 0 || moveList.Count <= 2 {
		return // No diversity for main thread or too few moves
	}
	
	// Different ordering strategies based on thread ID
	switch threadState.threadID % 4 {
	case 0:
		// Thread 0: Standard ordering (already done)
		return
		
	case 1:
		// Thread 1: Swap pairs of non-TT moves to explore different lines
		// Skip TT move if it exists (index 0)
		startIdx := 0
		if moveList.Count > 0 && (moveList.Moves[0].From == ttMove.From && moveList.Moves[0].To == ttMove.To) {
			startIdx = 1
		}
		
		for i := startIdx; i < moveList.Count-1; i += 2 {
			// Swap adjacent pairs
			moveList.Moves[i], moveList.Moves[i+1] = moveList.Moves[i+1], moveList.Moves[i]
		}
		
	case 2:
		// Thread 2: Reverse order of quiet moves (after captures/promotions)
		// Find where quiet moves start
		quietStart := -1
		for i := 0; i < moveList.Count; i++ {
			if !moveList.Moves[i].IsCapture && moveList.Moves[i].Promotion == board.Empty {
				quietStart = i
				break
			}
		}
		
		if quietStart > 0 && quietStart < moveList.Count-1 {
			// Reverse quiet moves
			for i, j := quietStart, moveList.Count-1; i < j; i, j = i+1, j-1 {
				moveList.Moves[i], moveList.Moves[j] = moveList.Moves[j], moveList.Moves[i]
			}
		}
		
	case 3:
		// Thread 3: Rotate moves (except TT move) to start with different move
		startIdx := 0
		if moveList.Count > 0 && (moveList.Moves[0].From == ttMove.From && moveList.Moves[0].To == ttMove.To) {
			startIdx = 1
		}
		
		if startIdx < moveList.Count-2 {
			// Rotate by thread ID positions
			rotations := (threadState.threadID / 4) % (moveList.Count - startIdx)
			for r := 0; r < rotations; r++ {
				// Save first move after TT
				tmp := moveList.Moves[startIdx]
				// Shift all moves left
				for i := startIdx; i < moveList.Count-1; i++ {
					moveList.Moves[i] = moveList.Moves[i+1]
				}
				// Put first move at end
				moveList.Moves[moveList.Count-1] = tmp
			}
		}
	}
	
}

// isKillerMove checks if move is a killer using provided thread state
func (m *MinimaxEngine) isKillerMove(move board.Move, depth int, threadState *ThreadLocalState) bool {
	if depth < 0 || depth >= MaxKillerDepth {
		return false
	}

	// Check both killer slots at this depth
	return (move.From == threadState.killerTable[depth][0].From && move.To == threadState.killerTable[depth][0].To) ||
		(move.From == threadState.killerTable[depth][1].From && move.To == threadState.killerTable[depth][1].To)
}

// storeKiller stores a killer move using provided thread state
func (m *MinimaxEngine) storeKiller(move board.Move, depth int, threadState *ThreadLocalState) {
	if depth < 0 || depth >= MaxKillerDepth {
		return
	}

	if m.isKillerMove(move, depth, threadState) {
		return
	}

	// Shift killers and store new one
	threadState.killerTable[depth][1] = threadState.killerTable[depth][0]
	threadState.killerTable[depth][0] = move
}

// orderCapturesWithThreadState orders captures using thread state
func (m *MinimaxEngine) orderCapturesWithThreadState(b *board.Board, moveList *moves.MoveList, threadState *ThreadLocalState) {
	if moveList.Count <= 1 {
		return
	}

	// Reuse buffer if possible, only grow if needed to avoid excessive allocations
	if cap(threadState.moveOrderBuffer) < moveList.Count {
		threadState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		threadState.moveOrderBuffer = threadState.moveOrderBuffer[:moveList.Count]
	}

	// Score each capture
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := m.seeWithThreadState(b, move, threadState)
		threadState.moveOrderBuffer[i] = moveScore{index: i, score: int(score)}
	}

	// Sort captures by SEE score (highest first)
	for i := 0; i < moveList.Count-1; i++ {
		for j := i + 1; j < moveList.Count; j++ {
			if threadState.moveOrderBuffer[j].score > threadState.moveOrderBuffer[i].score {
				threadState.moveOrderBuffer[i], threadState.moveOrderBuffer[j] = threadState.moveOrderBuffer[j], threadState.moveOrderBuffer[i]
			}
		}
	}

	// Reuse reorder buffer if possible, only grow if needed
	if cap(threadState.reorderBuffer) < moveList.Count {
		threadState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		threadState.reorderBuffer = threadState.reorderBuffer[:moveList.Count]
	}

	// Reorder the actual moves using reusable buffer
	for i := 0; i < moveList.Count; i++ {
		origIndex := threadState.moveOrderBuffer[i].index
		threadState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], threadState.reorderBuffer)
}

// seeWithThreadState performs SEE calculation directly (no caching)
func (m *MinimaxEngine) seeWithThreadState(b *board.Board, move board.Move, threadState *ThreadLocalState) ai.EvaluationScore {
	// Calculate SEE directly - no caching (following Stockfish approach)
	// SEE values are too position-dependent for effective caching
	return ai.EvaluationScore(m.seeCalculator.SEE(b, move))
}



// min and max helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getCaptureScore calculates the capture score using SEE for accurate evaluation
// Higher scores indicate more valuable captures (better moves to try first)
// Move ordering priorities:
//  1. TT moves: 3,000,000+
//  2. Good captures (SEE > 0): 1,000,000+
//  3. Equal exchanges (SEE = 0): 900,000
//  4. Killer moves: 500,000
//  5. Good history moves: up to ~50,000
//  6. Slightly bad captures (SEE >= -100): 100,000+
//  7. Terrible captures (SEE < -100): 50,000+
//  8. Quiet moves: 0
//

func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move, threadState *ThreadLocalState) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0 // Non-captures get score 0
	}

	// Calculate SEE directly - no caching (following Stockfish approach)
	// SEE values are too position-dependent for effective caching
	seeValue := m.seeCalculator.SEE(b, move)

	// Get victim value for MVV-LVA tiebreaker
	victimValue := evaluation.PieceValues[move.Captured]
	if victimValue < 0 {
		victimValue = -victimValue
	}

	// Get attacker value for LVA (Least Valuable Attacker) tiebreaker
	attackerValue := evaluation.PieceValues[move.Piece]
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	// MVV-LVA tiebreaker calculation:
	// Higher victim value = better (MVV)
	// Lower attacker value = better (LVA) 
	// Formula: (victimValue * 10 - attackerValue) ensures proper ordering
	mvvLvaScore := (victimValue * 10) - attackerValue

	// Convert SEE value to move ordering score with proper tactical priorities
	// Use MVV-LVA as tiebreaker when SEE values are equal
	if seeValue > 0 {
		// Good captures: highest priority after TT moves
		// Add small MVV-LVA bonus for tiebreaking (max bonus = 8999, so this won't change category)
		return 1000000 + seeValue + mvvLvaScore
	} else if seeValue == 0 {
		// Equal exchanges: high priority, above killers
		// MVV-LVA tiebreaker: prefer capturing more valuable pieces with less valuable attackers
		return 900000 + mvvLvaScore
	} else if seeValue >= -100 {
		// Slightly bad captures: below killers but above history
		// Still might be tactical (pins, discoveries, etc.)
		return 100000 + seeValue + 100 + mvvLvaScore // Ensures positive score
	} else {
		// Terrible captures: below history but above quiet moves
		// Could be sacrifices leading to mate or forcing sequences
		return 25000 + seeValue + 1000 + mvvLvaScore // Ensures positive score, below history
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
		return
	}
	m.transpositionTable = NewTranspositionTable(sizeMB)
}

// SetTranspositionTableEnabled enables or disables transposition table usage
func (m *MinimaxEngine) SetTranspositionTableEnabled(enabled bool) {
	if m.transpositionTable != nil {
	}
}

// GetTranspositionTableStats returns transposition table statistics if available
func (m *MinimaxEngine) GetTranspositionTableStats() (hits, misses, collisions uint64, hitRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetStats()
	}
	return 0, 0, 0, 0
}

// GetDetailedTranspositionTableStats returns detailed TT statistics including fill rate and average depth
func (m *MinimaxEngine) GetDetailedTranspositionTableStats() (hits, misses, collisions, filled, averageDepth uint64, hitRate, fillRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetDetailedStats()
	}
	return 0, 0, 0, 0, 0, 0, 0
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	// Clear all thread-local state properly
	m.threadStates.Range(func(key, value interface{}) bool {
		if state, ok := value.(*ThreadLocalState); ok {
			// Clear killer table
			for i := 0; i < MaxKillerDepth; i++ {
				state.killerTable[i][0] = board.Move{}
				state.killerTable[i][1] = board.Move{}
			}
			// Reset search statistics
			state.searchStats = ai.SearchStats{}
			// Clear move order buffer to prevent stale data
			state.moveOrderBuffer = make([]moveScore, 0, 256)
		}
		return true
	})

	if m.transpositionTable != nil {
		m.transpositionTable.Clear()
	}
	if m.historyTable != nil {
		m.historyTable.Clear()
	}
}

// SetDebugMoveOrdering enables/disables move ordering debug tracking (for tests only)
func (m *MinimaxEngine) SetDebugMoveOrdering(enabled bool) {
	m.debugMoveOrdering = enabled
	// Clear debug info from all thread states when disabled
	if !enabled {
		m.threadStates.Range(func(key, value interface{}) bool {
			if state, ok := value.(*ThreadLocalState); ok {
				state.debugMoveOrder = nil
			}
			return true
		})
	}
}

// GetLastMoveOrder returns the move order from the last orderMoves call (for tests only)
// Note: In multi-threaded mode, this returns the order from the current goroutine
func (m *MinimaxEngine) GetLastMoveOrder() []board.Move {
	threadState := m.getThreadLocalState()
	return threadState.debugMoveOrder
}

// getHistoryScore returns the history score for a move using thread-local history table
func (m *MinimaxEngine) getHistoryScore(move board.Move, threadState *ThreadLocalState) int32 {
	// Use thread-local history table for diversity
	if threadState != nil && threadState.historyTable != nil {
		return threadState.historyTable.GetHistoryScore(move)
	}
	// Fallback to shared history table if no thread state
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}

// calculateAdaptiveWindow determines the appropriate aspiration window size
// based on search depth and position characteristics
func (m *MinimaxEngine) calculateAdaptiveWindow(depth int, baseWindow ai.EvaluationScore) ai.EvaluationScore {
	// Start with base window and adjust based on depth
	adaptiveWindow := baseWindow

	// Increase window size for deeper searches as they tend to be more volatile
	if depth >= 6 {
		adaptiveWindow += ai.EvaluationScore(depth * 10)
	}

	// Cap the maximum window size to prevent excessive re-searches
	maxWindow := ai.EvaluationScore(200)
	if adaptiveWindow > maxWindow {
		adaptiveWindow = maxWindow
	}

	return adaptiveWindow
}

// moveGivesCheck uses direct bitboard calculation for check detection
// This is the idiomatic approach used by strong engines like Stockfish
func (m *MinimaxEngine) moveGivesCheck(b *board.Board, move board.Move) bool {
	// Get piece type
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == board.Empty {
		return false
	}

	// Determine enemy king position
	var enemyKingBB board.Bitboard
	if piece < board.BlackPawn { // White piece
		enemyKingBB = b.GetPieceBitboard(board.BlackKing)
	} else { // Black piece
		enemyKingBB = b.GetPieceBitboard(board.WhiteKing)
	}

	if enemyKingBB == 0 {
		return false
	}

	kingSquare := enemyKingBB.LSB()
	toSquare := move.To.Rank*8 + move.To.File
	fromSquare := move.From.Rank*8 + move.From.File

	// 1. Direct check: Does the piece attack the king from its destination?
	if m.isDirectCheck(b, piece, toSquare, kingSquare) {
		return true
	}

	// 2. Discovered check: Moving this piece might uncover an attack
	if m.isDiscoveredCheck(b, fromSquare, toSquare, kingSquare, piece) {
		return true
	}

	// 3. Promotion check (special case)
	if move.Promotion != board.Empty {
		return m.isDirectCheck(b, move.Promotion, toSquare, kingSquare)
	}

	// 4. En passant discovered check (rare but possible)
	if move.IsEnPassant {
		return m.isEnPassantCheck(b, move, kingSquare)
	}

	return false
}

// isDirectCheck checks if a piece on a square directly attacks the king
func (m *MinimaxEngine) isDirectCheck(b *board.Board, piece board.Piece, fromSquare, kingSquare int) bool {
	switch piece {
	case board.WhitePawn:
		pawnAttacks := board.GetPawnAttacks(fromSquare, board.BitboardWhite)
		return pawnAttacks.HasBit(kingSquare)
	case board.BlackPawn:
		pawnAttacks := board.GetPawnAttacks(fromSquare, board.BitboardBlack)
		return pawnAttacks.HasBit(kingSquare)

	case board.WhiteKnight, board.BlackKnight:
		return board.GetKnightAttacks(fromSquare).HasBit(kingSquare)

	case board.WhiteBishop, board.BlackBishop:
		// Use magic bitboards with current board occupancy
		return board.GetBishopAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case board.WhiteRook, board.BlackRook:
		return board.GetRookAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case board.WhiteQueen, board.BlackQueen:
		return board.GetQueenAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case board.WhiteKing, board.BlackKing:
		return board.GetKingAttacks(fromSquare).HasBit(kingSquare)
	}

	return false
}

// isDiscoveredCheck detects if moving a piece discovers a check
func (m *MinimaxEngine) isDiscoveredCheck(b *board.Board, fromSquare, toSquare, kingSquare int, movingPiece board.Piece) bool {
	// Quick rejection: if piece isn't between any of our sliders and enemy king, no discovered check
	fromRank, fromFile := fromSquare/8, fromSquare%8
	kingRank, kingFile := kingSquare/8, kingSquare%8

	// Check if on same line as king
	onSameRank := fromRank == kingRank
	onSameFile := fromFile == kingFile
	onSameDiagonal := (fromRank - kingRank) == (fromFile - kingFile)
	onSameAntiDiag := (fromRank - kingRank) == -(fromFile - kingFile)

	if !onSameRank && !onSameFile && !onSameDiagonal && !onSameAntiDiag {
		return false // Can't discover check
	}

	// Create occupancy with piece removed from source and placed at destination
	occupancy := b.AllPieces.ClearBit(fromSquare).SetBit(toSquare)

	// Check our sliding pieces that could give discovered check
	if movingPiece < board.BlackPawn { // White piece moving
		if onSameRank || onSameFile {
			// Check white rooks and queens
			rooksQueens := b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen)
			rooksQueens = rooksQueens.ClearBit(fromSquare) // Don't count moving piece

			for rooksQueens != 0 {
				attackerSq := rooksQueens.LSB()
				rooksQueens = rooksQueens.ClearBit(attackerSq)

				if board.GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}

		if onSameDiagonal || onSameAntiDiag {
			// Check white bishops and queens
			bishopsQueens := b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen)
			bishopsQueens = bishopsQueens.ClearBit(fromSquare)

			for bishopsQueens != 0 {
				attackerSq := bishopsQueens.LSB()
				bishopsQueens = bishopsQueens.ClearBit(attackerSq)

				if board.GetBishopAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}
	} else { // Black piece moving
		if onSameRank || onSameFile {
			// Check black rooks and queens
			rooksQueens := b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen)
			rooksQueens = rooksQueens.ClearBit(fromSquare)

			for rooksQueens != 0 {
				attackerSq := rooksQueens.LSB()
				rooksQueens = rooksQueens.ClearBit(attackerSq)

				if board.GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}

		if onSameDiagonal || onSameAntiDiag {
			// Check black bishops and queens
			bishopsQueens := b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen)
			bishopsQueens = bishopsQueens.ClearBit(fromSquare)

			for bishopsQueens != 0 {
				attackerSq := bishopsQueens.LSB()
				bishopsQueens = bishopsQueens.ClearBit(attackerSq)

				if board.GetBishopAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}
	}

	return false
}

// isEnPassantCheck checks if en passant capture discovers check
func (m *MinimaxEngine) isEnPassantCheck(b *board.Board, move board.Move, kingSquare int) bool {
	// En passant can discover check on the rank
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// Calculate captured pawn square
	capturedPawnSquare := toSquare
	if move.From.Rank > move.To.Rank { // White capturing
		capturedPawnSquare -= 8
	} else { // Black capturing
		capturedPawnSquare += 8
	}

	// Create occupancy after en passant
	occupancy := b.AllPieces.ClearBit(fromSquare).ClearBit(capturedPawnSquare).SetBit(toSquare)

	// Check if any rook/queen can now attack the king
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece < board.BlackPawn { // White piece
		rooksQueens := b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen)
		for rooksQueens != 0 {
			attackerSq := rooksQueens.LSB()
			rooksQueens = rooksQueens.ClearBit(attackerSq)

			if board.GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
				return true
			}
		}
	} else { // Black piece
		rooksQueens := b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen)
		for rooksQueens != 0 {
			attackerSq := rooksQueens.LSB()
			rooksQueens = rooksQueens.ClearBit(attackerSq)

			if board.GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
				return true
			}
		}
	}

	return false
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}

// moveToDebugString converts a move to string for debugging (handles invalid moves)
func moveToDebugString(move board.Move) string {
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 ||
	   move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return "INVALID"
	}
	from := string('a'+rune(move.From.File)) + string('1'+rune(move.From.Rank))
	to := string('a'+rune(move.To.File)) + string('1'+rune(move.To.Rank))
	return from + to
}
