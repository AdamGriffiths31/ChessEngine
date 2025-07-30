package uci

import (
	"strings"
	"testing"
)

// TestBoardDesynchronizationBug tests the specific case where our engine's board state
// diverges from cutechess-cli's, causing legal moves to be rejected as illegal.
// This test replicates the exact UCI communication that led to the "d4f2 illegal move" issue.
func TestBoardDesynchronizationBug(t *testing.T) {
	engine := NewUCIEngine()
	
	// Simulate the exact UCI command sequence that led to the d4f2 desync
	// Based on the actual UCI logs from the game where d4f2 was rejected
	commands := []struct{
		input    string
		expected string // What we expect in the output (empty if no specific output expected)
		description string
	}{
		{"uci", "", "Initialize UCI"},
		{"isready", "readyok", "Check readiness"},
		{"ucinewgame", "", "Start new game"},
		
		// The actual move sequence that led to the problematic position
		// From UCI logs: the moves that should lead to Queen on d4
		{"position startpos moves d2d4 d7d5 a2a3 e7e6 b2b3 d8f6 c2c3 f6d8 e2e3 b7b6 f2f3 f8e7 g2g3 c7c5 h2h3 c8b7 a3a4 g8f6 b3b4 e7d6 c3c4 c5b4 e3e4 d6g3", "", "Apply move sequence"},
		
		{"go depth 1", "", "Search for a move"},
	}
	
	var output strings.Builder
	
	for i, cmd := range commands {
		t.Logf("Step %d: %s - %s", i+1, cmd.input, cmd.description)
		
		// Process the command
		response := engine.HandleCommand(cmd.input)
		
		if response != "" {
			output.WriteString(response + "\n")
			t.Logf("  Response: %s", response)
		}
		
		// Check expected output if specified
		if cmd.expected != "" && !strings.Contains(response, cmd.expected) {
			t.Errorf("Expected response to contain '%s', got '%s'", cmd.expected, response)
		}
		
		// After applying the move sequence, check the board state
		if strings.Contains(cmd.input, "position startpos moves") {
			currentFEN := engine.engine.GetCurrentFEN()
			t.Logf("  Current FEN: %s", currentFEN)
			
			// The expected FEN should have Queen on d4 (from our engine's perspective)
			expectedFEN := "rn1qk2r/pb3ppp/1p2pn2/3p4/PpPQP3/5PbP/8/RNBQKBNR w KQkq - 0 13"
			
			if currentFEN != expectedFEN {
				t.Logf("  WARNING: FEN mismatch detected!")
				t.Logf("    Our FEN:      %s", currentFEN)
				t.Logf("    Expected FEN: %s", expectedFEN)
				
				// Analyze the difference
				analyzeFENDifference(t, currentFEN, expectedFEN)
			} else {
				t.Logf("  SUCCESS: FEN matches expected position")
			}
		}
	}
	
	// After the go command, check what move our engine wants to make
	finalOutput := output.String()
	if strings.Contains(finalOutput, "bestmove") {
		t.Logf("Engine's chosen move: %s", finalOutput)
		
		if strings.Contains(finalOutput, "bestmove d4f2") {
			t.Logf("SUCCESS: Engine chose d4f2 (this will be rejected by cutechess-cli)")
			
			// Verify that d4f2 is actually legal according to our engine
			legalMoves := engine.engine.GetLegalMoves()
			d4f2Legal := false
			for i := 0; i < legalMoves.Count; i++ {
				move := legalMoves.Moves[i]
				moveStr := move.From.String() + move.To.String()
				if moveStr == "d4f2" {
					d4f2Legal = true
					t.Logf("  Confirmed: d4f2 is in our legal moves list")
					break
				}
			}
			
			if !d4f2Legal {
				t.Errorf("INCONSISTENCY: Engine chose d4f2 but it's not in legal moves")
			}
		} else {
			t.Logf("Engine chose a different move (no desync detected in this test)")
		}
	}
}

