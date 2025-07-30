package uci

import (
	"testing"
	"strings"
)

// TestMoveSequenceValidation validates the exact move sequence that led to the illegal move
func TestMoveSequenceValidation(t *testing.T) {
	// The exact move sequence from the UCI communication log
	moveSequence := "d2d4 c7c6 a2a3 d7d6 b2b3 c8f5 c2c3 d6d5 e2e3 a7a5 f2f3 f5g6 g2g3 h7h5 h2h3 d8c7 a3a4 g8f6 b3b4 g6b1 c3c4 b1g6 e3e4 e7e5 f3f4 f8b4 c1d2 g6e4 g3g4 h5g4 h3h4 d5c4 f4f5 b4d2 d1d2 c7e7 h4h5 c4c3 h5h6 c3d2"
	
	t.Logf("=== MOVE SEQUENCE VALIDATION ===")
	t.Logf("Full sequence: %s", moveSequence)
	
	// Split into individual moves
	moves := strings.Fields(moveSequence)
	t.Logf("Total moves: %d", len(moves))
	
	// Create UCI engine and apply the sequence
	engine := NewUCIEngine()
	
	// Initialize
	engine.HandleCommand("uci")
	engine.HandleCommand("isready")
	engine.HandleCommand("ucinewgame")
	
	// Apply the full move sequence
	positionCmd := "position startpos moves " + moveSequence
	t.Logf("Applying position command: %s", positionCmd)
	
	response := engine.HandleCommand(positionCmd)
	if response != "" {
		t.Logf("Position response: %s", response)
	}
	
	// Get the resulting position FEN
	// We need to access the engine's current FEN somehow
	// Let's check what position we ended up in
	
	t.Logf("\n=== ANALYZING FINAL POSITION ===")
	
	// The last move in the sequence was c3d2 (black pawn takes something on d2)
	lastMove := moves[len(moves)-1]
	t.Logf("Last move applied: %s", lastMove)
	
	// Now test if our engine generates the same illegal move
	response = engine.HandleCommand("go depth 1")
	t.Logf("Engine response: %s", response)
	
	if strings.Contains(response, "bestmove d4d2") {
		t.Logf("*** REPRODUCED: Engine chose illegal move d4d2 ***")
	} else {
		t.Logf("Engine chose different move: %s", response)
	}
	
	// Let's also manually verify each move in the sequence step by step
	t.Logf("\n=== STEP-BY-STEP VALIDATION ===")
	
	// Create a fresh engine
	stepEngine := NewUCIEngine()
	stepEngine.HandleCommand("uci")
	stepEngine.HandleCommand("isready")
	stepEngine.HandleCommand("ucinewgame")
	
	// Apply moves one by one and check for issues
	appliedMoves := []string{}
	
	for i, move := range moves {
		appliedMoves = append(appliedMoves, move)
		moveSeq := strings.Join(appliedMoves, " ")
		
		cmd := "position startpos moves " + moveSeq
		response := stepEngine.HandleCommand(cmd)
		
		if response != "" {
			t.Logf("  Move %d (%s): ERROR - %s", i+1, move, response)
		} else {
			t.Logf("  Move %d (%s): OK", i+1, move)
		}
		
		// Check some critical moves
		if i == len(moves)-5 { // 5 moves before the end
			t.Logf("    Position 5 moves before end:")
			// Could check legal moves here
		}
		
		if i == len(moves)-1 { // Final position
			t.Logf("    Final position reached")
			t.Logf("    Testing move generation...")
			
			testResponse := stepEngine.HandleCommand("go depth 1")
			if strings.Contains(testResponse, "d4d2") {
				t.Logf("    *** CONFIRMED: d4d2 illegal move reproduced in step-by-step test ***")
			}
		}
	}
}

// TestMoveParsingAccuracy tests if our move parsing matches cutechess expectations
func TestMoveParsingAccuracy(t *testing.T) {
	t.Logf("=== MOVE PARSING ACCURACY TEST ===")
	
	// Test some specific moves from the sequence that might be problematic
	problematicMoves := []string{
		"b4d2",  // Bishop takes something on d2
		"c3d2",  // Pawn takes on d2  
		"d4d2",  // The illegal queen move
	}
	
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("isready")
	
	for _, move := range problematicMoves {
		t.Logf("\nTesting move: %s", move)
		
		// Test in starting position first
		engine.HandleCommand("position startpos")
		
		// Try to apply just this move
		cmd := "position startpos moves " + move
		response := engine.HandleCommand(cmd)
		
		if response != "" {
			t.Logf("  Move %s caused error: %s", move, response)
		} else {
			t.Logf("  Move %s parsed successfully", move)
		}
	}
}

// TestCriticalPosition tests the position just before the illegal move
func TestCriticalPosition(t *testing.T) {
	t.Logf("=== CRITICAL POSITION TEST ===")
	
	// Moves up to the point just before the illegal move
	movesBeforeIllegal := "d2d4 c7c6 a2a3 d7d6 b2b3 c8f5 c2c3 d6d5 e2e3 a7a5 f2f3 f5g6 g2g3 h7h5 h2h3 d8c7 a3a4 g8f6 b3b4 g6b1 c3c4 b1g6 e3e4 e7e5 f3f4 f8b4 c1d2 g6e4 g3g4 h5g4 h3h4 d5c4 f4f5 b4d2 d1d2 c7e7 h4h5 c4c3 h5h6 c3d2"
	
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("isready")
	
	// Apply all moves up to the critical point
	cmd := "position startpos moves " + movesBeforeIllegal
	t.Logf("Setting up critical position...")
	engine.HandleCommand(cmd)
	
	// Now see what legal moves are available
	t.Logf("Requesting move from critical position...")
	response := engine.HandleCommand("go depth 1")
	
	t.Logf("Engine response: %s", response)
	
	if strings.Contains(response, "d4d2") {
		t.Errorf("BUG CONFIRMED: Engine generates illegal move d4d2 from critical position")
		t.Logf("This proves the move generation bug exists independent of move parsing")
	}
}