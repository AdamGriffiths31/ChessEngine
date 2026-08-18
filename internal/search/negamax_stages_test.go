package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// stageTTHash is an arbitrary non-zero hash used for stage TT fixtures. It must
// be non-zero because a zero Hash marks an empty transposition-table slot.
const stageTTHash uint64 = 0x123456789ABCDEF0

// stageMove is a valid, non-capture quiet move used across the stage tests.
func stageMove() board.Move {
	return board.Move{
		From: board.Square{File: 4, Rank: 1}, // e2
		To:   board.Square{File: 4, Rank: 3}, // e4
	}
}

// newEngineWithTT builds an engine with a small transposition table for probe
// and cutoff tests.
func newEngineWithTT() *MinimaxEngine {
	m := NewMinimaxEngine()
	m.SetTranspositionTableSize(1)
	return m
}

func TestProbeTT_ExactEntryCutoff(t *testing.T) {
	m := newEngineWithTT()
	move := stageMove()
	m.transpositionTable.Store(stageTTHash, 5, eval.EvaluationScore(100), EntryExact, move)

	beforeCutoffs := m.searchState.searchStats.TTCutoffs

	ttMove, score, alphaOut, betaOut, done := m.probeTT(stageTTHash, 3, 0, -1000, 1000)

	if !done {
		t.Fatalf("expected exact entry to produce a cutoff (done=true)")
	}
	if score != 100 {
		t.Errorf("score = %d, want 100", score)
	}
	if ttMove.From != move.From || ttMove.To != move.To {
		t.Errorf("ttMove = %v->%v, want %v->%v", ttMove.From, ttMove.To, move.From, move.To)
	}
	if alphaOut != -1000 || betaOut != 1000 {
		t.Errorf("bounds changed on exact cutoff: alpha=%d beta=%d", alphaOut, betaOut)
	}
	if m.searchState.searchStats.TTCutoffs != beforeCutoffs+1 {
		t.Errorf("TTCutoffs not incremented on exact cutoff")
	}
	if m.searchState.searchStats.PVNodes != 1 {
		t.Errorf("PVNodes = %d, want 1 (an exact TT entry resolves the node's true score, same as a PV node)", m.searchState.searchStats.PVNodes)
	}
}

func TestProbeTT_LowerBoundNarrowsAlpha(t *testing.T) {
	m := newEngineWithTT()
	m.transpositionTable.Store(stageTTHash, 5, eval.EvaluationScore(100), EntryLowerBound, stageMove())

	beforeCutoffs := m.searchState.searchStats.TTCutoffs

	// ttScore 100 is between alpha (0) and beta (1000): no cutoff, alpha raised.
	_, _, alphaOut, betaOut, done := m.probeTT(stageTTHash, 3, 0, 0, 1000)

	if done {
		t.Fatalf("lower-bound score inside the window must not cut off")
	}
	if alphaOut != 100 {
		t.Errorf("alphaOut = %d, want 100 (narrowed)", alphaOut)
	}
	if betaOut != 1000 {
		t.Errorf("betaOut = %d, want 1000 (unchanged)", betaOut)
	}
	if m.searchState.searchStats.TTCutoffs != beforeCutoffs {
		t.Errorf("TTCutoffs must not increment on narrowing")
	}
}

func TestProbeTT_LowerBoundCutoffAtBeta(t *testing.T) {
	m := newEngineWithTT()
	m.transpositionTable.Store(stageTTHash, 5, eval.EvaluationScore(2000), EntryLowerBound, stageMove())

	beforeCutoffs := m.searchState.searchStats.TTCutoffs

	// ttScore 2000 >= beta 1000: fail-high cutoff.
	_, score, _, _, done := m.probeTT(stageTTHash, 3, 0, 0, 1000)

	if !done || score != 2000 {
		t.Fatalf("expected lower-bound cutoff score=2000 done=true, got score=%d done=%v", score, done)
	}
	if m.searchState.searchStats.TTCutoffs != beforeCutoffs+1 {
		t.Errorf("TTCutoffs not incremented on lower-bound cutoff")
	}
	if m.searchState.searchStats.CutNodes != 1 {
		t.Errorf("CutNodes = %d, want 1 (a lower-bound cutoff is a fail-high, same node type as a move-loop beta cutoff)", m.searchState.searchStats.CutNodes)
	}
}

func TestProbeTT_UpperBoundNarrowsBeta(t *testing.T) {
	m := newEngineWithTT()
	m.transpositionTable.Store(stageTTHash, 5, eval.EvaluationScore(100), EntryUpperBound, stageMove())

	// ttScore 100 is above alpha (0) and below beta (1000): no cutoff, beta lowered.
	_, _, alphaOut, betaOut, done := m.probeTT(stageTTHash, 3, 0, 0, 1000)

	if done {
		t.Fatalf("upper-bound score inside the window must not cut off")
	}
	if betaOut != 100 {
		t.Errorf("betaOut = %d, want 100 (narrowed)", betaOut)
	}
	if alphaOut != 0 {
		t.Errorf("alphaOut = %d, want 0 (unchanged)", alphaOut)
	}
}

