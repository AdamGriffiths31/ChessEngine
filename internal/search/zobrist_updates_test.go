package search

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// zobristSeedBase + the FEN index seeds each playout, so failures reproduce.
const zobristSeedBase = 20240615

// zobristTestFENs covers the state that feeds the hash: full/partial/no
// castling rights, en passant for both colors, imminent promotions, and a
// sparse endgame. Other search tests reuse this list by index.
var zobristTestFENs = []string{
	// Standard starting position: full castling rights, no en passant.
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	// "Kiwipete": full castling rights, busy middlegame.
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	// Perft position 4: partial castling rights (black only, "kq").
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	// Partial castling rights (white kingside only, "K").
	"4k3/8/8/8/8/8/8/R3K2R w K - 0 1",
	// Perft "position 6": no castling rights, busy middlegame.
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	// White en passant capture available (d5 pawn can take e6).
	"4k3/8/8/3Pp3/8/8/8/4K3 w - e6 0 1",
	// Black en passant capture available (e4 pawn can take d3).
	"4k3/8/8/8/3Pp3/8/8/4K3 b - d3 0 1",
	// White promotion imminent (pawn on b7).
	"4k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
	// Black promotion imminent (pawn on b2).
	"4k3/8/8/8/8/8/1p6/4K3 b - - 0 1",
	// Sparse late endgame (king and pawn vs king).
	"8/8/4k3/8/8/4K3/4P3/8 w - - 0 1",
}

// maxZobristPlies caps the length of each random playout.
const maxZobristPlies = 40

// TestZobristIncrementalHashRoundTrip plays seeded random legal moves and
// checks at every ply that the incremental hash matches a from-scratch
// recompute, that unmake restores the pre-move hash, and that the null-move
// make/unmake round trip is exact.
func TestZobristIncrementalHashRoundTrip(t *testing.T) {
	for idx, fen := range zobristTestFENs {

		seed := int64(zobristSeedBase + idx)
		t.Run(fmt.Sprintf("fen%d", idx), func(t *testing.T) {
			runZobristPlayout(t, fen, seed)
		})
	}
}

func runZobristPlayout(t *testing.T, fen string, seed int64) {
	t.Helper()

	b := testutil.MustFromFEN(t, fen)
	m := NewMinimaxEngine()
	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	rng := rand.New(rand.NewSource(seed))
	player := playerFromSide(b.GetSideToMove())

	var playedMoves []string

	failf := func(format string, args ...interface{}) {
		t.Fatalf("fen=%q seed=%d moves=%v: %s", fen, seed, playedMoves, fmt.Sprintf(format, args...))
	}

	checkHash := func(label string) {
		want := book.GetPolyglotHash().HashPosition(b)
		got := b.GetHash()
		if got != want {
			failf("%s: incremental hash %#x != recomputed hash %#x", label, got, want)
		}
	}

	for ply := range maxZobristPlies {
		// Null-move round trip at this position.
		preNullHash := b.GetHash()
		nullUndo := b.MakeNullMove()
		if b.GetHash() == preNullHash {
			failf("ply %d: null move did not change hash", ply)
		}
		b.UnmakeNullMove(nullUndo)
		if b.GetHash() != preNullHash {
			failf("ply %d: hash after null-move unmake = %#x, want %#x", ply, b.GetHash(), preNullHash)
		}

		pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
		legalMoves := make([]board.Move, 0, pseudoMoves.Count)

		for i := range pseudoMoves.Count {
			move := pseudoMoves.Moves[i]

			preHash := b.GetHash()
			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				continue
			}

			if m.generator.IsKingInCheck(b, player) {
				b.UnmakeMove(undo)
				if b.GetHash() != preHash {
					failf("ply %d: hash after unmake of illegal move %s = %#x, want %#x", ply, moveString(move), b.GetHash(), preHash)
				}
				continue
			}

			// Legal candidate: make must give the right hash and unmake must restore it.
			checkHash(fmt.Sprintf("ply %d candidate %s after make", ply, moveString(move)))

			b.UnmakeMove(undo)
			if b.GetHash() != preHash {
				failf("ply %d: hash after unmake of legal move %s = %#x, want %#x", ply, moveString(move), b.GetHash(), preHash)
			}

			legalMoves = append(legalMoves, move)
		}
		movegen.ReleaseMoveList(pseudoMoves)

		if len(legalMoves) == 0 {
			break
		}

		chosen := legalMoves[rng.Intn(len(legalMoves))]
		preMoveHash := b.GetHash()

		undo, err := b.MakeMoveWithUndo(chosen)
		if err != nil {
			failf("ply %d: chosen move %s unexpectedly failed to make: %v", ply, moveString(chosen), err)
		}
		playedMoves = append(playedMoves, moveString(chosen))
		checkHash(fmt.Sprintf("ply %d chosen move %s after make", ply, moveString(chosen)))

		// Unmake and re-make the chosen move to continue the playout.
		b.UnmakeMove(undo)
		if b.GetHash() != preMoveHash {
			failf("ply %d: hash after unmake of chosen move %s = %#x, want %#x", ply, moveString(chosen), b.GetHash(), preMoveHash)
		}

		if _, err = b.MakeMoveWithUndo(chosen); err != nil {
			failf("ply %d: chosen move %s unexpectedly failed to re-make: %v", ply, moveString(chosen), err)
		}
		checkHash(fmt.Sprintf("ply %d chosen move %s after re-make", ply, moveString(chosen)))

		player = opposite(player)
	}
}
