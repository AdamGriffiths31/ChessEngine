package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestGenerateKingMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white king moves from initial position - should have no moves (blocked by pawns)
	whiteMoves := gen.GenerateKingMoves(b, White)
	if whiteMoves.Count != 0 {
		t.Errorf("Expected 0 white king moves from initial position, got %d", whiteMoves.Count)
	}
	
	// Test black king moves from initial position - should have no moves (blocked by pawns)
	blackMoves := gen.GenerateKingMoves(b, Black)
	if blackMoves.Count != 0 {
		t.Errorf("Expected 0 black king moves from initial position, got %d", blackMoves.Count)
	}
}

func TestGenerateKingMoves_CenterBoard(t *testing.T) {
	gen := NewGenerator()
	// White king on e4, empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/4K3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have 8 moves from center: all adjacent squares
	// d5, e5, f5, d4, f4, d3, e3, f3
	expectedMoves := 8
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d king moves from center, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"e4d5": false, "e4e5": false, "e4f5": false, // Up row
		"e4d4": false, "e4f4": false,                // Same row (left/right)
		"e4d3": false, "e4e3": false, "e4f3": false, // Down row
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

func TestGenerateKingMoves_Corner(t *testing.T) {
	gen := NewGenerator()
	// White king on a1 (corner), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/8/8/8/K7 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have 3 moves from corner: a2, b1, b2
	expectedMoves := 3
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d king moves from corner, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a1a2": false, "a1b1": false, "a1b2": false,
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

func TestGenerateKingMoves_Edge(t *testing.T) {
	gen := NewGenerator()
	// White king on a4 (edge), empty board otherwise
	b, _ := board.FromFEN("8/8/8/8/K7/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have 5 moves from edge: a5, b5, b4, b3, a3
	expectedMoves := 5
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d king moves from edge, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a4a5": false, "a4b5": false, "a4b4": false,
		"a4b3": false, "a4a3": false,
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

func TestGenerateKingMoves_BlockedByOwnPieces(t *testing.T) {
	gen := NewGenerator()
	// White king on e4, white pawns on d5, e5, f5
	b, _ := board.FromFEN("8/8/8/3PPP2/4K3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should not include moves to d5, e5, f5 (blocked by own pawns)
	// Available moves: d4, f4, d3, e3, f3 (5 moves)
	expectedMoves := 5
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d king moves when blocked by own pieces, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Verify blocked moves don't exist
	blockedSquares := []string{"d5", "e5", "f5"}
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" {
			for _, blocked := range blockedSquares {
				if move.To.String() == blocked {
					t.Errorf("King should not be able to move to %s (blocked by own piece)", blocked)
				}
			}
		}
	}
}

func TestGenerateKingMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// White king on e4, black pawns on d5, e5, f5 for captures
	b, _ := board.FromFEN("8/8/8/3ppp2/4K3/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have captures on d5, e5, f5, plus other normal moves
	foundCaptureD5 := false
	foundCaptureE5 := false
	foundCaptureF5 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.IsCapture {
			switch move.To.String() {
			case "d5":
				foundCaptureD5 = true
			case "e5":
				foundCaptureE5 = true
			case "f5":
				foundCaptureF5 = true
			}
		}
	}
	
	if !foundCaptureD5 {
		t.Error("Expected capture move Ke4xd5 not found")
	}
	if !foundCaptureE5 {
		t.Error("Expected capture move Ke4xe5 not found")
	}
	if !foundCaptureF5 {
		t.Error("Expected capture move Ke4xf5 not found")
	}
}

func TestGenerateKingMoves_CastlingFromInitialPosition(t *testing.T) {
	gen := NewGenerator()
	// Initial position but with pieces moved to allow castling
	b, _ := board.FromFEN("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have regular king moves plus castling moves
	foundKingsideCastling := false
	foundQueensideCastling := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e1" && move.IsCastling {
			if move.To.String() == "g1" {
				foundKingsideCastling = true
			}
			if move.To.String() == "c1" {
				foundQueensideCastling = true
			}
		}
	}
	
	if !foundKingsideCastling {
		t.Error("Expected kingside castling move O-O not found")
	}
	if !foundQueensideCastling {
		t.Error("Expected queenside castling move O-O-O not found")
	}
}

func TestGenerateKingMoves_CastlingBlocked(t *testing.T) {
	gen := NewGenerator()
	// King and rooks in position but pieces between them
	b, _ := board.FromFEN("r1b1k1nr/pppppppp/8/8/8/8/PPPPPPPP/R1B1K1NR w KQkq - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should not have castling moves (pieces in the way)
	foundCastling := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e1" && move.IsCastling {
			foundCastling = true
		}
	}
	
	if foundCastling {
		t.Error("Found castling move when pieces are blocking the path")
	}
}

