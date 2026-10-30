package eval

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func TestSEE_XRayBattery(t *testing.T) {
	t.Parallel()
	fen := "3r3k/8/8/3p4/8/8/3R4/3R3K w - - 0 1"
	move := board.Move{
		From:      board.Square{Rank: 1, File: 3},
		To:        board.Square{Rank: 4, File: 3},
		Piece:     board.WhiteRook,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	b := createBoardFromFEN(t, fen)
	calc := NewSEECalculator()

	result := calc.SEE(b, move)
	if result != 100 {
		t.Errorf("X-ray battery capture: expected 100, got %d", result)
	}
}

func TestSEE_SimpleCaptures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		move     board.Move
		expected int
	}{
		{
			name: "Pawn takes pawn - equal exchange",
			fen:  "4k3/8/8/4p3/3P4/8/8/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 3, File: 3},
				To:        board.Square{Rank: 4, File: 4},
				Piece:     board.WhitePawn,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100,
		},
		{
			name: "Queen takes pawn - undefended",
			fen:  "4k3/8/8/4p3/8/8/3Q4/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3},
				To:        board.Square{Rank: 4, File: 4},
				Piece:     board.WhiteQueen,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100,
		},
		{
			name: "Queen takes defended pawn - losing capture",
			fen:  "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3},
				To:        board.Square{Rank: 3, File: 3},
				Piece:     board.WhiteQueen,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: -800,
		},
		{
			name: "Rook takes queen - great capture",
			fen:  "4k3/8/8/4q3/8/8/3R4/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3},
				To:        board.Square{Rank: 4, File: 4},
				Piece:     board.WhiteRook,
				Captured:  board.BlackQueen,
				IsCapture: true,
			},
			expected: 900,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := createBoardFromFEN(t, tt.fen)
			calc := NewSEECalculator()

			result := calc.SEE(b, tt.move)
			if result != tt.expected {
				t.Errorf("SEE() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestSEE_ComplexExchanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		move     board.Move
		expected int
	}{
		{
			name: "Multiple attackers and defenders",
			fen:  "r3k2r/8/8/4n3/3BNB2/8/8/R3K2R w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 3, File: 3},
				To:        board.Square{Rank: 4, File: 4},
				Piece:     board.WhiteBishop,
				Captured:  board.BlackKnight,
				IsCapture: true,
			},
			expected: 320,
		},
		{
			name: "X-ray attack with queen behind rook",
			fen:  "8/8/8/4p3/8/8/3RQ3/8 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3},
				To:        board.Square{Rank: 4, File: 4},
				Piece:     board.WhiteRook,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := createBoardFromFEN(t, tt.fen)
			calc := NewSEECalculator()

			result := calc.SEE(b, tt.move)
			if result != tt.expected {
				t.Errorf("SEE() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestSEE_EnPassant(t *testing.T) {
	t.Parallel()
	b := board.NewBoard()
	b.SetPiece(4, 4, board.WhitePawn)
	b.SetPiece(4, 3, board.BlackPawn)
	b.SetEnPassantTarget(board.Square{Rank: 5, File: 3}, true)

	move := board.Move{
		From:        board.Square{Rank: 4, File: 4},
		To:          board.Square{Rank: 5, File: 3},
		Piece:       board.WhitePawn,
		Captured:    board.BlackPawn,
		IsEnPassant: true,
		IsCapture:   true,
	}

	calc := NewSEECalculator()
	result := calc.SEE(b, move)

	expected := 100
	if result != expected {
		t.Errorf("En passant SEE() = %d, expected %d", result, expected)
	}
}

func TestSEE_PromotionCaptures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fen      string
		move     board.Move
		expected int
	}{
		{
			name: "Queen promotion capture, recaptured by rook",
			fen:  "r3k3/1P6/8/r7/8/8/8/7K w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 6, File: 1},
				To:        board.Square{Rank: 7, File: 0},
				Piece:     board.WhitePawn,
				Captured:  board.BlackRook,
				Promotion: board.WhiteQueen,
				IsCapture: true,
			},
			expected: 400,
		},
		{
			name: "Queen promotion capture, square undefended",
			fen:  "n3k3/1P6/8/8/8/8/8/7K w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 6, File: 1},
				To:        board.Square{Rank: 7, File: 0},
				Piece:     board.WhitePawn,
				Captured:  board.BlackKnight,
				Promotion: board.WhiteQueen,
				IsCapture: true,
			},
			expected: 1120,
		},
		{
			name: "Underpromotion capture to knight, recaptured by king",
			fen:  "2bk4/1P6/8/8/8/8/8/7K w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 6, File: 1},
				To:        board.Square{Rank: 7, File: 2},
				Piece:     board.WhitePawn,
				Captured:  board.BlackBishop,
				Promotion: board.WhiteKnight,
				IsCapture: true,
			},
			expected: 230,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := createBoardFromFEN(t, tt.fen)
			calc := NewSEECalculator()

			result := calc.SEE(b, tt.move)
			if result != tt.expected {
				t.Errorf("SEE() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestSEE_NonCapture(t *testing.T) {
	t.Parallel()
	b := board.NewBoard()
	b.SetPiece(1, 3, board.WhiteRook)

	move := board.Move{
		From:      board.Square{Rank: 1, File: 3},
		To:        board.Square{Rank: 3, File: 3},
		Piece:     board.WhiteRook,
		Captured:  board.Empty,
		IsCapture: false,
	}

	calc := NewSEECalculator()
	result := calc.SEE(b, move)

	if result != 0 {
		t.Errorf("Non-capture SEE() = %d, expected 0", result)
	}
}

func BenchmarkSEE_SimpleCapture(b *testing.B) {
	testBoard := createBoardFromFEN(b, "8/8/8/4p3/3P4/8/8/8 w - - 0 1")
	calc := NewSEECalculator()
	move := board.Move{
		From:      board.Square{Rank: 3, File: 3},
		To:        board.Square{Rank: 4, File: 4},
		Piece:     board.WhitePawn,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	b.ResetTimer()
	for range b.N {
		calc.SEE(testBoard, move)
	}
}

func BenchmarkSEE_ComplexPosition(b *testing.B) {
	testBoard := createBoardFromFEN(b, "r3k2r/8/8/4n3/3BNB2/8/8/R3K2R w - - 0 1")
	calc := NewSEECalculator()
	move := board.Move{
		From:      board.Square{Rank: 3, File: 3},
		To:        board.Square{Rank: 4, File: 4},
		Piece:     board.WhiteBishop,
		Captured:  board.BlackKnight,
		IsCapture: true,
	}

	b.ResetTimer()
	for range b.N {
		calc.SEE(testBoard, move)
	}
}

func createBoardFromFEN(t testing.TB, fen string) *board.Board {
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN %s: %v", fen, err)
		return nil
	}

	if testing.Short() {
		return b
	}

	testName := t.Name()
	if strings.Contains(testName, "Benchmark") {
		return b
	}

	t.Logf("Board from FEN: %s", fen)
	for rank := 7; rank >= 0; rank-- {
		line := fmt.Sprintf("%d ", rank+1)
		for file := range 8 {
			piece := b.GetPiece(rank, file)
			if piece == board.Empty {
				line += "."
			} else {
				line += string(piece)
			}
		}
		t.Logf("%s", line)
	}
	t.Logf("  abcdefgh")

	return b
}
