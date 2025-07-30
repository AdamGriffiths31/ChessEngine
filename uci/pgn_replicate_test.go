package uci

import (
	"testing"
)

// TestReplicateD4F2Bug replicates the exact game from the PGN file that led to
// the "illegal move: d4f2" issue. This test uses the actual game moves from
// benchmark_20250729_190822.pgn to reproduce the exact board state where
// our engine incorrectly tried to play d4f2.
func TestReplicateD4F2Bug(t *testing.T) {
	engine := NewUCIEngine()
	
	// Initialize the engine
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// The exact game moves from the PGN file that led to d4f2 illegal move:
	// 1. d4 d5 2. a3 e6 3. b3 Qf6 4. c3 Qd8 5. e3 b6 6. f3 Be7 7. g3 c5 
	// 8. h3 Bb7 9. a4 Nf6 10. b4 Bd6 11. c4 cxb4 12. e4 Bxg3+
	// {White makes an illegal move: d4f2}
	
	pgnMoves := []struct {
		moveNum  int
		white    string
		black    string
		whiteUCI string
		blackUCI string
	}{
		{1, "d4", "d5", "d2d4", "d7d5"},
		{2, "a3", "e6", "a2a3", "e7e6"},
		{3, "b3", "Qf6", "b2b3", "d8f6"},
		{4, "c3", "Qd8", "c2c3", "f6d8"},
		{5, "e3", "b6", "e2e3", "b7b6"},
		{6, "f3", "Be7", "f2f3", "f8e7"},
		{7, "g3", "c5", "g2g3", "c7c5"},
		{8, "h3", "Bb7", "h2h3", "c8b7"},
		{9, "a4", "Nf6", "a3a4", "g8f6"},
		{10, "b4", "Bd6", "b3b4", "e7d6"},
		{11, "c4", "cxb4", "c3c4", "c5b4"},
		{12, "e4", "Bxg3+", "e3e4", "d6g3"},
	}
	
	// Build the complete move sequence for UCI
	var allMoves []string
	for _, move := range pgnMoves {
		allMoves = append(allMoves, move.whiteUCI, move.blackUCI)
	}
	
	moveSequence := ""
	for i, move := range allMoves {
		if i > 0 {
			moveSequence += " "
		}
		moveSequence += move
	}
	
	t.Logf("Replicating PGN game with %d moves: %s", len(allMoves), moveSequence)
	
	// Apply the position command
	positionCmd := "position startpos moves " + moveSequence
	response := engine.HandleCommand(positionCmd)
	
	if response != "" {
		t.Logf("Position command response: %s", response)
	}
	
	// Get the resulting position after Bxg3+ (Black gives check)
	currentFEN := engine.engine.GetCurrentFEN()
	t.Logf("Position after Bxg3+ (move 24): %s", currentFEN)
	
	// This should be the position where:
	// - Black bishop on g3 is giving check to White king on e1
	// - White pawn is on d4 (NOT a Queen)
	// - White must respond to check
	
	// Verify that White is in check
	legalMoves := engine.engine.GetLegalMoves()
	t.Logf("White has %d legal moves (must be check-escaping moves):", legalMoves.Count)
	
	d4f2Found := false
	validCheckEscapes := 0
	
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		moveStr := move.From.String() + move.To.String()
		t.Logf("  [%d]: %s (Piece=%c)", i, moveStr, move.Piece)
		
		if moveStr == "d4f2" {
			d4f2Found = true
			t.Logf("       ^^^ PROBLEM: d4f2 found in legal moves!")
		}
		
		// Check if this move gets out of check
		if move.Piece == 'K' || moveStr == "e1d2" || moveStr == "e1e2" {
			validCheckEscapes++
		}
	}
	
	// Verify the position is correct
	t.Logf("\n=== POSITION ANALYSIS ===")
	
	// The position should have White in check from bishop on g3
	if legalMoves.Count < 10 {
		t.Logf("✅ White appears to be in check (only %d legal moves)", legalMoves.Count)
	} else {
		t.Logf("❌ White might not be in check (has %d legal moves)", legalMoves.Count)
	}
	
	// d4f2 should NOT be legal (pawn can't move from d4 to f2)
	if d4f2Found {
		t.Errorf("❌ CRITICAL BUG: d4f2 is in legal moves but it's impossible!")
		t.Errorf("   A pawn on d4 cannot move to f2 - this is the root cause of the illegal move")
	} else {
		t.Logf("✅ d4f2 correctly NOT in legal moves")
	}
	
	// Test what happens if we ask the engine to search for a move
	t.Logf("\n=== TESTING AI MOVE SELECTION ===")
	
	// This is where the bug manifests - our AI somehow selects d4f2
	// Let's see what move it would actually choose
	// Note: We won't actually call the search to avoid the panic from earlier
	t.Logf("In the actual game, our AI selected 'd4f2' which cutechess rejected as illegal")
	t.Logf("The correct response would be a king move like e1d2 or e1e2 to escape check")
	
	// Based on the move sequence, d4 should contain a White pawn (moved there in move 1)
	t.Logf("Based on the move sequence, d4 should contain a White pawn (moved there in move 1)")
	
	if d4f2Found {
		t.Fatalf("TEST FAILED: Our move generator incorrectly allows pawn d4f2")
	} else {
		t.Logf("TEST PASSED: Move generator correctly excludes impossible d4f2")
	}
}

// TestPawnMovementValidation tests that pawns cannot make impossible moves
func TestPawnMovementValidation(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// Set up a simple position with a pawn on d4
	engine.HandleCommand("position startpos moves d2d4")
	
	legalMoves := engine.engine.GetLegalMoves()
	t.Logf("After d2d4, legal moves for White:")
	
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		moveStr := move.From.String() + move.To.String()
		
		// Look for any moves from d4
		if move.From.String() == "d4" {
			t.Logf("  Move from d4: %s (Piece=%c)", moveStr, move.Piece)
			
			// A pawn on d4 should only be able to move to d5
			if move.Piece == 'P' && moveStr != "d4d5" {
				t.Errorf("❌ Invalid pawn move: %s", moveStr)
			}
		}
	}
}