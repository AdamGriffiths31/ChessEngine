package board

import (
	"testing"
)

func TestBitboardPieces(t *testing.T) {
	board := NewBoard()
	
	// Add a white pawn
	board.SetPiece(1, 4, WhitePawn) // e2
	
	// Check piece count using bitboard
	count := board.getPieceCountFromBitboard(WhitePawn)
	if count != 1 {
		t.Errorf("Expected 1 white pawn, got %d", count)
	}
	
	// Check piece position using bitboard
	square := FileRankToSquare(4, 1)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(square) {
		t.Error("White pawn not at expected position")
	}
	
	// Move the pawn
	board.SetPiece(1, 4, Empty)
	board.SetPiece(3, 4, WhitePawn) // e4
	
	// Check piece count unchanged
	count = board.getPieceCountFromBitboard(WhitePawn)
	if count != 1 {
		t.Errorf("Expected 1 white pawn after move, got %d", count)
	}
	
	// Check piece at new position
	newSquare := FileRankToSquare(4, 3)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(newSquare) {
		t.Error("White pawn not at expected position after move")
	}
	
	// Check piece not at old position
	if board.PieceBitboards[WhitePawnIndex].HasBit(square) {
		t.Error("White pawn still at old position after move")
	}
}

func TestBitboardFromFEN(t *testing.T) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}
	
	// Check piece counts using bitboards
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
		count := board.getPieceCountFromBitboard(tc.piece)
		if count != tc.expected {
			t.Errorf("Expected %d %c pieces, got %d", tc.expected, tc.piece, count)
		}
	}
}

func TestBitboardCapture(t *testing.T) {
	board := NewBoard()
	
	// Set up a simple position with a white pawn and black piece to capture
	board.SetPiece(4, 4, WhitePawn) // e5
	board.SetPiece(5, 5, BlackKnight) // f6
	
	// Check initial counts using bitboards
	if board.getPieceCountFromBitboard(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}
	if board.getPieceCountFromBitboard(BlackKnight) != 1 {
		t.Errorf("Expected 1 black knight, got %d", board.getPieceCountFromBitboard(BlackKnight))
	}
	
	// Simulate capture: white pawn takes black knight
	board.SetPiece(5, 5, WhitePawn) // Pawn captures knight
	board.SetPiece(4, 4, Empty)     // Remove pawn from original square
	
	// Check counts after capture
	if board.getPieceCountFromBitboard(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn after capture, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}
	if board.getPieceCountFromBitboard(BlackKnight) != 0 {
		t.Errorf("Expected 0 black knights after capture, got %d", board.getPieceCountFromBitboard(BlackKnight))
	}
	
	// Verify piece positions using bitboards
	captureSquare := FileRankToSquare(5, 5)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(captureSquare) {
		t.Error("White pawn not at expected position after capture")
	}
	
	if board.PieceBitboards[BlackKnightIndex].PopCount() != 0 {
		t.Error("Black knight bitboard should be empty after capture")
	}
}

func TestBitboardMultiplePieces(t *testing.T) {
	board := NewBoard()
	
	// Add multiple white pawns
	positions := []struct{ rank, file int }{
		{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7},
	}
	
	for _, pos := range positions {
		board.SetPiece(pos.rank, pos.file, WhitePawn)
	}
	
	// Check count using bitboard
	if board.getPieceCountFromBitboard(WhitePawn) != 8 {
		t.Errorf("Expected 8 white pawns, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}
	
	// Verify all positions are tracked in bitboard
	for _, expectedPos := range positions {
		square := FileRankToSquare(expectedPos.file, expectedPos.rank)
		if !board.PieceBitboards[WhitePawnIndex].HasBit(square) {
			t.Errorf("Pawn at rank %d, file %d not found in bitboard", expectedPos.rank, expectedPos.file)
		}
	}
}

func TestBitboardRemoval(t *testing.T) {
	board := NewBoard()
	
	// Add multiple pieces
	board.SetPiece(0, 0, WhiteRook)
	board.SetPiece(0, 7, WhiteRook)
	board.SetPiece(7, 0, BlackRook)
	board.SetPiece(7, 7, BlackRook)
	
	// Check initial state using bitboards
	if board.getPieceCountFromBitboard(WhiteRook) != 2 {
		t.Errorf("Expected 2 white rooks, got %d", board.getPieceCountFromBitboard(WhiteRook))
	}
	if board.getPieceCountFromBitboard(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks, got %d", board.getPieceCountFromBitboard(BlackRook))
	}
	
	// Remove one white rook
	board.SetPiece(0, 0, Empty)
	
	if board.getPieceCountFromBitboard(WhiteRook) != 1 {
		t.Errorf("Expected 1 white rook after removal, got %d", board.getPieceCountFromBitboard(WhiteRook))
	}
	
	// Verify remaining white rook is at correct position
	remainingSquare := FileRankToSquare(7, 0)
	if !board.PieceBitboards[WhiteRookIndex].HasBit(remainingSquare) {
		t.Error("Remaining white rook not at expected position")
	}
	
	// Verify removed rook is not in bitboard
	removedSquare := FileRankToSquare(0, 0)
	if board.PieceBitboards[WhiteRookIndex].HasBit(removedSquare) {
		t.Error("Removed white rook still in bitboard")
	}
	
	// Black rooks should be unchanged
	if board.getPieceCountFromBitboard(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks unchanged, got %d", board.getPieceCountFromBitboard(BlackRook))
	}
}