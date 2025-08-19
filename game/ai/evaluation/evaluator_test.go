package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	if evaluator == nil {
		t.Fatal("NewEvaluator should not return nil")
	}

	if evaluator.GetName() != "Evaluator" {
		t.Errorf("Expected name 'Evaluator', got '%s'", evaluator.GetName())
	}
}

func TestEvaluateEmptyBoard(t *testing.T) {
	evaluator := NewEvaluator()
	b := board.NewBoard()

	// Empty board should have score 0 (always from White's perspective)
	score := evaluator.Evaluate(b)

	if score != 0 {
		t.Errorf("Expected score 0 for empty board, got %d", score)
	}
}

func TestEvaluateMaterialAndPST(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		expected    int
		description string
	}{
		{
			name:        "balanced fianchetto position",
			fen:         "rn1qkbnr/pbpppppp/1p6/8/8/6P1/PPPPPPBP/RNBQK1NR w KQkq - 0 1",
			expected:    0,
			description: "Equal material, symmetric development (g3+Bg2 vs b6+Bb7)",
		},
		{
			name:        "symmetric pawn promotion race",
			fen:         "8/PPPPPPPP/8/8/8/8/pppppppp/8 w - - 0 1",
			expected:    0,
			description: "White pawns on 7th rank vs black pawns on 2nd rank (symmetric near-promotion)",
		},
		{
			name:        "starting position",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expected:    0,
			description: "Standard chess starting position (perfectly symmetric)",
		},
		{
			name:        "endgame king and pawns",
			fen:         "7k/5ppp/8/8/8/8/5PPP/7K w - - 0 1",
			expected:    0,
			description: "Symmetric endgame with kings on h-files and 3 pawns each",
		},
		{
			name:        "bishops only",
			fen:         "8/8/2b2b2/8/8/2B2B2/8/8 w - - 0 1",
			expected:    0,
			description: "Symmetric bishop placement (white on c3,f3 vs black on c6,f6)",
		},
	}

	evaluator := NewEvaluator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.evaluateMaterialAndPST(b)

			if score != tt.expected {
				t.Errorf("Expected material + PST score %d, got %d", tt.expected, score)
			}

			t.Logf("Material + PST score: %d (%s)", score, tt.description)
		})
	}
}

func TestGetPositionalBonus(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		rank     int
		file     int
		piece    board.Piece
		expected int
	}{
		{
			name:     "black bishop on b7",
			fen:      "8/1b6/8/8/8/8/8/8 w - - 0 1",
			rank:     6,
			file:     1,
			piece:    board.BlackBishop,
			expected: -5,
		},
		{
			name:     "black rook on b2",
			fen:      "8/8/8/8/8/8/1r6/8 w - - 0 1",
			rank:     1,
			file:     1,
			piece:    board.BlackRook,
			expected: -10,
		},
		{
			name:     "black king on e8",
			fen:      "4k3/8/8/8/8/8/8/8 w - - 0 1",
			rank:     7,
			file:     4,
			piece:    board.BlackKing,
			expected: 0,
		},
		{
			name:     "black pawn on g2",
			fen:      "8/8/8/8/8/8/6p1/8 w - - 0 1",
			rank:     1,
			file:     6,
			piece:    board.BlackPawn,
			expected: -50,
		},
		{
			name:     "white bishop on g2",
			fen:      "8/8/8/8/8/8/6B1/8 w - - 0 1",
			rank:     1,
			file:     6,
			piece:    board.WhiteBishop,
			expected: 5,
		},
		{
			name:     "white rook on d7",
			fen:      "8/3R4/8/8/8/8/8/8 w - - 0 1",
			rank:     6,
			file:     3,
			piece:    board.WhiteRook,
			expected: 10,
		},
		{
			name:     "black knight on d2",
			fen:      "8/8/8/8/8/8/3n4/8 w - - 0 1",
			rank:     1,
			file:     3,
			piece:    board.BlackKnight,
			expected: -5,
		},
		{
			name:     "black king on h1",
			fen:      "8/8/8/8/8/8/8/7k w - - 0 1",
			rank:     0,
			file:     7,
			piece:    board.BlackKing,
			expected: 30,
		},
		{
			name:     "white king on h8",
			fen:      "7K/8/8/8/8/8/8/8 w - - 0 1",
			rank:     7,
			file:     7,
			piece:    board.WhiteKing,
			expected: -30,
		},
		{
			name:     "black pawn on e7",
			fen:      "8/4p3/8/8/8/8/8/8 w - - 0 1",
			rank:     6,
			file:     4,
			piece:    board.BlackPawn,
			expected: 20,
		},
		{
			name:     "white pawn on e2",
			fen:      "8/8/8/8/8/8/4P3/8 w - - 0 1",
			rank:     1,
			file:     4,
			piece:    board.WhitePawn,
			expected: -20,
		},
		{
			name:     "black pawn on a7",
			fen:      "8/p7/8/8/8/8/8/8 w - - 0 1",
			rank:     6,
			file:     0,
			piece:    board.BlackPawn,
			expected: -5,
		},
		{
			name:     "white rook on h7",
			fen:      "8/7R/8/8/8/8/8/8 w - - 0 1",
			rank:     6,
			file:     7,
			piece:    board.WhiteRook,
			expected: 5,
		},
		{
			name:     "white knight on e4",
			fen:      "8/8/8/8/4N3/8/8/8 w - - 0 1",
			rank:     3,
			file:     4,
			piece:    board.WhiteKnight,
			expected: 20,
		},
		{
			name:     "black knight on a1",
			fen:      "8/8/8/8/8/8/8/n7 w - - 0 1",
			rank:     0,
			file:     0,
			piece:    board.BlackKnight,
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create board from FEN to verify piece placement
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			// Verify the piece is where we expect it
			actualPiece := b.GetPiece(tt.rank, tt.file)
			if actualPiece != tt.piece {
				t.Fatalf("Expected piece %v at rank %d, file %d, but got %v", tt.piece, tt.rank, tt.file, actualPiece)
			}

			// Test the positional bonus function
			got := getPositionalBonus(tt.piece, tt.rank, tt.file)
			if got != tt.expected {
				t.Errorf("getPositionalBonus(%v, %d, %d) = %d, want %d", tt.piece, tt.rank, tt.file, got, tt.expected)
			}
		})
	}
}
