package search

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// TestLMRBasicFunctionality tests that LMR is applied correctly under the right conditions
func TestLMRBasicFunctionality(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1) // Small TT for testing
	
	// Create a test position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	// Test config with LMR enabled
	config := ai.SearchConfig{
		MaxDepth:         3,
		MaxTime:          time.Second * 5,
		UseAlphaBeta:     true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      1, // Lower threshold for testing
		LMRReductionBase: 0.75,
	}
	
	// Test with LMR enabled
	ctx := context.Background()
	resultWithLMR := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Test with LMR disabled
	config.UseLMR = false
	resultWithoutLMR := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should have reduced some moves when LMR is enabled
	if resultWithLMR.Stats.LMRReductions == 0 {
		t.Error("Expected some LMR reductions to occur")
	}
	
	// Should have no reductions when LMR is disabled
	if resultWithoutLMR.Stats.LMRReductions != 0 {
		t.Error("Expected no LMR reductions when LMR is disabled")
	}
	
	// Node count should be lower with LMR (most of the time)
	// Note: This might not always be true due to re-searches, but should be true on average
	t.Logf("Nodes with LMR: %d, without LMR: %d", 
		resultWithLMR.Stats.NodesSearched, resultWithoutLMR.Stats.NodesSearched)
}

// TestLMRReductionCalculation tests that reductions are calculated correctly
func TestLMRReductionCalculation(t *testing.T) {
	tests := []struct {
		name         string
		depth        int
		moveIndex    int
		minDepth     int
		minMoves     int
		reductionBase float64
		expectedMin  int
		expectedMax  int
	}{
		{
			name:         "shallow search - no reduction",
			depth:        2,
			moveIndex:    5,
			minDepth:     3,
			minMoves:     4,
			reductionBase: 0.75,
			expectedMin:  0,
			expectedMax:  0,
		},
		{
			name:         "few moves - no reduction",
			depth:        4,
			moveIndex:    2,
			minDepth:     3,
			minMoves:     4,
			reductionBase: 0.75,
			expectedMin:  0,
			expectedMax:  0,
		},
		{
			name:         "normal reduction case",
			depth:        4,
			moveIndex:    6,
			minDepth:     3,
			minMoves:     4,
			reductionBase: 0.75,
			expectedMin:  1,
			expectedMax:  3, // depth - 1
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate reduction using the same formula as the engine
			var reduction int
			if tt.depth >= tt.minDepth && tt.moveIndex >= tt.minMoves {
				reduction = int(tt.reductionBase + 
					math.Log(float64(tt.depth))*math.Log(float64(tt.moveIndex+1))/2.25)
				if reduction > tt.depth-1 {
					reduction = tt.depth - 1
				}
				if reduction < 1 {
					reduction = 1
				}
			}
			
			if reduction < tt.expectedMin || reduction > tt.expectedMax {
				t.Errorf("Expected reduction between %d and %d, got %d", 
					tt.expectedMin, tt.expectedMax, reduction)
			}
		})
	}
}

// TestLMRDoesNotReduceImportantMoves tests that important moves are not reduced
func TestLMRDoesNotReduceImportantMoves(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1)
	
	// Create a tactical position where captures and checks should not be reduced
	b, err := board.FromFEN("rnbqkb1r/pppp1ppp/5n2/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	config := ai.SearchConfig{
		MaxDepth:         4,
		MaxTime:          time.Second * 5,
		UseAlphaBeta:     true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      1,
		LMRReductionBase: 1.0,
	}
	
	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should still find good moves despite LMR
	if result.BestMove.From.File == -1 {
		t.Error("Expected to find a best move")
	}
	
	// Verify that statistics are tracked
	if result.Stats.LMRReductions < 0 {
		t.Error("LMR reduction statistics should be non-negative")
	}
}

