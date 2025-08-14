package moves

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestBitboardVsArrayMoveGeneration compares bitboard and array-based move generation
func TestBitboardVsArrayMoveGeneration(t *testing.T) {
	testPositions := []struct {
		name string
		fen  string
	}{
		{
			name: "starting_position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "kiwipete_position",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		},
		{
			name: "endgame_position",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		},
		{
			name: "promotion_position",
			fen:  "8/P1P5/K7/8/8/8/p1p5/k7 w - - 0 1",
		},
		{
			name: "en_passant_position",
			fen:  "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
		},
		{
			name: "castling_position",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
		},
	}

	for _, pos := range testPositions {
		t.Run(pos.name, func(t *testing.T) {
			b, err := board.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN %s: %v", pos.fen, err)
			}

			// Test both White and Black move generation
			for _, player := range []Player{White, Black} {
				t.Run(fmt.Sprintf("%s_to_move", player.String()), func(t *testing.T) {
					// Generate moves using existing generator
					generator := NewGenerator()
					arrayMoves := generator.GenerateAllMoves(b, player)

					// Generate moves using bitboard generator
					bitboardGenerator := NewBitboardMoveGenerator()
					bitboardMoves := bitboardGenerator.GenerateAllMovesBitboard(b, player)

					// Compare move counts
					if arrayMoves.Count != bitboardMoves.Count {
						t.Errorf("Move count mismatch: array=%d, bitboard=%d",
							arrayMoves.Count, bitboardMoves.Count)

						// Debug output
						t.Logf("Array moves (%d):", arrayMoves.Count)
						for i := 0; i < arrayMoves.Count; i++ {
							t.Logf("  %s", formatMove(arrayMoves.Moves[i]))
						}

						t.Logf("Bitboard moves (%d):", bitboardMoves.Count)
						for i := 0; i < bitboardMoves.Count; i++ {
							t.Logf("  %s", formatMove(bitboardMoves.Moves[i]))
						}
					}

					// Compare move sets
					arraySet := moveSetFromList(arrayMoves)
					bitboardSet := moveSetFromList(bitboardMoves)

					// Find moves in array but not in bitboard
					for move := range arraySet {
						if !bitboardSet[move] {
							t.Errorf("Move %s found in array generator but not in bitboard generator", move)
						}
					}

					// Find moves in bitboard but not in array
					for move := range bitboardSet {
						if !arraySet[move] {
							t.Errorf("Move %s found in bitboard generator but not in array generator", move)
						}
					}

					// Cleanup
					ReleaseMoveList(arrayMoves)
					ReleaseMoveList(bitboardMoves)
				})
			}
		})
	}
}

// TestBitboardPawnMoveGeneration specifically tests pawn move generation
func TestBitboardPawnMoveGeneration(t *testing.T) {
	testCases := []struct {
		name               string
		fen                string
		expectedWhiteMoves int
		expectedBlackMoves int
	}{
		{
			name:               "starting_position_pawns",
			fen:                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedWhiteMoves: 16, // 8 single pushes + 8 double pushes
			expectedBlackMoves: 16, // 8 single pushes + 8 double pushes
		},
		{
			name:               "pawn_captures",
			fen:                "8/8/8/3pP3/2PpP3/8/8/8 w - d6 0 1",
			expectedWhiteMoves: 5, // 2 pushes + 2 captures + 1 en passant
			expectedBlackMoves: 3, // 2 captures + 1 push
		},
		{
			name:               "promotion_pawns",
			fen:                "8/P1P5/8/8/8/8/p1p5/8 w - - 0 1",
			expectedWhiteMoves: 8, // 2 pawns × 4 promotions each
			expectedBlackMoves: 8, // 2 pawns × 4 promotions each
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			bitboardGen := NewBitboardMoveGenerator()

			// Test white pawns
			whiteMoves := GetMoveList()
			bitboardGen.generatePawnMovesBitboard(b, White, whiteMoves)
			if whiteMoves.Count != tc.expectedWhiteMoves {
				t.Errorf("White pawn moves: expected %d, got %d", tc.expectedWhiteMoves, whiteMoves.Count)
				for i := 0; i < whiteMoves.Count; i++ {
					t.Logf("  %s", formatMove(whiteMoves.Moves[i]))
				}
			}

			// Test black pawns
			blackMoves := GetMoveList()
			bitboardGen.generatePawnMovesBitboard(b, Black, blackMoves)
			if blackMoves.Count != tc.expectedBlackMoves {
				t.Errorf("Black pawn moves: expected %d, got %d", tc.expectedBlackMoves, blackMoves.Count)
				for i := 0; i < blackMoves.Count; i++ {
					t.Logf("  %s", formatMove(blackMoves.Moves[i]))
				}
			}

			ReleaseMoveList(whiteMoves)
			ReleaseMoveList(blackMoves)
		})
	}
}

