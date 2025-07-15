package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestGenerateQueenMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white queen moves from initial position - should have no moves (blocked by pawns)
	whiteMoves := gen.GenerateQueenMoves(b, White)
	if whiteMoves.Count != 0 {
		t.Errorf("Expected 0 white queen moves from initial position, got %d", whiteMoves.Count)
	}
	
	// Test black queen moves from initial position - should have no moves (blocked by pawns)
	blackMoves := gen.GenerateQueenMoves(b, Black)
	if blackMoves.Count != 0 {
		t.Errorf("Expected 0 black queen moves from initial position, got %d", blackMoves.Count)
	}
}

func TestGenerateQueenMoves_CenterBoard(t *testing.T) {
	gen := NewGenerator()
	// White queen on e4, empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/4Q3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should have 27 moves: combines rook (14) + bishop (13) moves from center
	// Rook moves: 4 directions × variable distances (e5,e6,e7,e8 + e3,e2,e1 + f4,g4,h4 + d4,c4,b4,a4) = 14
	// Bishop moves: 4 diagonals × variable distances (f5,g6,h7 + d5,c6,b7,a8 + f3,g2,h1 + d3,c2,b1) = 13
	expectedMoves := 27
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d queen moves from center, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check some specific moves exist (both rook-like and bishop-like)
	expectedMovesMap := map[string]bool{
		// Rook-like moves
		"e4e8": false, "e4e1": false, "e4h4": false, "e4a4": false,
		// Bishop-like moves
		"e4h7": false, "e4a8": false, "e4h1": false, "e4b1": false,
	}
	
	for _, move := range whiteMoves.Moves {
		moveStr := move.From.String() + move.To.String()
		if _, exists := expectedMovesMap[moveStr]; exists {
			expectedMovesMap[moveStr] = true
		}
	}
	
	for moveStr, found := range expectedMovesMap {
		if !found {
			t.Errorf("Expected move %s not found", moveStr)
		}
	}
}

func TestGenerateQueenMoves_Corner(t *testing.T) {
	gen := NewGenerator()
	// White queen on a1 (corner), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/8/8/8/Q7 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should have 21 moves: rook (14) + bishop (7) from corner
	// Rook moves: up (a2-a8) = 7, right (b1-h1) = 7, total = 14
	// Bishop moves: up-right diagonal (b2-h8) = 7
	expectedMoves := 21
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d queen moves from corner, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateQueenMoves_BlockedByOwnPieces(t *testing.T) {
	gen := NewGenerator()
	// White queen on e4, white pawns blocking on e5, f4, f5, and d3
	b, _ := board.FromFEN("8/8/8/4P1P1/4QP2/3P4/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Looking at the board position "8/8/8/4P1P1/4QP2/3P4/8/8":
	// Rank 5: e5=P, g5=P  
	// Rank 4: e4=Q, f4=P
	// Rank 3: d3=P
	// So from e4, the available moves are:
	// Down: e3,e2,e1 (3 moves) 
	// Left: d4,c4,b4,a4 (4 moves)
	// Up-left: d5,c6,b7,a8 (4 moves) 
	// Down-right: f3,g2,h1 (3 moves)
	// Up-right: f5 is blocked by pawn at g5, wait... f5 should be available, g5 has the pawn
	// Let me re-examine: g5 has pawn, so f5 should be available, then g6,h7 should also be available
	// Up-right: f5,g6,h7 (3 moves)
	// Total: 3 + 4 + 4 + 3 + 3 = 17 moves - this matches!
	expectedMoves := 17
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d queen moves when blocked by own pieces, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Verify blocked moves don't exist
	blockedSquares := []string{"e5", "f4", "d3"} // Removed f5 since g5 has the pawn, not f5
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" {
			for _, blocked := range blockedSquares {
				if move.To.String() == blocked {
					t.Errorf("Queen should not be able to move to %s (blocked by own piece)", blocked)
				}
			}
		}
	}
}

func TestGenerateQueenMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// White queen on e4, black pawns on e6, f4, f5, and d3 for captures
	b, _ := board.FromFEN("8/8/4p3/5p2/4Qp2/3p4/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should have captures on e6, f4, f5, d3, plus other normal moves
	foundCaptureE6 := false
	foundCaptureF4 := false
	foundCaptureF5 := false
	foundCaptureD3 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.IsCapture {
			switch move.To.String() {
			case "e6":
				foundCaptureE6 = true
			case "f4":
				foundCaptureF4 = true
			case "f5":
				foundCaptureF5 = true
			case "d3":
				foundCaptureD3 = true
			}
		}
	}
	
	if !foundCaptureE6 {
		t.Error("Expected capture move Qe4xe6 not found")
	}
	if !foundCaptureF4 {
		t.Error("Expected capture move Qe4xf4 not found")
	}
	if !foundCaptureF5 {
		t.Error("Expected capture move Qe4xf5 not found")
	}
	if !foundCaptureD3 {
		t.Error("Expected capture move Qe4xd3 not found")
	}
}

