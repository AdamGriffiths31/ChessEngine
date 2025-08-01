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

func TestKingCache(t *testing.T) {
	gen := NewGenerator()

	// Test initial state - cache should be invalid
	if gen.kingCacheValid {
		t.Error("Expected king cache to be invalid initially")
	}
	if gen.whiteKingPos != nil {
		t.Error("Expected white king position to be nil initially")
	}
	if gen.blackKingPos != nil {
		t.Error("Expected black king position to be nil initially")
	}

	// Load a standard starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// Test cache initialization through findKing
	whiteKingPos := gen.findKing(b, White)
	if whiteKingPos == nil {
		t.Error("Expected to find white king")
	}

	// Cache should now be valid
	if !gen.kingCacheValid {
		t.Error("Expected king cache to be valid after findKing call")
	}

	// Verify cached positions
	expectedWhiteKing := board.Square{File: 4, Rank: 0} // e1
	expectedBlackKing := board.Square{File: 4, Rank: 7} // e8

	if gen.whiteKingPos == nil || *gen.whiteKingPos != expectedWhiteKing {
		t.Errorf("Expected white king at %v, got %v", expectedWhiteKing, gen.whiteKingPos)
	}
	if gen.blackKingPos == nil || *gen.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king at %v, got %v", expectedBlackKing, gen.blackKingPos)
	}

	// Test that subsequent findKing calls use cache (should return same pointer)
	whiteKingPos2 := gen.findKing(b, White)
	if whiteKingPos != whiteKingPos2 {
		t.Error("Expected findKing to return cached position (same pointer)")
	}

	// Test cache update when king moves
	move := board.Move{
		From:      board.Square{File: 4, Rank: 0}, // e1
		To:        board.Square{File: 5, Rank: 0}, // f1
		Piece:     board.WhiteKing,
		IsCapture: false,
		Captured:  board.Empty,
		Promotion: board.Empty,
	}

	// Update the cache (simulating what happens during move execution)
	gen.updateKingCache(move)

	// Verify white king position was updated
	expectedNewWhiteKing := board.Square{File: 5, Rank: 0} // f1
	if gen.whiteKingPos == nil || *gen.whiteKingPos != expectedNewWhiteKing {
		t.Errorf("Expected white king at %v after move, got %v", expectedNewWhiteKing, gen.whiteKingPos)
	}

	// Black king position should remain unchanged
	if gen.blackKingPos == nil || *gen.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king to remain at %v, got %v", expectedBlackKing, gen.blackKingPos)
	}

	// Test cache initialization through findKing in GenerateAllMoves context
	gen2 := NewGenerator()
	moves := gen2.GenerateAllMoves(b, White)

	// Trigger cache initialization by calling findKing
	whiteKingPos3 := gen2.findKing(b, White)
	blackKingPos3 := gen2.findKing(b, Black)

	// Cache should be initialized after findKing calls
	if !gen2.kingCacheValid {
		t.Error("Expected king cache to be valid after findKing calls")
	}
	if gen2.whiteKingPos == nil || *gen2.whiteKingPos != expectedWhiteKing {
		t.Errorf("Expected white king cached at %v, got %v", expectedWhiteKing, gen2.whiteKingPos)
	}
	if gen2.blackKingPos == nil || *gen2.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king cached at %v, got %v", expectedBlackKing, gen2.blackKingPos)
	}

	// Verify the positions returned are correct
	if whiteKingPos3 == nil || *whiteKingPos3 != expectedWhiteKing {
		t.Errorf("Expected findKing to return white king at %v, got %v", expectedWhiteKing, whiteKingPos3)
	}
	if blackKingPos3 == nil || *blackKingPos3 != expectedBlackKing {
		t.Errorf("Expected findKing to return black king at %v, got %v", expectedBlackKing, blackKingPos3)
	}

	// Verify we got some moves
	if moves.Count == 0 {
		t.Error("Expected to generate some moves")
	}
}
