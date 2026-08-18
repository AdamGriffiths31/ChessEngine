package board

import (
	"testing"
)

func TestFromFEN_ValidCases(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

// TestFromFEN_EnPassantField exercises FromFEN's 4th (en-passant) field in
// isolation, including the malformed inputs discovered by FuzzFromFEN/
// FuzzParseEPD (Task 7.6, batch C report): any en-passant token that is not
// "-" or a valid 2-char square with rank 3/6 must return an error, never
// panic.
func TestFromFEN_EnPassantField(t *testing.T) {
	t.Parallel()
	const boardPart = "8/8/8/8/8/8/8/8"
	testCases := []struct {
		name             string
		fen              string
		wantErr          bool
		wantHasEnPassant bool
		wantFile         int
		wantRank         int
	}{
		{name: "dash_no_en_passant", fen: boardPart + " w - - 0 1", wantErr: false, wantHasEnPassant: false},
		{name: "valid_e3", fen: boardPart + " w - e3 0 1", wantErr: false, wantHasEnPassant: true, wantFile: 4, wantRank: 2},
		{name: "valid_e6", fen: boardPart + " w - e6 0 1", wantErr: false, wantHasEnPassant: true, wantFile: 4, wantRank: 5},
		// Fuzz-found crasher (batch C report, 7.6): a 1-char en-passant
		// field indexes enPassantStr[1] out of range.
		{name: "one_char_crasher", fen: boardPart + " w - e 0 1", wantErr: true},
		// Empty field, e.g. from a doubled space collapsing parts[3] to "".
		{name: "empty_field", fen: boardPart + " w -  0 1", wantErr: true},
		{name: "three_char", fen: boardPart + " w - e34 0 1", wantErr: true},
		{name: "out_of_range_square", fen: boardPart + " w - z9 0 1", wantErr: true},
		{name: "valid_shape_wrong_rank", fen: boardPart + " w - e4 0 1", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := FromFEN(tc.fen)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("FromFEN(%q): expected error, got valid board", tc.fen)
				}
				if b != nil {
					t.Errorf("FromFEN(%q): expected nil board on error, got %+v", tc.fen, b)
				}
				return
			}
			if err != nil {
				t.Fatalf("FromFEN(%q): unexpected error: %v", tc.fen, err)
			}
			square, hasEnPassant := b.GetEnPassantTarget()
			if hasEnPassant != tc.wantHasEnPassant {
				t.Fatalf("FromFEN(%q): hasEnPassant = %v, want %v", tc.fen, hasEnPassant, tc.wantHasEnPassant)
			}
			if hasEnPassant {
				if square.File != tc.wantFile || square.Rank != tc.wantRank {
					t.Errorf("FromFEN(%q): en passant square = %+v, want {File:%d Rank:%d}", tc.fen, square, tc.wantFile, tc.wantRank)
				}
			}
		})
	}
}

func TestBoardGetSetPiece(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(0, 0, WhiteKing)
	piece := board.GetPiece(0, 0)
	if piece != WhiteKing {
		t.Errorf("Expected %c, got %c", WhiteKing, piece)
	}
}

func TestIsValidPiece(t *testing.T) {
	t.Parallel()
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

func TestBitboardSynchronization(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(0, 0, WhiteRook) // a1
	board.SetPiece(7, 7, BlackKing) // h8
	board.SetPiece(3, 4, WhitePawn) // e4

	if !board.GetPieceBitboard(WhiteRook).HasBit(FileRankToSquare(0, 0)) {
		t.Error("White rook bitboard should have a1 set")
	}
	if !board.GetPieceBitboard(BlackKing).HasBit(FileRankToSquare(7, 7)) {
		t.Error("Black king bitboard should have h8 set")
	}
	if !board.GetPieceBitboard(WhitePawn).HasBit(FileRankToSquare(4, 3)) {
		t.Error("White pawn bitboard should have e4 set")
	}

	whitePieces := board.GetColorBitboard(BitboardWhite)
	if !whitePieces.HasBit(FileRankToSquare(0, 0)) || !whitePieces.HasBit(FileRankToSquare(4, 3)) {
		t.Error("White pieces bitboard should include white rook and pawn")
	}

	blackPieces := board.GetColorBitboard(BitboardBlack)
	if !blackPieces.HasBit(FileRankToSquare(7, 7)) {
		t.Error("Black pieces bitboard should include black king")
	}

	if board.AllPieces.PopCount() != 3 {
		t.Errorf("All pieces bitboard should have 3 pieces, got %d", board.AllPieces.PopCount())
	}
}

func TestSetPieceUpdatesAllRepresentations(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(3, 4, WhiteQueen) // e4

	if board.GetPiece(3, 4) != WhiteQueen {
		t.Error("Array representation should have white queen on e4")
	}

	if !board.GetPieceBitboard(WhiteQueen).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Bitboard representation should have white queen on e4")
	}

	if board.getPieceCountFromBitboard(WhiteQueen) != 1 {
		t.Error("Should have exactly one white queen")
	}

	board.SetPiece(3, 4, BlackRook) // e4

	if board.GetPieceBitboard(WhiteQueen).HasBit(FileRankToSquare(4, 3)) {
		t.Error("White queen should be removed from bitboard")
	}

	if !board.GetPieceBitboard(BlackRook).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Black rook should be added to bitboard")
	}

	if board.GetPiece(3, 4) != BlackRook {
		t.Error("Array representation should have black rook on e4")
	}
}