func TestProbeTT_UpperBoundCutoffAtAlpha(t *testing.T) {
	m := newEngineWithTT()
	m.transpositionTable.Store(stageTTHash, 5, eval.EvaluationScore(-2000), EntryUpperBound, stageMove())

	beforeAllNodes := m.searchState.searchStats.AllNodes

	// ttScore -2000 <= alpha 0: fail-low cutoff.
	_, score, _, _, done := m.probeTT(stageTTHash, 3, 0, 0, 1000)

	if !done || score != -2000 {
		t.Fatalf("expected upper-bound cutoff score=-2000 done=true, got score=%d done=%v", score, done)
	}
	if m.searchState.searchStats.AllNodes != beforeAllNodes+1 {
		t.Errorf("AllNodes = %d, want %d (an upper-bound cutoff is a fail-low, same node type as a move-loop !alphaImproved node)", m.searchState.searchStats.AllNodes, beforeAllNodes+1)
	}
}

func TestProbeTT_ShallowEntryDoesNotCutOff(t *testing.T) {
	m := newEngineWithTT()
	m.transpositionTable.Store(stageTTHash, 2, eval.EvaluationScore(100), EntryExact, stageMove())

	// Entry depth 2 < requested depth 5: only the move is usable, no cutoff.
	ttMove, _, alphaOut, betaOut, done := m.probeTT(stageTTHash, 5, 0, 0, 1000)

	if done {
		t.Fatalf("shallow entry must not cut off")
	}
	if !ttMove.IsValid() {
		t.Errorf("shallow entry should still return the TT move")
	}
	if alphaOut != 0 || betaOut != 1000 {
		t.Errorf("bounds changed on shallow entry: alpha=%d beta=%d", alphaOut, betaOut)
	}
}

func TestTryNullMove_DeclinesInCheck(t *testing.T) {
	m := NewMinimaxEngine()
	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	// Outer conditions satisfied (depth>=3, staticEval>=beta, beta in range) but
	// inCheck=true, so null move must decline without touching the null-move stat.
	score, done := m.tryNullMove(context.Background(), b, movegen.White, 5, 100, 1, 0, 200, true)

	if done {
		t.Fatalf("null move must decline when in check")
	}
	if score != 0 {
		t.Errorf("declined null move score = %d, want 0", score)
	}
	if m.searchState.searchStats.NullMoves != 0 {
		t.Errorf("NullMoves incremented while in check: %d", m.searchState.searchStats.NullMoves)
	}
}

func TestTryNullMove_DeclinesNearMateBound(t *testing.T) {
	m := NewMinimaxEngine()
	b := testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	// beta at MateScore is outside the (beta < MateScore-threshold) guard, so the
	// outer condition fails and null move is not attempted.
	score, done := m.tryNullMove(context.Background(), b, movegen.White, 5, eval.MateScore, 1, 0, eval.MateScore, false)

	if done {
		t.Fatalf("null move must decline near mate bounds")
	}
	if score != 0 {
		t.Errorf("declined null move score = %d, want 0", score)
	}
	if m.searchState.searchStats.NullMoves != 0 {
		t.Errorf("NullMoves incremented near mate bound: %d", m.searchState.searchStats.NullMoves)
	}
}

func TestTryNullMove_CutoffIncrementsCutNodes(t *testing.T) {
	m := NewMinimaxEngine()
	// White is up two queens with no other pieces on the board: giving Black a
	// free move (the null move) still leaves Black hopelessly lost, so the
	// reduced-depth verification search reliably fails high against a very
	// low beta.
	b := testutil.MustFromFEN(t, "6k1/8/8/8/8/8/8/QQ2K3 w - - 0 1")
	m.searchState.board = b
	staticEval := m.evaluator.Evaluate(b)

	beforeCutoffs := m.searchState.searchStats.NullCutoffs
	score, done := m.tryNullMove(context.Background(), b, movegen.White, 5, eval.EvaluationScore(0), 1, 0, staticEval, false)

	if !done {
		t.Fatalf("expected a null-move cutoff, got done=false (score=%d)", score)
	}
	if m.searchState.searchStats.NullCutoffs != beforeCutoffs+1 {
		t.Errorf("NullCutoffs not incremented on cutoff")
	}
	if m.searchState.searchStats.CutNodes != 1 {
		t.Errorf("CutNodes = %d, want 1 (a null-move cutoff is a fail-high, same node type as a move-loop beta cutoff)", m.searchState.searchStats.CutNodes)
	}
}

func TestTryRazoring_DeclinesWhenDisabled(t *testing.T) {
	m := NewMinimaxEngine()
	m.searchState.searchParams.RazoringEnabled = false

	// Conditions that would otherwise trigger razoring (low depth, staticEval far
	// below alpha), but razoring is disabled so it must decline before any
	// quiescence probe and without incrementing the attempt counter.
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
	// White is down two queens with no other pieces on the board: the
	// position is hopelessly lost, so the quiescence verification search
	// reliably confirms a fail-low against a generous alpha.
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

	m.recordCutoff(move, depth, ply, 1, stageTTHash, eval.EvaluationScore(50))

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

func TestRecordCutoff_SkipsKillerAndHistoryForCapture(t *testing.T) {
	m := newEngineWithTT()
	move := stageMove()
	move.IsCapture = true // capture: killer/history must be skipped
	const ply = 3
	const depth = 4

	m.recordCutoff(move, depth, ply, 2, stageTTHash, eval.EvaluationScore(50))

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
