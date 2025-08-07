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

func TestUnmakeMove(t *testing.T) {
	tests := []struct {
		name        string
		setupFEN    string
		move        Move
		expectedFEN string
		description string
	}{
		{
			name:        "Simple pawn move",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Moving and unmaking e2e4 should restore starting position",
		},
		{
			name:        "Capture move",
			setupFEN:    "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			move:        Move{From: Square{4, 3}, To: Square{3, 4}, Piece: WhitePawn, Captured: BlackPawn, IsCapture: true},
			expectedFEN: "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			description: "Capturing and unmaking exd5 should restore both pawns",
		},
		{
			name:        "White kingside castling",
			setupFEN:    "rnbqk2r/pppp1ppp/4pn2/8/1b6/4PN2/PPPPBPPP/RNBQK2R w KQkq - 2 5",
			move:        Move{From: Square{4, 0}, To: Square{6, 0}, Piece: WhiteKing, IsCastling: true},
			expectedFEN: "rnbqk2r/pppp1ppp/4pn2/8/1b6/4PN2/PPPPBPPP/RNBQK2R w KQkq - 2 5",
			description: "Castling and unmaking should restore king and rook",
		},
		{
			name:        "White queenside castling",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R3KBNR w KQkq - 0 1",
			move:        Move{From: Square{4, 0}, To: Square{2, 0}, Piece: WhiteKing, IsCastling: true},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R3KBNR w KQkq - 0 1",
			description: "Queenside castling and unmaking should restore positions",
		},
		{
			name:        "Black kingside castling",
			setupFEN:    "rnbqk2r/pppppppp/5n2/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 4 3",
			move:        Move{From: Square{4, 7}, To: Square{6, 7}, Piece: BlackKing, IsCastling: true},
			expectedFEN: "rnbqk2r/pppppppp/5n2/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 4 3",
			description: "Black castling and unmaking should work correctly",
		},
		{
			name:        "White en passant capture",
			setupFEN:    "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
			move:        Move{From: Square{4, 4}, To: Square{5, 5}, Piece: WhitePawn, Captured: BlackPawn, IsEnPassant: true},
			expectedFEN: "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
			description: "En passant capture and unmake should restore captured pawn",
		},
		{
			name:        "Black en passant capture",
			setupFEN:    "rnbqkbnr/pppp1ppp/8/8/3pP3/8/PPP2PPP/RNBQKBNR b KQkq e3 0 2",
			move:        Move{From: Square{3, 3}, To: Square{4, 2}, Piece: BlackPawn, Captured: WhitePawn, IsEnPassant: true},
			expectedFEN: "rnbqkbnr/pppp1ppp/8/8/3pP3/8/PPP2PPP/RNBQKBNR b KQkq e3 0 2",
			description: "Black en passant and unmake",
		},
		{
			name:        "White pawn promotion to queen",
			setupFEN:    "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			move:        Move{From: Square{7, 6}, To: Square{7, 7}, Piece: WhitePawn, Promotion: WhiteQueen},
			expectedFEN: "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			description: "Promotion and unmake should restore pawn",
		},
		{
			name:        "Black pawn promotion with capture",
			setupFEN:    "rnbqkbnr/pppppp1p/8/8/8/8/PPPPPPp1/RNBQKBNR b KQkq - 0 1",
			move:        Move{From: Square{6, 1}, To: Square{7, 0}, Piece: BlackPawn, Captured: WhiteRook, Promotion: BlackQueen, IsCapture: true},
			expectedFEN: "rnbqkbnr/pppppp1p/8/8/8/8/PPPPPPp1/RNBQKBNR b KQkq - 0 1",
			description: "Promotion with capture should restore both pieces",
		},
		{
			name:        "White pawn promotion to knight",
			setupFEN:    "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			move:        Move{From: Square{7, 6}, To: Square{7, 7}, Piece: WhitePawn, Promotion: WhiteKnight},
			expectedFEN: "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			description: "Promotion to knight and unmake should restore pawn",
		},
		{
			name:        "Black pawn promotion to rook",
			setupFEN:    "rnbqkbnr/pppppp1p/8/8/8/8/PPPPPPp1/RNBQKBNR b KQkq - 0 1",
			move:        Move{From: Square{6, 1}, To: Square{6, 0}, Piece: BlackPawn, Promotion: BlackRook},
			expectedFEN: "rnbqkbnr/pppppp1p/8/8/8/8/PPPPPPp1/RNBQKBNR b KQkq - 0 1",
			description: "Promotion to rook and unmake should restore pawn",
		},
		{
			name:        "White pawn promotion to bishop with capture",
			setupFEN:    "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			move:        Move{From: Square{7, 6}, To: Square{6, 7}, Piece: WhitePawn, Captured: BlackKnight, Promotion: WhiteBishop, IsCapture: true},
			expectedFEN: "rnbqkbn1/pppppppP/8/8/8/8/PPPPPP1P/RNBQKBNR w KQq - 0 1",
			description: "Promotion to bishop with capture should restore both pieces",
		},
		{
			name:        "Castling rights lost when queenside rook captured",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{3, 0}, To: Square{0, 7}, Piece: WhiteQueen, Captured: BlackRook, IsCapture: true},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Capturing rook should affect castling rights, then restore them",
		},
		{
			name:        "Castling rights lost when kingside rook captured",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{3, 0}, To: Square{7, 7}, Piece: WhiteQueen, Captured: BlackRook, IsCapture: true},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Capturing kingside rook should affect castling rights, then restore them",
		},
		{
			name:        "Double pawn push sets en passant target",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Double pawn push should set en passant target, then restore to none",
		},
		{
			name:        "Black double pawn push sets en passant target",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
			move:        Move{From: Square{3, 6}, To: Square{3, 4}, Piece: BlackPawn},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
			description: "Black double pawn push should set en passant target, then restore to none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup board
			board, err := FromFEN(tt.setupFEN)
			if err != nil {
				t.Fatalf("Failed to setup board: %v", err)
			}

			// Store initial state for comparison
			initialFEN := board.ToFEN()

			// Make the move
			undo, err := board.MakeMoveWithUndo(tt.move)
			if err != nil {
				t.Fatalf("Failed to make move: %v", err)
			}

			// Verify move was made (board changed)
			afterMoveFEN := board.ToFEN()
			if afterMoveFEN == initialFEN {
				t.Errorf("Board didn't change after move")
			}

			// Unmake the move
			board.UnmakeMove(undo)

			// Get final FEN
			finalFEN := board.ToFEN()

			// Compare with expected
			if finalFEN != tt.expectedFEN {
				t.Errorf("%s\nExpected: %s\nGot:      %s", 
					tt.description, tt.expectedFEN, finalFEN)
			}
		})
	}
}

