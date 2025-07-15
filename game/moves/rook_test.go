package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestGenerateRookMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white rook moves from initial position - should have no moves (blocked by pawns)
	whiteMoves := gen.GenerateRookMoves(b, White)
	if whiteMoves.Count != 0 {
		t.Errorf("Expected 0 white rook moves from initial position, got %d", whiteMoves.Count)
	}
	
	// Test black rook moves from initial position - should have no moves (blocked by pawns)
	blackMoves := gen.GenerateRookMoves(b, Black)
	if blackMoves.Count != 0 {
		t.Errorf("Expected 0 black rook moves from initial position, got %d", blackMoves.Count)
	}
}

func TestGenerateRookMoves_CenterBoard(t *testing.T) {
	gen := NewGenerator()
	// White rook on e4, empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/4R3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should have 14 moves: 4 directions × variable distances
	// Up: e5, e6, e7, e8 (4 moves)
	// Down: e3, e2, e1 (3 moves)
	// Right: f4, g4, h4 (3 moves)
	// Left: d4, c4, b4, a4 (4 moves)
	expectedMoves := 14
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d rook moves from center, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"e4e5": false, "e4e6": false, "e4e7": false, "e4e8": false, // Up
		"e4e3": false, "e4e2": false, "e4e1": false,             // Down
		"e4f4": false, "e4g4": false, "e4h4": false,             // Right
		"e4d4": false, "e4c4": false, "e4b4": false, "e4a4": false, // Left
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

func TestGenerateRookMoves_Corner(t *testing.T) {
	gen := NewGenerator()
	// White rook on a1 (corner), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/8/8/8/R7 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should have 14 moves: 2 directions only
	// Up: a2, a3, a4, a5, a6, a7, a8 (7 moves)
	// Right: b1, c1, d1, e1, f1, g1, h1 (7 moves)
	expectedMoves := 14
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d rook moves from corner, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateRookMoves_BlockedByOwnPieces(t *testing.T) {
	gen := NewGenerator()
	// White rook on e4, white pawns blocking on e5 and f4
	b, _ := board.FromFEN("8/8/8/4P3/4RP2/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should not include moves to e5 (blocked by own pawn) or f4 (blocked by own pawn)
	// Available moves: e3, e2, e1 (down), d4, c4, b4, a4 (left)
	expectedMoves := 7
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d rook moves when blocked by own pieces, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Verify blocked moves don't exist
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && (move.To.String() == "e5" || move.To.String() == "f4") {
			t.Errorf("Rook should not be able to move to %s (blocked by own piece)", move.To.String())
		}
	}
}

func TestGenerateRookMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// White rook on e4, black pawns on e6 and g4 for captures
	b, _ := board.FromFEN("8/8/4p3/8/4R1p1/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should have captures exf6 and gxg4, plus other normal moves
	foundCaptureE6 := false
	foundCaptureG4 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "e6" && move.IsCapture {
			foundCaptureE6 = true
		}
		if move.From.String() == "e4" && move.To.String() == "g4" && move.IsCapture {
			foundCaptureG4 = true
		}
	}
	
	if !foundCaptureE6 {
		t.Error("Expected capture move Re4xe6 not found")
	}
	if !foundCaptureG4 {
		t.Error("Expected capture move Re4xg4 not found")
	}
}

func TestGenerateRookMoves_CaptureBlocksPath(t *testing.T) {
	gen := NewGenerator()
	// White rook on e4, black pawn on e6, empty e7 and e8
	b, _ := board.FromFEN("8/8/4p3/8/4R3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should be able to capture on e6 but NOT move to e7 or e8 (path blocked by capture)
	foundCaptureE6 := false
	foundMoveE7 := false
	foundMoveE8 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "e6" {
			foundCaptureE6 = true
		}
		if move.From.String() == "e4" && move.To.String() == "e7" {
			foundMoveE7 = true
		}
		if move.From.String() == "e4" && move.To.String() == "e8" {
			foundMoveE8 = true
		}
	}
	
	if !foundCaptureE6 {
		t.Error("Expected capture move Re4xe6 not found")
	}
	if foundMoveE7 {
		t.Error("Unexpected move Re4-e7 found (should be blocked by capture)")
	}
	if foundMoveE8 {
		t.Error("Unexpected move Re4-e8 found (should be blocked by capture)")
	}
}

func TestGenerateRookMoves_EdgePositions(t *testing.T) {
	gen := NewGenerator()
	
	// Test rook on edge (h4)
	b, _ := board.FromFEN("8/8/8/8/7R/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Should have moves in 3 directions only (no moves to the right off board)
	// Up: h5, h6, h7, h8 (4 moves)
	// Down: h3, h2, h1 (3 moves)
	// Left: g4, f4, e4, d4, c4, b4, a4 (7 moves)
	expectedMoves := 14
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d rook moves from edge position, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateRookMoves_MultipleRooks(t *testing.T) {
	gen := NewGenerator()
	// Two white rooks on a1 and h8
	b, _ := board.FromFEN("7R/8/8/8/8/8/8/R7 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Each rook should have 14 moves (7 in each of 2 directions)
	// Total: 28 moves
	expectedMoves := 28
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d total moves for two rooks, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestSlidingMoves_DirectionalLogic(t *testing.T) {
	gen := NewGenerator()
	// Test the sliding moves logic directly with a rook on d4
	b, _ := board.FromFEN("8/8/8/8/3R4/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateRookMoves(b, White)
	
	// Verify moves in each direction
	upMoves := 0
	downMoves := 0
	leftMoves := 0
	rightMoves := 0
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "d4" {
			to := move.To.String()
			switch {
			case to[0] == 'd' && to[1] > '4': // Up
				upMoves++
			case to[0] == 'd' && to[1] < '4': // Down
				downMoves++
			case to[1] == '4' && to[0] < 'd': // Left
				leftMoves++
			case to[1] == '4' && to[0] > 'd': // Right
				rightMoves++
			}
		}
	}
	
	// Should have 4 moves up, 3 down, 3 left, 4 right
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
}