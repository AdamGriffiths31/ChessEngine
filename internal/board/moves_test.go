package board

import (
	"testing"
)

func TestMakeMoveFromEmptySquareFails(t *testing.T) {
	t.Parallel()
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	move := Move{From: Square{File: 4, Rank: 3}, To: Square{File: 4, Rank: 4}, Piece: Empty, Promotion: Empty}
	if err := board.MakeMove(move); err == nil {
		t.Error("MakeMove from an empty square succeeded, want an error")
	}
}

func TestUnmakeMove(t *testing.T) {
	t.Parallel()
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
			board, err := FromFEN(tt.setupFEN)
			if err != nil {
				t.Fatalf("Failed to setup board: %v", err)
			}

			initialFEN := board.ToFEN()

			undo, err := board.MakeMoveWithUndo(tt.move)
			if err != nil {
				t.Fatalf("Failed to make move: %v", err)
			}

			afterMoveFEN := board.ToFEN()
			if afterMoveFEN == initialFEN {
				t.Errorf("Board didn't change after move")
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

func TestUnmakeMoveWithMissingPiece(t *testing.T) {
	t.Parallel()
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	move := Move{
		From: Square{4, 1},
		To:   Square{4, 3},
		// Piece is intentionally not set
	}

	// MakeMoveWithUndo should fill in the piece
	undo, err := board.MakeMoveWithUndo(move)
	if err != nil {
		t.Fatalf("MakeMoveWithUndo failed: %v", err)
	}

	if undo.Move.Piece != WhitePawn {
		t.Errorf("MakeMoveWithUndo didn't set piece correctly: got %c", undo.Move.Piece)
	}

	board.UnmakeMove(undo)

	if board.GetPiece(1, 4) != WhitePawn {
		t.Error("Pawn not restored after unmake")
	}
}

func BenchmarkUnmakeMove(b *testing.B) {
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to create board from FEN: %v", err)
	}
	move := Move{From: Square{4, 1}, To: Square{4, 3}, Piece: WhitePawn}

	b.ResetTimer()
	for range b.N {
		undo, err := board.MakeMoveWithUndo(move)
		if err != nil {
			b.Fatalf("Failed to make move: %v", err)
		}
		board.UnmakeMove(undo)
	}
}

// A null move passes the turn: side flips, half-move clock ticks, en passant is
// cleared and castling rights are kept. Unmaking restores the exact position.
func TestNullMove(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fen  string
		want string // FEN after the null move
	}{
		{"white to move",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 1 1"},
		{"black to move bumps the full-move number",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 5 10",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 6 11"},
		{"en passant target is cleared",
			"rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
			"rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR b KQkq - 1 3"},
		{"castling rights are kept",
			"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 15 20",
			"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R b KQkq - 16 20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("FromFEN: %v", err)
			}

			undo := board.MakeNullMove()
			if got := board.ToFEN(); got != tt.want {
				t.Errorf("after null move:\n got %s\nwant %s", got, tt.want)
			}

			board.UnmakeNullMove(undo)
			if got := board.ToFEN(); got != tt.fen {
				t.Errorf("after unmake:\n got %s\nwant %s", got, tt.fen)
			}
		})
	}
}