func TestUnmakeMoveStateRestoration(t *testing.T) {
	// Test that all board state is properly restored
	board, _ := FromFEN("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 5 10")
	
	// Store original state
	origCastling := board.castlingRights
	origEnPassant := board.enPassantTarget
	origHalfMove := board.halfMoveClock
	origFullMove := board.fullMoveNumber
	origSideToMove := board.sideToMove

	// Make a move that affects state
	move := Move{
		From: Square{4, 0}, 
		To: Square{6, 0}, 
		Piece: WhiteKing, 
		IsCastling: true,
	}
	
	undo, _ := board.MakeMoveWithUndo(move)
	
	// State should have changed
	if board.castlingRights == origCastling {
		t.Error("Castling rights didn't change after king move")
	}
	if board.sideToMove == origSideToMove {
		t.Error("Side to move didn't change")
	}
	
	// Unmake the move
	board.UnmakeMove(undo)
	
	// All state should be restored
	if board.castlingRights != origCastling {
		t.Errorf("Castling rights not restored: expected %s, got %s", 
			origCastling, board.castlingRights)
	}
	if board.halfMoveClock != origHalfMove {
		t.Errorf("Half move clock not restored: expected %d, got %d", 
			origHalfMove, board.halfMoveClock)
	}
	if board.fullMoveNumber != origFullMove {
		t.Errorf("Full move number not restored: expected %d, got %d", 
			origFullMove, board.fullMoveNumber)
	}
	if board.sideToMove != origSideToMove {
		t.Errorf("Side to move not restored: expected %s, got %s", 
			origSideToMove, board.sideToMove)
	}
	if !equalEnPassant(board.enPassantTarget, origEnPassant) {
		t.Error("En passant target not restored")
	}
}