// TestBitboardKnightMoveGeneration specifically tests knight move generation
func TestBitboardKnightMoveGeneration(t *testing.T) {
	testCases := []struct {
		name          string
		fen           string
		expectedMoves int
		player        Player
	}{
		{
			name:          "knight_center_board",
			fen:           "8/8/8/3N4/8/8/8/8 w - - 0 1",
			expectedMoves: 8, // Knight in center has 8 moves
			player:        White,
		},
		{
			name:          "knight_corner",
			fen:           "8/8/8/8/8/8/8/N7 w - - 0 1",
			expectedMoves: 2, // Knight in corner has 2 moves
			player:        White,
		},
		{
			name:          "knight_surrounded_by_pawns",
			fen:           "8/8/8/2PPP3/2PNP3/2PPP3/8/8 w - - 0 1",
			expectedMoves: 8, // Knight can jump over surrounding pawns to all 8 target squares
			player:        White,
		},
		{
			name:          "knight_actually_blocked",
			fen:           "8/8/2P1P3/1P3P2/3N4/1P3P2/2P1P3/8 w - - 0 1",
			expectedMoves: 0, // Knight target squares all occupied by friendly pieces
			player:        White,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			bitboardGen := NewBitboardMoveGenerator()
			moves := GetMoveList()
			bitboardGen.GenerateKnightMovesBitboard(b, tc.player, moves)

			if moves.Count != tc.expectedMoves {
				t.Errorf("Knight moves: expected %d, got %d", tc.expectedMoves, moves.Count)
				for i := 0; i < moves.Count; i++ {
					t.Logf("  %s", formatMove(moves.Moves[i]))
				}
			}

			ReleaseMoveList(moves)
		})
	}
}

// TestBitboardSlidingPieceGeneration tests rook, bishop, and queen move generation
func TestBitboardSlidingPieceGeneration(t *testing.T) {
	testCases := []struct {
		name          string
		fen           string
		piece         string
		expectedMoves int
		player        Player
	}{
		{
			name:          "rook_center_empty",
			fen:           "8/8/8/3R4/8/8/8/8 w - - 0 1",
			piece:         "rook",
			expectedMoves: 14, // 7 horizontal + 7 vertical
			player:        White,
		},
		{
			name:          "bishop_center_empty",
			fen:           "8/8/8/3B4/8/8/8/8 w - - 0 1",
			piece:         "bishop",
			expectedMoves: 13, // 7+4+2 = 13 diagonal moves
			player:        White,
		},
		{
			name:          "queen_center_empty",
			fen:           "8/8/8/3Q4/8/8/8/8 w - - 0 1",
			piece:         "queen",
			expectedMoves: 27, // 14 (rook) + 13 (bishop)
			player:        White,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			bitboardGen := NewBitboardMoveGenerator()
			moves := GetMoveList()

			switch tc.piece {
			case "rook":
				bitboardGen.generateRookMovesBitboard(b, tc.player, moves)
			case "bishop":
				bitboardGen.generateBishopMovesBitboard(b, tc.player, moves)
			case "queen":
				bitboardGen.generateQueenMovesBitboard(b, tc.player, moves)
			}

			if moves.Count != tc.expectedMoves {
				t.Errorf("%s moves: expected %d, got %d", tc.piece, tc.expectedMoves, moves.Count)
				for i := 0; i < moves.Count; i++ {
					t.Logf("  %s", formatMove(moves.Moves[i]))
				}
			}

			ReleaseMoveList(moves)
		})
	}
}

