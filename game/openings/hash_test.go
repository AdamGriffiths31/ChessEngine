package openings

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestSpecificPositionHashes tests that various chess positions produce the expected Zobrist hash values
func TestSpecificPositionHashes(t *testing.T) {
	zobrist := GetPolyglotHash()

	tests := []struct {
		name        string
		fen         string
		expectedKey uint64
		moves       []string // moves to reach this position from starting position
	}{
		{
			name:        "Starting position",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedKey: 0x463b96181691fc9c,
			moves:       []string{},
		},
		{
			name:        "After e2e4",
			fen:         "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			expectedKey: 0x823c9b50fd114196,
			moves:       []string{"e2e4"},
		},
		{
			name:        "After e2e4 d7d5",
			fen:         "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			expectedKey: 0x0756b94461c50fb0,
			moves:       []string{"e2e4", "d7d5"},
		},
		{
			name:        "After e2e4 d7d5 e4e5",
			fen:         "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2",
			expectedKey: 0x662fafb965db29d4,
			moves:       []string{"e2e4", "d7d5", "e4e5"},
		},
		{
			name:        "After e2e4 d7d5 e4e5 f7f5",
			fen:         "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
			expectedKey: 0x22a48b5a8e47ff78,
			moves:       []string{"e2e4", "d7d5", "e4e5", "f7f5"},
		},
		{
			name:        "After e2e4 d7d5 e4e5 f7f5 e1e2",
			fen:         "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b kq - 1 3",
			expectedKey: 0x652a607ca3f242c1,
			moves:       []string{"e2e4", "d7d5", "e4e5", "f7f5", "e1e2"},
		},
		{
			name:        "After e2e4 d7d5 e4e5 f7f5 e1e2 e8f7",
			fen:         "rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 2 4",
			expectedKey: 0x00fdd303c946bdd9,
			moves:       []string{"e2e4", "d7d5", "e4e5", "f7f5", "e1e2", "e8f7"},
		},
		{
			name:        "After a2a4 b7b5 h2h4 b5b4 c2c4",
			fen:         "rnbqkbnr/p1pppppp/8/8/PpP4P/8/1P1PPPP1/RNBQKBNR b KQkq c3 0 3",
			expectedKey: 0x3c8123ea7b067637,
			moves:       []string{"a2a4", "b7b5", "h2h4", "b5b4", "c2c4"},
		},
		{
			name:        "After a2a4 b7b5 h2h4 b5b4 c2c4 b4c3 a1a3",
			fen:         "rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 1 4",
			expectedKey: 0x5c3f9b829b279560,
			moves:       []string{"a2a4", "b7b5", "h2h4", "b5b4", "c2c4", "b4c3", "a1a3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test 1: Create position from FEN and check hash
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create position from FEN: %v", err)
			}

			actualHash := zobrist.HashPosition(b)
			if actualHash != tt.expectedKey {
				t.Errorf("Hash mismatch for position from FEN\nExpected: 0x%016x\nGot:      0x%016x",
					tt.expectedKey, actualHash)
			}

			// Test 2: Reach position by playing moves from starting position
			if len(tt.moves) > 0 {
				startBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
				if err != nil {
					t.Fatalf("Failed to create starting position: %v", err)
				}

				for i, moveStr := range tt.moves {
					move, err := board.ParseSimpleMove(moveStr)
					if err != nil {
						t.Fatalf("Failed to parse move %s at index %d: %v", moveStr, i, err)
					}

					err = startBoard.MakeMove(move)
					if err != nil {
						t.Fatalf("Failed to make move %s at index %d: %v", moveStr, i, err)
					}
				}

				hashAfterMoves := zobrist.HashPosition(startBoard)
				if hashAfterMoves != tt.expectedKey {
					t.Errorf("(%s vs %s)Hash mismatch after playing moves\nExpected: 0x%016x\nGot:      0x%016x",
						startBoard.ToFEN(), tt.fen, tt.expectedKey, hashAfterMoves)
				}

				// Verify the final FEN matches expected (optional but useful)
				finalFEN := startBoard.ToFEN()
				if finalFEN != tt.fen {
					t.Errorf("FEN mismatch after playing moves\nExpected: %s\nGot:      %s",
						tt.fen, finalFEN)
				}
			}
		})
	}
}

