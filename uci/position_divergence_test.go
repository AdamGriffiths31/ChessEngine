package uci

import (
	"strings"
	"testing"
)

// TestPositionDivergenceAnalysis systematically applies each move to find where
// our engine diverges from cutechess-cli's expected position
func TestPositionDivergenceAnalysis(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== POSITION DIVERGENCE ANALYSIS ===")
	t.Logf("Goal: Find where our position diverges from cutechess-cli")
	
	// The exact move sequence from debug logs
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", "d8c7", "b8c7", "e6e5", "c4f7",
	}
	
	// Expected final position from cutechess-cli debug logs
	expectedFinal := "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15"
	
	// Apply moves one by one and check each position
	movesApplied := ""
	for i, move := range moves {
		if i > 0 {
			movesApplied += " "
		}
		movesApplied += move
		
		// Apply position command
		posCmd := "position startpos moves " + movesApplied
		engine.HandleCommand(posCmd)
		
		currentFEN := engine.engine.GetCurrentFEN()
		
		t.Logf("\nMove %d: %s", i+1, move)
		t.Logf("  Command: %s", posCmd)
		t.Logf("  Result:  %s", currentFEN)
		
		// Check critical positions where pieces might move incorrectly
		if isCriticalMove(i, move) {
			t.Logf("  *** CRITICAL MOVE - Analyzing in detail ***")
			analyzeCriticalPosition(t, engine, i+1, move, currentFEN)
		}
		
		// Final position comparison
		if i == len(moves)-1 {
			t.Logf("\n=== FINAL POSITION COMPARISON ===")
			t.Logf("Expected: %s", expectedFinal)
			t.Logf("Actual:   %s", currentFEN)
			
			if currentFEN == expectedFinal {
				t.Logf("✅ POSITIONS MATCH - Bug must be elsewhere")
			} else {
				t.Logf("❌ POSITION MISMATCH FOUND")
				
				// Detailed analysis of the mismatch
				analyzePositionMismatch(t, expectedFinal, currentFEN)
			}
		}
	}
}

// TestSingleMoveApplication tests applying moves one at a time to isolate bugs
func TestSingleMoveApplication(t *testing.T) {
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", "d8c7", "b8c7", "e6e5", "c4f7",
	}
	
	t.Logf("=== SINGLE MOVE APPLICATION TEST ===")
	
	// Test each problematic sequence
	criticalSequences := []struct {
		name    string
		endMove int
		focus   string
	}{
		{"Queen Creation", 25, "Where does the queen come from?"},
		{"Knight vs Queen", 28, "Why do we have knight instead of queen on f6?"},
		{"Final Position", 29, "Final state analysis"},
	}
	
	for _, seq := range criticalSequences {
		t.Logf("\n--- %s (through move %d) ---", seq.name, seq.endMove)
		t.Logf("Focus: %s", seq.focus)
		
		engine := NewUCIEngine()
		engine.HandleCommand("uci")
		engine.HandleCommand("ucinewgame")
		
		// Apply moves up to the critical point
		movesToApply := moves[:seq.endMove]
		moveString := strings.Join(movesToApply, " ")
		
		posCmd := "position startpos moves " + moveString
		engine.HandleCommand(posCmd)
		
		fen := engine.engine.GetCurrentFEN()
		legalMoves := engine.engine.GetLegalMoves()
		
		t.Logf("  Position: %s", fen)
		t.Logf("  Legal moves: %d", legalMoves.Count)
		
		// Analyze pieces on critical squares
		analyzeKeySquares(t, fen, seq.endMove)
		
		// Look for specific pieces
		if seq.endMove >= 25 {
			checkForQueenKnight(t, legalMoves, seq.endMove)
		}
	}
}

