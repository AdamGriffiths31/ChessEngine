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

// zobristSeedBase is the fixed base seed used to derive a per-FEN RNG seed
// (base + index into zobristTestFENs). Keeping it fixed makes any failure
// reproducible by re-running the test with the same index.
const zobristSeedBase = 20240615

// zobristTestFENs is a diverse set of legal positions exercising the pieces
// of board state that feed the Zobrist hash: full/partial/no castling
// rights, en passant availability for both colors, imminent promotions for
// both colors, and a sparse late endgame.
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

// TestZobristIncrementalHashRoundTrip plays a bounded sequence of random
// legal moves from a diverse set of positions and checks, at every ply,
// that the incrementally maintained Zobrist hash (board.Board.GetHash,
// updated via MinimaxEngine.GetHashDelta) matches a from-scratch recompute
// (book.ZobristHash.HashPosition), and that unmaking a move restores the
// exact pre-move hash. It also checks the null-move make/unmake round trip
// at every position visited.
func TestZobristIncrementalHashRoundTrip(t *testing.T) {
	for idx, fen := range zobristTestFENs {
		fen := fen
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

	for ply := 0; ply < maxZobristPlies; ply++ {
		// Null-move round trip check at the current position, before any
		// real move is made this ply.
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

		for i := 0; i < pseudoMoves.Count; i++ {
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

			// Legal candidate: verify make produced the right hash, then
			// verify unmake restores it, before deciding whether to play it.
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

		// Exercise the unmake path for the chosen move too, then re-make it
		// so the playout continues from the resulting position.
		b.UnmakeMove(undo)
		if b.GetHash() != preMoveHash {
			failf("ply %d: hash after unmake of chosen move %s = %#x, want %#x", ply, moveString(chosen), b.GetHash(), preMoveHash)
		}

		undo, err = b.MakeMoveWithUndo(chosen)
		if err != nil {
			failf("ply %d: chosen move %s unexpectedly failed to re-make: %v", ply, moveString(chosen), err)
		}
		_ = undo
		checkHash(fmt.Sprintf("ply %d chosen move %s after re-make", ply, moveString(chosen)))

		player = opposite(player)
	}
}

// playerFromSide converts the board's side-to-move string ("w"/"b") into a
// movegen.Player.
func playerFromSide(side string) movegen.Player {
	if side == "w" {
		return movegen.White
	}
	return movegen.Black
}

// opposite returns the other player.
func opposite(p movegen.Player) movegen.Player {
	if p == movegen.White {
		return movegen.Black
	}
	return movegen.White
}

// moveString renders a move as coordinate notation (e.g. "e2e4", "a7a8q")
// for failure messages.
func moveString(move board.Move) string {
	s := move.From.String() + move.To.String()
	if move.Promotion != board.Empty {
		s += string(promotionLetter(move.Promotion))
	}
	return s
}

// promotionLetter returns the lowercase algebraic letter for a promotion
// piece, for use in coordinate-notation move strings.
func promotionLetter(p board.Piece) rune {
	switch p {
	case board.WhiteQueen, board.BlackQueen:
		return 'q'
	case board.WhiteRook, board.BlackRook:
		return 'r'
	case board.WhiteBishop, board.BlackBishop:
		return 'b'
	case board.WhiteKnight, board.BlackKnight:
		return 'n'
	default:
		return '?'
	}
}