// TestBitboardCastlingGeneration tests castling move generation
func TestBitboardCastlingGeneration(t *testing.T) {
	testCases := []struct {
		name          string
		fen           string
		expectedMoves int
		player        Player
	}{
		{
			name:          "all_castling_available",
			fen:           "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			expectedMoves: 7, // 5 king moves + 2 castling
			player:        White,
		},
		{
			name:          "no_castling_rights",
			fen:           "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1",
			expectedMoves: 5, // 5 king moves only
			player:        White,
		},
		{
			name:          "kingside_only",
			fen:           "r3k2r/8/8/8/8/8/8/R3K2R w K - 0 1",
			expectedMoves: 6, // 5 king moves + 1 kingside castling
			player:        White,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			bitboardGen := NewBitboardMoveGenerator()
			moves := GetMoveList()
			bitboardGen.generateKingMovesBitboard(b, tc.player, moves)

			if moves.Count != tc.expectedMoves {
				t.Errorf("King moves: expected %d, got %d", tc.expectedMoves, moves.Count)
				for i := 0; i < moves.Count; i++ {
					t.Logf("  %s", formatMove(moves.Moves[i]))
				}
			}

			ReleaseMoveList(moves)
		})
	}
}

// Helper functions

// formatMove formats a move for display
func formatMove(move board.Move) string {
	from := fmt.Sprintf("%c%d", 'a'+move.From.File, move.From.Rank+1)
	to := fmt.Sprintf("%c%d", 'a'+move.To.File, move.To.Rank+1)

	result := from + to

	if move.Promotion != board.Empty {
		result += string(move.Promotion)
	}

	if move.IsCastling {
		result += " (castling)"
	}

	if move.IsEnPassant {
		result += " (en passant)"
	}

	if move.IsCapture {
		result += " (capture)"
	}

	return result
}

// moveSetFromList creates a set of move strings from a move list
func moveSetFromList(moves *MoveList) map[string]bool {
	set := make(map[string]bool)
	for i := 0; i < moves.Count; i++ {
		// Create a normalized string representation for comparison
		move := moves.Moves[i]
		key := fmt.Sprintf("%d%d%d%d%c",
			move.From.File, move.From.Rank,
			move.To.File, move.To.Rank,
			move.Promotion)
		set[key] = true
	}
	return set
}

// Benchmark tests to measure performance improvement

func BenchmarkBitboardMoveGeneration(b *testing.B) {
	positions := []struct {
		name string
		fen  string
	}{
		{"starting", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"},
		{"endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
	}

	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			board, err := board.FromFEN(pos.fen)
			if err != nil {
				b.Fatalf("Failed to parse FEN: %v", err)
			}

			bitboardGen := NewBitboardMoveGenerator()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				moves := bitboardGen.GenerateAllMovesBitboard(board, White)
				ReleaseMoveList(moves)
			}
		})
	}
}

func BenchmarkArrayMoveGeneration(b *testing.B) {
	positions := []struct {
		name string
		fen  string
	}{
		{"starting", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"},
		{"endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
	}

	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			board, err := board.FromFEN(pos.fen)
			if err != nil {
				b.Fatalf("Failed to parse FEN: %v", err)
			}

			generator := NewGenerator()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				moves := generator.GenerateAllMoves(board, White)
				ReleaseMoveList(moves)
			}
		})
	}
}

// BenchmarkMoveGenerationComparison compares bitboard vs array performance
func BenchmarkMoveGenerationComparison(b *testing.B) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	board, err := board.FromFEN(fen)
	if err != nil {
		b.Fatalf("Failed to parse FEN: %v", err)
	}

	b.Run("Array", func(b *testing.B) {
		generator := NewGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			moves := generator.GenerateAllMoves(board, White)
			ReleaseMoveList(moves)
		}
	})

	b.Run("Bitboard", func(b *testing.B) {
		bitboardGen := NewBitboardMoveGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			moves := bitboardGen.GenerateAllMovesBitboard(board, White)
			ReleaseMoveList(moves)
		}
	})
}