func TestUnmakeMoveWithMissingPiece(t *testing.T) {
	// Test error handling when move.Piece is not set
	board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Make move without setting Piece field
	move := Move{
		From: Square{4, 1}, 
		To: Square{4, 3},
		// Piece is intentionally not set
	}
	
	// MakeMoveWithUndo should fill in the piece
	undo, err := board.MakeMoveWithUndo(move)
	if err != nil {
		t.Fatalf("MakeMoveWithUndo failed: %v", err)
	}
	
	// Verify piece was set in undo
	if undo.Move.Piece != WhitePawn {
		t.Errorf("MakeMoveWithUndo didn't set piece correctly: got %c", undo.Move.Piece)
	}
	
	// Unmake should work correctly
	board.UnmakeMove(undo)
	
	// Verify pawn is back
	if board.GetPiece(1, 4) != WhitePawn {
		t.Error("Pawn not restored after unmake")
	}
}

func TestUnmakeMoveConsistencyCheck(t *testing.T) {
	// Test multiple make/unmake cycles maintain consistency
	board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	originalFEN := board.ToFEN()
	
	moves := []Move{
		{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn}, // e4
		{From: Square{4, 6}, To: Square{4, 4}, Piece: BlackPawn}, // e5
		{From: Square{6, 0}, To: Square{5, 2}, Piece: WhiteKnight}, // Nf3
		{From: Square{1, 7}, To: Square{2, 5}, Piece: BlackKnight}, // Nc6
	}
	
	// Make and unmake each move
	for i, move := range moves {
		undo, err := board.MakeMoveWithUndo(move)
		if err != nil {
			t.Fatalf("Move %d failed: %v", i, err)
		}
		
		board.UnmakeMove(undo)
		
		currentFEN := board.ToFEN()
		if currentFEN != originalFEN {
			t.Errorf("Board state corrupted after move %d make/unmake cycle\nExpected: %s\nGot: %s", 
				i, originalFEN, currentFEN)
		}
	}
}

func TestUnmakeMoveComplexPosition(t *testing.T) {
	// Test from the actual problematic position in the bug
	board, _ := FromFEN("rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14")
	originalFEN := board.ToFEN()
	
	// Try the illegal move that was causing issues
	move := Move{
		From: Square{3, 3}, // d4
		To: Square{4, 2},   // e3
		Piece: WhiteQueen,
		Captured: Empty, // No piece on e3
	}
	
	undo, err := board.MakeMoveWithUndo(move)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}
	
	// Unmake the move
	board.UnmakeMove(undo)
	
	// Verify board is restored exactly
	finalFEN := board.ToFEN()
	if finalFEN != originalFEN {
		t.Errorf("Board not restored correctly\nExpected: %s\nGot:      %s", 
			originalFEN, finalFEN)
	}
}

// Helper function to compare en passant targets
func equalEnPassant(a, b *Square) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.File == b.File && a.Rank == b.Rank
}

