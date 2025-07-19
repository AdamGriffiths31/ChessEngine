package board

import (
	"testing"
)

func TestPieceLists(t *testing.T) {
	board := NewBoard()
	
	// Add a white pawn
	board.SetPiece(1, 4, WhitePawn) // e2
	
	pawns := board.GetPieceList(WhitePawn)
	if len(pawns) != 1 {
		t.Errorf("Expected 1 white pawn, got %d", len(pawns))
	}
	
	if pawns[0].File != 4 || pawns[0].Rank != 1 {
		t.Error("White pawn not at expected position")
	}
	
	// Check piece count
	if board.GetPieceCount(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn count, got %d", board.GetPieceCount(WhitePawn))
	}
	
	// Move the pawn
	board.SetPiece(1, 4, Empty)
	board.SetPiece(3, 4, WhitePawn) // e4
	
	pawns = board.GetPieceList(WhitePawn)
	if len(pawns) != 1 {
		t.Errorf("Expected 1 white pawn after move, got %d", len(pawns))
	}
	
	if pawns[0].File != 4 || pawns[0].Rank != 3 {
		t.Error("White pawn not at expected position after move")
	}
	
	// Check piece count unchanged
	if board.GetPieceCount(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn count after move, got %d", board.GetPieceCount(WhitePawn))
	}
}

func TestPieceListsFromFEN(t *testing.T) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}
	
	// Check piece counts
	testCases := []struct {
		piece    Piece
		expected int
	}{
		{WhitePawn, 8},
		{BlackPawn, 8},
		{WhiteRook, 2},
		{BlackRook, 2},
		{WhiteKnight, 2},
		{BlackKnight, 2},
		{WhiteBishop, 2},
		{BlackBishop, 2},
		{WhiteQueen, 1},
		{BlackQueen, 1},
		{WhiteKing, 1},
		{BlackKing, 1},
	}
	
	for _, tc := range testCases {
		count := board.GetPieceCount(tc.piece)
		if count != tc.expected {
			t.Errorf("Expected %d %c pieces, got %d", tc.expected, tc.piece, count)
		}
		
		// Verify the piece list matches the count
		list := board.GetPieceList(tc.piece)
		if len(list) != tc.expected {
			t.Errorf("Expected %d %c pieces in list, got %d", tc.expected, tc.piece, len(list))
		}
	}
}

func TestPieceListsCapture(t *testing.T) {
	board := NewBoard()
	
	// Set up a simple position with a white pawn and black piece to capture
	board.SetPiece(4, 4, WhitePawn) // e5
	board.SetPiece(5, 5, BlackKnight) // f6
	
	// Check initial counts
	if board.GetPieceCount(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn, got %d", board.GetPieceCount(WhitePawn))
	}
	if board.GetPieceCount(BlackKnight) != 1 {
		t.Errorf("Expected 1 black knight, got %d", board.GetPieceCount(BlackKnight))
	}
	
	// Simulate capture: white pawn takes black knight
	board.SetPiece(5, 5, WhitePawn) // Pawn captures knight
	board.SetPiece(4, 4, Empty)     // Remove pawn from original square
	
	// Check counts after capture
	if board.GetPieceCount(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn after capture, got %d", board.GetPieceCount(WhitePawn))
	}
	if board.GetPieceCount(BlackKnight) != 0 {
		t.Errorf("Expected 0 black knights after capture, got %d", board.GetPieceCount(BlackKnight))
	}
	
	// Verify piece lists
	pawns := board.GetPieceList(WhitePawn)
	if len(pawns) != 1 || pawns[0].File != 5 || pawns[0].Rank != 5 {
		t.Error("White pawn not at expected position after capture")
	}
	
	knights := board.GetPieceList(BlackKnight)
	if len(knights) != 0 {
		t.Error("Black knight list should be empty after capture")
	}
}

func TestPieceListsMultiplePieces(t *testing.T) {
	board := NewBoard()
	
	// Add multiple white pawns
	positions := []struct{ rank, file int }{
		{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7},
	}
	
	for _, pos := range positions {
		board.SetPiece(pos.rank, pos.file, WhitePawn)
	}
	
	// Check count
	if board.GetPieceCount(WhitePawn) != 8 {
		t.Errorf("Expected 8 white pawns, got %d", board.GetPieceCount(WhitePawn))
	}
	
	// Check list
	pawns := board.GetPieceList(WhitePawn)
	if len(pawns) != 8 {
		t.Errorf("Expected 8 pawns in list, got %d", len(pawns))
	}
	
	// Verify all positions are tracked
	for _, expectedPos := range positions {
		found := false
		for _, actualPos := range pawns {
			if actualPos.Rank == expectedPos.rank && actualPos.File == expectedPos.file {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Pawn at rank %d, file %d not found in piece list", expectedPos.rank, expectedPos.file)
		}
	}
}

func TestPieceListsRemoval(t *testing.T) {
	board := NewBoard()
	
	// Add multiple pieces
	board.SetPiece(0, 0, WhiteRook)
	board.SetPiece(0, 7, WhiteRook)
	board.SetPiece(7, 0, BlackRook)
	board.SetPiece(7, 7, BlackRook)
	
	// Check initial state
	if board.GetPieceCount(WhiteRook) != 2 {
		t.Errorf("Expected 2 white rooks, got %d", board.GetPieceCount(WhiteRook))
	}
	if board.GetPieceCount(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks, got %d", board.GetPieceCount(BlackRook))
	}
	
	// Remove one white rook
	board.SetPiece(0, 0, Empty)
	
	if board.GetPieceCount(WhiteRook) != 1 {
		t.Errorf("Expected 1 white rook after removal, got %d", board.GetPieceCount(WhiteRook))
	}
	
	// Verify remaining white rook is at correct position
	whiteRooks := board.GetPieceList(WhiteRook)
	if len(whiteRooks) != 1 || whiteRooks[0].File != 7 || whiteRooks[0].Rank != 0 {
		t.Error("Remaining white rook not at expected position")
	}
	
	// Black rooks should be unchanged
	if board.GetPieceCount(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks unchanged, got %d", board.GetPieceCount(BlackRook))
	}
}