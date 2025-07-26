package board

import (
	"testing"
)

func TestParseSquare(t *testing.T) {
	testCases := []struct {
		notation string
		expected Square
		hasError bool
	}{
		{"a1", Square{File: 0, Rank: 0}, false},
		{"e4", Square{File: 4, Rank: 3}, false},
		{"h8", Square{File: 7, Rank: 7}, false},
		{"a9", Square{}, true},  // out of bounds
		{"i1", Square{}, true},  // out of bounds
		{"e", Square{}, true},   // too short
		{"e44", Square{}, true}, // too long
	}

	for _, tc := range testCases {
		t.Run(tc.notation, func(t *testing.T) {
			result, err := ParseSquare(tc.notation)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for notation %q, but got none", tc.notation)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for notation %q, but got: %v", tc.notation, err)
				}
				if result != tc.expected {
					t.Errorf("Expected %+v, got %+v", tc.expected, result)
				}
			}
		})
	}
}

func TestSquareString(t *testing.T) {
	testCases := []struct {
		square   Square
		expected string
	}{
		{Square{File: 0, Rank: 0}, "a1"},
		{Square{File: 4, Rank: 3}, "e4"},
		{Square{File: 7, Rank: 7}, "h8"},
	}

	for _, tc := range testCases {
		result := tc.square.String()
		if result != tc.expected {
			t.Errorf("Square %+v: expected %q, got %q", tc.square, tc.expected, result)
		}
	}
}

func TestParseSimpleMove(t *testing.T) {
	testCases := []struct {
		notation string
		hasError bool
		expected Move
	}{
		{
			"e2e4",
			false,
			Move{
				From:      Square{File: 4, Rank: 1},
				To:        Square{File: 4, Rank: 3},
				Promotion: Empty,
			},
		},
		{
			"a7a8Q",
			false,
			Move{
				From:      Square{File: 0, Rank: 6},
				To:        Square{File: 0, Rank: 7},
				Promotion: WhiteQueen,
			},
		},
		{
			"O-O",
			false,
			Move{
				IsCastling: true,
				Promotion:  Empty,
			},
		},
		{
			"O-O-O",
			false,
			Move{
				IsCastling: true,
				Promotion:  Empty,
			},
		},
		{"e9e4", true, Move{}},   // invalid square
		{"e2", true, Move{}},     // too short
		{"e2e4e5", true, Move{}}, // too long
	}

	for _, tc := range testCases {
		t.Run(tc.notation, func(t *testing.T) {
			result, err := ParseSimpleMove(tc.notation)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for notation %q, but got none", tc.notation)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for notation %q, but got: %v", tc.notation, err)
				}
				if result.From != tc.expected.From {
					t.Errorf("From square: expected %+v, got %+v", tc.expected.From, result.From)
				}
				if result.To != tc.expected.To {
					t.Errorf("To square: expected %+v, got %+v", tc.expected.To, result.To)
				}
				if result.Promotion != tc.expected.Promotion {
					t.Errorf("Promotion: expected %c, got %c", tc.expected.Promotion, result.Promotion)
				}
				if result.IsCastling != tc.expected.IsCastling {
					t.Errorf("IsCastling: expected %t, got %t", tc.expected.IsCastling, result.IsCastling)
				}
			}
		})
	}
}

func TestMakeMove(t *testing.T) {
	// Test basic pawn move
	board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	move := Move{
		From:      Square{File: 4, Rank: 1}, // e2
		To:        Square{File: 4, Rank: 3}, // e4
		Piece:     Empty,                    // Tell MakeMove to get piece from board
		Promotion: Empty,
	}

	err := board.MakeMove(move)
	if err != nil {
		t.Errorf("Expected no error making move, got: %v", err)
	}

	// Check that the piece moved correctly
	if board.GetPiece(1, 4) != Empty {
		t.Errorf("Expected e2 to be empty after move, got: %c", board.GetPiece(1, 4))
	}
	if board.GetPiece(3, 4) != WhitePawn {
		t.Errorf("Expected e4 to have white pawn after move, got: %c", board.GetPiece(3, 4))
	}
}

func TestBoardToFEN(t *testing.T) {
	// Test initial position
	board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	fen := board.ToFEN()
	expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	if fen != expected {
		t.Errorf("Expected FEN %q, got %q", expected, fen)
	}
}
