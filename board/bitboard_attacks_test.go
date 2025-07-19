package board

import (
	"testing"
)

func TestIsSquareAttackedByColor(t *testing.T) {
	board := NewBoard()
	
	// Set up a test position
	board.SetPiece(1, 1, WhitePawn)   // b2
	board.SetPiece(2, 2, WhiteKnight) // c3
	board.SetPiece(0, 0, WhiteRook)   // a1
	board.SetPiece(3, 3, WhiteBishop) // d4
	board.SetPiece(4, 4, WhiteQueen)  // e5
	board.SetPiece(0, 4, WhiteKing)   // e1
	
	testCases := []struct {
		square   string
		color    BitboardColor
		expected bool
		reason   string
	}{
		{"c3", BitboardWhite, true, "pawn on b2 attacks c3"},
		{"a3", BitboardWhite, true, "pawn on b2 attacks a3"},
		{"b1", BitboardWhite, true, "knight on c3 attacks b1"},
		{"d5", BitboardWhite, true, "knight on c3 attacks d5"},
		{"a2", BitboardWhite, true, "rook on a1 attacks a2"},
		{"c1", BitboardWhite, true, "rook on a1 attacks c1"},
		{"f2", BitboardWhite, true, "bishop on d4 attacks f2"},
		{"c5", BitboardWhite, true, "bishop on d4 attacks c5"},
		{"e6", BitboardWhite, true, "queen on e5 attacks e6"},
		{"a5", BitboardWhite, true, "queen on e5 attacks a5"},
		{"d1", BitboardWhite, true, "king on e1 attacks d1"},
		{"f1", BitboardWhite, true, "king on e1 attacks f1"},
		{"h8", BitboardWhite, true, "queen on e5 attacks h8"},
		{"a8", BitboardWhite, true, "rook on a1 attacks a8"},
	}
	
	for _, tc := range testCases {
		square := StringToSquare(tc.square)
		result := board.IsSquareAttackedByColor(square, tc.color)
		if result != tc.expected {
			t.Errorf("IsSquareAttackedByColor(%s, %s): expected %v, got %v - %s", 
				tc.square, colorName(tc.color), tc.expected, result, tc.reason)
		}
	}
}

func TestPawnAttacks(t *testing.T) {
	board := NewBoard()
	
	// White pawn attacks
	board.SetPiece(3, 4, WhitePawn) // e4
	
	// Test squares that should be attacked by white pawn
	attackedSquares := []string{"d5", "f5"}
	for _, square := range attackedSquares {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("White pawn on e4 should attack %s", square)
		}
	}
	
	// Test squares that should NOT be attacked
	notAttackedSquares := []string{"e5", "d4", "f4", "e3"}
	for _, square := range notAttackedSquares {
		sq := StringToSquare(square)
		if board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("White pawn on e4 should NOT attack %s", square)
		}
	}
	
	// Black pawn attacks
	board.SetPiece(4, 4, BlackPawn) // e5
	
	// Test squares that should be attacked by black pawn
	blackAttackedSquares := []string{"d4", "f4"}
	for _, square := range blackAttackedSquares {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardBlack) {
			t.Errorf("Black pawn on e5 should attack %s", square)
		}
	}
}

func TestKnightAttacks(t *testing.T) {
	board := NewBoard()
	board.SetPiece(3, 4, WhiteKnight) // e4
	
	expectedAttacks := []string{"d2", "f2", "c3", "g3", "c5", "g5", "d6", "f6"}
	for _, square := range expectedAttacks {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Knight on e4 should attack %s", square)
		}
	}
	
	// Test squares that should NOT be attacked
	notAttacked := []string{"e4", "e3", "e5", "d4", "f4"}
	for _, square := range notAttacked {
		sq := StringToSquare(square)
		if board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Knight on e4 should NOT attack %s", square)
		}
	}
}

func TestSlidingPieceAttacks(t *testing.T) {
	board := NewBoard()
	
	// Test rook attacks
	board.SetPiece(0, 0, WhiteRook) // a1
	
	// Rook should attack entire first rank and a-file
	rookAttacks := []string{"a2", "a8", "b1", "h1"}
	for _, square := range rookAttacks {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Rook on a1 should attack %s", square)
		}
	}
	
	// Test bishop attacks
	board.SetPiece(3, 3, WhiteBishop) // d4
	
	bishopAttacks := []string{"a1", "c3", "e5", "f6", "g7", "h8", "c5", "b6", "a7"}
	for _, square := range bishopAttacks {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Bishop on d4 should attack %s", square)
		}
	}
	
	// Test queen attacks (combination of rook and bishop)
	board.SetPiece(3, 4, WhiteQueen) // e4
	
	queenAttacks := []string{"e1", "e8", "a4", "h4", "b1", "h7", "a8", "g2"}
	for _, square := range queenAttacks {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Queen on e4 should attack %s", square)
		}
	}
}

