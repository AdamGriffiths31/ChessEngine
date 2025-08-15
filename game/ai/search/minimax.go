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
	// MinEval represents the minimum possible evaluation score
	MinEval = ai.EvaluationScore(-1000000)
	// MaxKillerDepth is the maximum depth for killer move tables
	MaxKillerDepth = 64
	// MateDistanceThreshold is the threshold for detecting mate distances
	MateDistanceThreshold = 1000
)

// LMRTable is a pre-calculated reduction table for Late Move Reductions
// Indexed by [depth][moveCount] to get reduction amount
var LMRTable [16][64]int

func init() {
	for depth := 1; depth < 16; depth++ {
		for moveCount := 1; moveCount < 64; moveCount++ {
			LMRTable[depth][moveCount] = int(math.Log(float64(depth)) * math.Log(float64(moveCount)) / 1.8)
		}
	}
}

// ThreadSearchParams holds thread-specific search parameters for Lazy SMP diversity
type ThreadSearchParams struct {
	LMRDivisor           float64
	NullMoveReduction    int
	HistoryHighThreshold int32
	HistoryMedThreshold  int32
	HistoryLowThreshold  int32
}

// getThreadSearchParams returns thread-specific search parameters based on threadID
func getThreadSearchParams(threadID int) ThreadSearchParams {

	baseParams := ThreadSearchParams{
		LMRDivisor:           1.8,
		NullMoveReduction:    2,
		HistoryHighThreshold: 2000,
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,
	}

	if threadID == 0 || threadID == -1 {
		return baseParams
	}

	switch threadID % 4 {
	case 1:
		return ThreadSearchParams{
			LMRDivisor:           1.6,
			NullMoveReduction:    3,
			HistoryHighThreshold: 1500,
			HistoryMedThreshold:  300,
			HistoryLowThreshold:  -700,
		}
	case 2:
		return ThreadSearchParams{
			LMRDivisor:           2.2,
			NullMoveReduction:    2,
			HistoryHighThreshold: 2800,
			HistoryMedThreshold:  800,
			HistoryLowThreshold:  -300,
		}
	case 3:
		return ThreadSearchParams{
			LMRDivisor:           1.4,
			NullMoveReduction:    4,
			HistoryHighThreshold: 1200,
			HistoryMedThreshold:  200,
			HistoryLowThreshold:  -900,
		}
	default:
		return baseParams
	}
}

// ThreadLocalState contains state for single-threaded search
// Kept for backwards compatibility and single-threaded mode
type ThreadLocalState struct {
	killerTable     [MaxKillerDepth][2]board.Move
	moveOrderBuffer []moveScore
	reorderBuffer   []board.Move
	searchStats     ai.SearchStats
	searchParams    ThreadSearchParams
	debugMoveOrder  []board.Move
	historyTable    *HistoryTable
	threadID        int
}

