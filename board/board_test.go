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

// Bitboard integration tests

func TestBitboardSynchronization(t *testing.T) {
	board := NewBoard()

	// Set some pieces and verify bitboards are updated
	board.SetPiece(0, 0, WhiteRook) // a1
	board.SetPiece(7, 7, BlackKing) // h8
	board.SetPiece(3, 4, WhitePawn) // e4

	// Check that bitboards have the correct pieces set
	if !board.GetPieceBitboard(WhiteRook).HasBit(FileRankToSquare(0, 0)) {
		t.Error("White rook bitboard should have a1 set")
	}
	if !board.GetPieceBitboard(BlackKing).HasBit(FileRankToSquare(7, 7)) {
		t.Error("Black king bitboard should have h8 set")
	}
	if !board.GetPieceBitboard(WhitePawn).HasBit(FileRankToSquare(4, 3)) {
		t.Error("White pawn bitboard should have e4 set")
	}

	// Check color bitboards
	whitePieces := board.GetColorBitboard(BitboardWhite)
	if !whitePieces.HasBit(FileRankToSquare(0, 0)) || !whitePieces.HasBit(FileRankToSquare(4, 3)) {
		t.Error("White pieces bitboard should include white rook and pawn")
	}

	blackPieces := board.GetColorBitboard(BitboardBlack)
	if !blackPieces.HasBit(FileRankToSquare(7, 7)) {
		t.Error("Black pieces bitboard should include black king")
	}

	// Check all pieces bitboard
	if board.AllPieces.PopCount() != 3 {
		t.Errorf("All pieces bitboard should have 3 pieces, got %d", board.AllPieces.PopCount())
	}
}

func TestSetPieceUpdatesAllRepresentations(t *testing.T) {
	board := NewBoard()

	// Set a piece
	board.SetPiece(3, 4, WhiteQueen) // e4

	// Verify array representation
	if board.GetPiece(3, 4) != WhiteQueen {
		t.Error("Array representation should have white queen on e4")
	}

	// Verify bitboard representation
	if !board.GetPieceBitboard(WhiteQueen).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Bitboard representation should have white queen on e4")
	}

	// Verify piece count using bitboard
	if board.getPieceCountFromBitboard(WhiteQueen) != 1 {
		t.Error("Should have exactly one white queen")
	}

	// Replace with different piece
	board.SetPiece(3, 4, BlackRook) // e4

	// Verify old piece is removed
	if board.GetPieceBitboard(WhiteQueen).HasBit(FileRankToSquare(4, 3)) {
		t.Error("White queen should be removed from bitboard")
	}

	// Verify new piece is added
	if !board.GetPieceBitboard(BlackRook).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Black rook should be added to bitboard")
	}

	// Verify array representation
	if board.GetPiece(3, 4) != BlackRook {
		t.Error("Array representation should have black rook on e4")
	}
}

func TestRemovePieceUpdatesAllRepresentations(t *testing.T) {
	board := NewBoard()

	// Set a piece then remove it
	board.SetPiece(3, 4, WhiteBishop) // e4
	board.SetPiece(3, 4, Empty)       // Remove piece

	// Verify array representation
	if board.GetPiece(3, 4) != Empty {
		t.Error("Array representation should be empty on e4")
	}

	// Verify bitboard representation
	if board.GetPieceBitboard(WhiteBishop).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Bitboard representation should not have white bishop on e4")
	}

	// Verify all pieces bitboard
	if board.AllPieces.HasBit(FileRankToSquare(4, 3)) {
		t.Error("All pieces bitboard should not have e4 set")
	}

	// Verify piece count using bitboard
	if board.getPieceCountFromBitboard(WhiteBishop) != 0 {
		t.Error("Should have no white bishops")
	}
}

