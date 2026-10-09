// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// State contains transient state for a single search operation
type State struct {
	killerTable      [MaxKillerDepth][2]board.Move
	moveOrderBuffers [MaxKillerDepth][]moveScore // Per-ply buffers for move ordering
	searchStats      SearchStats
	searchParams     Params
	searchCancelled  bool

	// Per-search invariants, set at the start of FindBestMove/runIterativeDeepening.
	board     *board.Board   // Board being searched (single instance, moves made/unmade on it)
	player    movegen.Player // Side to move at the current node
	config    SearchConfig   // Search configuration for this search
	pv        *pvTable       // Principal variation storage (allocated per search)
	collectPV bool           // Whether the current node should record into pv
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, and SEE-based move ordering. Opening book probing
// is not part of this engine; callers (internal/player, internal/uci) consult
// internal/book.Prober before invoking FindBestMove.
type MinimaxEngine struct {
	evaluator          eval.Evaluator
	generator          *movegen.Generator
	transpositionTable *TranspositionTable
	zobrist            *book.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *eval.SEECalculator
	searchState        State

	// Repetition detection
	zobristHistory    [MaxGamePly]uint64
	zobristHistoryPly uint16
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	engine := &MinimaxEngine{
		evaluator:          eval.NewEvaluator(),
		generator:          movegen.NewGenerator(),
		transpositionTable: nil,
		zobrist:            book.GetPolyglotHash(),
		historyTable:       NewHistoryTable(),
		seeCalculator:      eval.NewSEECalculator(),
		searchState: State{
			killerTable:  [MaxKillerDepth][2]board.Move{},
			searchParams: getParams(),
		},
	}

	return engine
}

// tryMove makes move on the search board and verifies it is legal (does not
// leave the mover's own king in check). Illegal pseudo-legal moves unmake and
// return (undo, false) - an expected outcome of scanning pseudo-legal moves,
// not an error.
//
// player must be the side making move, captured in a local variable BEFORE
// this call: m.searchState.player is scratch state mutated by null-move
// pruning, razoring, and LMR research within a node, so a live read can name
// the wrong side.
//
// MakeMoveWithUndo errors mean the generator emitted a malformed move; that
// panics loudly rather than silently skipping, so movegen regressions surface.
func (m *MinimaxEngine) tryMove(move board.Move, player movegen.Player) (undo board.MoveUndo, ok bool) {
	b := m.searchState.board

	undo, err := b.MakeMoveWithUndo(move)
	if err != nil {
		panic(fmt.Sprintf("search: movegen produced an unplayable move %+v (player=%v) at FEN %q: %v", move, player, b.ToFEN(), err))
	}

	if m.generator.IsKingInCheck(b, player) {
		b.UnmakeMove(undo)
		return undo, false
	}

	return undo, true
}

// FindBestMove searches for the best move using minimax. Opening book
// probing is the caller's responsibility (see internal/book.Prober);
// this always performs a full search.
//
// searchStats is reset unconditionally every call - callers reuse one engine
// across a whole game's moves, and stats must describe exactly one search.
// The transposition and history tables intentionally persist across calls;
// ClearSearchState resets them for a genuinely new game.
//
// Repetition history resets per call too unless the caller supplies
// config.RepetitionHistory with the real game's hashes; without it the
// search cannot recognize repetitions that happened before this call
// (see that field's doc comment).
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player movegen.Player, config SearchConfig) SearchResult {
	m.searchState.searchStats = SearchStats{}

	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	if len(config.RepetitionHistory) > 0 {
		m.seedRepetitionHistory(config.RepetitionHistory)
	} else {
		m.setupRepetitionHistory(b.GetHash())
	}

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
func (m *MinimaxEngine) SetEvaluator(eval eval.Evaluator) {
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

// TTStats returns aggregated transposition table statistics if a
// transposition table is configured; otherwise it returns the zero value.
func (m *MinimaxEngine) TTStats() TTStats {
	if m.transpositionTable != nil {
		return m.transpositionTable.Stats()
	}
	return TTStats{}
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
		m.searchState.moveOrderBuffers[i] = nil
	}
	m.searchState.searchStats = SearchStats{}
	m.searchState.searchCancelled = false

	m.searchState.board = nil
	m.searchState.player = movegen.White
	m.searchState.config = SearchConfig{}
	m.searchState.pv = nil
	m.searchState.collectPV = false

	// Clear repetition history
	m.zobristHistoryPly = 0

	if m.transpositionTable != nil {
		m.transpositionTable.Clear()
	}
	if m.historyTable != nil {
		m.historyTable.Clear()
	}
}
