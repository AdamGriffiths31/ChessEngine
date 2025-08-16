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
	maxDepth := 5
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

					// Use single generator instance for optimal cache performance
					generator := NewGenerator()
					result := PerftWithGenerator(b, depthTest.Depth, White, generator)
					if result != depthTest.Nodes {
						t.Errorf("Position %s at depth %d: expected %d nodes, got %d",
							position.Name, depthTest.Depth, depthTest.Nodes, result)
					}

				})
			}
		})
	}
}
