package uci

import (
	"strings"
	"testing"
)

// TestExactIllegalF6F7Position uses the EXACT position command from the debug logs
// that led to the illegal f6f7 move. This matches the actual cutechess-cli command.
func TestExactIllegalF6F7Position(t *testing.T) {
	engine := NewUCIEngine()
	
	// Initialize the engine
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== EXACT CUTECHESS-CLI POSITION REPLICATION ===")
	
	// This is the EXACT position command from the debug logs
	// From: [UCI-DEBUG] 2025/07/29 22:26:31.703446 CMD-PARSE: Command='position', Args=[startpos moves c2c4 g8f6 g1f3 a7a6 b1c3 a6a5 a2a3 a5a4 d2d4 b7b6 c1f4 b6b5 c4b5 c7c6 e2e3 c6c5 f4b8 c5c4 a1c1 d7d6 f1c4 d6d5 c3d5 e7e6 d5c7 d8c7 b8c7 e6e5 c4f7]
	exactPositionCmd := "position startpos moves c2c4 g8f6 g1f3 a7a6 b1c3 a6a5 a2a3 a5a4 d2d4 b7b6 c1f4 b6b5 c4b5 c7c6 e2e3 c6c5 f4b8 c5c4 a1c1 d7d6 f1c4 d6d5 c3d5 e7e6 d5c7 d8c7 b8c7 e6e5 c4f7"
	
	t.Logf("Applying exact position command:")
	t.Logf("%s", exactPositionCmd)
	
	response := engine.HandleCommand(exactPositionCmd)
	if response != "" {
		t.Logf("Position response: %s", response)
	}
	
	// Get the resulting position
	actualFEN := engine.engine.GetCurrentFEN()
	t.Logf("\n=== POSITION COMPARISON ===")
	t.Logf("Actual FEN:   %s", actualFEN)
	
	// From debug logs, the expected position was:
	// [UCI-DEBUG] 2025/07/29 22:26:31.705817 Move 15 search starting - Position: r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15, Player: Black
	expectedFEN := "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15"
	t.Logf("Expected FEN: %s", expectedFEN)
	
	if actualFEN == expectedFEN {
		t.Logf("✅ EXACT MATCH! Position replicated correctly")
	} else {
		t.Logf("❌ MISMATCH! Position differs from debug logs")
		
		// Analyze the differences character by character
		t.Logf("\n=== DETAILED DIFFERENCE ANALYSIS ===")
		if len(actualFEN) != len(expectedFEN) {
			t.Logf("Length difference: actual=%d, expected=%d", len(actualFEN), len(expectedFEN))
		}
		
		minLen := len(actualFEN)
		if len(expectedFEN) < minLen {
			minLen = len(expectedFEN)
		}
		
		for i := 0; i < minLen; i++ {
			if actualFEN[i] != expectedFEN[i] {
				t.Logf("First difference at position %d:", i)
				t.Logf("  Actual:   '%c' (char %d)", actualFEN[i], actualFEN[i])
				t.Logf("  Expected: '%c' (char %d)", expectedFEN[i], expectedFEN[i])
				
				// Show context around the difference
				start := i - 10
				if start < 0 {
					start = 0
				}
				end := i + 10
				if end > minLen {
					end = minLen
				}
				
				t.Logf("  Context - Actual:   '%s'", actualFEN[start:end])
				t.Logf("  Context - Expected: '%s'", expectedFEN[start:end])
				break
			}
		}
	}
	
	// Get legal moves to see if f6f7 is present
	legalMoves := engine.engine.GetLegalMoves()
	t.Logf("\n=== LEGAL MOVES ANALYSIS ===")
	t.Logf("Position has %d legal moves:", legalMoves.Count)
	
	f6f7Found := false
	queenOnF6 := false
	knightOnF6 := false
	
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		moveStr := move.From.String() + move.To.String()
		t.Logf("  [%d]: %s (From=%s, To=%s, Piece=%d, Captured=%d)", 
			i, moveStr, move.From.String(), move.To.String(), move.Piece, move.Captured)
		
		// Check what's on f6
		if move.From.String() == "f6" {
			if move.Piece == 113 { // Black Queen
				queenOnF6 = true
			} else if move.Piece == 110 { // Black Knight
				knightOnF6 = true
			}
		}
		
		if moveStr == "f6f7" {
			f6f7Found = true
			t.Logf("       ^^^ ILLEGAL MOVE FOUND!")
			t.Logf("       This is the move that cutechess-cli rejected")
		}
	}
	
	// Determine what piece is actually on f6
	t.Logf("\n=== F6 SQUARE ANALYSIS ===")
	if queenOnF6 {
		t.Logf("✅ Black Queen found on f6 (matches debug logs)")
	} else if knightOnF6 {
		t.Logf("❌ Black Knight found on f6 (doesn't match debug logs)")
	} else {
		t.Logf("❌ No piece found moving from f6")
	}
	
	t.Logf("\n=== EXPECTED VS ACTUAL PIECE PLACEMENT ===")
	// Parse the 6th rank from both FENs
	actualRank6 := getRank6FromFEN(actualFEN)
	expectedRank6 := getRank6FromFEN(expectedFEN)
	
	t.Logf("Rank 6 (actual):   '%s'", actualRank6)
	t.Logf("Rank 6 (expected): '%s'", expectedRank6)
	
	if actualRank6 == expectedRank6 {
		t.Logf("✅ Rank 6 matches expected")
	} else {
		t.Logf("❌ Rank 6 differs - this explains the f6f7 discrepancy")
	}
	
	t.Logf("\n=== CONCLUSION ===")
	if f6f7Found && queenOnF6 {
		t.Logf("🎯 SUCCESS: Replicated the exact illegal move scenario!")
		t.Logf("   - f6f7 move is present in legal moves")
		t.Logf("   - Black Queen is on f6 (matches debug logs)")
		t.Logf("   - This explains why our AI selected f6f7")
		t.Logf("   - But cutechess-cli rejected it as illegal")
	} else if !f6f7Found {
		t.Logf("❌ FAILED: f6f7 move not found in legal moves")
		t.Logf("   This suggests position replication failed")
	} else if knightOnF6 {
		t.Logf("❌ FAILED: Knight on f6 instead of Queen")
		t.Logf("   Position replication produced different result")
	}
}

