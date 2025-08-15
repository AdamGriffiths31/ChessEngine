package evaluation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewSEECalculator(t *testing.T) {
	calc := NewSEECalculator()
	if calc == nil {
		t.Fatal("NewSEECalculator() returned nil")
	}
}

func TestSEE_SimpleCaptures(t *testing.T) {
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
				From:      board.Square{Rank: 3, File: 3}, // d4
				To:        board.Square{Rank: 4, File: 4}, // e5
				Piece:     board.WhitePawn,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100, // Gain a pawn
		},
		{
			name: "Queen takes pawn - undefended",
			fen:  "4k3/8/8/4p3/8/8/3Q4/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3}, // d2
				To:        board.Square{Rank: 4, File: 4}, // e5
				Piece:     board.WhiteQueen,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100, // Gain a pawn
		},
		{
			name: "Queen takes defended pawn - losing capture",
			fen:  "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1", // Pawn on e5 can defend d4, kings added
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3}, // d2
				To:        board.Square{Rank: 3, File: 3}, // d4
				Piece:     board.WhiteQueen,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: -800, // Lose queen (900) for pawn (100)
		},
		{
			name: "Rook takes queen - great capture",
			fen:  "4k3/8/8/4q3/8/8/3R4/4K3 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3}, // d2
				To:        board.Square{Rank: 4, File: 4}, // e5
				Piece:     board.WhiteRook,
				Captured:  board.BlackQueen,
				IsCapture: true,
			},
			expected: 900, // Gain a queen
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
				From:      board.Square{Rank: 3, File: 3}, // d4 bishop
				To:        board.Square{Rank: 4, File: 4}, // e5 knight
				Piece:     board.WhiteBishop,
				Captured:  board.BlackKnight,
				IsCapture: true,
			},
			// Analysis: Only Bxe5 possible - black rooks blocked by king on e8
			// Net: Just gain knight = +320
			expected: 320,
		},
		{
			name: "X-ray attack with queen behind rook",
			fen:  "8/8/8/4p3/8/8/3RQ3/8 w - - 0 1",
			move: board.Move{
				From:      board.Square{Rank: 1, File: 3}, // d2 rook
				To:        board.Square{Rank: 4, File: 4}, // e5
				Piece:     board.WhiteRook,
				Captured:  board.BlackPawn,
				IsCapture: true,
			},
			expected: 100, // Just gain the pawn (no defenders)
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
	// Set up en passant position
	b := board.NewBoard()
	// Place white pawn on e5 and black pawn on d5 (that just moved two squares)
	b.SetPiece(4, 4, board.WhitePawn)                     // e5
	b.SetPiece(4, 3, board.BlackPawn)                     // d5
	b.SetEnPassantTarget(&board.Square{Rank: 5, File: 3}) // d6

	move := board.Move{
		From:        board.Square{Rank: 4, File: 4}, // e5
		To:          board.Square{Rank: 5, File: 3}, // d6
		Piece:       board.WhitePawn,
		Captured:    board.BlackPawn,
		IsEnPassant: true,
		IsCapture:   true,
	}

	calc := NewSEECalculator()
	result := calc.SEE(b, move)

	// Should gain a pawn (100)
	expected := 100
	if result != expected {
		t.Errorf("En passant SEE() = %d, expected %d", result, expected)
	}
}

func TestSEE_NonCapture(t *testing.T) {
	b := board.NewBoard()
	b.SetPiece(1, 3, board.WhiteRook) // d2

	move := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 3, File: 3}, // d4
		Piece:     board.WhiteRook,
		Captured:  board.Empty,
		IsCapture: false,
	}

	calc := NewSEECalculator()
	result := calc.SEE(b, move)

	// Non-captures should return 0
	if result != 0 {
		t.Errorf("Non-capture SEE() = %d, expected 0", result)
	}
}

