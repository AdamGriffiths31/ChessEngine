package uci

import (
	"strings"
	"testing"
)

func TestBestMoveFormat(t *testing.T) {
	// Test the exact format of our bestmove responses to find any issues
	engine := NewUCIEngine()
	
	// Set up the position where the illegal move occurs
	engine.HandleCommand("uci")
	engine.HandleCommand("isready")
	engine.HandleCommand("ucinewgame")
	
	// Apply the exact move sequence from the logs
	moveSequence := "c2c4 e7e5 a2a3 b8c6 b2b3 g8f6 d2d3 d7d5 e2e3 c8e6 f2f3 a7a6 g2g3 f8d6 h2h3 d8d7 a3a4 e5e4 b3b4 d6g3 e1d2 g3e5 d3d4 e4f3 e3e4 d5e4 h3h4 d7d4 f1d3 d4d3"
	positionCmd := "position startpos moves " + moveSequence
	
	t.Logf("Setting up position with: %s", positionCmd)
	engine.HandleCommand(positionCmd)
	
	// Now ask for a move from this position
	t.Logf("Requesting move from position...")
	response := engine.HandleCommand("go depth 1")
	
	t.Logf("Response: '%s'", response)
	
	// Check for any formatting issues
	lines := strings.Split(response, "\n")
	for i, line := range lines {
		t.Logf("Line %d: '%s' (length: %d)", i, line, len(line))
		
		if strings.HasPrefix(line, "bestmove ") {
			t.Logf("Found bestmove line: '%s'", line)
			
			// Check for trailing spaces
			if strings.HasSuffix(line, " ") {
				t.Errorf("PROBLEM: bestmove line has trailing space!")
			}
			
			// Extract the move part
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				move := parts[1]
				t.Logf("Move: '%s' (length: %d)", move, len(move))
				
				// Check move format
				if len(move) != 4 && len(move) != 5 { // 4 for normal, 5 for promotion
					t.Errorf("PROBLEM: Move has unexpected length: %d", len(move))
				}
				
				// Check for non-printable characters
				for j, r := range move {
					if r < 32 || r > 126 {
						t.Errorf("PROBLEM: Move contains non-printable character at position %d: %d", j, r)
					}
				}
			}
		}
	}
}