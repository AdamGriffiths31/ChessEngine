//go:build long

package moves

import (
	"fmt"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestBitboardPerformance tests the performance on a challenging perft test
func TestBitboardPerformance(t *testing.T) {
	// Start with depth 4 for faster testing, can increase later
	testDepth := 4
	expectedNodes := int64(4085603) // Kiwipete depth 4
	
	// Kiwipete position FEN
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
	
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse Kiwipete FEN: %v", err)
	}

	fmt.Printf("Starting Kiwipete depth %d performance test...\n", testDepth)
	fmt.Printf("Expected nodes: %d\n", expectedNodes)
	
	// Note: The generator now uses bitboard implementation by default
	fmt.Printf("\n=== High-Performance Bitboard Move Generation ===\n")
	generator := NewGenerator()
	start := time.Now()
	result := PerftWithGenerator(b, testDepth, White, generator)
	duration := time.Since(start)
	
	// Verify correctness
	if result != expectedNodes {
		t.Errorf("Kiwipete depth %d: expected %d nodes, got %d", testDepth, expectedNodes, result)
	}
	
	fmt.Printf("Results:\n")
	fmt.Printf("  Nodes: %d\n", result)
	fmt.Printf("  Time: %v\n", duration)
	fmt.Printf("  Nodes per second: %.0f\n", float64(result)/duration.Seconds())
	
	// Record timing result
	recordTimingResult("kiwipete_depth4", result, duration)
	
	// Cleanup
	generator.Release()
}

// TestKiwipeteDepth5Performance runs a deeper test with the optimized generator
func TestKiwipeteDepth5Performance(t *testing.T) {
	// Depth 5 for more intensive testing
	testDepth := 5
	expectedNodes := int64(193690690) // Kiwipete depth 5
	
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
	
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse Kiwipete FEN: %v", err)
	}

	fmt.Printf("Starting Kiwipete depth %d performance test...\n", testDepth)
	fmt.Printf("Expected nodes: %d\n", expectedNodes)
	
	// Use the default generator (now uses bitboards)
	generator := NewGenerator()
	defer generator.Release()
	
	start := time.Now()
	result := PerftWithGenerator(b, testDepth, White, generator)
	duration := time.Since(start)
	
	// Verify correctness
	if result != expectedNodes {
		t.Errorf("Kiwipete depth %d: expected %d nodes, got %d", testDepth, expectedNodes, result)
	}
	
	// Report timing results
	fmt.Printf("Depth %d results:\n", testDepth)
	fmt.Printf("  Nodes: %d\n", result)
	fmt.Printf("  Time: %v\n", duration)
	fmt.Printf("  Nodes per second: %.0f\n", float64(result)/duration.Seconds())
	
	// Record results
	recordTimingResult(fmt.Sprintf("kiwipete_depth%d", testDepth), result, duration)
}

func recordComparisonResult(testName string, nodes int64, arrayDuration, bitboardDuration time.Duration, speedup float64) {
	fmt.Printf("\n=== COMPARISON RESULT ===\n")
	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("Nodes: %d\n", nodes)
	fmt.Printf("Array Duration: %v\n", arrayDuration)
	fmt.Printf("Bitboard Duration: %v\n", bitboardDuration)
	fmt.Printf("Speedup: %.2fx\n", speedup)
	fmt.Printf("Array NPS: %.0f\n", float64(nodes)/arrayDuration.Seconds())
	fmt.Printf("Bitboard NPS: %.0f\n", float64(nodes)/bitboardDuration.Seconds())
	fmt.Printf("========================\n")
}