func TestSEE_PawnAttackers(t *testing.T) {
	// Setup board with pawns attacking e5
	b := board.NewBoard()
	b.SetPiece(3, 3, board.WhitePawn) // d4 - attacks e5
	b.SetPiece(3, 5, board.WhitePawn) // f4 - attacks e5
	b.SetPiece(5, 3, board.BlackPawn) // d6 - attacks e5
	b.SetPiece(5, 5, board.BlackPawn) // f6 - attacks e5

	e5Square := 4*8 + 4 // e5
	whiteAttackers := b.GetAttackersToSquare(e5Square, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(e5Square, board.BitboardBlack)

	// Should find 2 white pawns and 2 black pawns attacking e5
	expectedWhiteCount := 2
	expectedBlackCount := 2
	actualWhiteCount := whiteAttackers.PopCount()
	actualBlackCount := blackAttackers.PopCount()

	if actualWhiteCount != expectedWhiteCount {
		t.Errorf("Expected %d white pawn attackers, got %d", expectedWhiteCount, actualWhiteCount)
	}
	if actualBlackCount != expectedBlackCount {
		t.Errorf("Expected %d black pawn attackers, got %d", expectedBlackCount, actualBlackCount)
	}
}

func TestSEE_KnightAttackers(t *testing.T) {
	// Setup board with knights attacking e5
	b := board.NewBoard()
	b.SetPiece(2, 3, board.WhiteKnight) // d3 - attacks e5
	b.SetPiece(6, 5, board.BlackKnight) // f7 - attacks e5

	e5Square := 4*8 + 4 // e5
	whiteAttackers := b.GetAttackersToSquare(e5Square, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(e5Square, board.BitboardBlack)

	// Should find 1 white knight and 1 black knight attacking e5
	expectedWhiteCount := 1
	expectedBlackCount := 1
	actualWhiteCount := whiteAttackers.PopCount()
	actualBlackCount := blackAttackers.PopCount()

	if actualWhiteCount != expectedWhiteCount {
		t.Errorf("Expected %d white knight attackers, got %d", expectedWhiteCount, actualWhiteCount)
	}
	if actualBlackCount != expectedBlackCount {
		t.Errorf("Expected %d black knight attackers, got %d", expectedBlackCount, actualBlackCount)
	}
}

func TestSEE_SlidingPieceAttackers(t *testing.T) {
	// Setup board with sliding pieces attacking e5
	b := board.NewBoard()
	b.SetPiece(4, 0, board.WhiteRook)   // a5 - attacks e5 horizontally
	b.SetPiece(7, 4, board.BlackRook)   // e8 - attacks e5 vertically
	b.SetPiece(2, 2, board.WhiteBishop) // c3 - attacks e5 diagonally
	b.SetPiece(6, 6, board.BlackQueen)  // g7 - attacks e5 diagonally

	e5Square := 4*8 + 4 // e5
	whiteAttackers := b.GetAttackersToSquare(e5Square, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(e5Square, board.BitboardBlack)

	// Should find 2 white sliding pieces and 2 black sliding pieces attacking e5
	expectedWhiteCount := 2
	expectedBlackCount := 2
	actualWhiteCount := whiteAttackers.PopCount()
	actualBlackCount := blackAttackers.PopCount()

	if actualWhiteCount != expectedWhiteCount {
		t.Errorf("Expected %d white sliding piece attackers, got %d", expectedWhiteCount, actualWhiteCount)
	}
	if actualBlackCount != expectedBlackCount {
		t.Errorf("Expected %d black sliding piece attackers, got %d", expectedBlackCount, actualBlackCount)
	}
}

func TestSEE_GetPieceValue(t *testing.T) {
	calc := NewSEECalculator()

	tests := []struct {
		piece    board.Piece
		expected int
	}{
		{board.WhitePawn, 100},
		{board.BlackPawn, 100},
		{board.WhiteKnight, 320},
		{board.BlackKnight, 320},
		{board.WhiteBishop, 330},
		{board.BlackBishop, 330},
		{board.WhiteRook, 500},
		{board.BlackRook, 500},
		{board.WhiteQueen, 900},
		{board.BlackQueen, 900},
		{board.WhiteKing, 10000},
		{board.BlackKing, 10000},
		{board.Empty, 0},
	}

	for _, tt := range tests {
		result := calc.getPieceValue(tt.piece)
		if result != tt.expected {
			t.Errorf("getPieceValue(%v) = %d, expected %d", tt.piece, result, tt.expected)
		}
	}
}

func TestSEE_IsWhitePiece(t *testing.T) {
	calc := NewSEECalculator()

	tests := []struct {
		piece    board.Piece
		expected bool
	}{
		{board.WhitePawn, true},
		{board.WhiteKnight, true},
		{board.WhiteBishop, true},
		{board.WhiteRook, true},
		{board.WhiteQueen, true},
		{board.WhiteKing, true},
		{board.BlackPawn, false},
		{board.BlackKnight, false},
		{board.BlackBishop, false},
		{board.BlackRook, false},
		{board.BlackQueen, false},
		{board.BlackKing, false},
		{board.Empty, false},
	}

	for _, tt := range tests {
		result := calc.isWhitePiece(tt.piece)
		if result != tt.expected {
			t.Errorf("isWhitePiece(%v) = %t, expected %t", tt.piece, result, tt.expected)
		}
	}
}

// Benchmark tests
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
	for i := 0; i < b.N; i++ {
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
	for i := 0; i < b.N; i++ {
		calc.SEE(testBoard, move)
	}
}


// Helper function to create board from FEN using existing functionality
func createBoardFromFEN(t testing.TB, fen string) *board.Board {
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN %s: %v", fen, err)
		return nil
	}

	// Only print the board for non-benchmark tests
	if testing.Short() {
		return b
	}

	// Check if this is a benchmark test by looking at the test name
	testName := t.Name()
	if strings.Contains(testName, "Benchmark") {
		return b // Skip printing for benchmarks
	}

	// Debug: Print the board
	t.Logf("Board from FEN: %s", fen)
	for rank := 7; rank >= 0; rank-- {
		line := fmt.Sprintf("%d ", rank+1)
		for file := 0; file < 8; file++ {
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
