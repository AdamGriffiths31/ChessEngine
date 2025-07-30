package uci

import (
	"strings"
	"testing"
	
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestPieceIdentificationBugIsolated specifically tests the d5c7 move to see what piece
// is actually being moved vs what should be moved
func TestPieceIdentificationBugIsolated(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== PIECE IDENTIFICATION BUG ANALYSIS ===")
	
	// Build up to the critical move d5c7 (move 25)
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6", // Stop here - move 24, right before d5c7
	}
	
	// Apply moves up to the critical point
	movesApplied := strings.Join(moves, " ")
	posCmd := "position startpos moves " + movesApplied
	engine.HandleCommand(posCmd)
	
	currentFEN := engine.engine.GetCurrentFEN()
	t.Logf("Position before d5c7: %s", currentFEN)
	
	// Examine the board state manually to understand what pieces are where
	t.Logf("\n=== BOARD ANALYSIS BEFORE d5c7 ===")
	
	// Get access to the internal board (we need this for piece analysis)
	internalBoard := engine.engine.GetState().Board
	
	// Check critical squares
	criticalSquares := []string{"d5", "c7", "f6"}
	for _, square := range criticalSquares {
		piece := getBugPieceAtSquare(internalBoard, square)
		color := getBugPieceColorString(piece)
		t.Logf("Square %s: %s (%s)", square, bugDescribePiece(piece), color)
	}
	
	// Now test what happens when we try to apply d5c7
	t.Logf("\n=== TESTING d5c7 CONVERSION ===")
	
	// Manually convert the UCI move to see what piece it picks up
	uciMove := "d5c7"
	converter := NewMoveConverter()
	
	// This is the critical line - what piece does it find at d5?
	move, err := converter.FromUCIWithLogging(uciMove, internalBoard, engine.debugLogger)
	if err != nil {
		t.Errorf("Failed to convert UCI move %s: %v", uciMove, err)
		return
	}
	
	t.Logf("\n=== MOVE CONVERSION RESULT ===")
	t.Logf("UCI move: %s", uciMove)
	t.Logf("From square: %s", move.From.String())
	t.Logf("To square: %s", move.To.String()) 
	t.Logf("Piece being moved: %c (%s)", move.Piece, bugDescribePiece(byte(move.Piece)))
	t.Logf("Piece color: %s", getBugPieceColorString(byte(move.Piece)))
	t.Logf("Captured piece: %c (%s)", move.Captured, bugDescribePiece(byte(move.Captured)))
	
	// THE CRITICAL TEST: Is this the right piece?
	t.Logf("\n=== PIECE EXPECTATION ANALYSIS ===")
	
	// Based on the game sequence, what piece SHOULD be moving?
	// The knight that moved g8→f6 should now be on f6
	// The move d5c7 in chess notation suggests a piece on d5 moving to c7
	// BUT, in our previous analysis, we saw the f6 knight should move to c7
	
	knightOnF6 := getBugPieceAtSquare(internalBoard, "f6")
	pieceOnD5 := getBugPieceAtSquare(internalBoard, "d5")
	
	t.Logf("Piece on f6: %c (%s %s)", knightOnF6, getBugPieceColorString(knightOnF6), bugDescribePiece(knightOnF6))
	t.Logf("Piece on d5: %c (%s %s)", pieceOnD5, getBugPieceColorString(pieceOnD5), bugDescribePiece(pieceOnD5))
	
	// Test: What piece is the d5c7 move actually picking up?
	if byte(move.Piece) != knightOnF6 {
		t.Logf("🚨 BUG IDENTIFIED!")
		t.Logf("   Expected: Move should involve the Black Knight on f6")
		t.Logf("   Actual: Move is using piece %c (%s) from d5", move.Piece, bugDescribePiece(byte(move.Piece)))
		t.Logf("   This explains why the knight stays on f6 instead of moving to c7!")
		
		// Additional analysis
		if getBugPieceColorString(byte(move.Piece)) == "White" {
			t.Logf("🔍 ROOT CAUSE: A WHITE piece is being moved instead of the BLACK knight")
			t.Logf("   This means our engine thinks there's a White piece on d5 that can capture c7")
			t.Logf("   But the BLACK knight on f6 should be the one moving to c7")
		}
	} else {
		t.Logf("✅ UNEXPECTED: The correct piece is being selected")
		t.Logf("   This suggests the bug might be elsewhere in the process")
	}
	
	// Final test: Apply the move and see what happens
	t.Logf("\n=== TESTING MOVE APPLICATION ===")
	
	fullCmd := posCmd + " d5c7"
	engine.HandleCommand(fullCmd)
	
	finalFEN := engine.engine.GetCurrentFEN()
	t.Logf("Position after d5c7: %s", finalFEN)
	
	// Check if the f6 knight moved
	finalBoard := engine.engine.GetState().Board
	f6After := getBugPieceAtSquare(finalBoard, "f6")
	c7After := getBugPieceAtSquare(finalBoard, "c7")
	
	t.Logf("f6 after move: %c (%s)", f6After, bugDescribePiece(f6After))
	t.Logf("c7 after move: %c (%s)", c7After, bugDescribePiece(c7After))
	
	if f6After != 0 { // 0 means Empty
		t.Logf("❌ CONFIRMED BUG: The f6 knight did NOT move!")
		t.Logf("   Expected: f6 should be empty after knight moves to c7")
		t.Logf("   Actual: f6 still contains %c", f6After)
	} else {
		t.Logf("✅ f6 is now empty - knight did move")
	}
}

// Helper functions with unique names to avoid conflicts
func getBugPieceAtSquare(boardPtr *board.Board, square string) byte {
	// Convert square string (e.g., "d5") to rank, file
	if len(square) != 2 {
		return 0
	}
	
	file := int(square[0] - 'a') // a=0, b=1, c=2, etc.
	rank := int(square[1] - '1') // 1=0, 2=1, 3=2, etc.
	
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return 0
	}
	
	piece := boardPtr.GetPiece(rank, file)
	return byte(piece)
}

func getBugPieceColorString(piece byte) string {
	if piece >= 'A' && piece <= 'Z' {
		return "White"
	} else if piece >= 'a' && piece <= 'z' {
		return "Black"
	}
	return "Empty"
}

func bugDescribePiece(piece byte) string {
	switch piece {
	case 'K': return "King"
	case 'Q': return "Queen" 
	case 'R': return "Rook"
	case 'B': return "Bishop"
	case 'N': return "Knight"
	case 'P': return "Pawn"
	case 'k': return "King"
	case 'q': return "Queen"
	case 'r': return "Rook"
	case 'b': return "Bishop"
	case 'n': return "Knight"
	case 'p': return "Pawn"
	case 0: return "Empty"
	default: return "Unknown"
	}
}