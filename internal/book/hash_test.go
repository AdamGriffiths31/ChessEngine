package book

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// Tests in this package do not use t.Parallel(): GetPolyglotHash lazily
// initialises a package-level singleton without synchronisation, which
// `go test -race` flags as a data race when tests run concurrently.

// parseTestMove parses a plain four-character move such as "e2e4".
func parseTestMove(s string) (board.Move, error) {
	if len(s) != 4 {
		return board.Move{}, fmt.Errorf("unsupported move %q", s)
	}
	file := func(c byte) int { return int(c - 'a') }
	rank := func(c byte) int { return int(c - '1') }
	return board.Move{
		From:      board.Square{File: file(s[0]), Rank: rank(s[1])},
		To:        board.Square{File: file(s[2]), Rank: rank(s[3])},
		Promotion: board.Empty,
		Piece:     board.Empty,
	}, nil
}

// TestSpecificPositionHashes checks published Polyglot keys, both for the FEN
// directly and for the position reached by playing the listed moves from the
// start. The keys are the ground truth that opening books are looked up by.
func TestSpecificPositionHashes(t *testing.T) {
	zobrist := GetPolyglotHash()

	tests := []struct {
		name        string
		fen         string
		expectedKey uint64
		moves       []string
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
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create position from FEN: %v", err)
			}

			actualHash := zobrist.HashPosition(b)
			if actualHash != tt.expectedKey {
				t.Errorf("Hash mismatch for position from FEN\nExpected: 0x%016x\nGot:      0x%016x",
					tt.expectedKey, actualHash)
			}

			if len(tt.moves) > 0 {
				startBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
				if err != nil {
					t.Fatalf("Failed to create starting position: %v", err)
				}

				for i, moveStr := range tt.moves {
					move, err := parseTestMove(moveStr)
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

				finalFEN := startBoard.ToFEN()
				if finalFEN != tt.fen {
					t.Errorf("FEN mismatch after playing moves\nExpected: %s\nGot:      %s",
						tt.fen, finalFEN)
				}
			}
		})
	}
}
