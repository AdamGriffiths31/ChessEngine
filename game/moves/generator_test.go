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

func TestFindKing(t *testing.T) {
	gen := NewGenerator()

	// Load a standard starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// Test finding kings using bitboard lookup
	whiteKingPos := gen.findKing(b, White)
	if whiteKingPos == nil {
		t.Error("Expected to find white king")
	}

	blackKingPos := gen.findKing(b, Black)
	if blackKingPos == nil {
		t.Error("Expected to find black king")
	}

	// Verify positions are correct
	expectedWhiteKing := board.Square{File: 4, Rank: 0} // e1
	expectedBlackKing := board.Square{File: 4, Rank: 7} // e8

	if *whiteKingPos != expectedWhiteKing {
		t.Errorf("Expected white king at %v, got %v", expectedWhiteKing, *whiteKingPos)
	}
	if *blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king at %v, got %v", expectedBlackKing, *blackKingPos)
	}

	// Test with empty board - should return nil
	emptyBoard, _ := board.FromFEN("8/8/8/8/8/8/8/8 w - - 0 1")
	whiteKingPosEmpty := gen.findKing(emptyBoard, White)
	blackKingPosEmpty := gen.findKing(emptyBoard, Black)

	if whiteKingPosEmpty != nil {
		t.Error("Expected nil for white king on empty board")
	}
	if blackKingPosEmpty != nil {
		t.Error("Expected nil for black king on empty board")
	}

	// Test with custom position
	customBoard, _ := board.FromFEN("8/8/8/3k4/3K4/8/8/8 w - - 0 1")
	whiteKingCustom := gen.findKing(customBoard, White)
	blackKingCustom := gen.findKing(customBoard, Black)

	expectedWhiteCustom := board.Square{File: 3, Rank: 3} // d4
	expectedBlackCustom := board.Square{File: 3, Rank: 4} // d5

	if whiteKingCustom == nil || *whiteKingCustom != expectedWhiteCustom {
		t.Errorf("Expected white king at %v, got %v", expectedWhiteCustom, whiteKingCustom)
	}
	if blackKingCustom == nil || *blackKingCustom != expectedBlackCustom {
		t.Errorf("Expected black king at %v, got %v", expectedBlackCustom, blackKingCustom)
	}

	// Test move generation still works
	moves := gen.GenerateAllMoves(b, White)
	if moves.Count == 0 {
		t.Error("Expected to generate some moves")
	}
	ReleaseMoveList(moves)
}