func TestSlidingPieceAttacksWithBlockers(t *testing.T) {
	board := NewBoard()
	
	// Place rook and a blocker
	board.SetPiece(0, 0, WhiteRook) // a1
	board.SetPiece(0, 3, BlackPawn) // d1 (blocker)
	
	// Rook should attack up to the blocker (including blocker square)
	shouldAttack := []string{"b1", "c1", "d1"}
	for _, square := range shouldAttack {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Rook on a1 should attack %s (up to blocker)", square)
		}
	}
	
	// Rook should NOT attack beyond the blocker
	shouldNotAttack := []string{"e1", "f1", "g1", "h1"}
	for _, square := range shouldNotAttack {
		sq := StringToSquare(square)
		if board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("Rook on a1 should NOT attack %s (beyond blocker)", square)
		}
	}
}

func TestKingAttacks(t *testing.T) {
	board := NewBoard()
	board.SetPiece(3, 4, WhiteKing) // e4
	
	expectedAttacks := []string{"d3", "e3", "f3", "d4", "f4", "d5", "e5", "f5"}
	for _, square := range expectedAttacks {
		sq := StringToSquare(square)
		if !board.IsSquareAttackedByColor(sq, BitboardWhite) {
			t.Errorf("King on e4 should attack %s", square)
		}
	}
}

func TestGetAttackersToSquare(t *testing.T) {
	board := NewBoard()
	
	// Set up multiple attackers to e4
	board.SetPiece(2, 3, WhitePawn)   // d3 -> attacks e4 (pawn attacks diagonally forward)
	board.SetPiece(2, 2, WhiteKnight) // c3 -> attacks e4 (knight L-shape: 2 squares in one direction, 1 in perpendicular)
	board.SetPiece(3, 0, WhiteRook)   // a4 -> attacks e4 (rook attacks horizontally)
	board.SetPiece(1, 1, WhiteBishop) // b2 -> doesn't attack e4 (not on diagonal), just for testing other functionality
	
	e4Square := StringToSquare("e4")
	attackers := board.GetAttackersToSquare(e4Square, BitboardWhite)
	
	expectedAttackers := []string{"d3", "c3", "a4"}
	actualCount := attackers.PopCount()
	if actualCount != len(expectedAttackers) {
		t.Errorf("Expected %d attackers to e4, got %d", len(expectedAttackers), actualCount)
		// Debug output
		actualAttackers := attackers.BitList()
		for _, sq := range actualAttackers {
			t.Logf("Actual attacker at: %s", SquareToString(sq))
		}
	}
	
	// Check individual attackers
	for _, expectedSquare := range expectedAttackers {
		sq := StringToSquare(expectedSquare)
		if !attackers.HasBit(sq) {
			t.Errorf("Expected %s to be an attacker of e4", expectedSquare)
		}
	}
}

func TestIsInCheck(t *testing.T) {
	// Test starting position - no check
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse starting position: %v", err)
	}
	
	if board.IsInCheck(BitboardWhite) {
		t.Error("White should not be in check in starting position")
	}
	if board.IsInCheck(BitboardBlack) {
		t.Error("Black should not be in check in starting position")
	}
	
	// Test position with white in check
	board = NewBoard()
	board.SetPiece(0, 4, WhiteKing)  // e1
	board.SetPiece(7, 4, BlackRook)  // e8 - attacks white king
	
	if !board.IsInCheck(BitboardWhite) {
		t.Error("White should be in check from black rook")
	}
	if board.IsInCheck(BitboardBlack) {
		t.Error("Black should not be in check")
	}
	
	// Test position with black in check
	board = NewBoard()
	board.SetPiece(7, 4, BlackKing)   // e8
	board.SetPiece(6, 3, WhiteBishop) // d7 - attacks black king on e8
	
	if board.IsInCheck(BitboardWhite) {
		t.Error("White should not be in check")
	}
	if !board.IsInCheck(BitboardBlack) {
		t.Error("Black should be in check from white bishop")
	}
}