// TestIncrementalHashUpdate tests that hash is correctly updated incrementally
func TestIncrementalHashUpdate(t *testing.T) {
	zobrist := GetPolyglotHash()

	// Start from initial position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Verify starting hash
	startHash := zobrist.HashPosition(b)
	if startHash != 0x463b96181691fc9c {
		t.Errorf("Starting position hash mismatch\nExpected: 0x%016x\nGot:      0x%016x",
			0x463b96181691fc9c, startHash)
	}

	// Sequence of moves and expected hashes
	moveSequence := []struct {
		move        string
		expectedKey uint64
	}{
		{"e2e4", 0x823c9b50fd114196},
		{"d7d5", 0x0756b94461c50fb0},
		{"e4e5", 0x662fafb965db29d4},
		{"f7f5", 0x22a48b5a8e47ff78},
		{"e1e2", 0x652a607ca3f242c1},
		{"e8f7", 0x00fdd303c946bdd9},
	}

	for i, ms := range moveSequence {
		move, err := board.ParseSimpleMove(ms.move)
		if err != nil {
			t.Fatalf("Failed to parse move %s at step %d: %v", ms.move, i+1, err)
		}

		err = b.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %s at step %d: %v", ms.move, i+1, err)
		}

		actualHash := zobrist.HashPosition(b)
		if actualHash != ms.expectedKey {
			t.Errorf("Hash mismatch after move %s (step %d)\nExpected: 0x%016x\nGot:      0x%016x",
				ms.move, i+1, ms.expectedKey, actualHash)
		}
	}
}

// TestHashConsistency tests that the same position always produces the same hash
func TestHashConsistency(t *testing.T) {
	zobrist := GetPolyglotHash()

	testFENs := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		"rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 0 4",
		"rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 0 4",
	}

	for _, fen := range testFENs {
		// Create the same position multiple times
		hashes := make([]uint64, 5)
		for i := 0; i < 5; i++ {
			b, err := board.FromFEN(fen)
			if err != nil {
				t.Fatalf("Failed to create position from FEN: %v", err)
			}
			hashes[i] = zobrist.HashPosition(b)
		}

		// All hashes should be identical
		for i := 1; i < len(hashes); i++ {
			if hashes[i] != hashes[0] {
				t.Errorf("Inconsistent hash for position %s\nFirst:  0x%016x\nOther:  0x%016x (iteration %d)",
					fen, hashes[0], hashes[i], i+1)
			}
		}
	}
}


// TestCastlingRightsHashEffect tests that castling rights affect the hash
func TestCastlingRightsHashEffect(t *testing.T) {
	zobrist := GetPolyglotHash()

	// Positions that differ only in castling rights
	fenPairs := []struct {
		fen1 string
		fen2 string
		desc string
	}{
		{
			"rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b kq - 0 3",
			"rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b - - 0 3",
			"with and without black castling rights",
		},
		{
			"rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 0 4",
			"rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b kq - 0 4",
			"with and without white kingside castling",
		},
	}

	for _, pair := range fenPairs {
		b1, err := board.FromFEN(pair.fen1)
		if err != nil {
			t.Fatalf("Failed to create position 1 for %s: %v", pair.desc, err)
		}

		b2, err := board.FromFEN(pair.fen2)
		if err != nil {
			t.Fatalf("Failed to create position 2 for %s: %v", pair.desc, err)
		}

		hash1 := zobrist.HashPosition(b1)
		hash2 := zobrist.HashPosition(b2)

		if hash1 == hash2 {
			t.Errorf("Positions %s should have different hashes", pair.desc)
		}
	}
}

// TestSideToMoveHashEffect tests that side to move affects hash correctly
func TestSideToMoveHashEffect(t *testing.T) {
	zobrist := GetPolyglotHash()

	// Same position with different side to move
	b1, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create position with white to move: %v", err)
	}

	b2, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create position with black to move: %v", err)
	}

	hash1 := zobrist.HashPosition(b1)
	hash2 := zobrist.HashPosition(b2)

	if hash1 == hash2 {
		t.Error("Positions with different side to move should have different hashes")
	}
}
