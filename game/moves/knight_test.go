package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestGenerateKnightMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white knight moves from initial position
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have 4 moves: Nb1-a3, Nb1-c3, Ng1-f3, Ng1-h3
	expectedMoves := 4
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d white knight moves from initial position, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"b1a3": false, "b1c3": false,
		"g1f3": false, "g1h3": false,
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
	
	// Test black knight moves from initial position
	blackMoves := gen.GenerateKnightMoves(b, Black)
	
	// Should have 4 moves: Nb8-a6, Nb8-c6, Ng8-f6, Ng8-h6
	expectedMoves = 4
	if blackMoves.Count != expectedMoves {
		t.Errorf("Expected %d black knight moves from initial position, got %d", expectedMoves, blackMoves.Count)
	}
}

func TestGenerateKnightMoves_CenterBoard(t *testing.T) {
	gen := NewGenerator()
	// White knight on e4, empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/4N3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have 8 moves from center: all L-shaped moves
	// d6, f6, c5, g5, c3, g3, d2, f2
	expectedMoves := 8
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d knight moves from center, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"e4d6": false, "e4f6": false, // Up 2, left/right 1
		"e4c5": false, "e4g5": false, // Up 1, left/right 2
		"e4c3": false, "e4g3": false, // Down 1, left/right 2
		"e4d2": false, "e4f2": false, // Down 2, left/right 1
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

func TestGenerateKnightMoves_Corner(t *testing.T) {
	gen := NewGenerator()
	// White knight on a1 (corner), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/8/8/8/N7 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have 2 moves from corner: b3, c2
	expectedMoves := 2
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d knight moves from corner, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a1b3": false, "a1c2": false,
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

func TestGenerateKnightMoves_Edge(t *testing.T) {
	gen := NewGenerator()
	// White knight on a4 (edge), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/N7/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have 4 moves from edge: b6, c5, c3, b2
	expectedMoves := 4
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d knight moves from edge, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a4b6": false, "a4c5": false,
		"a4c3": false, "a4b2": false,
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

func TestGenerateKnightMoves_BlockedByOwnPieces(t *testing.T) {
	gen := NewGenerator()
	// White knight on e4, white pawns on d6 and f6
	b, _ := board.FromFEN("8/8/3P1P2/8/4N3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should not include moves to d6 and f6 (blocked by own pawns)
	// Available moves: c5, g5, c3, g3, d2, f2 (6 moves)
	expectedMoves := 6
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d knight moves when blocked by own pieces, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Verify blocked moves don't exist
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && (move.To.String() == "d6" || move.To.String() == "f6") {
			t.Errorf("Knight should not be able to move to %s (blocked by own piece)", move.To.String())
		}
	}
}

func TestGenerateKnightMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// White knight on e4, black pawns on d6 and f6 for captures
	b, _ := board.FromFEN("8/8/3p1p2/8/4N3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have captures on d6 and f6, plus other normal moves
	foundCaptureD6 := false
	foundCaptureF6 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "d6" && move.IsCapture {
			foundCaptureD6 = true
		}
		if move.From.String() == "e4" && move.To.String() == "f6" && move.IsCapture {
			foundCaptureF6 = true
		}
	}
	
	if !foundCaptureD6 {
		t.Error("Expected capture move Ne4xd6 not found")
	}
	if !foundCaptureF6 {
		t.Error("Expected capture move Ne4xf6 not found")
	}
}

func TestGenerateKnightMoves_CanJumpOverPieces(t *testing.T) {
	gen := NewGenerator()
	// White knight on e4, surrounded by pieces but target squares empty
	b, _ := board.FromFEN("8/8/8/3ppp2/3pNp2/3ppp2/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Knight should still be able to reach all 8 target squares despite being surrounded
	// d6, f6, c5, g5, c3, g3, d2, f2
	expectedMoves := 8
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d knight moves (can jump over pieces), got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestGenerateKnightMoves_MultipleKnights(t *testing.T) {
	gen := NewGenerator()
	// Two white knights on b1 and g1 (like initial position)
	b, _ := board.FromFEN("8/8/8/8/8/8/8/1N4N1 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Each knight should have 3 moves
	// b1: a3, c3, d2 (3 moves)
	// g1: f3, h3, e2 (3 moves)
	// Total: 6 moves
	expectedMoves := 6
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d total moves for two knights, got %d", expectedMoves, whiteMoves.Count)
	}
}

func TestKnightMoves_AllEightDirections(t *testing.T) {
	gen := NewGenerator()
	// Test the L-shaped moves logic with a knight on d4
	b, _ := board.FromFEN("8/8/8/8/3N4/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKnightMoves(b, White)
	
	// Should have 8 moves: all L-shaped directions
	if whiteMoves.Count != 8 {
		t.Errorf("Expected 8 knight moves from d4, got %d", whiteMoves.Count)
	}
	
	// Verify each specific L-shaped move
	expectedMoves := []string{
		"d4e6", "d4c6", // Up 2, right/left 1
		"d4e2", "d4c2", // Down 2, right/left 1
		"d4f5", "d4b5", // Right/left 2, up 1
		"d4f3", "d4b3", // Right/left 2, down 1
	}
	
	foundMoves := make(map[string]bool)
	for _, move := range whiteMoves.Moves {
		moveStr := move.From.String() + move.To.String()
		foundMoves[moveStr] = true
	}
	
	for _, expectedMove := range expectedMoves {
		if !foundMoves[expectedMove] {
			t.Errorf("Expected knight move %s not found", expectedMove)
		}
	}
}

func TestGenerateKnightMoves_NearBoardEdges(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name          string
		fen           string
		expectedCount int
		position      string
	}{
		{"Knight on h1", "8/8/8/8/8/8/8/7N w - - 0 1", 2, "h1"},
		{"Knight on a8", "N7/8/8/8/8/8/8/8 w - - 0 1", 2, "a8"},
		{"Knight on h8", "7N/8/8/8/8/8/8/8 w - - 0 1", 2, "h8"},
		{"Knight on d1", "8/8/8/8/8/8/8/3N4 w - - 0 1", 4, "d1"},
		{"Knight on d8", "3N4/8/8/8/8/8/8/8 w - - 0 1", 4, "d8"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := board.FromFEN(tt.fen)
			whiteMoves := gen.GenerateKnightMoves(b, White)
			
			if whiteMoves.Count != tt.expectedCount {
				t.Errorf("Expected %d knight moves from %s, got %d", tt.expectedCount, tt.position, whiteMoves.Count)
			}
		})
	}
}