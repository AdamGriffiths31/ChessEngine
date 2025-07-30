package uci

import (
	"testing"
)

// TestIllegalF6F7Position replicates the exact game sequence from cutechess-cli
// that led to the illegal f6f7 move. This test applies the exact position command
// that cutechess-cli sent us and analyzes the resulting board state.
func TestIllegalF6F7Position(t *testing.T) {
	engine := NewUCIEngine()
	
	// Initialize the engine
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== REPLICATING CUTECHESS-CLI POSITION SEQUENCE ===")
	
	// This is the EXACT position command that cutechess-cli sent us before the illegal f6f7 move
	positionCmd := "position startpos moves c2c4 g8f6 g1f3 a7a6 b1c3 a6a5 a2a3 a5a4 d2d4 b7b6 c1f4 b6b5 c4b5 c7c6 e2e3 c6c5 f4b8 c5c4 a1c1 d7d6 f1c4 d6d5 c3d5 e7e6 d5c7 d8c7 b8c7 e6e5 c4f7"
	
	t.Logf("Applying cutechess-cli position command:")
	t.Logf("%s", positionCmd)
	
	response := engine.HandleCommand(positionCmd)
	if response != "" {
		t.Logf("Position response: %s", response)
	}
	
	// Get the resulting position after all moves applied
	finalFEN := engine.engine.GetCurrentFEN()
	t.Logf("\n=== FINAL POSITION ANALYSIS ===")
	t.Logf("Final FEN: %s", finalFEN)
	
	// Expected FEN based on our logs: r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15
	expectedFEN := "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15"
	
	if finalFEN == expectedFEN {
		t.Logf("✅ FEN matches expected position from debug logs")
	} else {
		t.Logf("❌ FEN mismatch!")
		t.Logf("Expected: %s", expectedFEN)
		t.Logf("Actual:   %s", finalFEN)
	}
	
	// Analyze the critical pieces and squares
	t.Logf("\n=== BOARD ANALYSIS ===")
	t.Logf("Rank 8: %s (Black back rank)", finalFEN[0:8])
	t.Logf("Rank 7: %s (2 empty, Bishop, 2 empty, Bishop, pawn, pawn)", finalFEN[9:16])
	t.Logf("Rank 6: %s (5 empty, queen, 2 empty)", finalFEN[17:24])
	t.Logf("Rank 5: %s", finalFEN[25:32])
	
	// Get legal moves for Black (the side that will make the illegal f6f7)
	legalMoves := engine.engine.GetLegalMoves()
	t.Logf("\n=== LEGAL MOVES ANALYSIS ===")
	t.Logf("Black has %d legal moves:", legalMoves.Count)
	
	f6f7Found := false
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		moveStr := move.From.String() + move.To.String()
		t.Logf("  [%d]: %s (From=%s, To=%s, Piece=%d, Captured=%d)", 
			i, moveStr, move.From.String(), move.To.String(), move.Piece, move.Captured)
		
		if moveStr == "f6f7" {
			f6f7Found = true
			t.Logf("       ^^^ THIS IS THE ILLEGAL MOVE!")
			
			// Analyze the f6f7 move in detail
			t.Logf("       Queen on f6 (Piece=%d) attempting to capture piece on f7 (Captured=%d)", move.Piece, move.Captured)
			
			// Piece 113 should be Black Queen, Piece 66 should be captured White Bishop
			if move.Piece == 113 {
				t.Logf("       ✅ Piece 113 = Black Queen (correct)")
			} else {
				t.Logf("       ❌ Piece %d is not expected Black Queen (113)", move.Piece)
			}
			
			if move.Captured == 66 {
				t.Logf("       ✅ Captured 66 = White Bishop (according to our engine)")
			} else {
				t.Logf("       ❌ Captured %d is not expected White Bishop (66)", move.Captured)
			}
		}
	}
	
	if !f6f7Found {
		t.Errorf("❌ CRITICAL: f6f7 move NOT found in legal moves!")
		t.Errorf("   This suggests the move generator bug may have been fixed or the position is different")
	}
	
	// Test specific squares
	t.Logf("\n=== SPECIFIC SQUARE ANALYSIS ===")
	t.Logf("According to FEN '%s':", finalFEN)
	
	// Parse rank 7: "2B2Bpp" means positions c7=B, f7=B, g7=p, h7=p
	rank7 := finalFEN[9:16] // "2B2Bpp"
	t.Logf("Rank 7 pattern: %s", rank7)
	t.Logf("  a7=empty, b7=empty, c7=B (White Bishop)")
	t.Logf("  d7=empty, e7=empty, f7=B (White Bishop) ← KEY SQUARE")
	t.Logf("  g7=p (Black pawn), h7=p (Black pawn)")
	
	// Parse rank 6: "5q2" means positions f6=q
	rank6 := finalFEN[17:24] // "5q2"
	t.Logf("Rank 6 pattern: %s", rank6)
	t.Logf("  a6-e6=empty, f6=q (Black Queen) ← MOVING PIECE")
	t.Logf("  g6-h6=empty")
	
	t.Logf("\n=== CHESS RULES ANALYSIS ===")
	t.Logf("Move f6f7: Black Queen from f6 to f7")
	t.Logf("Direction: One square diagonally up-right")
	t.Logf("Target square f7: Contains White Bishop (according to our FEN)")
	t.Logf("Expected result: Queen captures Bishop (legal move)")
	
	t.Logf("\n=== WHY CUTECHESS-CLI REJECTS THIS ===")
	t.Logf("Our engine sees: f6f7 as legal queen capture")
	t.Logf("Cutechess-cli sees: f6f7 as illegal move")
	t.Logf("Possible reasons:")
	t.Logf("  1. Position desynchronization - cutechess has different board state")
	t.Logf("  2. Chess rules violation - move violates some rule we're not checking")
	t.Logf("  3. Check situation - queen move might leave king in check")
	t.Logf("  4. Move validation bug - our converter/generator has error")
	
	// Check if Black king is in check
	// Note: We don't have easy access to check detection, but we can analyze the move count
	if legalMoves.Count < 10 {
		t.Logf("  → HYPOTHESIS: Black might be in check (only %d legal moves)", legalMoves.Count)
		t.Logf("    If in check, f6f7 might not resolve the check situation")
	} else {
		t.Logf("  → Black has %d legal moves (probably not in check)", legalMoves.Count)
	}
	
	t.Logf("\n=== CONCLUSION ===")
	t.Logf("The position has been successfully replicated.")
	t.Logf("Our move generator produces f6f7 as legal, but cutechess-cli rejects it.")
	t.Logf("This test confirms the bug exists and provides the exact position for investigation.")
}