func TestRemovePieceUpdatesAllRepresentations(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	board.SetPiece(3, 4, WhiteBishop) // e4
	board.SetPiece(3, 4, Empty)

	if board.GetPiece(3, 4) != Empty {
		t.Error("Array representation should be empty on e4")
	}

	if board.GetPieceBitboard(WhiteBishop).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Bitboard representation should not have white bishop on e4")
	}

	if board.AllPieces.HasBit(FileRankToSquare(4, 3)) {
		t.Error("All pieces bitboard should not have e4 set")
	}

	if board.getPieceCountFromBitboard(WhiteBishop) != 0 {
		t.Error("Should have no white bishops")
	}
}

func TestFENToBitboards(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse starting position FEN: %v", err)
	}

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

	if board.WhitePieces.PopCount() != 16 {
		t.Errorf("Expected 16 white pieces, got %d", board.WhitePieces.PopCount())
	}
	if board.BlackPieces.PopCount() != 16 {
		t.Errorf("Expected 16 black pieces, got %d", board.BlackPieces.PopCount())
	}
	if board.AllPieces.PopCount() != 32 {
		t.Errorf("Expected 32 total pieces, got %d", board.AllPieces.PopCount())
	}

	rank1and2 := RankMask(0) | RankMask(1)
	if (board.WhitePieces & rank1and2) != board.WhitePieces {
		t.Error("All white pieces should be on ranks 1-2")
	}

	rank7and8 := RankMask(6) | RankMask(7)
	if (board.BlackPieces & rank7and8) != board.BlackPieces {
		t.Error("All black pieces should be on ranks 7-8")
	}
}

func TestComplexPositionBitboards(t *testing.T) {
	t.Parallel()
	fen := "r2qkb1r/pb1p1ppp/1pn1pn2/8/2PP4/2N1PN2/PP3PPP/R1BQKB1R w KQkq - 0 8"
	board, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse complex position FEN: %v", err)
	}

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
	t.Parallel()
	board := NewBoard()

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
				if board.AllPieces.HasBit(square) {
					t.Errorf("Square %s should be empty in bitboards but is occupied", SquareToString(square))
				}
			} else {
				if !board.GetPieceBitboard(arrayPiece).HasBit(square) {
					t.Errorf("Square %s should have %c in bitboard", SquareToString(square), arrayPiece)
				}

				if !board.AllPieces.HasBit(square) {
					t.Errorf("Square %s should be set in all pieces bitboard", SquareToString(square))
				}
			}
		}
	}
}

func TestPieceToBitboardIndex(t *testing.T) {
	t.Parallel()
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

func TestIncrementalEvalInitialization(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		fen              string
		expectedMaterial int
		expectedPST      int
	}{
		{
			name:             "empty board",
			fen:              "8/8/8/8/8/8/8/8 w - - 0 1",
			expectedMaterial: 0,
			expectedPST:      0,
		},
		{
			name:             "starting position",
			fen:              "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedMaterial: 0, // Symmetric position
			expectedPST:      0, // Symmetric PST bonuses
		},
		{
			name:             "white advantage",
			fen:              "8/8/8/8/8/8/4P3/8 w - - 0 1", // White pawn on e2
			expectedMaterial: 100,
			expectedPST:      -20, // e2 pawn gets -20 from PST
		},
		{
			name:             "black advantage",
			fen:              "8/4p3/8/8/8/8/8/8 w - - 0 1", // Black pawn on e7
			expectedMaterial: -100,
			expectedPST:      20, // e7 black pawn gets bonus (flipped)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			board, err := FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			if board.GetMaterialScore() != tc.expectedMaterial {
				t.Errorf("Expected material score %d, got %d", tc.expectedMaterial, board.GetMaterialScore())
			}

			if board.GetPSTScore() != tc.expectedPST {
				t.Errorf("Expected PST score %d, got %d", tc.expectedPST, board.GetPSTScore())
			}
		})
	}
}

