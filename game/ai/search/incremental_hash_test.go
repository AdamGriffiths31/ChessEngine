package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

func TestIncrementalHashCorrectness(t *testing.T) {
	// Set up a board in the starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse starting position: %v", err)
	}

	// Set up zobrist hashing using MinimaxEngine as hash updater
	engine := NewMinimaxEngine()
	zobrist := openings.GetPolyglotHash()
	b.SetHashUpdater(engine)
	b.InitializeHashFromPosition(zobrist.HashPosition)

	// Get initial full hash
	initialHashFull := zobrist.HashPosition(b)
	initialHashIncremental := b.GetHash()

	if initialHashFull != initialHashIncremental {
		t.Errorf("Initial hash mismatch: full=%d, incremental=%d", initialHashFull, initialHashIncremental)
	}

	// Test various moves
	testMoves := []struct {
		name string
		move board.Move
	}{
		{
			name: "pawn move",
			move: board.Move{
				From:  board.Square{File: 4, Rank: 1}, // e2
				To:    board.Square{File: 4, Rank: 3}, // e4
				Piece: board.WhitePawn,
			},
		},
	}

	for _, test := range testMoves {
		t.Run(test.name, func(t *testing.T) {
			// Make move using incremental updates
			undo, err := b.MakeMoveWithUndo(test.move)
			if err != nil {
				t.Fatalf("Failed to make move: %v", err)
			}

			// Calculate hash using full recalculation
			fullHash := zobrist.HashPosition(b)
			incrementalHash := b.GetHash()

			if fullHash != incrementalHash {
				t.Errorf("Hash mismatch after %s: full=%d, incremental=%d", test.name, fullHash, incrementalHash)
			}

			// Unmake move
			b.UnmakeMove(undo)

			// Verify hash is restored correctly
			restoredHash := b.GetHash()
			expectedHash := zobrist.HashPosition(b)

			if restoredHash != expectedHash {
				t.Errorf("Hash mismatch after unmake %s: restored=%d, expected=%d", test.name, restoredHash, expectedHash)
			}

			if restoredHash != initialHashIncremental {
				t.Errorf("Hash not restored to initial after unmake %s: restored=%d, initial=%d", test.name, restoredHash, initialHashIncremental)
			}
		})
	}
}

func TestIncrementalHashComplexMoves(t *testing.T) {
	// Test with a more complex position
	b, err := board.FromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse test position: %v", err)
	}

	engine := NewMinimaxEngine()
	zobrist := openings.GetPolyglotHash()
	b.SetHashUpdater(engine)
	b.InitializeHashFromPosition(zobrist.HashPosition)

	// Test capture move
	captureMove := board.Move{
		From:      board.Square{File: 5, Rank: 2}, // f3
		To:        board.Square{File: 7, Rank: 2}, // h3
		Piece:     board.WhiteQueen,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	undo, err := b.MakeMoveWithUndo(captureMove)
	if err != nil {
		t.Fatalf("Failed to make capture move: %v", err)
	}

	fullHash := zobrist.HashPosition(b)
	incrementalHash := b.GetHash()

	if fullHash != incrementalHash {
		t.Errorf("Hash mismatch after capture: full=%d, incremental=%d", fullHash, incrementalHash)
	}

	b.UnmakeMove(undo)

	// Verify restoration
	restoredHash := b.GetHash()
	expectedHash := zobrist.HashPosition(b)

	if restoredHash != expectedHash {
		t.Errorf("Hash mismatch after unmake capture: restored=%d, expected=%d", restoredHash, expectedHash)
	}
}

func BenchmarkIncrementalHashUpdate(b *testing.B) {
	// Set up board
	testBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to parse starting position: %v", err)
	}

	engine := NewMinimaxEngine()
	testBoard.SetHashUpdater(engine)
	testBoard.InitializeHashFromPosition(openings.GetPolyglotHash().HashPosition)

	move := board.Move{
		From:  board.Square{File: 4, Rank: 1}, // e2
		To:    board.Square{File: 4, Rank: 3}, // e4
		Piece: board.WhitePawn,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		undo, err := testBoard.MakeMoveWithUndo(move)
		if err != nil {
			b.Fatalf("Failed to make move: %v", err)
		}
		_ = testBoard.GetHash() // Access the incremental hash
		testBoard.UnmakeMove(undo)
	}
}

func BenchmarkFullHashRecalculation(b *testing.B) {
	// Set up board
	testBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("Failed to parse starting position: %v", err)
	}

	zobrist := openings.GetPolyglotHash()
	move := board.Move{
		From:  board.Square{File: 4, Rank: 1}, // e2
		To:    board.Square{File: 4, Rank: 3}, // e4
		Piece: board.WhitePawn,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		undo, err := testBoard.MakeMoveWithUndo(move)
		if err != nil {
			b.Fatalf("Failed to make move: %v", err)
		}
		_ = zobrist.HashPosition(testBoard) // Full recalculation
		testBoard.UnmakeMove(undo)
	}
}