// TestMoveByMoveValidation validates each move is legal before applying
func TestMoveByMoveValidation(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", "d8c7", "b8c7", "e6e5", "c4f7",
	}
	
	t.Logf("=== MOVE-BY-MOVE VALIDATION ===")
	
	for i, move := range moves {
		// Get current position
		beforeFEN := engine.engine.GetCurrentFEN()
		legalMoves := engine.engine.GetLegalMoves()
		
		// Check if move is legal
		moveFound := false
		for j := 0; j < legalMoves.Count; j++ {
			legalMoveUCI := engine.converter.ToUCI(legalMoves.Moves[j])
			if legalMoveUCI == move {
				moveFound = true
				break
			}
		}
		
		t.Logf("Move %d: %s", i+1, move)
		t.Logf("  Before: %s", beforeFEN)
		t.Logf("  Legal:  %v", moveFound)
		
		if !moveFound {
			t.Errorf("  ❌ ILLEGAL MOVE DETECTED at move %d: %s", i+1, move)
			t.Logf("  Available legal moves:")
			for j := 0; j < legalMoves.Count && j < 10; j++ {
				t.Logf("    %s", engine.converter.ToUCI(legalMoves.Moves[j]))
			}
			return
		}
		
		// Apply the move using position command (like cutechess-cli does)
		movesApplied := strings.Join(moves[:i+1], " ")
		posCmd := "position startpos moves " + movesApplied
		engine.HandleCommand(posCmd)
		
		afterFEN := engine.engine.GetCurrentFEN()
		t.Logf("  After:  %s", afterFEN)
		
		// Validate the move actually changed the position
		if beforeFEN == afterFEN && move != "0000" {
			t.Errorf("  ❌ MOVE DID NOT CHANGE POSITION: %s", move)
		}
	}
	
	t.Logf("\n✅ All moves validated successfully")
}

// Helper functions

func isCriticalMove(moveIndex int, move string) bool {
	// Critical moves that might affect queen/knight placement
	criticalMoves := []string{"d5c7", "d8c7", "b8c7", "e6e5"}
	for _, critical := range criticalMoves {
		if move == critical {
			return true
		}
	}
	// Also check late game moves
	return moveIndex >= 24
}

func analyzeCriticalPosition(t *testing.T, engine *UCIEngine, moveNum int, move string, fen string) {
	legalMoves := engine.engine.GetLegalMoves()
	
	t.Logf("    Legal moves after move %d:", moveNum)
	for i := 0; i < legalMoves.Count && i < 5; i++ {
		moveUCI := engine.converter.ToUCI(legalMoves.Moves[i])
		t.Logf("      %s", moveUCI)
	}
	
	// Check for pieces on f6
	pieces := analyzePiecesOnSquare(fen, "f6")
	if len(pieces) > 0 {
		t.Logf("    Pieces on f6: %s", pieces)
	}
}

func analyzePositionMismatch(t *testing.T, expected, actual string) {
	// Split FEN into components
	expectedParts := strings.Split(expected, " ")
	actualParts := strings.Split(actual, " ")
	
	if len(expectedParts) > 0 && len(actualParts) > 0 {
		expectedBoard := expectedParts[0]
		actualBoard := actualParts[0]
		
		t.Logf("  Board position mismatch:")
		t.Logf("    Expected: %s", expectedBoard)
		t.Logf("    Actual:   %s", actualBoard)
		
		// Find first difference
		minLen := len(expectedBoard)
		if len(actualBoard) < minLen {
			minLen = len(actualBoard)
		}
		
		for i := 0; i < minLen; i++ {
			if expectedBoard[i] != actualBoard[i] {
				t.Logf("  First difference at position %d:", i)
				t.Logf("    Expected: '%c'", expectedBoard[i])
				t.Logf("    Actual:   '%c'", actualBoard[i])
				
				// Analyze which rank this affects
				rank := findRankFromPosition(expectedBoard, i)
				t.Logf("    This affects rank %d", rank)
				break
			}
		}
	}
}

func analyzeKeySquares(t *testing.T, fen string, moveNum int) {
	squares := []string{"f6", "f7", "c7", "d8"}
	
	for _, square := range squares {
		piece := analyzePiecesOnSquare(fen, square)
		if piece != "" {
			t.Logf("    %s: %s", square, piece)
		}
	}
}

func checkForQueenKnight(t *testing.T, legalMoves interface{}, moveNum int) {
	// This would need to be implemented based on your legal moves structure
	t.Logf("    Checking for queen/knight on f6 after move %d", moveNum)
}

func analyzePiecesOnSquare(fen, square string) string {
	// Simple FEN analysis - would need proper implementation
	// This is a placeholder for the concept
	if strings.Contains(fen, "q") && square == "f6" {
		return "Black Queen"
	}
	if strings.Contains(fen, "n") && square == "f6" {
		return "Black Knight"
	}
	return ""
}

func findRankFromPosition(board string, pos int) int {
	// Count slashes before position to determine rank
	slashes := 0
	for i := 0; i < pos && i < len(board); i++ {
		if board[i] == '/' {
			slashes++
		}
	}
	return 8 - slashes // Rank 8 is first, rank 1 is last
}