func TestUnmakeMoveSequence(t *testing.T) {
	// Test making/unmaking multiple special moves in sequence
	board, _ := FromFEN("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1")
	originalFEN := board.ToFEN()
	
	// Sequence: simple moves to test multiple make/unmake cycles
	moves := []Move{
		// White castles kingside
		{From: Square{4, 0}, To: Square{6, 0}, Piece: WhiteKing, IsCastling: true},
		// Black moves pawn
		{From: Square{4, 6}, To: Square{4, 4}, Piece: BlackPawn},
		// White moves pawn
		{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn},
	}
	
	var undos []MoveUndo
	
	// Make all moves
	for i, move := range moves {
		undo, err := board.MakeMoveWithUndo(move)
		if err != nil {
			t.Fatalf("Move %d failed: %v", i, err)
		}
		undos = append(undos, undo)
	}
	
	// Unmake all moves in reverse order
	for i := len(undos) - 1; i >= 0; i-- {
		board.UnmakeMove(undos[i])
	}
	
	// Should be back to original position
	finalFEN := board.ToFEN()
	if finalFEN != originalFEN {
		t.Errorf("Sequence make/unmake failed\nExpected: %s\nGot:      %s", 
			originalFEN, finalFEN)
	}
}

func TestUnmakeMoveEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFEN    string
		move        Move
		expectedFEN string
		description string
	}{
		{
			name:        "King in check position",
			setupFEN:    "rnb1kbnr/pppp1ppp/8/4p3/6Pq/8/PPPPP1PP/RNBQKBNR w KQkq - 1 3",
			move:        Move{From: Square{6, 1}, To: Square{6, 2}, Piece: WhitePawn},
			expectedFEN: "rnb1kbnr/pppp1ppp/8/4p3/6Pq/8/PPPPP1PP/RNBQKBNR w KQkq - 1 3",
			description: "Move in check position should restore correctly",
		},
		{
			name:        "Piece on corner square",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{0, 0}, To: Square{0, 3}, Piece: WhiteRook},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Moving corner rook should restore properly",
		},
		{
			name:        "Maximum material position",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{1, 0}, To: Square{2, 2}, Piece: WhiteKnight},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Full material position should restore correctly",
		},
		{
			name:        "Minimal material position", 
			setupFEN:    "8/8/8/8/8/8/8/K6k w - - 0 1",
			move:        Move{From: Square{0, 0}, To: Square{1, 0}, Piece: WhiteKing},
			expectedFEN: "8/8/8/8/8/8/8/K6k w - - 0 1",
			description: "King-only endgame should restore correctly",
		},
		{
			name:        "Piece on edge of board",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:        Move{From: Square{7, 1}, To: Square{7, 3}, Piece: WhitePawn},
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Edge pawn move should restore correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := FromFEN(tt.setupFEN)
			if err != nil {
				t.Fatalf("Failed to setup board: %v", err)
			}

			undo, err := board.MakeMoveWithUndo(tt.move)
			if err != nil {
				t.Fatalf("Failed to make move: %v", err)
			}

			board.UnmakeMove(undo)

			finalFEN := board.ToFEN()
			if finalFEN != tt.expectedFEN {
				t.Errorf("%s\nExpected: %s\nGot:      %s", 
					tt.description, tt.expectedFEN, finalFEN)
			}
		})
	}
}

// Benchmark to ensure performance isn't degraded
func BenchmarkUnmakeMove(b *testing.B) {
	board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	move := Move{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		undo, _ := board.MakeMoveWithUndo(move)
		board.UnmakeMove(undo)
	}
}

