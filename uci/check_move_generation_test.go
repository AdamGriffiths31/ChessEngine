package uci

import (
	"testing"
)

// TestKnightMoveInCheckBug tests if the move generator incorrectly allows knight moves
// when the king is in check and the knight move doesn't resolve the check
func TestKnightMoveInCheckBug(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== TESTING KNIGHT MOVE GENERATION WHEN KING IN CHECK ===")
	
	// Set up the exact position after Bxf7+ where the bug occurs
	// FEN: r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15
	// Black king on e8, White bishop on f7 (giving check), Black knight on f6
	positionCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	engine.HandleCommand(positionCmd)
	
	currentFEN := engine.engine.GetCurrentFEN()
	t.Logf("Position set to: %s", currentFEN)
	t.Logf("Black king is in check from White bishop on f7")
	t.Logf("Black knight is on f6")
	
	// Generate all legal moves
	t.Logf("\n=== GENERATING LEGAL MOVES ===")
	
	// Use the engine's internal move generation
	// We need to call the AI to get legal moves, but we'll capture them from the debug output
	engine.HandleCommand("go depth 1")
	
	t.Logf("\n=== ANALYSIS ===")
	t.Logf("Legal moves should only include:")
	t.Logf("1. King moves that get out of check: e8d7, e8e7, e8f7 (captures bishop), e8d8")
	t.Logf("2. NO knight moves should be legal because f6f7 doesn't resolve the check")
	t.Logf("3. f6f7 should NOT be in the legal moves list")
	
	// The actual analysis will be visible in the debug logs
	t.Logf("\nCheck the debug logs above to see if f6f7 was incorrectly included in legal moves")
}

// TestSpecificKnightMoveValidation tests if we can manually validate the f6f7 move
func TestSpecificKnightMoveValidation(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== TESTING SPECIFIC f6f7 MOVE VALIDATION ===")
	
	// Set up the position after Bxf7+
	positionCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	engine.HandleCommand(positionCmd)
	
	// Try to apply the f6f7 move and see what happens
	t.Logf("Attempting to apply f6f7 move...")
	
	// This should fail or be rejected
	illegalMoveCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15 moves f6f7"
	engine.HandleCommand(illegalMoveCmd)
	
	finalFEN := engine.engine.GetCurrentFEN()
	t.Logf("FEN after attempting f6f7: %s", finalFEN)
	
	// Check if the position changed
	expectedFEN := "r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	if finalFEN == expectedFEN {
		t.Logf("✅ GOOD: f6f7 was rejected - position unchanged")
	} else {
		t.Logf("❌ BUG: f6f7 was accepted - position changed!")
		t.Logf("Expected: %s", expectedFEN)  
		t.Logf("Actual:   %s", finalFEN)
	}
}

// TestLegalMovesInCheck specifically tests what moves are considered legal when in check
func TestLegalMovesInCheck(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== TESTING LEGAL MOVES WHEN KING IN CHECK ===")
	
	// Set up the position after Bxf7+
	positionCmd := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
	engine.HandleCommand(positionCmd)
	
	t.Logf("Position: r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15")
	t.Logf("Black to move, king in check from bishop on f7")
	
	// Test each expected legal move individually
	expectedLegalMoves := []string{"e8d7", "e8e7", "e8f7", "e8d8"}
	expectedIllegalMoves := []string{"f6f7", "f6d7", "f6e4", "f6g4", "f6h5"}
	
	t.Logf("\n=== TESTING EXPECTED LEGAL MOVES ===")
	for _, move := range expectedLegalMoves {
		testMove := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15 moves " + move
		
		// Create a separate engine instance for each test
		testEngine := NewUCIEngine()
		testEngine.HandleCommand("uci")
		testEngine.HandleCommand("ucinewgame")
		testEngine.HandleCommand(testMove)
		
		resultFEN := testEngine.engine.GetCurrentFEN()
		originalFEN := "r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
		
		if resultFEN != originalFEN {
			t.Logf("✅ %s: LEGAL (position changed)", move)
		} else {
			t.Logf("❌ %s: REJECTED (should be legal)", move)
		}
	}
	
	t.Logf("\n=== TESTING EXPECTED ILLEGAL MOVES ===")
	for _, move := range expectedIllegalMoves {
		testMove := "position fen r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15 moves " + move
		
		// Create a separate engine instance for each test
		testEngine := NewUCIEngine()
		testEngine.HandleCommand("uci")
		testEngine.HandleCommand("ucinewgame")
		testEngine.HandleCommand(testMove)
		
		resultFEN := testEngine.engine.GetCurrentFEN()
		originalFEN := "r1b1kb1r/5Bpp/5n2/1P2p3/p2P4/P1N1P3/1P3PPP/2RQKB1R b Kkq - 0 15"
		
		if resultFEN == originalFEN {
			t.Logf("✅ %s: CORRECTLY REJECTED", move)
		} else {
			t.Logf("❌ %s: INCORRECTLY ACCEPTED (this is the bug!)", move)
		}
	}
}