func TestFENToBitboards(t *testing.T) {
	testCases := []struct {
		name     string
		fen      string
		piece    Piece
		square   int
		expected bool
	}{
		{
			name:     "starting position white king",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			piece:    WhiteKing,
			square:   FileRankToSquare(4, 0), // e1
			expected: true,
		},
		{
			name:     "starting position black queen",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			piece:    BlackQueen,
			square:   FileRankToSquare(3, 7), // d8
			expected: true,
		},
		{
			name:     "middle game position",
			fen:      "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
			piece:    WhiteBishop,
			square:   FileRankToSquare(2, 3), // c4
			expected: true,
		},
		{
			name:     "empty square in middle game",
			fen:      "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
			piece:    WhitePawn,
			square:   FileRankToSquare(3, 3), // d4
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			board, err := FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			hasPiece := board.GetPieceBitboard(tc.piece).HasBit(tc.square)
			if hasPiece != tc.expected {
				t.Errorf("Expected piece %c on square %s to be %v, got %v",
					tc.piece, SquareToString(tc.square), tc.expected, hasPiece)
			}
		})
	}
}

func TestStartingPositionBitboards(t *testing.T) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse starting position FEN: %v", err)
	}

	// Test piece counts
	expectedCounts := map[Piece]int{
		WhitePawn:   8,
		WhiteRook:   2,
		WhiteKnight: 2,
		WhiteBishop: 2,
		WhiteQueen:  1,
		WhiteKing:   1,
		BlackPawn:   8,
		BlackRook:   2,
		BlackKnight: 2,
		BlackBishop: 2,
		BlackQueen:  1,
		BlackKing:   1,
	}

	for piece, expectedCount := range expectedCounts {
		actualCount := board.GetPieceBitboard(piece).PopCount()
		if actualCount != expectedCount {
			t.Errorf("Expected %d %c pieces, got %d", expectedCount, piece, actualCount)
		}
	}

	// Test color bitboards
	if board.WhitePieces.PopCount() != 16 {
		t.Errorf("Expected 16 white pieces, got %d", board.WhitePieces.PopCount())
	}
	if board.BlackPieces.PopCount() != 16 {
		t.Errorf("Expected 16 black pieces, got %d", board.BlackPieces.PopCount())
	}
	if board.AllPieces.PopCount() != 32 {
		t.Errorf("Expected 32 total pieces, got %d", board.AllPieces.PopCount())
	}

	// Test that white pieces occupy ranks 1-2
	rank1and2 := RankMask(0) | RankMask(1)
	if (board.WhitePieces & rank1and2) != board.WhitePieces {
		t.Error("All white pieces should be on ranks 1-2")
	}

	// Test that black pieces occupy ranks 7-8
	rank7and8 := RankMask(6) | RankMask(7)
	if (board.BlackPieces & rank7and8) != board.BlackPieces {
		t.Error("All black pieces should be on ranks 7-8")
	}
}

func TestComplexPositionBitboards(t *testing.T) {
	// Test a complex middle-game position
	fen := "r2qkb1r/pb1p1ppp/1pn1pn2/8/2PP4/2N1PN2/PP3PPP/R1BQKB1R w KQkq - 0 8"
	board, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse complex position FEN: %v", err)
	}

	// Verify specific piece positions
	testCases := []struct {
		piece    Piece
		square   string
		expected bool
	}{
		{WhiteQueen, "d1", true},
		{BlackQueen, "d8", true},
		{WhitePawn, "c4", true},
		{WhitePawn, "d4", true},
		{BlackPawn, "b6", true},
		{BlackPawn, "e6", true},
		{WhiteKnight, "c3", true},
		{WhiteKnight, "f3", true},
		{BlackKnight, "c6", true},
		{BlackKnight, "f6", true},
		{WhitePawn, "e5", false}, // Should be empty
		{BlackPawn, "d5", false}, // Should be empty
	}

	for _, tc := range testCases {
		square := StringToSquare(tc.square)
		hasPiece := board.GetPieceBitboard(tc.piece).HasBit(square)
		if hasPiece != tc.expected {
			t.Errorf("Expected piece %c on %s to be %v, got %v",
				tc.piece, tc.square, tc.expected, hasPiece)
		}
	}
}

