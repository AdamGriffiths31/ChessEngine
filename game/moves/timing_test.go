package moves

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestKiwipeteDepth6Timing(t *testing.T) {
	// Kiwipete position FEN
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
	expectedNodes := int64(8031647685)
	
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse Kiwipete FEN: %v", err)
	}

	fmt.Printf("Starting Kiwipete depth 6 timing test...\n")
	fmt.Printf("Expected nodes: %d\n", expectedNodes)
	
	// Time the perft calculation
	start := time.Now()
	result := Perft(b, 6, White)
	duration := time.Since(start)
	
	// Verify correctness
	if result != expectedNodes {
		t.Errorf("Kiwipete depth 6: expected %d nodes, got %d", expectedNodes, result)
	}
	
	// Report timing results
	fmt.Printf("Kiwipete depth 6 results:\n")
	fmt.Printf("  Nodes: %d\n", result)
	fmt.Printf("  Time: %v\n", duration)
	fmt.Printf("  Nodes per second: %.0f\n", float64(result)/duration.Seconds())
	
	// Record results to file
	recordTimingResult("kiwipete_depth6", result, duration)
}

func recordTimingResult(testName string, nodes int64, duration time.Duration) {
	// Simple logging to stdout
	fmt.Printf("\n=== TIMING RESULT ===\n")
	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("Nodes: %d\n", nodes)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Nodes/sec: %.0f\n", float64(nodes)/duration.Seconds())
	fmt.Printf("====================\n")
	
	// Write results to markdown file
	file, err := os.OpenFile("timing_results.md", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening timing results file: %v\n", err)
		return
	}
	defer file.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("## %s - %s\n\n", testName, timestamp)
	entry += fmt.Sprintf("- **Nodes**: %d\n", nodes)
	entry += fmt.Sprintf("- **Duration**: %v\n", duration)
	entry += fmt.Sprintf("- **Nodes/sec**: %.0f\n\n", float64(nodes)/duration.Seconds())
	
	if _, err := file.WriteString(entry); err != nil {
		fmt.Printf("Error writing to timing results file: %v\n", err)
	}
}