// Cleanup performs any necessary cleanup operations for the engine
func (m *MinimaxEngine) Cleanup() {
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, SEE-based move ordering, and opening book support
// Now thread-safe with per-thread state management
type MinimaxEngine struct {
	evaluator          ai.Evaluator
	generator          *moves.Generator
	bookService        *openings.BookLookupService
	transpositionTable *TranspositionTable
	zobrist            *openings.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *evaluation.SEECalculator
	debugMoveOrdering  bool
	threadStates       sync.Map
	globalStats        ai.SearchStats
	statsMutex         sync.Mutex
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

	engine.initializeThreadStates(16)

	return engine
}

// initializeThreadStates pre-allocates thread-local state for parallel search
// This prevents expensive on-demand allocation during search and reduces contention
func (m *MinimaxEngine) initializeThreadStates(maxThreads int) {
	for threadID := 0; threadID < maxThreads; threadID++ {
		key := generateThreadKey(threadID)

		threadState := &ThreadLocalState{
			killerTable:     [MaxKillerDepth][2]board.Move{},
			moveOrderBuffer: make([]moveScore, 512),
			reorderBuffer:   make([]board.Move, 512),
			searchStats:     ai.SearchStats{},
			searchParams:    getThreadSearchParams(threadID),
			debugMoveOrder:  make([]board.Move, 0, 512),
			historyTable:    NewHistoryTable(),
			threadID:        threadID,
		}

		m.threadStates.Store(key, threadState)
	}

	mainKey := "thread_main"
	mainState := &ThreadLocalState{
		killerTable:     [MaxKillerDepth][2]board.Move{},
		moveOrderBuffer: make([]moveScore, 512),
		reorderBuffer:   make([]board.Move, 512),
		searchStats:     ai.SearchStats{},
		searchParams:    getThreadSearchParams(-1),
		historyTable:    NewHistoryTable(),
		threadID:        -1,
		debugMoveOrder:  make([]board.Move, 0, 512),
	}
	m.threadStates.Store(mainKey, mainState)
}

// fastCopyBoard creates an optimized deep copy of a board for thread isolation
func (m *MinimaxEngine) fastCopyBoard(original *board.Board) *board.Board {
	if original == nil {
		return nil
	}

	newBoard := &board.Board{}

	newBoard.SetCastlingRights(original.GetCastlingRights())
	newBoard.SetHalfMoveClock(original.GetHalfMoveClock())
	newBoard.SetFullMoveNumber(original.GetFullMoveNumber())
	newBoard.SetSideToMove(original.GetSideToMove())

	if epTarget := original.GetEnPassantTarget(); epTarget != nil {
		newTarget := &board.Square{
			File: epTarget.File,
			Rank: epTarget.Rank,
		}
		newBoard.SetEnPassantTarget(newTarget)
	}

	for i := 0; i < 12; i++ {
		newBoard.PieceBitboards[i] = original.PieceBitboards[i]
	}

	newBoard.WhitePieces = original.WhitePieces
	newBoard.BlackPieces = original.BlackPieces
	newBoard.AllPieces = original.AllPieces

	for i := 0; i < 64; i++ {
		newBoard.Mailbox[i] = original.Mailbox[i]
	}

	newBoard.SetHash(original.GetHash())
	newBoard.SetHashUpdater(m)

	return newBoard
}

// getThreadLocalState returns state for single-threaded search
// For multi-threaded search, workers will have their own WorkerState
func (m *MinimaxEngine) getThreadLocalState() *ThreadLocalState {
	const singleThreadKey = "single_thread"

	if state, exists := m.threadStates.Load(singleThreadKey); exists {
		threadState, ok := state.(*ThreadLocalState)
		if !ok {
			panic("invalid thread state type")
		}
		return threadState
	}

	newState := &ThreadLocalState{
		moveOrderBuffer: make([]moveScore, 256),
		debugMoveOrder:  make([]board.Move, 0),
	}

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
	key := generateThreadKey(threadID)

	if state, exists := m.threadStates.Load(key); exists {
		threadState, ok := state.(*ThreadLocalState)
		if !ok {
			panic("invalid thread state type")
		}
		return threadState
	}

	fmt.Printf("WARNING: Creating thread state on-demand for worker %d\n", threadID)
	newState := &ThreadLocalState{
		moveOrderBuffer: make([]moveScore, 256),
		searchParams:    getThreadSearchParams(threadID),
		debugMoveOrder:  make([]board.Move, 0),
	}

	m.threadStates.Store(key, newState)
	return newState
}

// getThreadLocalStateFromContext returns thread state based on context
// If context has threadID, returns thread-specific state; otherwise returns single-thread state
func (m *MinimaxEngine) getThreadLocalStateFromContext(ctx context.Context) *ThreadLocalState {
	if threadID, ok := ctx.Value("threadID").(int); ok {
		return m.getThreadSpecificState(threadID)
	}

	return m.getThreadLocalState()
}

// AggregateThreadStats collects statistics from all thread-local states
// Should be called after search completion for accurate reporting
func (m *MinimaxEngine) AggregateThreadStats() ai.SearchStats {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.globalStats = ai.SearchStats{}

	m.threadStates.Range(func(_, value interface{}) bool {
		threadState, ok := value.(*ThreadLocalState)
		if !ok {
			return true
		}
		stats := threadState.searchStats

		m.globalStats.NodesSearched += stats.NodesSearched
		m.globalStats.LMRReductions += stats.LMRReductions
		m.globalStats.LMRReSearches += stats.LMRReSearches
		m.globalStats.LMRNodesSkipped += stats.LMRNodesSkipped

		threadState.searchStats = ai.SearchStats{}

		return true
	})

	return m.globalStats
}

// GetHashDelta implements the board.HashUpdater interface
// Calculates the zobrist hash delta for a move from old state to new state
func (m *MinimaxEngine) GetHashDelta(b *board.Board, move board.Move, oldState board.State) uint64 {
	var hashDelta uint64

	hashDelta ^= m.zobrist.GetSideKey()

	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	if move.Piece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(move.Piece)
		hashDelta ^= m.zobrist.GetPieceKey(fromSquare, pieceIndex)
	}

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

	if move.IsCapture && move.Captured != board.Empty {
		capturedPieceIndex := m.zobrist.GetPieceIndex(move.Captured)
		if move.IsEnPassant {
			var captureRank int
			if move.Piece == board.WhitePawn {
				captureRank = 4
			} else {
				captureRank = 3
			}
			captureSquare := captureRank*8 + move.To.File
			hashDelta ^= m.zobrist.GetPieceKey(captureSquare, capturedPieceIndex)
		} else {
			hashDelta ^= m.zobrist.GetPieceKey(toSquare, capturedPieceIndex)
		}
	}

	if move.IsCastling {
		var rookFrom, rookTo int
		switch move.To.File {
		case 6:
			rookFrom = move.From.Rank*8 + 7
			rookTo = move.From.Rank*8 + 5
		case 2:
			rookFrom = move.From.Rank*8 + 0
			rookTo = move.From.Rank*8 + 3
		}

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

	if oldState.CastlingRights != b.GetCastlingRights() {
		oldRights := m.zobrist.GetCastlingKey(oldState.CastlingRights)
		newRights := m.zobrist.GetCastlingKey(b.GetCastlingRights())
		hashDelta ^= oldRights ^ newRights
	}

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
	var pawnRank int
	var pawnPiece board.Piece

	if sideToMove == "b" {
		pawnRank = 4
		pawnPiece = board.BlackPawn
	} else {
		pawnRank = 3
		pawnPiece = board.WhitePawn
	}

	epFile := epTarget.File

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
		return fmt.Errorf("failed to load opening books: %w", err)
	}

	m.bookService = service

	return nil
}