func TestBitboardConsistency(t *testing.T) {
	board := NewBoard()

	// Set up a random position
	moves := []struct {
		rank, file int
		piece      Piece
	}{
		{0, 0, WhiteRook},
		{0, 4, WhiteKing},
		{1, 0, WhitePawn},
		{1, 1, WhitePawn},
		{3, 3, WhiteQueen},
		{7, 0, BlackRook},
		{7, 4, BlackKing},
		{6, 0, BlackPawn},
		{6, 1, BlackPawn},
		{4, 4, BlackQueen},
	}

	for _, move := range moves {
		board.SetPiece(move.rank, move.file, move.piece)
	}

	// Verify that derived bitboards match the sum of piece bitboards
	var calculatedWhite, calculatedBlack Bitboard

	for i := WhitePawnIndex; i <= WhiteKingIndex; i++ {
		calculatedWhite |= board.PieceBitboards[i]
	}

	for i := BlackPawnIndex; i <= BlackKingIndex; i++ {
		calculatedBlack |= board.PieceBitboards[i]
	}

	if board.WhitePieces != calculatedWhite {
		t.Error("White pieces bitboard doesn't match sum of white piece bitboards")
	}

	if board.BlackPieces != calculatedBlack {
		t.Error("Black pieces bitboard doesn't match sum of black piece bitboards")
	}

	if board.AllPieces != (calculatedWhite | calculatedBlack) {
		t.Error("All pieces bitboard doesn't match union of color bitboards")
	}

	// Verify that array and bitboard representations are consistent
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			square := FileRankToSquare(file, rank)
			arrayPiece := board.GetPiece(rank, file)

			if arrayPiece == Empty {
				// Should not be set in any piece bitboard
				if board.AllPieces.HasBit(square) {
					t.Errorf("Square %s should be empty in bitboards but is occupied", SquareToString(square))
				}
			} else {
				// Should be set in the correct piece bitboard
				if !board.GetPieceBitboard(arrayPiece).HasBit(square) {
					t.Errorf("Square %s should have %c in bitboard", SquareToString(square), arrayPiece)
				}

				// Should be set in all pieces bitboard
				if !board.AllPieces.HasBit(square) {
					t.Errorf("Square %s should be set in all pieces bitboard", SquareToString(square))
				}
			}
		}
	}
}

func TestPieceToBitboardIndex(t *testing.T) {
	testCases := []struct {
		piece         Piece
		expectedIndex int
	}{
		{WhitePawn, WhitePawnIndex},
		{WhiteRook, WhiteRookIndex},
		{WhiteKnight, WhiteKnightIndex},
		{WhiteBishop, WhiteBishopIndex},
		{WhiteQueen, WhiteQueenIndex},
		{WhiteKing, WhiteKingIndex},
		{BlackPawn, BlackPawnIndex},
		{BlackRook, BlackRookIndex},
		{BlackKnight, BlackKnightIndex},
		{BlackBishop, BlackBishopIndex},
		{BlackQueen, BlackQueenIndex},
		{BlackKing, BlackKingIndex},
		{Empty, -1},
		{'x', -1}, // Invalid piece
	}

	for _, tc := range testCases {
		actualIndex := PieceToBitboardIndex(tc.piece)
		if actualIndex != tc.expectedIndex {
			t.Errorf("Expected index %d for piece %c, got %d", tc.expectedIndex, tc.piece, actualIndex)
		}
	}
}

// Benchmark tests for performance comparison
func BenchmarkSetPieceWithBitboards(b *testing.B) {
	board := NewBoard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rank := i % 8
		file := (i / 8) % 8
		piece := WhitePawn
		if i%2 == 0 {
			piece = BlackPawn
		}
		board.SetPiece(rank, file, piece)
	}
}

func BenchmarkGetPieceBitboard(b *testing.B) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to create board from FEN: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.GetPieceBitboard(WhitePawn)
	}
}

func BenchmarkGetColorBitboard(b *testing.B) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to create board from FEN: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = board.GetColorBitboard(BitboardWhite)
	}
}
