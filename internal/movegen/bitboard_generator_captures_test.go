package movegen

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// TestGenerateCapturesMatchesFilteredPseudoLegal verifies that GenerateCaptures
// produces exactly the captures and promotions that quiescence search previously
// obtained by generating all pseudo-legal moves and filtering.
func TestGenerateCapturesMatchesFilteredPseudoLegal(t *testing.T) {
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
			name: "capture_promotion_position",
			fen:  "1n1q4/P1P5/K7/8/8/8/p1p5/1N1Q3k w - - 0 1",
		},
		{
			name: "en_passant_position",
			fen:  "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
		},
		{
			name: "black_en_passant_position",
			fen:  "rnbqkbnr/pppp1ppp/8/8/3pP3/8/PPP2PPP/RNBQKBNR b KQkq e3 0 3",
		},
		{
			name: "queen_heavy_middlegame",
			fen:  "r1bq1rk1/pp2nppp/2n1p3/2ppP3/3P4/2PB1N2/PP1N1PPP/R1BQ1RK1 w - - 0 1",
		},
	}

	for _, pos := range testPositions {
		t.Run(pos.name, func(t *testing.T) {
			b, err := board.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN %s: %v", pos.fen, err)
			}

			for _, player := range []Player{White, Black} {
				t.Run(fmt.Sprintf("%s_to_move", player.String()), func(t *testing.T) {
					generator := NewBitboardMoveGenerator()

					allMoves := generator.GeneratePseudoLegalMoves(b, player)
					defer ReleaseMoveList(allMoves)

					expected := make(map[string]bool)
					for i := 0; i < allMoves.Count; i++ {
						move := allMoves.Moves[i]
						if move.IsCapture || move.Promotion != board.Empty {
							expected[formatMove(move)] = true
						}
					}

					captures := generator.GenerateCaptures(b, player)
					defer ReleaseMoveList(captures)

					actual := make(map[string]bool)
					for i := 0; i < captures.Count; i++ {
						actual[formatMove(captures.Moves[i])] = true
					}

					if len(actual) != captures.Count {
						t.Errorf("GenerateCaptures produced duplicate moves: %d moves, %d unique",
							captures.Count, len(actual))
					}

					for move := range expected {
						if !actual[move] {
							t.Errorf("Move %s expected from GenerateCaptures but missing", move)
						}
					}

					for move := range actual {
						if !expected[move] {
							t.Errorf("Move %s from GenerateCaptures is not a pseudo-legal capture/promotion", move)
						}
					}
				})
			}
		})
	}
}