func TestIncrementalEvalSetPiece(t *testing.T) {
	t.Parallel()
	board := NewBoard()

	// Initially empty - scores should be 0
	if board.GetMaterialScore() != 0 || board.GetPSTScore() != 0 {
		t.Errorf("Empty board should have 0 scores, got material=%d pst=%d",
			board.GetMaterialScore(), board.GetPSTScore())
	}

	// Add a white knight to e4 (rank=3, file=4)
	board.SetPiece(3, 4, WhiteKnight)

	// Knight value is 320, e4 knight PST bonus is 20
	expectedMaterial := 320
	expectedPST := 20

	if board.GetMaterialScore() != expectedMaterial {
		t.Errorf("After adding knight, expected material %d, got %d",
			expectedMaterial, board.GetMaterialScore())
	}

	if board.GetPSTScore() != expectedPST {
		t.Errorf("After adding knight, expected PST %d, got %d",
			expectedPST, board.GetPSTScore())
	}

	// Replace with black queen
	board.SetPiece(3, 4, BlackQueen)

	// Black queen value is -900, e4 black queen PST is -5
	expectedMaterial = -900
	expectedPST = -5

	if board.GetMaterialScore() != expectedMaterial {
		t.Errorf("After replacing with queen, expected material %d, got %d",
			expectedMaterial, board.GetMaterialScore())
	}

	if board.GetPSTScore() != expectedPST {
		t.Errorf("After replacing with queen, expected PST %d, got %d",
			expectedPST, board.GetPSTScore())
	}

	// Remove piece
	board.SetPiece(3, 4, Empty)

	if board.GetMaterialScore() != 0 || board.GetPSTScore() != 0 {
		t.Errorf("After removing piece, should have 0 scores, got material=%d pst=%d",
			board.GetMaterialScore(), board.GetPSTScore())
	}
}

func TestIncrementalEvalMoveUnmake(t *testing.T) {
	t.Parallel()
	board, err := FromFEN("8/8/8/8/8/8/4P3/8 w - - 0 1") // White pawn on e2
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	initialMaterial := board.GetMaterialScore()
	initialPST := board.GetPSTScore()

	// Make a move: e2 to e4
	move := Move{
		From:  Square{Rank: 1, File: 4},
		To:    Square{Rank: 3, File: 4},
		Piece: WhitePawn,
	}

	undo, err := board.MakeMoveWithUndo(move)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	// After move, pawn is on e4 - different PST value
	// e2 has PST -20, e4 has PST 20
	expectedMaterialAfter := 100 // Still same material
	expectedPSTAfter := 20       // New PST value for e4

	if board.GetMaterialScore() != expectedMaterialAfter {
		t.Errorf("After move, expected material %d, got %d",
			expectedMaterialAfter, board.GetMaterialScore())
	}

	if board.GetPSTScore() != expectedPSTAfter {
		t.Errorf("After move, expected PST %d, got %d",
			expectedPSTAfter, board.GetPSTScore())
	}

	board.UnmakeMove(undo)

	if board.GetMaterialScore() != initialMaterial {
		t.Errorf("After unmake, expected material %d, got %d",
			initialMaterial, board.GetMaterialScore())
	}

	if board.GetPSTScore() != initialPST {
		t.Errorf("After unmake, expected PST %d, got %d",
			initialPST, board.GetPSTScore())
	}
}

func TestIncrementalEvalCaptures(t *testing.T) {
	t.Parallel()
	// Position with potential capture
	board, err := FromFEN("8/8/8/8/4p3/8/4P3/8 w - - 0 1") // White pawn on e2, black pawn on e4
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Verify initial state is balanced (100 - 100)
	if board.GetMaterialScore() != 0 {
		t.Errorf("Initial material should be 0, got %d", board.GetMaterialScore())
	}

	// White pawn captures black pawn (illegal in real chess, but tests the mechanics)
	// Manually simulate: remove black pawn, move white pawn
	board.SetPiece(3, 4, Empty)     // Remove black pawn from e4
	board.SetPiece(1, 4, Empty)     // Remove white pawn from e2
	board.SetPiece(3, 4, WhitePawn) // Place white pawn on e4

	expectedMaterial := 100
	expectedPST := 20 // e4 white pawn PST

	if board.GetMaterialScore() != expectedMaterial {
		t.Errorf("After capture, expected material %d, got %d",
			expectedMaterial, board.GetMaterialScore())
	}

	if board.GetPSTScore() != expectedPST {
		t.Errorf("After capture, expected PST %d, got %d",
			expectedPST, board.GetPSTScore())
	}
}

func TestIncrementalEvalComplexPosition(t *testing.T) {
	t.Parallel()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	board, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Calculate expected material manually by iterating
	expectedMaterial := 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := board.GetPiece(rank, file)
			if piece != Empty {
				pieceValue := 0
				switch piece {
				case WhitePawn:
					pieceValue = 100
				case WhiteKnight:
					pieceValue = 320
				case WhiteBishop:
					pieceValue = 330
				case WhiteRook:
					pieceValue = 500
				case WhiteQueen:
					pieceValue = 900
				case BlackPawn:
					pieceValue = -100
				case BlackKnight:
					pieceValue = -320
				case BlackBishop:
					pieceValue = -330
				case BlackRook:
					pieceValue = -500
				case BlackQueen:
					pieceValue = -900
				}
				expectedMaterial += pieceValue
			}
		}
	}

	if board.GetMaterialScore() != expectedMaterial {
		t.Errorf("Complex position material mismatch: expected %d, got %d",
			expectedMaterial, board.GetMaterialScore())
	}

	totalScore := board.GetMaterialScore() + board.GetPSTScore()
	if totalScore == 0 {
		t.Log("Warning: total score is 0 in asymmetric position")
	}
}
