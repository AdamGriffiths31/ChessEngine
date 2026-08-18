package movegen

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestPerft_StandardPositions(t *testing.T) {
	maxDepth := 5
	if envDepth := os.Getenv("PERFT_MAX_DEPTH"); envDepth != "" {
		if depth, err := strconv.Atoi(envDepth); err == nil {
			maxDepth = depth
		}
	}

	// In short mode, cap depth to keep the suite fast, unless
	// PERFT_MAX_DEPTH already set an even lower cap (env var always wins
	// when it asks for less work than the short-mode default).
	if testing.Short() && maxDepth > 3 {
		maxDepth = 3
	}

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
			player := playerFromSide(b.GetSideToMove())

			for _, depthTest := range position.Depths {
				t.Run(fmt.Sprintf("depth_%d", depthTest.Depth), func(t *testing.T) {
					if depthTest.Depth > maxDepth {
						t.Skipf("Skipping depth %d (max depth: %d)", depthTest.Depth, maxDepth)
					}

					// Use single generator instance for optimal cache performance
					generator := NewGenerator()
					result := PerftWithGenerator(b, depthTest.Depth, player, generator)
					if result != depthTest.Nodes {
						t.Errorf("Position %s at depth %d: expected %d nodes, got %d",
							position.Name, depthTest.Depth, depthTest.Nodes, result)
					}

				})
			}
		})
	}
}

// playerFromSide converts the board's side-to-move string ("w"/"b") into a
// Player, so perft is run for whichever side the FEN actually specifies
// instead of assuming White.
func playerFromSide(side string) Player {
	if side == "w" {
		return White
	}
	return Black
}
