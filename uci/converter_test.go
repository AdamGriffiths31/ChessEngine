package uci

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestMoveConverter_ToUCI(t *testing.T) {
	converter := NewMoveConverter()

	tests := []struct {
		name     string
		move     board.Move
		expected string
	}{
		{
			name: "simple pawn move",
			move: board.Move{
				From:      board.Square{File: 4, Rank: 1}, // e2
				To:        board.Square{File: 4, Rank: 3}, // e4
				Promotion: board.Empty,
			},
			expected: "e2e4",
		},
		{
			name: "knight move",
			move: board.Move{
				From:      board.Square{File: 1, Rank: 0}, // b1
				To:        board.Square{File: 2, Rank: 2}, // c3
				Promotion: board.Empty,
			},
			expected: "b1c3",
		},
		{
			name: "promotion to queen",
			move: board.Move{
				From:      board.Square{File: 0, Rank: 6}, // a7
				To:        board.Square{File: 0, Rank: 7}, // a8
				Promotion: board.WhiteQueen,
			},
			expected: "a7a8q",
		},
		{
			name: "promotion to knight",
			move: board.Move{
				From:      board.Square{File: 7, Rank: 1}, // h2
				To:        board.Square{File: 7, Rank: 0}, // h1
				Promotion: board.BlackKnight,
			},
			expected: "h2h1n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToUCI(tt.move)
			if result != tt.expected {
				t.Errorf("ToUCI() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMoveConverter_FromUCI(t *testing.T) {
	converter := NewMoveConverter()

	tests := []struct {
		name        string
		uciMove     string
		boardFEN    string
		expectedErr bool
		validate    func(board.Move) bool
	}{
		{
			name:     "simple pawn move",
			uciMove:  "e2e4",
			boardFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			validate: func(m board.Move) bool {
				return m.From.File == 4 && m.From.Rank == 1 &&
					m.To.File == 4 && m.To.Rank == 3 &&
					m.Piece == board.WhitePawn
			},
		},
		{
			name:     "knight move",
			uciMove:  "g1f3",
			boardFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			validate: func(m board.Move) bool {
				return m.From.File == 6 && m.From.Rank == 0 &&
					m.To.File == 5 && m.To.Rank == 2 &&
					m.Piece == board.WhiteKnight
			},
		},
		{
			name:     "promotion",
			uciMove:  "a7a8q",
			boardFEN: "rnbqkbnr/Pppppppp/8/8/8/8/1PPPPPPP/RNBQKBNR w KQkq - 0 1",
			validate: func(m board.Move) bool {
				return m.From.File == 0 && m.From.Rank == 6 &&
					m.To.File == 0 && m.To.Rank == 7 &&
					m.Piece == board.WhitePawn &&
					m.Promotion == board.WhiteQueen
			},
		},
		{
			name:        "invalid move format",
			uciMove:     "e2",
			boardFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedErr: true,
		},
		{
			name:        "invalid square",
			uciMove:     "z9e4",
			boardFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedErr: true,
		},
		{
			name:        "no piece on from square",
			uciMove:     "e4e5",
			boardFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.boardFEN)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			move, err := converter.FromUCI(tt.uciMove, b)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("FromUCI() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FromUCI() unexpected error: %v", err)
				return
			}

			if tt.validate != nil && !tt.validate(move) {
				t.Errorf("FromUCI() move validation failed: %+v", move)
			}
		})
	}
}

func TestSquareToUCI(t *testing.T) {
	tests := []struct {
		name     string
		square   board.Square
		expected string
	}{
		{
			name:     "a1",
			square:   board.Square{File: 0, Rank: 0},
			expected: "a1",
		},
		{
			name:     "h8",
			square:   board.Square{File: 7, Rank: 7},
			expected: "h8",
		},
		{
			name:     "e4",
			square:   board.Square{File: 4, Rank: 3},
			expected: "e4",
		},
		{
			name:     "d5",
			square:   board.Square{File: 3, Rank: 4},
			expected: "d5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := squareToUCI(tt.square)
			if result != tt.expected {
				t.Errorf("squareToUCI() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseUCISquare(t *testing.T) {
	tests := []struct {
		name        string
		uciSquare   string
		expected    board.Square
		expectedErr bool
	}{
		{
			name:      "a1",
			uciSquare: "a1",
			expected:  board.Square{File: 0, Rank: 0},
		},
		{
			name:      "h8",
			uciSquare: "h8",
			expected:  board.Square{File: 7, Rank: 7},
		},
		{
			name:      "e4",
			uciSquare: "e4",
			expected:  board.Square{File: 4, Rank: 3},
		},
		{
			name:        "invalid length",
			uciSquare:   "a",
			expectedErr: true,
		},
		{
			name:        "out of bounds file",
			uciSquare:   "z1",
			expectedErr: true,
		},
		{
			name:        "out of bounds rank",
			uciSquare:   "a9",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUCISquare(tt.uciSquare)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("parseUCISquare() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseUCISquare() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("parseUCISquare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsePromotionPiece(t *testing.T) {
	tests := []struct {
		name          string
		promotionChar byte
		originalPiece board.Piece
		expected      board.Piece
	}{
		{
			name:          "white pawn to queen",
			promotionChar: 'q',
			originalPiece: board.WhitePawn,
			expected:      board.WhiteQueen,
		},
		{
			name:          "black pawn to queen",
			promotionChar: 'q',
			originalPiece: board.BlackPawn,
			expected:      board.BlackQueen,
		},
		{
			name:          "white pawn to knight",
			promotionChar: 'n',
			originalPiece: board.WhitePawn,
			expected:      board.WhiteKnight,
		},
		{
			name:          "black pawn to rook",
			promotionChar: 'r',
			originalPiece: board.BlackPawn,
			expected:      board.BlackRook,
		},
		{
			name:          "white pawn to bishop",
			promotionChar: 'b',
			originalPiece: board.WhitePawn,
			expected:      board.WhiteBishop,
		},
		{
			name:          "invalid promotion defaults to queen",
			promotionChar: 'x',
			originalPiece: board.WhitePawn,
			expected:      board.WhiteQueen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePromotionPiece(tt.promotionChar, tt.originalPiece)
			if result != tt.expected {
				t.Errorf("parsePromotionPiece() = %v, want %v", result, tt.expected)
			}
		})
	}
}