// FindBestMove searches for the best move using minimax with optional opening book
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	result := ai.SearchResult{
		Stats: ai.SearchStats{},
	}

	threadState := m.getThreadLocalStateFromContext(ctx)

	if config.UseOpeningBook {
		if err := m.initializeBookService(config); err != nil {
			fmt.Printf("Warning: failed to initialize opening book service: %v\n", err)
			m.bookService = nil
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

	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	startTime := time.Now()

	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

	if m.historyTable != nil {
		m.historyTable.Age()
	}

	if config.NumThreads > 1 {
		return m.lazySMPSearch(ctx, b, player, config)
	}

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
	if b == nil {
		return ai.SearchResult{
			BestMove: board.Move{},
			Score:    -ai.MateScore,
			Stats:    ai.SearchStats{},
		}
	}

	if config.UseOpeningBook {
		if err := m.initializeBookService(config); err != nil {
			fmt.Printf("Warning: failed to initialize opening book service: %v\n", err)
			m.bookService = nil
		}
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

	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}
	if m.historyTable != nil {
		m.historyTable.Age()
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		isCheck := m.generator.IsKingInCheck(b, player)
		panic(fmt.Sprintf("THREADING BUG: Main thread found no legal moves! Player=%v, InCheck=%v, FEN=%s",
			player, isCheck, b.ToFEN()))
	}

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

	for workerID := 0; workerID < config.NumThreads; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					fmt.Printf("[PANIC-worker-%d] Worker thread panicked: %v\n", id, r)
					panic(r)
				}
			}()

			workerBoard := m.fastCopyBoard(b)
			if workerBoard == nil {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - fastCopyBoard returned nil!", id))
			}

			threadState := m.getThreadSpecificState(id)
			if threadState == nil {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - getThreadSpecificState returned nil!", id))
			}

			originalFEN := b.ToFEN()
			workerFEN := workerBoard.ToFEN()
			if originalFEN != workerFEN {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - Board copy mismatch! Original: %s, Worker: %s",
					id, originalFEN, workerFEN))
			}

			if threadState.historyTable != nil {
				threadState.historyTable.Age()
			}

			result := m.runIterativeDeepening(ctx, workerBoard, player, config, threadState, startTime, id)

			if result.Stats.Depth == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - runIterativeDeepening returned depth 0! Nodes=%d, Score=%d",
					id, result.Stats.NodesSearched, result.Score))
			}

			if result.Stats.NodesSearched == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - runIterativeDeepening searched 0 nodes! Depth=%d, Score=%d",
					id, result.Stats.Depth, result.Score))
			}

			workerResult := lazySMPWorkerResult{
				bestMove: result.BestMove,
				score:    result.Score,
				depth:    result.Stats.Depth,
				stats:    result.Stats,
				workerID: id,
			}

			if workerResult.depth == 0 {
				panic(fmt.Sprintf("THREADING BUG: Worker %d - About to send result with depth 0!", id))
			}

			resultChan <- workerResult
		}(workerID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return m.selectBestLazySMPResult(resultChan, startTime)
}

