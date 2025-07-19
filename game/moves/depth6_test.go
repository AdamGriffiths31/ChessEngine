//go:build long

package moves

import (
	"fmt"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestKiwipeteDepth6Performance runs the ultimate performance test - depth 6 with over 8 billion nodes
func TestKiwipeteDepth6Performance(t *testing.T) {
	// Depth 6 - the ultimate test with 8+ billion nodes
	testDepth := 6
	expectedNodes := int64(8031647685) // Kiwipete depth 6
	
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
	
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse Kiwipete FEN: %v", err)
	}

	fmt.Printf("🚀 Starting Kiwipete depth %d ULTIMATE performance test...\n", testDepth)
	fmt.Printf("Expected nodes: %s\n", formatLargeNumber(expectedNodes))
	fmt.Printf("This will test over 8 billion positions!\n\n")
	
	// Use the high-performance bitboard generator
	generator := NewGenerator()
	defer generator.Release()
	
	fmt.Printf("⏱️  Starting timer...\n")
	start := time.Now()
	result := PerftWithGenerator(b, testDepth, White, generator)
	duration := time.Since(start)
	
	// Verify correctness
	if result != expectedNodes {
		t.Errorf("Kiwipete depth %d: expected %d nodes, got %d", testDepth, expectedNodes, result)
	}
	
	// Report timing results
	fmt.Printf("\n🎉 DEPTH %d RESULTS:\n", testDepth)
	fmt.Printf("  ✅ Nodes: %s\n", formatLargeNumber(result))
	fmt.Printf("  ⏱️  Time: %v\n", duration)
	fmt.Printf("  🚀 Nodes per second: %s\n", formatLargeNumber(int64(float64(result)/duration.Seconds())))
	fmt.Printf("  📊 Minutes: %.2f\n", duration.Minutes())
	
	// Calculate performance metrics
	nps := float64(result) / duration.Seconds()
	fmt.Printf("\n📈 PERFORMANCE ANALYSIS:\n")
	fmt.Printf("  • Processing speed: %s nodes/second\n", formatLargeNumber(int64(nps)))
	fmt.Printf("  • Million nodes/sec: %.1f MN/s\n", nps/1_000_000)
	fmt.Printf("  • Billion nodes processed in: %.1f minutes\n", duration.Minutes())
	
	// Record results
	recordTimingResult(fmt.Sprintf("kiwipete_depth%d_ULTIMATE", testDepth), result, duration)
	
	fmt.Printf("\n🎊 ULTIMATE TEST COMPLETED SUCCESSFULLY!\n")
}

// Helper function to format large numbers with commas
func formatLargeNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}
	
	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}