package uci

import (
	"strings"
	"testing"
)

// TestExactGameReproduction plays the exact game sequence from the PGN file
// to see if we can reproduce the f6f7 bug
func TestExactGameReproduction(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== REPRODUCING EXACT GAME FROM PGN ===")
	
	// Exact game sequence from the PGN file (converted to UCI format)
	// 1. c4 Nf6 2. Nf3 a6 3. Nc3 a5 4. a3 a4 5. d4 b6 6. Bf4 b5 
	// 7. cxb5 c6 8. e3 c5 9. Bxb8 c4 10. Rc1 d6 11. Bxc4 d5 
	// 12. Nxd5 e6 13. Nc7+ Qxc7 14. Bxc7 e5 15. Bxf7+
	gameSequence := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4",
		"d2d4", "b7b6", "c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", 
		"f4b8", "c5c4", "a1c1", "d7d6", "b8c4", "d6d5", "f3d5", "e7e6",
		"d5c7", "d8c7", "c4c7", "e6e5", "c7f7", // This is Bxf7+
	}
	
	// Apply all moves up to Bxf7+
	t.Logf("Applying %d moves from the exact game sequence...", len(gameSequence))
	
	var currentPos string
	for i, move := range gameSequence {
		moveCmd := "position startpos moves " + strings.Join(gameSequence[:i+1], " ")
		engine.HandleCommand(moveCmd)
		
		currentPos = engine.engine.GetCurrentFEN()
		t.Logf("Move %d: %s -> %s", i+1, move, currentPos)
	}
	
	t.Logf("\n=== CRITICAL POSITION AFTER Bxf7+ ===")
	t.Logf("Final position: %s", currentPos)
	t.Logf("Black to move, should be in check from White bishop on f7")
	
	// Check what pieces are where
	t.Logf("\n=== PIECE ANALYSIS ===")
	// In standard FEN: r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15
	// Let's see what we actually have
	
	// Generate legal moves to see what options Black has
	t.Logf("\n=== LEGAL MOVES ANALYSIS ===")
	engine.HandleCommand("go depth 1")
	
	t.Logf("\nNow let's see what move our AI would choose...")
	t.Logf("If our AI chooses f6f7, we've reproduced the bug!")
	t.Logf("If our AI chooses e8f7, then the real game had a different position")
}

// TestPositionComparison compares our reproduced position with the expected position
func TestPositionComparison(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== POSITION COMPARISON TEST ===")
	
	// Apply the game sequence
	gameSequence := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4",
		"d2d4", "b7b6", "c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", 
		"f4b8", "c5c4", "a1c1", "d7d6", "b8c4", "d6d5", "f3d5", "e7e6",
		"d5c7", "d8c7", "c4c7", "e6e5", "c7f7",
	}
	
	moveCmd := "position startpos moves " + strings.Join(gameSequence, " ")
	engine.HandleCommand(moveCmd)
	
	actualFEN := engine.engine.GetCurrentFEN()
	expectedFEN := "r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	
	t.Logf("Expected FEN: %s", expectedFEN)
	t.Logf("Actual FEN:   %s", actualFEN)
	
	if actualFEN == expectedFEN {
		t.Logf("✅ MATCH: Position matches expected FEN exactly")
		t.Logf("This means our move sequence replication is correct")
	} else {
		t.Logf("❌ MISMATCH: Position differs from expected")
		t.Logf("This could explain why we get different AI moves")
		
		// Analyze the differences
		expectedParts := strings.Split(expectedFEN, " ")
		actualParts := strings.Split(actualFEN, " ")
		
		if len(expectedParts) >= 1 && len(actualParts) >= 1 {
			t.Logf("\nBoard position comparison:")
			t.Logf("Expected: %s", expectedParts[0])
			t.Logf("Actual:   %s", actualParts[0])
			
			if expectedParts[0] != actualParts[0] {
				t.Logf("❌ BOARD DIFFERENCE FOUND!")
			}
		}
		
		if len(expectedParts) >= 2 && len(actualParts) >= 2 {
			t.Logf("\nSide to move:")
			t.Logf("Expected: %s", expectedParts[1])
			t.Logf("Actual:   %s", actualParts[1])
		}
		
		if len(expectedParts) >= 3 && len(actualParts) >= 3 {
			t.Logf("\nCastling rights:")
			t.Logf("Expected: %s", expectedParts[2])
			t.Logf("Actual:   %s", actualParts[2])
		}
	}
}

// TestAIChoiceInExactPosition tests what move the AI chooses from the exact reproduced position
func TestAIChoiceInExactPosition(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== AI CHOICE TEST IN EXACT POSITION ===")
	
	// Set up the exact position after Bxf7+
	positionCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	engine.HandleCommand(positionCmd)
	
	t.Logf("Position set to exact FEN from debug logs")
	t.Logf("Black to move, in check from bishop on f7")
	
	// Ask AI to choose a move
	t.Logf("\n=== ASKING AI FOR BEST MOVE ===")
	searchResponse := engine.HandleCommand("go depth 3")
	
	// Parse the bestmove from response
	lines := strings.Split(searchResponse, "\n")
	var bestMove string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "bestmove ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
			break
		}
	}
	
	t.Logf("AI selected move: %s", bestMove)
	
	if bestMove == "f6f7" {
		t.Logf("🚨 BUG REPRODUCED: AI chose f6f7 (the illegal move from PGN)")
		t.Logf("This confirms our engine has the bug")
	} else if bestMove == "e8f7" {
		t.Logf("✅ EXPECTED: AI chose e8f7 (King captures bishop)")
		t.Logf("This is the correct legal move")
	} else {
		t.Logf("🔍 DIFFERENT: AI chose %s (neither f6f7 nor e8f7)", bestMove)
		t.Logf("This is unexpected - investigating...")
	}
	
	// Verify the chosen move is legal
	t.Logf("\n=== MOVE LEGALITY VERIFICATION ===")
	testCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15 moves " + bestMove
	
	testEngine := NewUCIEngine()
	testEngine.HandleCommand("uci")
	testEngine.HandleCommand("ucinewgame")
	testEngine.HandleCommand(testCmd)
	
	finalFEN := testEngine.engine.GetCurrentFEN()
	originalFEN := "r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	
	if finalFEN != originalFEN {
		t.Logf("✅ LEGAL: Move %s was applied successfully", bestMove)
		t.Logf("Result: %s", finalFEN)
	} else {
		t.Logf("❌ ILLEGAL: Move %s was rejected", bestMove)
		t.Logf("Position unchanged")
	}
}