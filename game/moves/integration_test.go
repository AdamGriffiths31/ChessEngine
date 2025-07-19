package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestGeneratorIntegration tests that the bitboard generator works correctly
func TestGeneratorIntegration(t *testing.T) {
	testPositions := []struct {
		name string
		fen  string
		expectedWhite int
		expectedBlack int
	}{
		{
			name: "starting_position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedWhite: 20, // 16 pawn moves + 4 knight moves
			expectedBlack: 20,
		},
		{
			name: "kiwipete_position", 
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			expectedWhite: 48,
			expectedBlack: 43,
		},
	}

	for _, pos := range testPositions {
		t.Run(pos.name, func(t *testing.T) {
			b, err := board.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN %s: %v", pos.fen, err)
			}

			gen := NewGenerator()
			defer gen.Release()

			// Test White moves
			whiteMoves := gen.GenerateAllMoves(b, White)
			if whiteMoves.Count != pos.expectedWhite {
				t.Errorf("White move count: expected %d, got %d", 
					pos.expectedWhite, whiteMoves.Count)
			}

			// Test Black moves  
			blackMoves := gen.GenerateAllMoves(b, Black)
			if blackMoves.Count != pos.expectedBlack {
				t.Errorf("Black move count: expected %d, got %d", 
					pos.expectedBlack, blackMoves.Count)
			}

			// Cleanup
			ReleaseMoveList(whiteMoves)
			ReleaseMoveList(blackMoves)
		})
	}
}


// BenchmarkBitboardGeneration benchmarks the bitboard generator
func BenchmarkBitboardGeneration(b *testing.B) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	board, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Failed to parse FEN: %v", err)
	}

	gen := NewGenerator()
	defer gen.Release()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		moves := gen.GenerateAllMoves(board, White)
		ReleaseMoveList(moves)
	}
}