func TestNullMove(t *testing.T) {
	// Test null move make/unmake with various positions
	tests := []struct {
		name        string
		setupFEN    string
		description string
	}{
		{
			name:        "Starting position white to move",
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			description: "Null move from starting position should switch to black",
		},
		{
			name:        "Starting position black to move", 
			setupFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 5 10",
			description: "Null move from black should switch to white and increment full moves",
		},
		{
			name:        "Position with en passant target",
			setupFEN:    "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
			description: "Null move should clear en passant target",
		},
		{
			name:        "Position with all castling rights",
			setupFEN:    "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 15 20",
			description: "Null move should preserve castling rights",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			board, err := FromFEN(test.setupFEN)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			// Store original state
			origSideToMove := board.sideToMove
			origCastling := board.castlingRights
			origEnPassant := board.enPassantTarget
			origHalfMove := board.halfMoveClock
			origFullMove := board.fullMoveNumber

			// Make null move
			undo := board.MakeNullMove()

			// Verify the null move changed what it should
			if origSideToMove == "w" {
				if board.sideToMove != "b" {
					t.Errorf("Expected side to move to change from w to b")
				}
				if board.fullMoveNumber != origFullMove {
					t.Errorf("Expected full move number to stay %d when white makes null move, got %d", 
						origFullMove, board.fullMoveNumber)
				}
			} else {
				if board.sideToMove != "w" {
					t.Errorf("Expected side to move to change from b to w")
				}
				if board.fullMoveNumber != origFullMove+1 {
					t.Errorf("Expected full move number to increment from %d to %d when black makes null move, got %d", 
						origFullMove, origFullMove+1, board.fullMoveNumber)
				}
			}

			// Half move clock should increment
			if board.halfMoveClock != origHalfMove+1 {
				t.Errorf("Expected half move clock to increment from %d to %d, got %d", 
					origHalfMove, origHalfMove+1, board.halfMoveClock)
			}

			// En passant target should be cleared
			if board.enPassantTarget != nil {
				t.Errorf("Expected en passant target to be cleared, got %v", board.enPassantTarget)
			}

			// Castling rights should be preserved
			if board.castlingRights != origCastling {
				t.Errorf("Expected castling rights to be preserved as %q, got %q", 
					origCastling, board.castlingRights)
			}

			// Unmake null move
			board.UnmakeNullMove(undo)

			// Verify everything is restored
			if board.sideToMove != origSideToMove {
				t.Errorf("Expected side to move to be restored to %q, got %q", 
					origSideToMove, board.sideToMove)
			}
			if board.castlingRights != origCastling {
				t.Errorf("Expected castling rights to be restored to %q, got %q", 
					origCastling, board.castlingRights)
			}
			if (board.enPassantTarget == nil) != (origEnPassant == nil) {
				t.Errorf("Expected en passant target restoration mismatch")
			}
			if board.enPassantTarget != nil && origEnPassant != nil {
				if *board.enPassantTarget != *origEnPassant {
					t.Errorf("Expected en passant target to be restored to %v, got %v", 
						*origEnPassant, *board.enPassantTarget)
				}
			}
			if board.halfMoveClock != origHalfMove {
				t.Errorf("Expected half move clock to be restored to %d, got %d", 
					origHalfMove, board.halfMoveClock)
			}
			if board.fullMoveNumber != origFullMove {
				t.Errorf("Expected full move number to be restored to %d, got %d", 
					origFullMove, board.fullMoveNumber)
			}

			// Verify board position unchanged (all pieces should be in same place)
			finalFEN := board.ToFEN()
			if finalFEN != test.setupFEN {
				t.Errorf("Expected board to be fully restored to original FEN %q, got %q", 
					test.setupFEN, finalFEN)
			}
		})
	}
}

func TestNullMoveConstant(t *testing.T) {
	// Test that the NullMove constant has expected properties
	if NullMove.From.File != -1 || NullMove.From.Rank != -1 {
		t.Errorf("Expected NullMove.From to be (-1, -1), got (%d, %d)", 
			NullMove.From.File, NullMove.From.Rank)
	}
	if NullMove.To.File != -1 || NullMove.To.Rank != -1 {
		t.Errorf("Expected NullMove.To to be (-1, -1), got (%d, %d)", 
			NullMove.To.File, NullMove.To.Rank)
	}
	if NullMove.Piece != Empty {
		t.Errorf("Expected NullMove.Piece to be Empty, got %v", NullMove.Piece)
	}
	if NullMove.IsCapture || NullMove.IsCastling || NullMove.IsEnPassant {
		t.Errorf("Expected all NullMove flags to be false")
	}
}
