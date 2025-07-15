package game

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestMoveParserParseMove(t *testing.T) {
	parser := NewMoveParser(White)
	gameBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	testCases := []struct {
		notation     string
		expectedErr  string
		hasError     bool
		expectMove   bool
	}{
		{"e2e4", "", false, true},
		{"quit", "QUIT", true, false},
		{"reset", "RESET", true, false},
		{"fen", "FEN", true, false},
		{"o-o", "", false, true},
		{"o-o-o", "", false, true},
		{"e7e8q", "", false, true},
		{"invalid", "algebraic notation not fully implemented - use coordinate notation (e.g., e2e4)", true, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.notation, func(t *testing.T) {
			move, err := parser.ParseMove(tc.notation, gameBoard)
			
			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for notation %q, but got none", tc.notation)
				} else if err.Error() != tc.expectedErr {
					t.Errorf("Expected error %q, got %q", tc.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for notation %q, but got: %v", tc.notation, err)
				}
				if tc.expectMove && move.Promotion != board.Empty {
					// For non-promotion moves, promotion should be Empty
					if tc.notation != "e7e8q" && move.Promotion != board.Empty {
						t.Errorf("Expected promotion to be Empty for notation %q, got %c", tc.notation, move.Promotion)
					}
				}
			}
		})
	}
}

func TestMoveParserSetCurrentPlayer(t *testing.T) {
	parser := NewMoveParser(White)
	
	if parser.currentPlayer != White {
		t.Errorf("Expected initial player to be White, got %v", parser.currentPlayer)
	}
	
	parser.SetCurrentPlayer(Black)
	if parser.currentPlayer != Black {
		t.Errorf("Expected player to be Black after setting, got %v", parser.currentPlayer)
	}
}

func TestMoveParserParseCastling(t *testing.T) {
	parser := NewMoveParser(White)
	gameBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white kingside castling
	move, err := parser.ParseMove("o-o", gameBoard)
	if err != nil {
		t.Errorf("Expected no error for white kingside castling, got: %v", err)
	}
	if !move.IsCastling {
		t.Errorf("Expected move to be castling")
	}
	
	// Test black castling
	parser.SetCurrentPlayer(Black)
	move, err = parser.ParseMove("o-o-o", gameBoard)
	if err != nil {
		t.Errorf("Expected no error for black queenside castling, got: %v", err)
	}
	if !move.IsCastling {
		t.Errorf("Expected move to be castling")
	}
}

func TestMoveParserCharToPiece(t *testing.T) {
	testCases := []struct {
		player   Player
		char     byte
		expected board.Piece
		hasError bool
	}{
		{White, 'q', board.WhiteQueen, false},
		{White, 'r', board.WhiteRook, false},
		{White, 'b', board.WhiteBishop, false},
		{White, 'n', board.WhiteKnight, false},
		{Black, 'q', board.BlackQueen, false},
		{Black, 'r', board.BlackRook, false},
		{Black, 'b', board.BlackBishop, false},
		{Black, 'n', board.BlackKnight, false},
		{White, 'x', board.Empty, true},
	}
	
	for _, tc := range testCases {
		parser := NewMoveParser(tc.player)
		result, err := parser.charToPiece(tc.char)
		
		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for char %c with player %v, but got none", tc.char, tc.player)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for char %c with player %v, but got: %v", tc.char, tc.player, err)
			}
			if result != tc.expected {
				t.Errorf("Expected piece %c, got %c", tc.expected, result)
			}
		}
	}
}