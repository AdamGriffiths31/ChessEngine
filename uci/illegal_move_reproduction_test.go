package uci

import (
	"strings"
	"testing"
)

// TestD4E3IllegalMoveReproduction tests the exact position where d4e3 illegal move occurs
func TestD4E3IllegalMoveReproduction(t *testing.T) {
	// Exact FEN position from UCI logs where d4e3 illegal move was generated
	fenPosition := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	
	// Create engine instance
	engine := NewUCIEngine()
	
	// Set up the position
	positionCmd := "position fen " + fenPosition
	engine.HandleCommand(positionCmd)
	
	// Request a move at depth 1 to reproduce the bug
	response := engine.HandleCommand("go depth 1")
	
	// Check if the illegal move d4e3 is still being returned
	if strings.Contains(response, "bestmove d4e3") {
		t.Errorf("BUG REPRODUCED: Engine returned illegal move d4e3 in position with king in check")
		t.Errorf("Position: %s", fenPosition)
		t.Errorf("Response: %s", response)
		
		// Log the position analysis for debugging
		t.Logf("Position analysis:")
		t.Logf("- White King on d2 is in check from Black Bishop on f4")
		t.Logf("- Queen on d4 moving to e3 does NOT resolve the check")
		t.Logf("- Legal moves should only be King moves: d2e1, d2c2, d2e2, d2d3")
	} else {
		t.Logf("SUCCESS: Engine did not return illegal move d4e3")
		t.Logf("Engine response: %s", response)
	}
}

// TestKnownIllegalMovePositions tests all documented illegal move positions
func TestKnownIllegalMovePositions(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		illegalMove string
		description string
	}{
		{
			name:        "D4E3_Queen_Doesnt_Resolve_Check",
			fen:         "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14",
			illegalMove: "d4e3",
			description: "Queen on d4 moving to e3 doesn't resolve check from bishop on f4",
		},
		// Add more cases as they are discovered
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewUCIEngine()
			
			// Set up position
			positionCmd := "position fen " + tc.fen
			engine.HandleCommand(positionCmd)
			
			// Get engine's move choice
			response := engine.HandleCommand("go depth 1")
			
			// Check if illegal move is returned
			if strings.Contains(response, "bestmove "+tc.illegalMove) {
				t.Errorf("BUG: Engine returned illegal move %s", tc.illegalMove)
				t.Errorf("Position: %s", tc.fen)
				t.Errorf("Description: %s", tc.description)
				t.Errorf("Engine response: %s", response)
			} else {
				t.Logf("PASS: Engine avoided illegal move %s", tc.illegalMove)
				t.Logf("Engine chose: %s", response)
			}
		})
	}
}

// TestLegalKingMovesInCheck verifies that when king is in check, only legal moves are generated
func TestLegalKingMovesInCheck(t *testing.T) {
	// Position where White King on d2 is in check from Black Bishop on f4
	fenPosition := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	
	engine := NewUCIEngine()
	engine.HandleCommand("position fen " + fenPosition)
	
	// Get the engine's internal legal moves (this would require exposing the method)
	// For now, we test through the go command
	response := engine.HandleCommand("go depth 1")
	
	// Expected legal moves when king is in check:
	// - d2e1 (King moves)
	// - d2c2 (King moves) 
	// - d2e2 (King moves)
	// - d2d3 (King moves)
	// NOT d4e3 (Queen move that doesn't resolve check)
	
	expectedLegalMoves := []string{"d2e1", "d2c2", "d2e2", "d2d3"}
	
	// The response should contain one of the expected legal moves
	foundLegalMove := false
	var chosenMove string
	
	for _, move := range expectedLegalMoves {
		if strings.Contains(response, "bestmove "+move) {
			foundLegalMove = true
			chosenMove = move
			break
		}
	}
	
	if !foundLegalMove {
		t.Errorf("Engine did not choose a legal king move when in check")
		t.Errorf("Expected one of: %v", expectedLegalMoves)
		t.Errorf("Engine response: %s", response)
	} else {
		t.Logf("SUCCESS: Engine chose legal king move: %s", chosenMove)
	}
}