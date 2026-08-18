package board

import (
	"testing"
)

func TestBitboardPieces(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(1, 4, WhitePawn) // e2

	count := board.getPieceCountFromBitboard(WhitePawn)
	if count != 1 {
		t.Errorf("Expected 1 white pawn, got %d", count)
	}

	square := FileRankToSquare(4, 1)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(square) {
		t.Error("White pawn not at expected position")
	}

	board.SetPiece(1, 4, Empty)
	board.SetPiece(3, 4, WhitePawn) // e4

	count = board.getPieceCountFromBitboard(WhitePawn)
	if count != 1 {
		t.Errorf("Expected 1 white pawn after move, got %d", count)
	}

	newSquare := FileRankToSquare(4, 3)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(newSquare) {
		t.Error("White pawn not at expected position after move")
	}

	if board.PieceBitboards[WhitePawnIndex].HasBit(square) {
		t.Error("White pawn still at old position after move")
	}
}

func TestBitboardFromFEN(t *testing.T) {
	t.Parallel()
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

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
	t.Parallel()
	board := NewBoard()

	board.SetPiece(4, 4, WhitePawn)   // e5
	board.SetPiece(5, 5, BlackKnight) // f6

	if board.getPieceCountFromBitboard(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}
	if board.getPieceCountFromBitboard(BlackKnight) != 1 {
		t.Errorf("Expected 1 black knight, got %d", board.getPieceCountFromBitboard(BlackKnight))
	}

	// Simulate capture: white pawn takes black knight
	board.SetPiece(5, 5, WhitePawn)
	board.SetPiece(4, 4, Empty) // Remove pawn from original square

	if board.getPieceCountFromBitboard(WhitePawn) != 1 {
		t.Errorf("Expected 1 white pawn after capture, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}
	if board.getPieceCountFromBitboard(BlackKnight) != 0 {
		t.Errorf("Expected 0 black knights after capture, got %d", board.getPieceCountFromBitboard(BlackKnight))
	}

	captureSquare := FileRankToSquare(5, 5)
	if !board.PieceBitboards[WhitePawnIndex].HasBit(captureSquare) {
		t.Error("White pawn not at expected position after capture")
	}

	if board.PieceBitboards[BlackKnightIndex].PopCount() != 0 {
		t.Error("Black knight bitboard should be empty after capture")
	}
}

func TestBitboardMultiplePieces(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	positions := []struct{ rank, file int }{
		{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7},
	}

	for _, pos := range positions {
		board.SetPiece(pos.rank, pos.file, WhitePawn)
	}

	if board.getPieceCountFromBitboard(WhitePawn) != 8 {
		t.Errorf("Expected 8 white pawns, got %d", board.getPieceCountFromBitboard(WhitePawn))
	}

	for _, expectedPos := range positions {
		square := FileRankToSquare(expectedPos.file, expectedPos.rank)
		if !board.PieceBitboards[WhitePawnIndex].HasBit(square) {
			t.Errorf("Pawn at rank %d, file %d not found in bitboard", expectedPos.rank, expectedPos.file)
		}
	}
}

func TestBitboardRemoval(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(0, 0, WhiteRook)
	board.SetPiece(0, 7, WhiteRook)
	board.SetPiece(7, 0, BlackRook)
	board.SetPiece(7, 7, BlackRook)

	if board.getPieceCountFromBitboard(WhiteRook) != 2 {
		t.Errorf("Expected 2 white rooks, got %d", board.getPieceCountFromBitboard(WhiteRook))
	}
	if board.getPieceCountFromBitboard(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks, got %d", board.getPieceCountFromBitboard(BlackRook))
	}

	board.SetPiece(0, 0, Empty)

	if board.getPieceCountFromBitboard(WhiteRook) != 1 {
		t.Errorf("Expected 1 white rook after removal, got %d", board.getPieceCountFromBitboard(WhiteRook))
	}

	remainingSquare := FileRankToSquare(7, 0)
	if !board.PieceBitboards[WhiteRookIndex].HasBit(remainingSquare) {
		t.Error("Remaining white rook not at expected position")
	}

	removedSquare := FileRankToSquare(0, 0)
	if board.PieceBitboards[WhiteRookIndex].HasBit(removedSquare) {
		t.Error("Removed white rook still in bitboard")
	}

	if board.getPieceCountFromBitboard(BlackRook) != 2 {
		t.Errorf("Expected 2 black rooks unchanged, got %d", board.getPieceCountFromBitboard(BlackRook))
	}
}