func TestGetPieceAttacks(t *testing.T) {
	board := NewBoard()
	
	// Test pawn attacks
	e4Square := StringToSquare("e4")
	pawnAttacks := board.GetPieceAttacks(WhitePawn, e4Square)
	
	expectedPawnAttacks := []int{StringToSquare("d5"), StringToSquare("f5")}
	if pawnAttacks.PopCount() != 2 {
		t.Errorf("White pawn should have 2 attacks, got %d", pawnAttacks.PopCount())
	}
	
	for _, sq := range expectedPawnAttacks {
		if !pawnAttacks.HasBit(sq) {
			t.Errorf("Pawn attacks should include %s", SquareToString(sq))
		}
	}
	
	// Test knight attacks
	knightAttacks := board.GetPieceAttacks(WhiteKnight, e4Square)
	if knightAttacks.PopCount() != 8 {
		t.Errorf("Knight should have 8 attacks from e4, got %d", knightAttacks.PopCount())
	}
	
	// Test empty piece
	emptyAttacks := board.GetPieceAttacks(Empty, e4Square)
	if emptyAttacks != 0 {
		t.Error("Empty piece should have no attacks")
	}
}

func TestGetAllAttackedSquares(t *testing.T) {
	board := NewBoard()
	
	// Set up some pieces
	board.SetPiece(1, 4, WhitePawn)   // e2
	board.SetPiece(2, 1, WhiteKnight) // b3
	board.SetPiece(0, 4, WhiteKing)   // e1
	
	attacks := board.GetAllAttackedSquares(BitboardWhite)
	
	// Should include pawn attacks
	if !attacks.HasBit(StringToSquare("d3")) || !attacks.HasBit(StringToSquare("f3")) {
		t.Error("Should include pawn attacks d3 and f3")
	}
	
	// Should include some knight attacks
	if !attacks.HasBit(StringToSquare("d4")) || !attacks.HasBit(StringToSquare("a5")) {
		t.Error("Should include knight attacks")
	}
	
	// Should include king attacks
	if !attacks.HasBit(StringToSquare("d1")) || !attacks.HasBit(StringToSquare("f1")) {
		t.Error("Should include king attacks")
	}
}

func TestIsSquareEmptyBitboard(t *testing.T) {
	board := NewBoard()
	
	// Empty square
	e4Square := StringToSquare("e4")
	if !board.IsSquareEmptyBitboard(e4Square) {
		t.Error("e4 should be empty in new board")
	}
	
	// Occupied square
	board.SetPiece(3, 4, WhitePawn) // e4
	if board.IsSquareEmptyBitboard(e4Square) {
		t.Error("e4 should not be empty after placing pawn")
	}
	
	// Invalid square
	if board.IsSquareEmptyBitboard(-1) {
		t.Error("Invalid square should not be considered empty")
	}
	if board.IsSquareEmptyBitboard(64) {
		t.Error("Invalid square should not be considered empty")
	}
}

func TestGetPieceOnSquare(t *testing.T) {
	board := NewBoard()
	
	// Empty square
	e4Square := StringToSquare("e4")
	piece := board.GetPieceOnSquare(e4Square)
	if piece != Empty {
		t.Errorf("Expected Empty on e4, got %c", piece)
	}
	
	// Occupied square
	board.SetPiece(3, 4, WhiteQueen) // e4
	piece = board.GetPieceOnSquare(e4Square)
	if piece != WhiteQueen {
		t.Errorf("Expected WhiteQueen on e4, got %c", piece)
	}
	
	// Invalid squares
	piece = board.GetPieceOnSquare(-1)
	if piece != Empty {
		t.Error("Invalid square should return Empty")
	}
	
	piece = board.GetPieceOnSquare(64)
	if piece != Empty {
		t.Error("Invalid square should return Empty")
	}
}

func TestAttackDetectionPerformance(t *testing.T) {
	// Set up a complex position
	board, err := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}
	
	// Test attack detection for all squares
	attackedCount := 0
	for square := 0; square < 64; square++ {
		if board.IsSquareAttackedByColor(square, BitboardWhite) {
			attackedCount++
		}
	}
	
	// Should have reasonable number of attacked squares
	if attackedCount < 10 || attackedCount > 40 {
		t.Errorf("Expected reasonable number of attacked squares, got %d", attackedCount)
	}
}

// Helper function for color names in tests
func colorName(color BitboardColor) string {
	if color == BitboardWhite {
		return "White"
	}
	return "Black"
}

// Benchmark tests
func BenchmarkIsSquareAttackedByColor(b *testing.B) {
	board, _ := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	square := StringToSquare("e4")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.IsSquareAttackedByColor(square, BitboardWhite)
	}
}

func BenchmarkGetAttackersToSquare(b *testing.B) {
	board, _ := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	square := StringToSquare("e4")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.GetAttackersToSquare(square, BitboardWhite)
	}
}

func BenchmarkIsInCheck(b *testing.B) {
	board, _ := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.IsInCheck(BitboardWhite)
	}
}

func BenchmarkGetAllAttackedSquares(b *testing.B) {
	board, _ := FromFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.GetAllAttackedSquares(BitboardWhite)
	}
}