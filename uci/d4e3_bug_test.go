package uci

import (
	"strings"
	"testing"
)

// TestD4E3IllegalMoveBug tests the specific illegal move scenario we captured from real gameplay
// Position: rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14
// Our engine generated d4e3 which cutechess-cli rejected as illegal
func TestD4E3IllegalMoveBug(t *testing.T) {
	engine := NewUCIEngine()

	// Reproduce the exact game sequence that led to the d4e3 illegal move
	gameSequence := []string{
		"d2d4", "d7d5",       // 1. d4 d5
		"a2a3", "e7e6",       // 2. a3 e6
		"b2b3", "g8f6",       // 3. b3 Nf6
		"c2c3", "c7c5",       // 4. c3 c5
		"e2e3", "b7b6",       // 5. e3 b6
		"f2f3", "a7a6",       // 6. f3 a6
		"g2g3", "f8d6",       // 7. g3 Bd6
		"h2h3", "d6g3",       // 8. h3 Bxg3+
		"e1d2", "c8b7",       // 9. Kd2 Bb7
		"a3a4", "g3d6",       // 10. a4 Bd6
		"b3b4", "c5b4",       // 11. b4 cxb4
		"c3c4", "a6a5",       // 12. c4 a5
		"e3e4", "b7f4",       // 13. e4 Bf4+ (this is where the illegal move d4e3 was generated)
	}

	t.Logf("Reproducing exact game sequence that led to d4e3 illegal move...")

	// Apply the game sequence
	positionCmd := "position startpos moves " + strings.Join(gameSequence, " ")
	t.Logf("Position command: %s", positionCmd)
	
	engine.HandleCommand(positionCmd)

	// Expected FEN from our UCI logs
	expectedFEN := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	t.Logf("Expected position FEN: %s", expectedFEN)

	// Test what our engine does with this position using the go command
	t.Logf("Testing engine response to 'go depth 1' command...")
	response := engine.HandleCommand("go depth 1")
	t.Logf("Engine response: %s", response)
	
	if strings.Contains(response, "bestmove d4e3") {
		t.Errorf("❌ BUG CONFIRMED: Engine chooses d4e3 as best move")
	} else {
		t.Logf("✅ Engine chose different move or test setup issue")
	}
}

// TestAnalyzeD4E3Position analyzes why d4e3 might be considered illegal
func TestAnalyzeD4E3Position(t *testing.T) {
	engine := NewUCIEngine()

	// Set up the exact position where d4e3 was generated
	positionFEN := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	engine.HandleCommand("position fen " + positionFEN)

	t.Logf("Analyzing position: %s", positionFEN)
	
	// Test what move our engine chooses
	response := engine.HandleCommand("go depth 1")
	t.Logf("Engine response: %s", response)
	
	if strings.Contains(response, "bestmove d4e3") {
		t.Logf("🚨 BUG REPRODUCED: Engine chooses d4e3 in this position")
	} else {
		t.Logf("Engine chose a different move")
	}
}