func TestGenerateKingMoves_CastlingKingMoved(t *testing.T) {
	gen := NewGenerator()
	// King not on starting square
	b, _ := board.FromFEN("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R4RK1 w kq - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should not have castling moves (king not on starting square)
	foundCastling := false
	
	for _, move := range whiteMoves.Moves {
		if move.IsCastling {
			foundCastling = true
		}
	}
	
	if foundCastling {
		t.Error("Found castling move when king is not on starting square")
	}
}

func TestGenerateKingMoves_CastlingNoRook(t *testing.T) {
	gen := NewGenerator()
	// King in position but no rooks
	b, _ := board.FromFEN("4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should not have castling moves (no rooks)
	foundCastling := false
	
	for _, move := range whiteMoves.Moves {
		if move.IsCastling {
			foundCastling = true
		}
	}
	
	if foundCastling {
		t.Error("Found castling move when rooks are missing")
	}
}

func TestGenerateKingMoves_AllDirections(t *testing.T) {
	gen := NewGenerator()
	// Test the king moves logic with a king on d4
	b, _ := board.FromFEN("8/8/8/8/3K4/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GenerateKingMoves(b, White)
	
	// Should have exactly 8 moves from d4
	if whiteMoves.Count != 8 {
		t.Errorf("Expected 8 king moves from d4, got %d", whiteMoves.Count)
	}
	
	// Verify each specific adjacent square
	expectedMoves := []string{
		"d4c5", "d4d5", "d4e5", // Up row
		"d4c4", "d4e4",         // Same row
		"d4c3", "d4d3", "d4e3", // Down row
	}
	
	foundMoves := make(map[string]bool)
	for _, move := range whiteMoves.Moves {
		moveStr := move.From.String() + move.To.String()
		foundMoves[moveStr] = true
	}
	
	for _, expectedMove := range expectedMoves {
		if !foundMoves[expectedMove] {
			t.Errorf("Expected king move %s not found", expectedMove)
		}
	}
}

func TestGenerateKingMoves_BlackKing(t *testing.T) {
	gen := NewGenerator()
	// Black king on e8, empty board
	b, _ := board.FromFEN("4k3/8/8/8/8/8/8/8 b - - 0 1")
	
	blackMoves := gen.GenerateKingMoves(b, Black)
	
	// Should have 5 moves from e8 (edge position)
	expectedMoves := 5
	if blackMoves.Count != expectedMoves {
		t.Errorf("Expected %d black king moves from edge, got %d", expectedMoves, blackMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"e8d8": false, "e8f8": false, // Same rank
		"e8d7": false, "e8e7": false, "e8f7": false, // Down rank
	}
	
	for _, move := range blackMoves.Moves {
		moveStr := move.From.String() + move.To.String()
		if _, exists := expectedMovesMap[moveStr]; exists {
			expectedMovesMap[moveStr] = true
		}
	}
	
	for moveStr, found := range expectedMovesMap {
		if !found {
			t.Errorf("Expected black king move %s not found", moveStr)
		}
	}
}

func TestGenerateKingMoves_BlackCastling(t *testing.T) {
	gen := NewGenerator()
	// Black king and rooks in castling position
	b, _ := board.FromFEN("r3k2r/8/8/8/8/8/8/R3K2R b kq - 0 1")
	
	blackMoves := gen.GenerateKingMoves(b, Black)
	
	// Should have regular moves plus both castling moves
	foundKingsideCastling := false
	foundQueensideCastling := false
	
	for _, move := range blackMoves.Moves {
		if move.From.String() == "e8" && move.IsCastling {
			if move.To.String() == "g8" {
				foundKingsideCastling = true
			}
			if move.To.String() == "c8" {
				foundQueensideCastling = true
			}
		}
	}
	
	if !foundKingsideCastling {
		t.Error("Expected black kingside castling move not found")
	}
	if !foundQueensideCastling {
		t.Error("Expected black queenside castling move not found")
	}
}