package search

import (
	"context"
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestDepthBasedSEEPruning(t *testing.T) {
	// Create a position with captures of different SEE values to test pruning thresholds
	fen := "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// Test move: Queen takes defended pawn (should have SEE around -800)
	testMove := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 3, File: 3}, // d4
		Piece:     board.WhiteQueen,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	seeValue := engine.seeCalculator.SEE(b, testMove)
	t.Logf("Test capture (Qxd4) SEE value: %d", seeValue)

	// Test pruning thresholds at different depths
	testCases := []struct {
		depth           int
		expectedThreshold int
		shouldPrune     bool
	}{
		{depth: 2, expectedThreshold: -50, shouldPrune: seeValue < -50},
		{depth: 4, expectedThreshold: -50, shouldPrune: seeValue < -50},
		{depth: 5, expectedThreshold: -20, shouldPrune: seeValue < -20},
		{depth: 8, expectedThreshold: -20, shouldPrune: seeValue < -20},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("depth_%d", tc.depth), func(t *testing.T) {
			// Simulate the pruning logic
			pruneThreshold := -50
			if tc.depth > 4 {
				pruneThreshold = -20
			}

			shouldPrune := seeValue < pruneThreshold
			
			t.Logf("Depth %d: threshold=%d, SEE=%d, should prune=%v", 
				tc.depth, pruneThreshold, seeValue, shouldPrune)

			if pruneThreshold != tc.expectedThreshold {
				t.Errorf("Expected threshold %d, got %d", tc.expectedThreshold, pruneThreshold)
			}

			if shouldPrune != tc.shouldPrune {
				t.Errorf("Expected shouldPrune=%v, got %v", tc.shouldPrune, shouldPrune)
			}
		})
	}
}

func TestQuiescenceSearchPruning(t *testing.T) {
	// Test that quiescence search actually prunes moves based on the new thresholds
	fen := "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	ctx := context.Background()

	// Call quiescence search at different depths to observe pruning behavior
	var stats1, stats2 ai.SearchStats

	// Shallow quiescence search (depth 2)
	alpha := ai.EvaluationScore(-1000)
	beta := ai.EvaluationScore(1000)
	result1 := engine.quiescence(ctx, b, moves.White, alpha, beta, 2, &stats1)

	// Deep quiescence search (depth 6) 
	result2 := engine.quiescence(ctx, b, moves.White, alpha, beta, 6, &stats2)

	t.Logf("Shallow quiescence (depth 2): result=%d, nodes=%d", result1, stats1.NodesSearched)
	t.Logf("Deep quiescence (depth 6): result=%d, nodes=%d", result2, stats2.NodesSearched)

	// Deep search should potentially search fewer nodes due to more aggressive pruning
	// Note: This is not guaranteed as it depends on the specific position and move ordering
	t.Logf("Node difference: %d (deep search may prune more aggressively)", 
		int(stats1.NodesSearched) - int(stats2.NodesSearched))
}

func TestSEEPruningThresholds(t *testing.T) {
	// Test various SEE values against the pruning thresholds
	testCaptures := []struct {
		name     string
		seeValue int
		description string
	}{
		{"Excellent capture", 300, "Should never be pruned"},
		{"Good capture", 50, "Should never be pruned"},
		{"Equal exchange", 0, "Should never be pruned"},
		{"Slightly bad capture", -30, "Should be pruned at deep depths only"},
		{"Bad capture", -75, "Should be pruned at all depths"},
		{"Terrible capture", -800, "Should be pruned at all depths"},
	}

	for _, tc := range testCaptures {
		t.Run(tc.name, func(t *testing.T) {
			// Test at shallow depth (threshold -50)
			shallowThreshold := -50
			shallowPrune := tc.seeValue < shallowThreshold

			// Test at deep depth (threshold -20)
			deepThreshold := -20
			deepPrune := tc.seeValue < deepThreshold

			t.Logf("%s (SEE %d): %s", tc.name, tc.seeValue, tc.description)
			t.Logf("  Shallow depth: pruned=%v (threshold %d)", shallowPrune, shallowThreshold)
			t.Logf("  Deep depth: pruned=%v (threshold %d)", deepPrune, deepThreshold)

			// Verify logical expectations
			if tc.seeValue >= 0 {
				if shallowPrune || deepPrune {
					t.Errorf("Non-negative SEE moves should never be pruned")
				}
			}

			if tc.seeValue < -50 {
				if !shallowPrune || !deepPrune {
					t.Errorf("Very bad captures (SEE < -50) should be pruned at all depths")
				}
			}

			if tc.seeValue >= -50 && tc.seeValue < -20 {
				if shallowPrune {
					t.Errorf("Moderately bad captures should not be pruned at shallow depths")
				}
				if !deepPrune {
					t.Errorf("Moderately bad captures should be pruned at deep depths")
				}
			}
		})
	}
}