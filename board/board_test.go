package board

import (
	"testing"
)

func TestFromFEN_ValidCases(t *testing.T) {
	testCases := []struct {
		name string
		fen  string
	}{
		{"initial_position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"empty_board", "8/8/8/8/8/8/8/8 w - - 0 1"},
		{"single_piece", "8/8/8/8/8/8/8/4K3 w - - 0 1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			board, err := FromFEN(tc.fen)
			if err != nil {
				t.Errorf("Expected valid FEN %q to parse successfully, got error: %v", tc.fen, err)
			}
			if board == nil {
				t.Errorf("Expected board to be non-nil for valid FEN %q", tc.fen)
			}
		})
	}
}

func TestFromFEN_InvalidCases(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		expectedErr string
	}{
		{"empty_string", "", "invalid FEN: missing board position"},
		{"too_many_ranks", "8/8/8/8/8/8/8/8/8 w - - 0 1", "invalid FEN: must have exactly 8 ranks"},
		{"too_few_ranks", "8/8/8/8/8/8/8 w - - 0 1", "invalid FEN: must have exactly 8 ranks"},
		{"invalid_piece", "8/8/8/8/8/8/8/4X3 w - - 0 1", "invalid FEN: invalid piece character"},
		{"too_many_files", "9/8/8/8/8/8/8/8 w - - 0 1", "invalid FEN: invalid piece character"},
		{"insufficient_files", "7/8/8/8/8/8/8/8 w - - 0 1", "invalid FEN: incorrect number of files in rank"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			board, err := FromFEN(tc.fen)
			if err == nil {
				t.Errorf("Expected FEN %q to return error, but got valid board", tc.fen)
			}
			if board != nil {
				t.Errorf("Expected board to be nil for invalid FEN %q", tc.fen)
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("Expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestBoardGetSetPiece(t *testing.T) {
	board := NewBoard()
	
	// Test setting and getting a piece
	board.SetPiece(0, 0, WhiteKing)
	piece := board.GetPiece(0, 0)
	if piece != WhiteKing {
		t.Errorf("Expected %c, got %c", WhiteKing, piece)
	}
	
	// Test out of bounds
	piece = board.GetPiece(-1, 0)
	if piece != Empty {
		t.Errorf("Expected empty for out of bounds, got %c", piece)
	}
	
	piece = board.GetPiece(8, 0)
	if piece != Empty {
		t.Errorf("Expected empty for out of bounds, got %c", piece)
	}
}

func TestIsValidPiece(t *testing.T) {
	validPieces := []Piece{
		WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
	}
	
	for _, piece := range validPieces {
		if !isValidPiece(piece) {
			t.Errorf("Expected %c to be valid", piece)
		}
	}
	
	invalidPieces := []Piece{'x', 'Y', '1', '.', ' '}
	for _, piece := range invalidPieces {
		if isValidPiece(piece) {
			t.Errorf("Expected %c to be invalid", piece)
		}
	}
}