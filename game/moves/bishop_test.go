package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestGenerateBishopMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white bishop moves from initial position - should have no moves (blocked by pawns)
	whiteMoves := gen.GenerateBishopMoves(b, White)
	if whiteMoves.Count != 0 {
		t.Errorf("Expected 0 white bishop moves from initial position, got %d", whiteMoves.Count)
	}
	
	// Test black bishop moves from initial position - should have no moves (blocked by pawns)
	blackMoves := gen.GenerateBishopMoves(b, Black)
	if blackMoves.Count != 0 {
		t.Errorf("Expected 0 black bishop moves from initial position, got %d", blackMoves.Count)
	}
}

func TestGenerateBishopMoves_CenterBoard(t *testing.T) {
	gen := NewGenerator()
	// White bishop on e4, empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/4B3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should have 13 moves: 4 diagonal directions × variable distances
	// Up-right: f5, g6, h7 (3 moves)
	// Up-left: d5, c6, b7, a8 (4 moves)
	// Down-right: f3, g2, h1 (3 moves)
	// Down-left: d3, c2, b1 (3 moves)
	expectedMoves := 13
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d bishop moves from center, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"e4f5": false, "e4g6": false, "e4h7": false, // Up-right
		"e4d5": false, "e4c6": false, "e4b7": false, "e4a8": false, // Up-left
		"e4f3": false, "e4g2": false, "e4h1": false, // Down-right
		"e4d3": false, "e4c2": false, "e4b1": false, // Down-left
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

func TestGenerateBishopMoves_Corner(t *testing.T) {
	gen := NewGenerator()
	// White bishop on a1 (corner), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/8/8/8/B7 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should have 7 moves: only up-right diagonal
	// Up-right: b2, c3, d4, e5, f6, g7, h8 (7 moves)
	expectedMoves := 7
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d bishop moves from corner, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a1b2": false, "a1c3": false, "a1d4": false, "a1e5": false,
		"a1f6": false, "a1g7": false, "a1h8": false,
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

func TestGenerateBishopMoves_BlockedByOwnPieces(t *testing.T) {
	gen := NewGenerator()
	// White bishop on e4, white pawns blocking on f5 and d3
	b, _ := board.FromFEN("8/8/8/5P2/4B3/3P4/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should not include moves to f5 (blocked by own pawn) or beyond f5 in that direction
	// Should not include moves to d3 (blocked by own pawn) or beyond d3 in that direction
	// Available moves: d5, c6, b7, a8 (up-left), f3, g2, h1 (down-right)
	expectedMoves := 7
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d bishop moves when blocked by own pieces, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Verify blocked moves don't exist
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && (move.To.String() == "f5" || move.To.String() == "d3") {
			t.Errorf("Bishop should not be able to move to %s (blocked by own piece)", move.To.String())
		}
		if move.From.String() == "e4" && (move.To.String() == "g6" || move.To.String() == "h7") {
			t.Errorf("Bishop should not be able to move to %s (path blocked by own piece)", move.To.String())
		}
		if move.From.String() == "e4" && (move.To.String() == "c2" || move.To.String() == "b1") {
			t.Errorf("Bishop should not be able to move to %s (path blocked by own piece)", move.To.String())
		}
	}
}

func TestGenerateBishopMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// White bishop on e4, black pawns on f5 and d3 for captures
	b, _ := board.FromFEN("8/8/8/5p2/4B3/3p4/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should have captures on f5 and d3, plus other normal moves
	foundCaptureF5 := false
	foundCaptureD3 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "f5" && move.IsCapture {
			foundCaptureF5 = true
		}
		if move.From.String() == "e4" && move.To.String() == "d3" && move.IsCapture {
			foundCaptureD3 = true
		}
	}
	
	if !foundCaptureF5 {
		t.Error("Expected capture move Be4xf5 not found")
	}
	if !foundCaptureD3 {
		t.Error("Expected capture move Be4xd3 not found")
	}
}

func TestGenerateBishopMoves_CaptureBlocksPath(t *testing.T) {
	gen := NewGenerator()
	// White bishop on e4, black pawn on f5, empty g6 and h7
	b, _ := board.FromFEN("8/8/8/5p2/4B3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should be able to capture on f5 but NOT move to g6 or h7 (path blocked by capture)
	foundCaptureF5 := false
	foundMoveG6 := false
	foundMoveH7 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "f5" {
			foundCaptureF5 = true
		}
		if move.From.String() == "e4" && move.To.String() == "g6" {
			foundMoveG6 = true
		}
		if move.From.String() == "e4" && move.To.String() == "h7" {
			foundMoveH7 = true
		}
	}
	
	if !foundCaptureF5 {
		t.Error("Expected capture move Be4xf5 not found")
	}
	if foundMoveG6 {
		t.Error("Unexpected move Be4-g6 found (should be blocked by capture)")
	}
	if foundMoveH7 {
		t.Error("Unexpected move Be4-h7 found (should be blocked by capture)")
	}
}

func TestGenerateBishopMoves_EdgePositions(t *testing.T) {
	gen := NewGenerator()
	
	// Test bishop on edge (h4)
	b, _ := board.FromFEN("8/8/8/8/7B/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should have moves in 2 diagonal directions only
	// Up-left: g5, f6, e7, d8 (4 moves)
	// Down-left: g3, f2, e1 (3 moves)
	expectedMoves := 7
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d bishop moves from edge position, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateBishopMoves_MultipleBishops(t *testing.T) {
	gen := NewGenerator()
	// Two white bishops on a1 and h8 - they're on the same diagonal so they block each other
	// Let's use different positions: a1 and a8 (different diagonals)
	b, _ := board.FromFEN("B7/8/8/8/8/8/8/B7 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Bishop on a1: 7 moves (up-right diagonal: b2,c3,d4,e5,f6,g7,h8)
	// Bishop on a8: 7 moves (down-right diagonal: b7,c6,d5,e4,f3,g2,h1)
	// Total: 14 moves (no overlap since they're on different diagonals)
	expectedMoves := 14
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d total moves for two bishops on different diagonals, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestBishopMoves_DiagonalDirections(t *testing.T) {
	gen := NewGenerator()
	// Test the diagonal moves logic with a bishop on d4
	b, _ := board.FromFEN("8/8/8/8/3B4/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Verify moves in each diagonal direction
	upRightMoves := 0
	upLeftMoves := 0
	downRightMoves := 0
	downLeftMoves := 0
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "d4" {
			fromFile, fromRank := int(move.From.File), int(move.From.Rank)
			toFile, toRank := int(move.To.File), int(move.To.Rank)
			
			if toFile > fromFile && toRank > fromRank { // Up-right
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
	
	// From d4: up-right (4 moves), up-left (3 moves), down-right (3 moves), down-left (3 moves)
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

func TestGenerateBishopMoves_ReusesSlidingLogic(t *testing.T) {
	gen := NewGenerator()
	// Test that bishop correctly reuses the sliding moves logic from rook
	b, _ := board.FromFEN("8/8/2p5/8/4B3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateBishopMoves(b, White)
	
	// Should be able to capture the pawn on c6 but not move beyond it
	foundCaptureC6 := false
	foundMoveB7 := false
	foundMoveA8 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "c6" && move.IsCapture {
			foundCaptureC6 = true
		}
		if move.From.String() == "e4" && move.To.String() == "b7" {
			foundMoveB7 = true
		}
		if move.From.String() == "e4" && move.To.String() == "a8" {
			foundMoveA8 = true
		}
	}
	
	if !foundCaptureC6 {
		t.Error("Expected capture move Be4xc6 not found")
	}
	if foundMoveB7 {
		t.Error("Unexpected move Be4-b7 found (should be blocked by capture)")
	}
	if foundMoveA8 {
		t.Error("Unexpected move Be4-a8 found (should be blocked by capture)")
	}
}