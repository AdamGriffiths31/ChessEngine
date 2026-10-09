package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// stageTTHash is an arbitrary non-zero hash (zero marks an empty TT slot).
const stageTTHash uint64 = 0x123456789ABCDEF0

// stageMove is a quiet e2e4.
func stageMove() board.Move {
	return board.Move{
		From: board.Square{File: 4, Rank: 1}, // e2
		To:   board.Square{File: 4, Rank: 3}, // e4
	}
}

// newEngineWithTT builds an engine with a small transposition table.
func newEngineWithTT() *MinimaxEngine {
	m := NewMinimaxEngine()
	m.SetTranspositionTableSize(1)
	return m
}

// Null-move pruning must decline, without counting an attempt, when the side to
// move is in check, beta is near the mate bound, or only a king and pawns
// remain (zugzwang risk).
func TestTryNullMove_Declines(t *testing.T) {
	tests := []struct {
		name         string
		fen          string
		beta         eval.EvaluationScore
		inCheck      bool
		wantZugzwang int64 // expected NullMoveZugzwangSkipped
	}{
		{"in check", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 100, true, 0},
		{"beta at the mate bound", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", eval.MateScore, false, 0},
		{"no non-pawn material", "6k1/8/8/8/8/8/P7/K7 w - - 0 1", 0, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMinimaxEngine()
			b := testutil.MustFromFEN(t, tt.fen)
			m.searchState.board = b
			staticEval := m.evaluator.Evaluate(b)
			if tt.beta == eval.MateScore {
				staticEval = eval.MateScore
			}

			score, done := m.tryNullMove(context.Background(), b, movegen.White, 5, tt.beta, 1, 0, 1, staticEval, tt.inCheck)

			if done || score != 0 {
				t.Errorf("tryNullMove = (%d, done=%v), want it to decline", score, done)
			}
			stats := m.searchState.searchStats
			if stats.NullMoves != 0 {
				t.Errorf("NullMoves = %d, want 0", stats.NullMoves)
			}
			if stats.NullMoveZugzwangSkipped != tt.wantZugzwang {
				t.Errorf("NullMoveZugzwangSkipped = %d, want %d", stats.NullMoveZugzwangSkipped, tt.wantZugzwang)
			}
		})
	}
}

