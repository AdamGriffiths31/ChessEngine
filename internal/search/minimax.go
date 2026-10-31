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
	moveOrderBuffers [MaxKillerDepth][]int
	searchStats      Stats
	searchParams     Params
	searchCancelled  bool

	staticEvalStack [MaxKillerDepth]eval.EvaluationScore

	board  *board.Board
	config Config
	pv     *pvTable
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, and SEE-based move ordering. Opening book probing
// is not part of this engine; callers (internal/uci) consult
// internal/book.Prober before invoking FindBestMove.
type MinimaxEngine struct {
	evaluator          eval.Evaluator
	generator          *movegen.Generator
	transpositionTable *TranspositionTable
	zobrist            *book.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *eval.SEECalculator
	searchState        State

	zobristHistory    [MaxGamePly]uint64
	zobristHistoryPly uint16
	repetitionRootPly uint16
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	engine := &MinimaxEngine{
		evaluator:     eval.NewEvaluator(),
		generator:     movegen.NewGenerator(),
		zobrist:       book.GetPolyglotHash(),
		historyTable:  NewHistoryTable(),
		seeCalculator: eval.NewSEECalculator(),
		searchState:   State{searchParams: getParams()},
	}

	return engine
}

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
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player movegen.Player, config Config) Result {
	if config.MaxTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.MaxTime)
		defer cancel()
	}

	m.searchState.searchStats = Stats{}

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

	m.historyTable.Age()

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

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	m.searchState.killerTable = [MaxKillerDepth][2]board.Move{}
	m.searchState.moveOrderBuffers = [MaxKillerDepth][]int{}
	m.searchState.searchStats = Stats{}
	m.searchState.searchCancelled = false

	m.searchState.board = nil
	m.searchState.config = Config{}
	m.searchState.pv = nil

	m.zobristHistoryPly = 0
	m.repetitionRootPly = 0

	if m.transpositionTable != nil {
		m.transpositionTable.Clear()
	}
	m.historyTable.Clear()
}
