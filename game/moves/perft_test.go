package moves

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPerft_StandardPositions(t *testing.T) {
	// Get configurable max depth from environment variable, default to 2
	maxDepth := 4
	if envDepth := os.Getenv("PERFT_MAX_DEPTH"); envDepth != "" {
		if depth, err := strconv.Atoi(envDepth); err == nil {
			maxDepth = depth
		}
	}

	// Load test data
	testData, err := LoadPerftTestData(GetTestDataPath())
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	for _, position := range testData.Positions {
		t.Run(position.Name, func(t *testing.T) {
			b, err := board.FromFEN(position.FEN)
			if err != nil {
				t.Fatalf("Failed to parse FEN %s: %v", position.FEN, err)
			}

			for _, depthTest := range position.Depths {
				t.Run(fmt.Sprintf("depth_%d", depthTest.Depth), func(t *testing.T) {
					// Skip tests beyond max depth
					if depthTest.Depth > maxDepth {
						t.Skipf("Skipping depth %d (max depth: %d)", depthTest.Depth, maxDepth)
					}

					result := Perft(b, depthTest.Depth, White)
					if result != depthTest.Nodes {
						t.Errorf("Position %s at depth %d: expected %d nodes, got %d",
							position.Name, depthTest.Depth, depthTest.Nodes, result)
					}
					
					// Test detailed statistics if available
					if depthTest.Captures > 0 || depthTest.EnPassant > 0 || depthTest.Castles > 0 || depthTest.Promotions > 0 {
						stats := PerftWithStats(b, depthTest.Depth, White)
						if stats.Nodes != depthTest.Nodes {
							t.Errorf("Position %s at depth %d: expected %d nodes, got %d (stats)",
								position.Name, depthTest.Depth, depthTest.Nodes, stats.Nodes)
						}
						if depthTest.Captures > 0 && stats.Captures != depthTest.Captures {
							t.Errorf("Position %s at depth %d: expected %d captures, got %d",
								position.Name, depthTest.Depth, depthTest.Captures, stats.Captures)
						}
						if depthTest.EnPassant > 0 && stats.EnPassant != depthTest.EnPassant {
							t.Errorf("Position %s at depth %d: expected %d en passant, got %d",
								position.Name, depthTest.Depth, depthTest.EnPassant, stats.EnPassant)
						}
						if depthTest.Castles > 0 && stats.Castles != depthTest.Castles {
							t.Errorf("Position %s at depth %d: expected %d castles, got %d",
								position.Name, depthTest.Depth, depthTest.Castles, stats.Castles)
						}
						if depthTest.Promotions > 0 && stats.Promotions != depthTest.Promotions {
							t.Errorf("Position %s at depth %d: expected %d promotions, got %d",
								position.Name, depthTest.Depth, depthTest.Promotions, stats.Promotions)
						}
					}
				})
			}
		})
	}
}