// Black is a queen down on a full board, so passing leaves it lost and the
// reduced-depth search fails high against a low beta. The board must be dense:
// sparse boards skip null-move cutoffs entirely (see params.go).
func TestTryNullMove_CutsOffWhenOpponentIsLost(t *testing.T) {
	m := NewMinimaxEngine()
	b := testutil.MustFromFEN(t, "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	m.searchState.board = b
	staticEval := m.evaluator.Evaluate(b)

	score, done := m.tryNullMove(context.Background(), b, movegen.White, 5, 0, 1, 0, 1, staticEval, false)

	if !done {
		t.Fatalf("expected a null-move cutoff, got done=false (score=%d)", score)
	}
	stats := m.searchState.searchStats
	if stats.NullCutoffs != 1 {
		t.Errorf("NullCutoffs = %d, want 1", stats.NullCutoffs)
	}
	if stats.CutNodes != 1 {
		t.Errorf("CutNodes = %d, want 1 (a null-move cutoff is a fail-high)", stats.CutNodes)
	}
}

func TestTryRazoring_DeclinesWhenDisabled(t *testing.T) {
	m := NewMinimaxEngine()
	m.searchState.searchParams.RazoringEnabled = false

	// Otherwise-qualifying conditions, but razoring is off: decline, no attempt counted.
	score, done := m.tryRazoring(context.Background(), movegen.White, 1, 100, 200, 1, 0, -500, false)

	if done {
		t.Fatalf("razoring must decline when disabled")
	}
	if score != 0 {
		t.Errorf("declined razoring score = %d, want 0", score)
	}
	if m.searchState.searchStats.RazoringAttempts != 0 {
		t.Errorf("RazoringAttempts incremented while disabled: %d", m.searchState.searchStats.RazoringAttempts)
	}
}

func TestTryRazoring_CutoffIncrementsAllNodes(t *testing.T) {
	m := NewMinimaxEngine()
	m.searchState.searchParams.RazoringEnabled = true
	// White is two queens down, so quiescence confirms a fail-low against a high alpha.
	b := testutil.MustFromFEN(t, "4k3/8/8/8/8/8/8/qq2K3 w - - 0 1")
	m.searchState.board = b
	m.searchState.player = movegen.White
	staticEval := m.evaluator.Evaluate(b)

	beforeCutoffs := m.searchState.searchStats.RazoringCutoffs
	score, done := m.tryRazoring(context.Background(), movegen.White, 1, eval.EvaluationScore(0), eval.EvaluationScore(2000), 1, 0, staticEval, false)

	if !done {
		t.Fatalf("expected a razoring cutoff, got done=false (score=%d)", score)
	}
	if m.searchState.searchStats.RazoringCutoffs != beforeCutoffs+1 {
		t.Errorf("RazoringCutoffs not incremented on cutoff")
	}
	if m.searchState.searchStats.AllNodes != 1 {
		t.Errorf("AllNodes = %d, want 1 (a razoring cutoff is a fail-low, same node type as a move-loop !alphaImproved node)", m.searchState.searchStats.AllNodes)
	}
}

func TestRecordCutoff_StoresKillerAndHistoryForQuietMove(t *testing.T) {
	m := newEngineWithTT()
	move := stageMove() // quiet, non-capture
	const ply = 3
	const depth = 4

	if m.getHistoryScore(move) != 0 {
		t.Fatalf("precondition: history score should start at 0")
	}

	list := &movegen.MoveList{}
	list.AddMove(move)
	m.recordCutoff(move, depth, ply, 1, stageTTHash, eval.EvaluationScore(50), list, 0)

	if !m.isKillerMove(move, ply) {
		t.Errorf("quiet move not stored as killer at ply %d", ply)
	}
	if m.getHistoryScore(move) <= 0 {
		t.Errorf("quiet move history not updated: %d", m.getHistoryScore(move))
	}
	if m.searchState.searchStats.TotalCutoffs != 1 {
		t.Errorf("TotalCutoffs = %d, want 1", m.searchState.searchStats.TotalCutoffs)
	}
	if m.searchState.searchStats.FirstMoveCutoffs != 1 {
		t.Errorf("FirstMoveCutoffs = %d, want 1 (legalMoveCount==1)", m.searchState.searchStats.FirstMoveCutoffs)
	}
	if m.searchState.searchStats.CutNodes != 1 {
		t.Errorf("CutNodes = %d, want 1", m.searchState.searchStats.CutNodes)
	}
}

func TestRecordCutoff_MalusesEarlierQuietMoves(t *testing.T) {
	m := newEngineWithTT()
	const ply = 3
	const depth = 4

	hero := stageMove() // quiet, non-capture
	earlier := board.Move{
		From:  board.Square{File: 0, Rank: 1}, // a2
		To:    board.Square{File: 0, Rank: 2}, // a3
		Piece: board.WhitePawn,
	}

	list := &movegen.MoveList{}
	list.AddMove(earlier)
	list.AddMove(hero)

	m.recordCutoff(hero, depth, ply, 2, stageTTHash, eval.EvaluationScore(50), list, 1)

	if m.getHistoryScore(hero) <= 0 {
		t.Errorf("cutoff move should earn a positive bonus, got %d", m.getHistoryScore(hero))
	}
	if m.getHistoryScore(earlier) >= 0 {
		t.Errorf("earlier quiet move should take a negative malus, got %d", m.getHistoryScore(earlier))
	}
}

func TestRecordCutoff_SkipsKillerAndHistoryForCapture(t *testing.T) {
	m := newEngineWithTT()
	move := stageMove()
	move.IsCapture = true // capture: killer/history must be skipped
	const ply = 3
	const depth = 4

	m.recordCutoff(move, depth, ply, 2, stageTTHash, eval.EvaluationScore(50), nil, 0)

	if m.isKillerMove(move, ply) {
		t.Errorf("capture must not be stored as killer")
	}
	if m.getHistoryScore(move) != 0 {
		t.Errorf("capture history should stay 0, got %d", m.getHistoryScore(move))
	}
	// Stats bookkeeping still runs for captures.
	if m.searchState.searchStats.TotalCutoffs != 1 {
		t.Errorf("TotalCutoffs = %d, want 1", m.searchState.searchStats.TotalCutoffs)
	}
	if m.searchState.searchStats.FirstMoveCutoffs != 0 {
		t.Errorf("FirstMoveCutoffs = %d, want 0 (legalMoveCount==2)", m.searchState.searchStats.FirstMoveCutoffs)
	}
}
