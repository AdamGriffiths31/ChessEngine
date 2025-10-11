// Package search provides chess move search algorithms and transposition table implementation.
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

// State contains transient state for a single search operation
type State struct {
	killerTable     [MaxKillerDepth][2]board.Move
	moveOrderBuffer []moveScore
	reorderBuffer   []board.Move
	searchStats     ai.SearchStats
	searchParams    Params
	searchCancelled bool
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, SEE-based move ordering, and opening book support
type MinimaxEngine struct {
	evaluator          ai.Evaluator
	generator          *moves.Generator
	bookService        *openings.BookLookupService
	transpositionTable *TranspositionTable
	zobrist            *openings.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *evaluation.SEECalculator
	searchState        State

	// Repetition detection
	zobristHistory    [MaxGamePly]uint64
	zobristHistoryPly uint16
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
		searchState: State{
			killerTable:     [MaxKillerDepth][2]board.Move{},
			moveOrderBuffer: make([]moveScore, 0, 512),
			reorderBuffer:   make([]board.Move, 0, 512),
			searchParams:    getParams(),
		},
	}

	return engine
}

// initializeBookService initializes the opening book service with simple configuration
func (m *MinimaxEngine) initializeBookService(config ai.SearchConfig) error {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		m.bookService = nil
		return nil
	}

	bookConfig := openings.BookConfig{
		Enabled:       true,
		BookFiles:     config.BookFiles,
		SelectionMode: openings.SelectBest, // Always pick best move
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
	// Initialize book service only once if not already done and still in opening phase
	moveNumber := b.GetFullMoveNumber()
	if config.UseOpeningBook && m.bookService == nil && moveNumber <= 10 {
		if err := m.initializeBookService(config); err != nil {
			fmt.Printf("Warning: failed to initialize opening book service: %v\n", err)
		}
	}

	// Check opening book only in first 10 moves
	if config.UseOpeningBook && m.bookService != nil && moveNumber <= 10 {
		bookMove, err := m.bookService.FindBookMove(b)
		if err == nil && bookMove != nil {
			return ai.SearchResult{
				BestMove: *bookMove,
				Score:    0,
				Stats:    ai.SearchStats{BookMoveUsed: true},
			}
		}
	}

	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	// Setup repetition detection with root position
	m.setupRepetitionHistory(b.GetHash())

	startTime := time.Now()

	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

	if m.historyTable != nil {
		m.historyTable.Age()
	}

	return m.runIterativeDeepening(ctx, b, player, config, startTime)
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

// GetTwoBucketStats returns statistics specific to the two-bucket collision resolution
func (m *MinimaxEngine) GetTwoBucketStats() (secondBucketUse uint64, secondBucketRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetTwoBucketStats()
	}
	return 0, 0
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	for i := 0; i < MaxKillerDepth; i++ {
		m.searchState.killerTable[i][0] = board.Move{}
		m.searchState.killerTable[i][1] = board.Move{}
	}
	m.searchState.searchStats = ai.SearchStats{}
	m.searchState.moveOrderBuffer = make([]moveScore, 0, 256)
	m.searchState.searchCancelled = false

	// Clear repetition history
	m.zobristHistoryPly = 0

	if m.transpositionTable != nil {
		m.transpositionTable.Clear()
	}
	if m.historyTable != nil {
		m.historyTable.Clear()
	}
}