// Helper function to extract rank 6 from FEN string
func getRank6FromFEN(fen string) string {
	parts := strings.Split(fen, " ")
	if len(parts) == 0 {
		return ""
	}
	
	ranks := strings.Split(parts[0], "/")
	if len(ranks) < 6 {
		return ""
	}
	
	// Rank 6 is index 2 (rank 8=0, rank 7=1, rank 6=2, ...)
	return ranks[2]
}


// TestMoveByMoveExactReplication applies moves one by one to debug where divergence occurs
func TestMoveByMoveExactReplication(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// The exact move sequence from debug logs
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6",
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", "d8c7", "b8c7", "e6e5", "c4f7",
	}
	
	t.Logf("=== MOVE-BY-MOVE EXACT REPLICATION ===")
	
	// Key positions to track
	keyMoves := []int{23, 24, 25, 26, 27, 28} // Last few moves
	
	movesApplied := ""
	for i, move := range moves {
		if i > 0 {
			movesApplied += " "
		}
		movesApplied += move
		
		posCmd := "position startpos moves " + movesApplied
		engine.HandleCommand(posCmd)
		
		currentFEN := engine.engine.GetCurrentFEN()
		
		// Focus on the critical final moves
		if contains(keyMoves, i) || i == len(moves)-1 {
			t.Logf("Move %d (%s): %s", i+1, move, currentFEN)
			
			// Check for queen/knight on f6 in final position
			if i == len(moves)-1 {
				legalMoves := engine.engine.GetLegalMoves()
				t.Logf("  Final position has %d legal moves", legalMoves.Count)
				
				for j := 0; j < legalMoves.Count; j++ {
					legalMove := legalMoves.Moves[j]
					moveStr := legalMove.From.String() + legalMove.To.String()
					
					if legalMove.From.String() == "f6" {
						if legalMove.Piece == 113 {
							t.Logf("  → Queen on f6 can move to %s", legalMove.To.String())
						} else if legalMove.Piece == 110 {
							t.Logf("  → Knight on f6 can move to %s", legalMove.To.String())
						}
					}
					
					if moveStr == "f6f7" {
						t.Logf("  → FOUND f6f7 in legal moves!")
					}
				}
			}
		}
	}
}

// Helper function to check if slice contains value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}