// runIterativeDeepening runs the core iterative deepening search
// Used for both single-threaded search and individual worker threads in Lazy SMP
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig, threadState *ThreadLocalState, startTime time.Time, workerID ...int) ai.SearchResult {
	threadID := "main"
	if len(workerID) > 0 {
		threadID = fmt.Sprintf("worker-%d", workerID[0])
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		isCheck := m.generator.IsKingInCheck(b, player)
		panic(fmt.Sprintf("THREADING BUG: No legal moves found in thread %s! Player=%v, InCheck=%v, FEN=%s",
			threadID, player, isCheck, b.ToFEN()))
	}

	var rootTTMove board.Move
	if m.transpositionTable != nil {
		hash := b.GetHash()
		if entry, found := m.transpositionTable.Probe(hash); found {
			rootTTMove = entry.BestMove
		}
	}

	m.orderRootMovesWithDiversity(b, legalMoves, rootTTMove, threadState)

	lastCompletedBestMove := legalMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)
	lastCompletedDepth := 0
	var finalStats ai.SearchStats

	startingDepth := 1
	if threadState != nil && threadState.threadID >= 0 {
		if threadState.threadID%2 == 1 {
			startingDepth = 2
		}
	}

	for currentDepth := startingDepth; currentDepth <= config.MaxDepth; currentDepth++ {
		select {
		case <-ctx.Done():
			finalStats.Time = time.Since(startTime)
			finalStats.Depth = lastCompletedDepth
			return ai.SearchResult{
				BestMove: lastCompletedBestMove,
				Score:    lastCompletedScore,
				Stats:    finalStats,
			}
		default:
		}

		if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
			break
		}

		bestScore := -ai.MateScore - 1
		bestMove := legalMoves.Moves[0]

		alpha := -ai.MateScore - 1
		beta := ai.MateScore + 1

		// Use aspiration windows for depths > 1
		useAspirationWindow := currentDepth > 1
		window := ai.EvaluationScore(50)
		
		if useAspirationWindow {
			alpha = lastCompletedScore - window
			beta = lastCompletedScore + window
		}

		// Keep trying with wider windows until we get a score within bounds
		for {
			tempBestScore := -ai.MateScore - 1
			tempBestMove := board.Move{}
			moveIndex := 0

			for _, move := range legalMoves.Moves[:legalMoves.Count] {
				
				undo, err := b.MakeMoveWithUndo(move)
				if err != nil {
					continue
				}

				var score ai.EvaluationScore
				var moveStats ai.SearchStats

				if moveIndex == 0 {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, threadState, &moveStats)
				} else {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -alpha-1, -alpha, currentDepth, config, threadState, &moveStats)

					if score > alpha && score < beta {
						score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, threadState, &moveStats)
					}
				}

				b.UnmakeMove(undo)
				

				if score > tempBestScore {
					tempBestScore = score
					tempBestMove = move

					if score > alpha {
						alpha = score
					}
				}

				moveIndex++

				if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
					break
				}
			}
			
			// Check if we need to re-search with wider window
			if !useAspirationWindow || (tempBestScore > lastCompletedScore - window && tempBestScore < lastCompletedScore + window) {
				// Score is within aspiration window or we're not using aspiration
				bestScore = tempBestScore
				bestMove = tempBestMove
				break
			}
			
			// Aspiration window failed - widen and retry
			if tempBestScore <= lastCompletedScore - window {
				// Fail low - score is worse than expected
				window *= 2
				alpha = lastCompletedScore - window
				// Keep beta the same to avoid re-searching moves that already failed high
			} else if tempBestScore >= lastCompletedScore + window {
				// Fail high - score is better than expected  
				window *= 2
				beta = lastCompletedScore + window
				// Keep alpha the same
				alpha = lastCompletedScore - window/2
			}
			
			// Safety: if window gets too large, disable aspiration
			if window > 1000 {
				useAspirationWindow = false
				alpha = -ai.MateScore - 1
				beta = ai.MateScore + 1
			}
		}

		finalStats.NodesSearched = threadState.searchStats.NodesSearched
		finalStats.Depth = currentDepth
		finalStats.LMRReductions = threadState.searchStats.LMRReductions
		finalStats.LMRReSearches = threadState.searchStats.LMRReSearches
		finalStats.LMRNodesSkipped = threadState.searchStats.LMRNodesSkipped

		if config.MaxTime == 0 || time.Since(startTime) < config.MaxTime {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			lastCompletedDepth = currentDepth
		}

		if bestScore >= ai.MateScore-1000 || bestScore <= -ai.MateScore+1000 {
			break
		}
	}

	finalStats.Time = time.Since(startTime)
	finalStats.Depth = lastCompletedDepth

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

	moveVotes := make(map[board.Move]float64)
	moveResults := make(map[board.Move]lazySMPWorkerResult)
	allResults := make([]lazySMPWorkerResult, 0, 16)

	var minScore, maxScore ai.EvaluationScore
	var maxDepth int
	firstResult := true

	for result := range resultChan {
		resultsReceived++
		allResults = append(allResults, result)

		if result.depth == 0 {
			panic(fmt.Sprintf("THREADING BUG: Main thread received result with depth 0 from worker-%d!", result.workerID))
		}

		if result.stats.NodesSearched == 0 {
			panic(fmt.Sprintf("THREADING BUG: Main thread received result with 0 nodes from worker-%d!", result.workerID))
		}

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

		aggregatedStats.NodesSearched += result.stats.NodesSearched
		aggregatedStats.LMRReductions += result.stats.LMRReductions
		aggregatedStats.LMRReSearches += result.stats.LMRReSearches
		aggregatedStats.LMRNodesSkipped += result.stats.LMRNodesSkipped

		if existing, exists := moveResults[result.bestMove]; !exists ||
			result.depth > existing.depth ||
			(result.depth == existing.depth && result.score > existing.score) {
			moveResults[result.bestMove] = result
		}
	}

	scoreRange := maxScore - minScore
	if scoreRange == 0 {
		scoreRange = 1
	}

	for _, result := range allResults {

		depthWeight := float64(result.depth) / float64(maxDepth)

		scoreWeight := float64(result.score-minScore) / float64(scoreRange)

		nodeWeight := math.Log10(float64(result.stats.NodesSearched)+1) / 100.0

		voteWeight := depthWeight*0.6 + scoreWeight*0.35 + nodeWeight*0.05

		moveVotes[result.bestMove] += voteWeight

	}

	var bestMove board.Move
	var bestVotes float64
	for move, votes := range moveVotes {
		if votes > bestVotes {
			bestVotes = votes
			bestMove = move
		}
	}

	bestResult := moveResults[bestMove]

	aggregatedStats.Depth = maxDepth

	aggregatedStats.Time = time.Since(startTime)

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

	inCheck := m.generator.IsKingInCheck(b, player)
	if inCheck && depth < originalMaxDepth {
		depth++
	}

	var ttMove board.Move
	hash := b.GetHash()

	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			ttMove = entry.BestMove

			if ttMove.From.File >= 0 && ttMove.From.File <= 7 &&
				ttMove.To.File >= 0 && ttMove.To.File <= 7 {
				if ttMove.From == ttMove.To {
					ttMove = board.Move{}
				}
			} else {
				ttMove = board.Move{}
			}

			minDepthRequired := depth
			if config.NumThreads > 1 {
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

	staticEval := m.evaluator.Evaluate(b)
	if !config.DisableNullMove && depth >= 3 &&
		staticEval >= beta &&
		beta < ai.MateScore-MateDistanceThreshold &&
		beta > -ai.MateScore+MateDistanceThreshold {
		if !inCheck {
			threadState.searchStats.NullMoves++

			nullReduction := threadState.searchParams.NullMoveReduction
			if depth >= 6 && nullReduction < 3 {
				nullReduction++
			}

			nullUndo := b.MakeNullMove()

			nullScore := -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-nullReduction, -beta, -beta+1, originalMaxDepth, config, threadState, stats)

			b.UnmakeNullMove(nullUndo)

			if nullScore >= beta {
				if nullScore < ai.MateScore-MateDistanceThreshold {
					threadState.searchStats.NullCutoffs++
					return beta
				}
			}
		}
	}

	if depth <= 0 {
		return m.quiescence(ctx, b, player, alpha, beta, originalMaxDepth-depth, threadState, stats)
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

	m.orderMoves(b, legalMoves, currentDepth, ttMove, threadState)

	bestScore := -ai.MateScore - 1
	bestMove := board.Move{}
	entryType := EntryUpperBound
	moveCount := 0
	
	// Track if we improved alpha to determine correct entry type
	alphaImproved := false

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		moveCount++
		var score ai.EvaluationScore

		reduction := 0

		if depth >= config.LMRMinDepth &&
			moveCount > config.LMRMinMoves &&
			!inCheck &&
			!move.IsCapture &&
			move.Promotion == board.Empty &&
			!m.isKillerMove(move, currentDepth, threadState) {

			givesCheck := board.MoveGivesCheck(b, move)

			if !givesCheck {
				reduction = int(math.Log(float64(min(depth, 15))) * math.Log(float64(min(moveCount, 63))) / threadState.searchParams.LMRDivisor)

				historyScore := m.getHistoryScore(move, threadState)
				if historyScore > threadState.searchParams.HistoryHighThreshold {
					reduction = 0
				} else if historyScore > threadState.searchParams.HistoryMedThreshold && reduction > 0 {
					newReduction := reduction * 2 / 3
					if newReduction >= 0 {
						reduction = newReduction
					}
				} else if historyScore < threadState.searchParams.HistoryLowThreshold && reduction >= 0 {
					newReduction := reduction * 4 / 3
					if newReduction < depth {
						reduction = newReduction
					} else {
						reduction = depth - 1
					}
				}

				if reduction >= depth {
					reduction = depth - 1
				}
				if reduction < 0 {
					reduction = 0
				}
			}
		}

		if reduction > 0 {
			threadState.searchStats.LMRReductions++

			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-reduction, -alpha-1, -alpha, originalMaxDepth, config, threadState, stats)

			if score > alpha {
				threadState.searchStats.LMRReSearches++

				score = -m.negamax(ctx, b, oppositePlayer(player),
					depth-1, -beta, -alpha, originalMaxDepth, config, threadState, stats)
			}
		} else {
			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1, -beta, -alpha, originalMaxDepth, config, threadState, stats)
		}

		b.UnmakeMove(undo)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if alpha >= beta {
				if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
					m.storeKiller(move, currentDepth, threadState)
				}

				if !move.IsCapture {
					if threadState != nil && threadState.historyTable != nil {
						threadState.historyTable.UpdateHistory(move, depth)
					} else {
						m.historyTable.UpdateHistory(move, depth)
					}
				}

				if m.transpositionTable != nil {
					m.transpositionTable.Store(hash, depth, bestScore, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	if m.transpositionTable != nil {
		// Determine correct entry type based on bounds
		if alphaImproved {
			// We found a move that improved alpha
			if bestScore >= beta {
				// This shouldn't happen as we return early on beta cutoff
				entryType = EntryLowerBound
			} else {
				// Score is between original alpha and beta
				entryType = EntryExact
			}
		} else {
			// No move improved alpha - this is an upper bound
			entryType = EntryUpperBound
		}
		
		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
}

// quiescence performs quiescence search using provided thread state
func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, depthFromRoot int, threadState *ThreadLocalState, stats *ai.SearchStats) ai.EvaluationScore {
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

	eval := m.evaluator.Evaluate(b)
	if player == moves.Black {
		eval = -eval
	}

	if !inCheck {
		if eval >= beta {
			return beta
		}
		if eval > alpha {
			alpha = eval
		}
	}

	allMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(allMoves)

	captureList := moves.GetMoveList()
	defer moves.ReleaseMoveList(captureList)

	for i := 0; i < allMoves.Count; i++ {
		move := allMoves.Moves[i]
		if move.IsCapture || move.Promotion != board.Empty {
			captureList.AddMove(move)
		}
	}

	if captureList.Count == 0 {
		if inCheck {
			return -ai.MateScore + ai.EvaluationScore(depthFromRoot)
		}
		return eval
	}

	m.orderCapturesWithThreadState(b, captureList, threadState)

	bestScore := eval
	for i := 0; i < captureList.Count; i++ {
		move := captureList.Moves[i]

		if !inCheck {
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

			if move.Promotion != board.Empty {
				captureValue += 800
			}

			margin := ai.EvaluationScore(200)
			if eval+captureValue+margin < alpha {
				continue
			}
		}

		if !inCheck && move.IsCapture {
			if seeScore := m.seeWithThreadState(b, move); seeScore < 0 {
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

	if cap(threadState.moveOrderBuffer) < moveList.Count {
		threadState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		threadState.moveOrderBuffer = threadState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := 0

		if move.From == ttMove.From && move.To == ttMove.To && move.Promotion == ttMove.Promotion {
			score = 3000000
		} else {
			if move.IsCapture {
				score = m.getCaptureScore(b, move)
			}

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

			if !move.IsCapture && m.isKillerMove(move, depth, threadState) {
				score = 500000
			}

			if !move.IsCapture && move.Promotion == board.Empty {
				score += int(m.getHistoryScore(move, threadState))
			}
		}

		threadState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	for i := 0; i < moveList.Count-1; i++ {
		for j := i + 1; j < moveList.Count; j++ {
			if threadState.moveOrderBuffer[j].score > threadState.moveOrderBuffer[i].score {
				threadState.moveOrderBuffer[i], threadState.moveOrderBuffer[j] = threadState.moveOrderBuffer[j], threadState.moveOrderBuffer[i]
			}
		}
	}

	if cap(threadState.reorderBuffer) < moveList.Count {
		threadState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		threadState.reorderBuffer = threadState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := threadState.moveOrderBuffer[i].index
		threadState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], threadState.reorderBuffer)

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
	m.orderMoves(b, moveList, 0, ttMove, threadState)

	if threadState == nil || threadState.threadID < 0 || moveList.Count <= 2 {
		return
	}

	switch threadState.threadID % 4 {
	case 0:
		return

	case 1:
		startIdx := 0
		if moveList.Count > 0 && (moveList.Moves[0].From == ttMove.From && moveList.Moves[0].To == ttMove.To) {
			startIdx = 1
		}

		for i := startIdx; i < moveList.Count-1; i += 2 {
			moveList.Moves[i], moveList.Moves[i+1] = moveList.Moves[i+1], moveList.Moves[i]
		}

	case 2:
		quietStart := -1
		for i := 0; i < moveList.Count; i++ {
			if !moveList.Moves[i].IsCapture && moveList.Moves[i].Promotion == board.Empty {
				quietStart = i
				break
			}
		}

		if quietStart > 0 && quietStart < moveList.Count-1 {
			for i, j := quietStart, moveList.Count-1; i < j; i, j = i+1, j-1 {
				moveList.Moves[i], moveList.Moves[j] = moveList.Moves[j], moveList.Moves[i]
			}
		}

	case 3:
		startIdx := 0
		if moveList.Count > 0 && (moveList.Moves[0].From == ttMove.From && moveList.Moves[0].To == ttMove.To) {
			startIdx = 1
		}

		if startIdx < moveList.Count-2 {
			rotations := (threadState.threadID / 4) % (moveList.Count - startIdx)
			for r := 0; r < rotations; r++ {
				tmp := moveList.Moves[startIdx]
				for i := startIdx; i < moveList.Count-1; i++ {
					moveList.Moves[i] = moveList.Moves[i+1]
				}
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

	threadState.killerTable[depth][1] = threadState.killerTable[depth][0]
	threadState.killerTable[depth][0] = move
}

// orderCapturesWithThreadState orders captures using thread state
func (m *MinimaxEngine) orderCapturesWithThreadState(b *board.Board, moveList *moves.MoveList, threadState *ThreadLocalState) {
	if moveList.Count <= 1 {
		return
	}

	if cap(threadState.moveOrderBuffer) < moveList.Count {
		threadState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		threadState.moveOrderBuffer = threadState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := m.seeWithThreadState(b, move)
		threadState.moveOrderBuffer[i] = moveScore{index: i, score: int(score)}
	}

	for i := 0; i < moveList.Count-1; i++ {
		for j := i + 1; j < moveList.Count; j++ {
			if threadState.moveOrderBuffer[j].score > threadState.moveOrderBuffer[i].score {
				threadState.moveOrderBuffer[i], threadState.moveOrderBuffer[j] = threadState.moveOrderBuffer[j], threadState.moveOrderBuffer[i]
			}
		}
	}

	if cap(threadState.reorderBuffer) < moveList.Count {
		threadState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		threadState.reorderBuffer = threadState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := threadState.moveOrderBuffer[i].index
		threadState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], threadState.reorderBuffer)
}

// seeWithThreadState performs SEE calculation directly (no caching)
func (m *MinimaxEngine) seeWithThreadState(b *board.Board, move board.Move) ai.EvaluationScore {
	return ai.EvaluationScore(m.seeCalculator.SEE(b, move))
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
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	seeValue := m.seeCalculator.SEE(b, move)

	victimValue := evaluation.PieceValues[move.Captured]
	if victimValue < 0 {
		victimValue = -victimValue
	}

	attackerValue := evaluation.PieceValues[move.Piece]
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	mvvLvaScore := (victimValue * 10) - attackerValue

	if seeValue > 0 {
		return 1000000 + seeValue + mvvLvaScore
	} else if seeValue == 0 {
		return 900000 + mvvLvaScore
	} else if seeValue >= -100 {
		return 100000 + seeValue + 100 + mvvLvaScore
	}
	return 25000 + seeValue + 1000 + mvvLvaScore
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
	m.threadStates.Range(func(_, value interface{}) bool {
		if state, ok := value.(*ThreadLocalState); ok {
			for i := 0; i < MaxKillerDepth; i++ {
				state.killerTable[i][0] = board.Move{}
				state.killerTable[i][1] = board.Move{}
			}
			state.searchStats = ai.SearchStats{}
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
	if !enabled {
		m.threadStates.Range(func(_, value interface{}) bool {
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
	if threadState != nil && threadState.historyTable != nil {
		return threadState.historyTable.GetHistoryScore(move)
	}
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}