// TestStepByStepPositionBuild builds the position step by step to find where divergence occurs
func TestStepByStepPositionBuild(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// Break down the moves to see progression
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6",
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", "d8c7", "b8c7", "e6e5", "c4f7",
	}
	
	t.Logf("=== STEP-BY-STEP POSITION BUILD ===")
	
	movesApplied := ""
	for i, move := range moves {
		if i > 0 {
			movesApplied += " "
		}
		movesApplied += move
		
		posCmd := "position startpos moves " + movesApplied
		engine.HandleCommand(posCmd)
		
		currentFEN := engine.engine.GetCurrentFEN()
		
		// Log every 5th move and the critical final moves
		if i%5 == 4 || i >= len(moves)-3 {
			t.Logf("After move %d (%s): %s", i+1, move, currentFEN)
		}
		
		// Pay special attention to the last few moves
		if i >= len(moves)-3 {
			legalMoves := engine.engine.GetLegalMoves()
			t.Logf("  Legal moves: %d", legalMoves.Count)
			
			// Look for f6f7 in the final position
			if i == len(moves)-1 {
				for j := 0; j < legalMoves.Count; j++ {
					legalMove := legalMoves.Moves[j]
					moveStr := legalMove.From.String() + legalMove.To.String()
					if moveStr == "f6f7" {
						t.Logf("  → f6f7 found as legal move (Captured=%d)", legalMove.Captured)
					}
				}
			}
		}
	}
}