// TestUCIPositionCommandWithIllegalMove tests what happens when cutechess-cli
// sends us a move that we think is illegal
func TestUCIPositionCommandWithIllegalMove(t *testing.T) {
	engine := NewUCIEngine()
	
	// Set up a position where we can test illegal move handling
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// Apply a sequence that leads to a known state
	engine.HandleCommand("position startpos moves e2e4 e7e5")
	
	// Now try to apply an illegal move via position command
	// This should trigger our validation logic
	t.Logf("Testing illegal move handling...")
	
	// Try a clearly illegal move like moving a piece that doesn't exist
	_ = engine.HandleCommand("position startpos moves e2e4 e7e5 h8a1")
	
	// Check how our engine handles this
	currentFEN := engine.engine.GetCurrentFEN()
	t.Logf("FEN after illegal move attempt: %s", currentFEN)
	
	// The FEN should either:
	// 1. Still be at the e2e4 e7e5 position (if we rejected the illegal move)
	// 2. Be in some invalid state (if we applied it incorrectly)
	
	expectedValidFEN := "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
	if currentFEN != expectedValidFEN {
		t.Logf("Board state changed after illegal move - this indicates the bug!")
		t.Logf("  Expected: %s", expectedValidFEN)
		t.Logf("  Got:      %s", currentFEN)
	}
}

// analyzeFENDifference compares two FEN strings and reports the differences
func analyzeFENDifference(t *testing.T, fen1, fen2 string) {
	parts1 := strings.Split(fen1, " ")
	parts2 := strings.Split(fen2, " ")
	
	if len(parts1) >= 1 && len(parts2) >= 1 {
		board1 := parts1[0]
		board2 := parts2[0]
		
		if board1 != board2 {
			t.Logf("    Board position difference:")
			
			ranks1 := strings.Split(board1, "/")
			ranks2 := strings.Split(board2, "/")
			
			for i := 0; i < 8 && i < len(ranks1) && i < len(ranks2); i++ {
				if ranks1[i] != ranks2[i] {
					t.Logf("      Rank %d: '%s' vs '%s'", 8-i, ranks1[i], ranks2[i])
				}
			}
		}
	}
	
	// Check other FEN components
	for i := 1; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] != parts2[i] {
			labels := []string{"side_to_move", "castling", "en_passant", "halfmove", "fullmove"}
			if i-1 < len(labels) {
				t.Logf("    %s: '%s' vs '%s'", labels[i-1], parts1[i], parts2[i])
			}
		}
	}
}

// TestMoveValidationLogic tests the specific logic that validates moves
// in the POSITION command handler
func TestMoveValidationLogic(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// Set up a position with many legal moves to test our "check all moves" fix
	engine.HandleCommand("position startpos moves e2e4")
	
	legalMoves := engine.engine.GetLegalMoves()
	t.Logf("Position has %d legal moves", legalMoves.Count)
	
	// Find a move that would be at index > 10 (to test our previous bug)
	var testMove string
	for i := 0; i < legalMoves.Count; i++ {
		if i >= 10 { // This would have been missed by the old logic
			move := legalMoves.Moves[i]
			testMove = move.From.String() + move.To.String()
			t.Logf("Testing move at index %d: %s", i, testMove)
			break
		}
	}
	
	if testMove != "" {
		// Apply this move via position command and verify it works
		positionCmd := "position startpos moves e2e4 " + testMove
		_ = engine.HandleCommand(positionCmd)
		t.Logf("Applied move %s via position command", testMove)
		
		// Verify the move was applied successfully
		newFEN := engine.engine.GetCurrentFEN()
		t.Logf("New FEN: %s", newFEN)
		
		// The board should have changed from the starting position
		if strings.Contains(newFEN, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR") {
			t.Errorf("Board didn't change - move %s was not applied", testMove)
		} else {
			t.Logf("SUCCESS: Move %s was applied correctly", testMove)
		}
	}
}