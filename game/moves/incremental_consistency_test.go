package moves

import (
	"math/rand"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestIncrementalEvalConsistency plays seeded-random legal games from varied
// positions and asserts after every make and unmake that the incrementally
// maintained material/PST scores match a from-scratch recomputation.
func TestIncrementalEvalConsistency(t *testing.T) {
	startFENs := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		// Sharp middlegame with castling available both sides
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		// Rook endgame
		"8/5pk1/8/3R4/6r1/8/5PK1/8 w - - 0 1",
		// Promotion race
		"8/P6k/8/8/8/8/p6K/8 w - - 0 1",
		// En passant heavy
		"rnbqkbnr/pp1ppppp/8/2pP4/8/8/PPP1PPPP/RNBQKBNR w KQkq c6 0 3",
	}

	generator := NewGenerator()
	rng := rand.New(rand.NewSource(20260703))

	for _, fen := range startFENs {
		b, err := board.FromFEN(fen)
		if err != nil {
			t.Fatalf("bad FEN %q: %v", fen, err)
		}

		player := White
		if b.GetSideToMove() == "b" {
			player = Black
		}

		for ply := 0; ply < 200; ply++ {
			legal := generator.GenerateAllMoves(b, player)
			if legal.Count == 0 {
				ReleaseMoveList(legal)
				break // mate or stalemate: game over
			}
			move := legal.Moves[rng.Intn(legal.Count)]
			ReleaseMoveList(legal)

			materialBefore := b.GetMaterialScore()
			pstBefore := b.GetPSTScore()

			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				t.Fatalf("%s ply %d: legal move failed to apply: %v", fen, ply, err)
			}

			assertScoresMatchRecompute(t, b, fen, ply)

			// Also verify unmake restores exactly, then re-make to continue
			b.UnmakeMove(undo)
			if b.GetMaterialScore() != materialBefore || b.GetPSTScore() != pstBefore {
				t.Fatalf("%s ply %d: unmake did not restore scores: material %d->%d, pst %d->%d",
					fen, ply, materialBefore, b.GetMaterialScore(), pstBefore, b.GetPSTScore())
			}
			if _, err := b.MakeMoveWithUndo(move); err != nil {
				t.Fatalf("%s ply %d: re-make failed: %v", fen, ply, err)
			}

			player = opposite(player)
		}
	}
}

func assertScoresMatchRecompute(t *testing.T, b *board.Board, fen string, ply int) {
	t.Helper()
	fresh, err := board.FromFEN(b.ToFEN())
	if err != nil {
		t.Fatalf("%s ply %d: ToFEN round-trip failed: %v", fen, ply, err)
	}
	fresh.InitializeEvalScoresFromPosition()

	if b.GetMaterialScore() != fresh.GetMaterialScore() {
		t.Fatalf("%s ply %d: incremental material %d != recomputed %d\n  position: %s",
			fen, ply, b.GetMaterialScore(), fresh.GetMaterialScore(), b.ToFEN())
	}
	if b.GetPSTScore() != fresh.GetPSTScore() {
		t.Fatalf("%s ply %d: incremental PST %d != recomputed %d\n  position: %s",
			fen, ply, b.GetPSTScore(), fresh.GetPSTScore(), b.ToFEN())
	}
}

func opposite(p Player) Player {
	if p == White {
		return Black
	}
	return White
}
