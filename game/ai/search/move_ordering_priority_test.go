package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMoveOrderingPriorities(t *testing.T) {
	// This test demonstrates the issue where terrible captures get higher priority than killer moves
	// Position where we have a terrible capture and a good quiet move (potential killer)
	fen := "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetDebugMoveOrdering(true)

	// Set up a killer move at depth 0 (simulate previous search finding Qd3 as good)
	killerMove := board.Move{
		From:  board.Square{Rank: 1, File: 3}, // d2
		To:    board.Square{Rank: 2, File: 3}, // d3
		Piece: board.WhiteQueen,
	}
	// Store killer move using the public API
	engine.storeKiller(killerMove, 0)

	generator := moves.NewGenerator()
	legalMoves := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(legalMoves)

	// Find the terrible capture and the killer move
	var terribleCapture, killerMoveFound board.Move
	var foundCapture, foundKiller bool

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		if move.Piece == board.WhiteQueen && move.IsCapture &&
			move.To.Rank == 3 && move.To.File == 3 { // Qxd4 - terrible capture
			terribleCapture = move
			foundCapture = true
		} else if move.From == killerMove.From && move.To == killerMove.To {
			killerMoveFound = move
			foundKiller = true
		}
	}

	if !foundCapture {
		t.Skip("Could not find the terrible capture Qxd4")
	}
	if !foundKiller {
		t.Skip("Could not find the killer move Qd3")
	}

	// Calculate scores manually
	terribleCaptureScore := engine.getCaptureScore(b, terribleCapture)
	killerScore := 500000 // This is the hardcoded killer move score in orderMoves

	seeValue := engine.seeCalculator.SEE(b, terribleCapture)
	t.Logf("Terrible capture (Qxd4) SEE: %d", seeValue)
	t.Logf("Terrible capture score: %d", terribleCaptureScore)
	t.Logf("Killer move score: %d", killerScore)

	// The issue: terrible capture should have lower priority than killer moves
	if terribleCaptureScore > killerScore {
		t.Errorf("ISSUE CONFIRMED: Terrible capture (score %d) has higher priority than killer move (score %d)",
			terribleCaptureScore, killerScore)
		t.Logf("This means the engine will try terrible captures before good quiet moves")
		t.Logf("Expected: Killer moves should have higher priority than clearly losing captures")
	}

	// Order moves and check actual ordering
	engine.orderMoves(b, legalMoves, 0, board.Move{})
	orderedMoves := engine.GetLastMoveOrder()

	var capturePos, killerPos int = -1, -1
	for i, move := range orderedMoves {
		if move.From == terribleCapture.From && move.To == terribleCapture.To {
			capturePos = i
		} else if move.From == killerMoveFound.From && move.To == killerMoveFound.To {
			killerPos = i
		}
	}

	t.Logf("Terrible capture position in ordered list: %d", capturePos)
	t.Logf("Killer move position in ordered list: %d", killerPos)

	if capturePos != -1 && killerPos != -1 && capturePos < killerPos {
		t.Errorf("ORDERING ISSUE: Terrible capture (pos %d) ordered before killer move (pos %d)",
			capturePos, killerPos)
	}
}

func TestMoveOrderingThresholds(t *testing.T) {
	// Test various SEE values and their resulting scores to show the threshold issue
	fen := "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// Create a mock move for testing different scenarios
	testMove := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 3, File: 3}, // d4
		Piece:     board.WhiteQueen,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	// Get the actual SEE value
	actualSEE := engine.seeCalculator.SEE(b, testMove)
	actualScore := engine.getCaptureScore(b, testMove)

	t.Logf("Actual terrible capture:")
	t.Logf("  SEE: %d", actualSEE)
	t.Logf("  Score: %d", actualScore)

	// Test threshold values
	testCases := []struct {
		name     string
		seeValue int
		priority string
	}{
		{"Excellent capture", 400, "Should be highest priority"},
		{"Good capture", 100, "Should be high priority"},
		{"Equal exchange", 0, "Should be above killers"},
		{"Slightly bad capture", -50, "Should be below killers, above history"},
		{"Bad capture", -200, "Should be below history, above quiet moves"},
		{"Terrible capture", -800, "Should be below history, above quiet moves"},
	}

	killerScore := 500000
	t.Logf("\nKiller move score: %d", killerScore)
	t.Logf("\nSEE threshold analysis:")

	historyScore := 50000 // Maximum history score
	t.Logf("Max history score: %d", historyScore)
	t.Logf("")

	for _, tc := range testCases {
		// Use the actual scoring function logic
		var score int
		if tc.seeValue > 0 {
			score = 1000000 + tc.seeValue
		} else if tc.seeValue == 0 {
			score = 900000
		} else if tc.seeValue >= -100 {
			score = 100000 + tc.seeValue + 100
		} else {
			score = 25000 + tc.seeValue + 1000
		}

		var status string
		if tc.seeValue > 0 || tc.seeValue == 0 {
			if score > killerScore {
				status = "✅ ABOVE killers (correct)"
			} else {
				status = "⚠️ Should be above killers"
			}
		} else if tc.seeValue >= -100 {
			if score < killerScore && score > historyScore {
				status = "✅ Between killers and history (correct)"
			} else if score > killerScore {
				status = "⚠️ Should be below killers"
			} else {
				status = "⚠️ Should be above history"
			}
		} else { // tc.seeValue < -100
			if score < historyScore && score > 0 {
				status = "✅ Below history, above quiet (correct)"
			} else if score > historyScore {
				status = "⚠️ Should be below history"
			} else {
				status = "⚠️ Should be above quiet moves"
			}
		}

		t.Logf("  %s (SEE %d): score %d - %s",
			tc.name, tc.seeValue, score, status)
	}

	// Show current implementation
	t.Logf("\n✅ CURRENT IMPLEMENTATION (Tactical Ordering):")
	t.Logf("  if seeValue > 0: return 1000000 + seeValue           // Good captures")
	t.Logf("  else if seeValue == 0: return 900000                 // Equal exchanges")
	t.Logf("  else if seeValue >= -100: return 100000 + seeValue + 100  // Below killers, above history")
	t.Logf("  else: return 25000 + seeValue + 1000                 // Below history, above quiet")
}
