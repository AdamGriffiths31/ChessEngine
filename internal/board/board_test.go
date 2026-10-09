package board

import (
	"testing"
)

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

// The en-passant field must be "-" or a square on rank 3 or 6; anything else is
// an error, never a panic.
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
		// Regression: a 1-char field once indexed out of range.
		{name: "one_char_crasher", fen: boardPart + " w - e 0 1", wantErr: true},
		// Doubled space leaves an empty field.
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

func TestPieceToBitboardIndexRejectsNonPieces(t *testing.T) {
	t.Parallel()
	for _, piece := range []Piece{Empty, 'x', ' '} {
		if got := PieceToBitboardIndex(piece); got != -1 {
			t.Errorf("PieceToBitboardIndex(%q) = %d, want -1", piece, got)
		}
	}
}

// SetPiece keeps the mailbox and every bitboard in step: place, replace, clear.
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

	board.SetPiece(3, 4, Empty)

	if board.GetPiece(3, 4) != Empty || board.GetPieceBitboard(BlackRook).HasBit(FileRankToSquare(4, 3)) {
		t.Error("Emptying a square should clear both the array and the bitboard")
	}
	if board.AllPieces.HasBit(FileRankToSquare(4, 3)) {
		t.Error("All pieces bitboard should not have e4 set")
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
	for rank := range 8 {
		for file := range 8 {
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

func BenchmarkSetPieceWithBitboards(b *testing.B) {
	board := NewBoard()
	b.ResetTimer()
	for i := range b.N {
		rank := i % 8
		file := (i / 8) % 8
		piece := WhitePawn
		if i%2 == 0 {
			piece = BlackPawn
		}
		board.SetPiece(rank, file, piece)
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
