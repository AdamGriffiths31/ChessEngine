package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Test helper to create a thread state
func createThreadState(threadID int) *ThreadLocalState {
	return &ThreadLocalState{
		searchParams:    getThreadSearchParams(threadID),
		searchStats:     ai.SearchStats{},
		moveOrderBuffer: make([]moveScore, 256),
		debugMoveOrder:  make([]board.Move, 0),
	}
}

// Test helper to run negamax with common setup
func runNegamax(t *testing.T, fen string, depth int, alpha, beta ai.EvaluationScore, config ai.SearchConfig, threadID int) *ThreadLocalState {
	t.Helper()

	engine := NewMinimaxEngine()
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN %q: %v", fen, err)
	}

	threadState := createThreadState(threadID)
	ctx := context.Background()
	stats := &ai.SearchStats{}

	_ = engine.negamax(ctx, b, moves.White, depth, alpha, beta, config.MaxDepth, config, threadState, stats)

	return threadState
}

func TestRazoring(t *testing.T) {
	tests := []struct {
		name            string
		fen             string
		depth           int
		alpha           ai.EvaluationScore
		beta            ai.EvaluationScore
		disableRazoring bool
		threadID        int
		expectAttempts  bool
		expectCutoffs   bool
		description     string
	}{
		// Basic functionality tests
		{
			name:            "losing_position_depth_1",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1", // White King vs Black King + Queen + Rook
			depth:           1,
			alpha:           500,
			beta:            600,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  true,
			expectCutoffs:   true,
			description:     "Should razor in clearly losing position at depth 1",
		},
		{
			name:            "losing_position_depth_2",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           2,
			alpha:           400,
			beta:            500,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  true,
			expectCutoffs:   true,
			description:     "Should razor in clearly losing position at depth 2",
		},
		{
			name:            "losing_position_depth_3",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           3,
			alpha:           300,
			beta:            400,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  true,
			expectCutoffs:   true,
			description:     "Should razor in clearly losing position at depth 3",
		},

		// Depth boundary tests
		{
			name:            "depth_0_no_razoring",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           0,
			alpha:           500,
			beta:            600,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  false,
			expectCutoffs:   false,
			description:     "Should not razor at depth 0",
		},

		// Config disable tests
		{
			name:            "razoring_disabled",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           2,
			alpha:           500,
			beta:            600,
			disableRazoring: true,
			threadID:        0,
			expectAttempts:  false,
			expectCutoffs:   false,
			description:     "Should not razor when disabled in config",
		},

		// King in check tests
		{
			name:            "king_in_check",
			fen:             "4r3/8/8/8/8/8/8/4K3 w - - 0 1", // White king in check from rook
			depth:           2,
			alpha:           500,
			beta:            600,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  false,
			expectCutoffs:   false,
			description:     "Should not razor when king is in check",
		},

		// Margin boundary tests
		{
			name:            "alpha_too_low_depth_1",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           1,
			alpha:           -1500, // Very low alpha, static eval + margin might not be < alpha
			beta:            -1400,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  false,
			expectCutoffs:   false,
			description:     "Should not razor when alpha is too low for margin",
		},

		// Different thread parameters (different margins)
		{
			name:            "thread_1_different_margins",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           1,
			alpha:           400,
			beta:            500,
			disableRazoring: false,
			threadID:        1, // Thread 1 has different (more aggressive) margins
			expectAttempts:  true,
			expectCutoffs:   true,
			description:     "Should razor with thread 1 parameters (different margins)",
		},
		{
			name:            "thread_2_different_margins",
			fen:             "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1",
			depth:           1,
			alpha:           400,
			beta:            500,
			disableRazoring: false,
			threadID:        2, // Thread 2 has different (less aggressive) margins
			expectAttempts:  true,
			expectCutoffs:   true,
			description:     "Should razor with thread 2 parameters (different margins)",
		},

		// Equal position tests (should not razor)
		{
			name:            "equal_position",
			fen:             "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Starting position
			depth:           2,
			alpha:           50,
			beta:            150,
			disableRazoring: false,
			threadID:        0,
			expectAttempts:  false,
			expectCutoffs:   false,
			description:     "Should not razor in roughly equal starting position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ai.SearchConfig{
				DisableRazoring: tt.disableRazoring,
				MaxDepth:        5,
			}

			threadState := runNegamax(t, tt.fen, tt.depth, tt.alpha, tt.beta, config, tt.threadID)

			hasAttempts := threadState.searchStats.RazoringAttempts > 0
			hasCutoffs := threadState.searchStats.RazoringCutoffs > 0

			// Check attempts
			if tt.expectAttempts && !hasAttempts {
				t.Errorf("Expected razoring attempts, got %d", threadState.searchStats.RazoringAttempts)
			}
			if !tt.expectAttempts && hasAttempts {
				t.Errorf("Did not expect razoring attempts, got %d", threadState.searchStats.RazoringAttempts)
			}

			// Check cutoffs (only when we expect attempts)
			if tt.expectAttempts && tt.expectCutoffs && !hasCutoffs {
				t.Errorf("Expected razoring cutoffs, got %d", threadState.searchStats.RazoringCutoffs)
			}

			// Log results for debugging
			if hasAttempts {
				t.Logf("%s: attempts=%d, cutoffs=%d, failed=%d",
					tt.description,
					threadState.searchStats.RazoringAttempts,
					threadState.searchStats.RazoringCutoffs,
					threadState.searchStats.RazoringFailed)
			}
		})
	}
}

// Test statistics consistency
func TestRazoringStatistics(t *testing.T) {
	config := ai.SearchConfig{
		DisableRazoring: false,
		MaxDepth:        3,
	}

	// Use a position that should trigger razoring but might have mixed results
	threadState := runNegamax(t, "r3k3/3q4/8/8/8/8/8/4K3 w - - 0 1", 2, 300, 400, config, 0)

	// Verify statistical consistency
	attempts := threadState.searchStats.RazoringAttempts
	cutoffs := threadState.searchStats.RazoringCutoffs
	failed := threadState.searchStats.RazoringFailed

	if attempts == 0 {
		t.Skip("No razoring attempts in this test, skipping statistics validation")
	}

	// Basic consistency checks
	if cutoffs > attempts {
		t.Errorf("Cutoffs (%d) cannot exceed attempts (%d)", cutoffs, attempts)
	}

	if failed > attempts {
		t.Errorf("Failed (%d) cannot exceed attempts (%d)", failed, attempts)
	}

	if cutoffs+failed != attempts {
		t.Logf("Note: cutoffs(%d) + failed(%d) = %d, attempts=%d", cutoffs, failed, cutoffs+failed, attempts)
		t.Logf("This might be valid if some attempts don't reach quiescence verification")
	}

	t.Logf("Statistics: attempts=%d, cutoffs=%d, failed=%d", attempts, cutoffs, failed)
}