func TestGenerateQueenMoves_CaptureBlocksPath(t *testing.T) {
	gen := NewGenerator()
	// White queen on e4, black pawn on f5, empty g6 and h7
	b, _ := board.FromFEN("8/8/8/5p2/4Q3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should be able to capture on f5 but NOT move to g6 or h7 (path blocked by capture)
	foundCaptureF5 := false
	foundMoveG6 := false
	foundMoveH7 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" {
			switch move.To.String() {
			case "f5":
				foundCaptureF5 = true
			case "g6":
				foundMoveG6 = true
			case "h7":
				foundMoveH7 = true
			}
		}
	}
	
	if !foundCaptureF5 {
		t.Error("Expected capture move Qe4xf5 not found")
	}
	if foundMoveG6 {
		t.Error("Unexpected move Qe4-g6 found (should be blocked by capture)")
	}
	if foundMoveH7 {
		t.Error("Unexpected move Qe4-h7 found (should be blocked by capture)")
	}
}

func TestGenerateQueenMoves_AllDirections(t *testing.T) {
	gen := NewGenerator()
	// Test the queen moves logic with a queen on d4
	b, _ := board.FromFEN("8/8/8/8/3Q4/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Verify moves in each of the 8 directions
	upMoves := 0
	downMoves := 0
	leftMoves := 0
	rightMoves := 0
	upRightMoves := 0
	upLeftMoves := 0
	downRightMoves := 0
	downLeftMoves := 0
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "d4" {
			fromFile, fromRank := int(move.From.File), int(move.From.Rank)
			toFile, toRank := int(move.To.File), int(move.To.Rank)
			
			if fromFile == toFile && toRank > fromRank { // Up
				upMoves++
			} else if fromFile == toFile && toRank < fromRank { // Down
				downMoves++
			} else if fromRank == toRank && toFile < fromFile { // Left
				leftMoves++
			} else if fromRank == toRank && toFile > fromFile { // Right
				rightMoves++
			} else if toFile > fromFile && toRank > fromRank { // Up-right
				upRightMoves++
			} else if toFile < fromFile && toRank > fromRank { // Up-left
				upLeftMoves++
			} else if toFile > fromFile && toRank < fromRank { // Down-right
				downRightMoves++
			} else if toFile < fromFile && toRank < fromRank { // Down-left
				downLeftMoves++
			}
		}
	}
	
	// From d4: up (4), down (3), left (3), right (4), up-right (4), up-left (3), down-right (3), down-left (3)
	if upMoves != 4 {
		t.Errorf("Expected 4 up moves, got %d", upMoves)
	}
	if downMoves != 3 {
		t.Errorf("Expected 3 down moves, got %d", downMoves)
	}
	if leftMoves != 3 {
		t.Errorf("Expected 3 left moves, got %d", leftMoves)
	}
	if rightMoves != 4 {
		t.Errorf("Expected 4 right moves, got %d", rightMoves)
	}
	if upRightMoves != 4 {
		t.Errorf("Expected 4 up-right moves, got %d", upRightMoves)
	}
	if upLeftMoves != 3 {
		t.Errorf("Expected 3 up-left moves, got %d", upLeftMoves)
	}
	if downRightMoves != 3 {
		t.Errorf("Expected 3 down-right moves, got %d", downRightMoves)
	}
	if downLeftMoves != 3 {
		t.Errorf("Expected 3 down-left moves, got %d", downLeftMoves)
	}
}

func TestGenerateQueenMoves_EdgePositions(t *testing.T) {
	gen := NewGenerator()
	
	// Test queen on edge (h4)
	b, _ := board.FromFEN("8/8/8/8/7Q/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should have moves in available directions from edge
	// Up: h5,h6,h7,h8 (4), Down: h3,h2,h1 (3), Left: g4,f4,e4,d4,c4,b4,a4 (7)
	// Up-left: g5,f6,e7,d8 (4), Down-left: g3,f2,e1 (3)
	// Total: 4+3+7+4+3 = 21 moves
	expectedMoves := 21
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d queen moves from edge position, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateQueenMoves_ReusesSlidingLogic(t *testing.T) {
	gen := NewGenerator()
	// Test that queen correctly reuses the sliding moves logic
	b, _ := board.FromFEN("8/8/2p5/8/4Q3/8/2p5/8 w - - 0 1")
	
	whiteMoves := gen.GenerateQueenMoves(b, White)
	
	// Should be able to capture pawns on c6 and c2 but not move beyond them
	foundCaptureC6 := false
	foundCaptureC2 := false
	foundMoveB7 := false
	foundMoveA8 := false
	foundMoveB1 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" {
			switch move.To.String() {
			case "c6":
				if move.IsCapture {
					foundCaptureC6 = true
				}
			case "c2":
				if move.IsCapture {
					foundCaptureC2 = true
				}
			case "b7":
				foundMoveB7 = true
			case "a8":
				foundMoveA8 = true
			case "b1":
				foundMoveB1 = true
			}
		}
	}
	
	if !foundCaptureC6 {
		t.Error("Expected capture move Qe4xc6 not found")
	}
	if !foundCaptureC2 {
		t.Error("Expected capture move Qe4xc2 not found")
	}
	if foundMoveB7 {
		t.Error("Unexpected move Qe4-b7 found (should be blocked by capture on c6)")
	}
	if foundMoveA8 {
		t.Error("Unexpected move Qe4-a8 found (should be blocked by capture on c6)")
	}
	if foundMoveB1 {
		t.Error("Unexpected move Qe4-b1 found (should be blocked by capture on c2)")
	}
}