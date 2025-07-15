package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("Expected generator to be non-nil")
	}
}

func TestGenerateAllMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white moves - should include pawns (16) + knights (4) + king (0) = 20 moves
	// Rooks, bishops, queens are blocked by pawns in initial position
	// King has no legal moves (blocked by pawns)
	whiteMoves := gen.GenerateAllMoves(b, White)
	if whiteMoves.Count != 20 {
		t.Errorf("Expected 20 white moves from initial position (16 pawn + 4 knight + 0 king), got %d", whiteMoves.Count)
	}
	
	// Test black moves - should include pawns (16) + knights (4) + king (0) = 20 moves
	blackMoves := gen.GenerateAllMoves(b, Black)
	if blackMoves.Count != 20 {
		t.Errorf("Expected 20 black moves from initial position (16 pawn + 4 knight + 0 king), got %d", blackMoves.Count)
	}
}

func TestGeneratePawnMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	whiteMoves := gen.GeneratePawnMoves(b, White)
	
	// Should have 16 moves: 8 pawns × 2 moves each (1 square and 2 squares forward)
	expectedMoves := 16
	if whiteMoves.Count != expectedMoves {
		t.Errorf("Expected %d white pawn moves, got %d", expectedMoves, whiteMoves.Count)
	}
	
	// Check specific moves exist
	expectedMovesMap := map[string]bool{
		"a2a3": false, "a2a4": false,
		"b2b3": false, "b2b4": false,
		"c2c3": false, "c2c4": false,
		"d2d3": false, "d2d4": false,
		"e2e3": false, "e2e4": false,
		"f2f3": false, "f2f4": false,
		"g2g3": false, "g2g4": false,
		"h2h3": false, "h2h4": false,
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

func TestGeneratePawnMoves_BlockedPawn(t *testing.T) {
	gen := NewGenerator()
	// Position with white pawn on e4, black pawn on e5 (blocking)
	b, _ := board.FromFEN("rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2")
	
	whiteMoves := gen.GeneratePawnMoves(b, White)
	
	// e4 pawn should have no forward moves (blocked by e5)
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" {
			t.Errorf("e4 pawn should be blocked, but found move: %s", move.To.String())
		}
	}
}

func TestGeneratePawnMoves_Captures(t *testing.T) {
	gen := NewGenerator()
	// Position with white pawn on e4, black pawns on d5 and f5 for captures
	b, _ := board.FromFEN("rnbqkbnr/ppp2ppp/8/3ppp2/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 3")
	
	whiteMoves := gen.GeneratePawnMoves(b, White)
	
	// Should have captures exd5 and exf5
	foundExd5 := false
	foundExf5 := false
	
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e4" && move.To.String() == "d5" && move.IsCapture {
			foundExd5 = true
		}
		if move.From.String() == "e4" && move.To.String() == "f5" && move.IsCapture {
			foundExf5 = true
		}
	}
	
	if !foundExd5 {
		t.Error("Expected capture move exd5 not found")
	}
	if !foundExf5 {
		t.Error("Expected capture move exf5 not found")
	}
}

func TestGeneratePawnMoves_Promotion(t *testing.T) {
	gen := NewGenerator()
	// White pawn on 7th rank ready to promote
	b, _ := board.FromFEN("8/4P3/8/8/8/8/8/8 w - - 0 1")
	
	whiteMoves := gen.GeneratePawnMoves(b, White)
	
	// Should have 4 promotion moves (Q, R, B, N)
	expectedCount := 4
	if whiteMoves.Count != expectedCount {
		t.Errorf("Expected %d promotion moves, got %d", expectedCount, whiteMoves.Count)
	}
	
	// Check that all promotion pieces are present
	promotionPieces := make(map[board.Piece]bool)
	for _, move := range whiteMoves.Moves {
		if move.From.String() == "e7" && move.To.String() == "e8" {
			promotionPieces[move.Promotion] = true
		}
	}
	
	expectedPieces := []board.Piece{
		board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight,
	}
	
	for _, piece := range expectedPieces {
		if !promotionPieces[piece] {
			t.Errorf("Expected promotion to %c not found", piece)
		}
	}
}

func TestMoveList_AddMove(t *testing.T) {
	ml := NewMoveList()
	
	if ml.Count != 0 {
		t.Errorf("Expected empty move list to have count 0, got %d", ml.Count)
	}
	
	move := board.Move{
		From: board.Square{File: 4, Rank: 1},
		To:   board.Square{File: 4, Rank: 3},
	}
	
	ml.AddMove(move)
	
	if ml.Count != 1 {
		t.Errorf("Expected move list count to be 1 after adding move, got %d", ml.Count)
	}
	
	if len(ml.Moves) != 1 {
		t.Errorf("Expected moves slice length to be 1, got %d", len(ml.Moves))
	}
}

func TestMoveList_Contains(t *testing.T) {
	ml := NewMoveList()
	
	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	ml.AddMove(move)
	
	if !ml.Contains(move) {
		t.Error("Expected move list to contain the added move")
	}
	
	differentMove := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 2},
		Promotion: board.Empty,
	}
	
	if ml.Contains(differentMove) {
		t.Error("Expected move list to not contain different move")
	}
}

func TestMovesEqual(t *testing.T) {
	move1 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	move2 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	if !MovesEqual(move1, move2) {
		t.Error("Expected identical moves to be equal")
	}
	
	move3 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 2},
		Promotion: board.Empty,
	}
	
	if MovesEqual(move1, move3) {
		t.Error("Expected different moves to not be equal")
	}
}