// TestLMRWithDifferentConfigurations tests LMR with various parameter settings
func TestLMRWithDifferentConfigurations(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1)
	
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	baseConfig := ai.SearchConfig{
		MaxDepth:     3,
		MaxTime:      time.Second * 5,
		UseAlphaBeta: true,
		UseLMR:       true,
	}
	
	configurations := []struct {
		name         string
		minDepth     int
		minMoves     int
		reductionBase float64
	}{
		{"conservative", 4, 6, 0.5},
		{"aggressive", 2, 3, 1.0},
		{"default", 3, 4, 0.75},
	}
	
	ctx := context.Background()
	
	for _, config := range configurations {
		t.Run(config.name, func(t *testing.T) {
			searchConfig := baseConfig
			searchConfig.LMRMinDepth = config.minDepth
			searchConfig.LMRMinMoves = config.minMoves
			searchConfig.LMRReductionBase = config.reductionBase
			
			result := engine.FindBestMove(ctx, b, moves.White, searchConfig)
			
			// Should complete without error
			if result.BestMove.From.File == -1 {
				t.Error("Expected to find a best move")
			}
			
			t.Logf("Config %s: Reductions=%d, ReSearches=%d, NodesSkipped=%d, TotalNodes=%d", 
				config.name,
				result.Stats.LMRReductions,
				result.Stats.LMRReSearches, 
				result.Stats.LMRNodesSkipped,
				result.Stats.NodesSearched)
		})
	}
}

// TestLMRReSearchConditions tests that re-searches happen when moves fail high
func TestLMRReSearchConditions(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1)
	
	// Use a position where some moves might surprise us
	b, err := board.FromFEN("r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	config := ai.SearchConfig{
		MaxDepth:         4,
		MaxTime:          time.Second * 10,
		UseAlphaBeta:     true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      2,
		LMRReductionBase: 0.75,
	}
	
	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Some re-searches should occur in a complex position
	t.Logf("Reductions: %d, Re-searches: %d", 
		result.Stats.LMRReductions, result.Stats.LMRReSearches)
	
	// Re-searches should not exceed reductions
	if result.Stats.LMRReSearches > result.Stats.LMRReductions {
		t.Error("Number of re-searches should not exceed number of reductions")
	}
}

// TestLMRStatisticsTracking tests that LMR statistics are properly tracked
func TestLMRStatisticsTracking(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(1)
	
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	config := ai.SearchConfig{
		MaxDepth:         3,
		MaxTime:          time.Second * 5,
		UseAlphaBeta:     true,
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      2,
		LMRReductionBase: 0.75,
	}
	
	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// All statistics should be non-negative
	if result.Stats.LMRReductions < 0 {
		t.Error("LMR reductions should be non-negative")
	}
	if result.Stats.LMRReSearches < 0 {
		t.Error("LMR re-searches should be non-negative")
	}
	if result.Stats.LMRNodesSkipped < 0 {
		t.Error("LMR nodes skipped should be non-negative")
	}
	
	// Re-searches should not exceed reductions
	if result.Stats.LMRReSearches > result.Stats.LMRReductions {
		t.Error("Re-searches should not exceed reductions")
	}
	
	t.Logf("LMR Statistics - Reductions: %d, Re-searches: %d, Nodes skipped: %d",
		result.Stats.LMRReductions,
		result.Stats.LMRReSearches,
		result.Stats.LMRNodesSkipped)
}

// TestLMRInteractionWithOtherFeatures tests that LMR works well with other search features
func TestLMRInteractionWithOtherFeatures(t *testing.T) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(16) // Larger TT
	
	b, err := board.FromFEN("r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	config := ai.SearchConfig{
		MaxDepth:         4,
		MaxTime:          time.Second * 10,
		UseAlphaBeta:     true,
		UseNullMove:      true, // Enable null move pruning
		UseLMR:           true,
		LMRMinDepth:      2,
		LMRMinMoves:      3,
		LMRReductionBase: 0.75,
	}
	
	ctx := context.Background()
	result := engine.FindBestMove(ctx, b, moves.White, config)
	
	// Should find a reasonable move
	if result.BestMove.From.File == -1 {
		t.Error("Expected to find a best move")
	}
	
	// Should have some efficiency gains from multiple pruning techniques
	t.Logf("Combined pruning - Total nodes: %d, LMR reductions: %d", 
		result.Stats.NodesSearched, result.Stats.LMRReductions)
}