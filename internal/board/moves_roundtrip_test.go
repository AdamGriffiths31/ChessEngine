package board_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

// roundtripSeedBase is the fixed base seed used to derive a per-FEN RNG seed
// (base + index into roundtripTestFENs). Keeping it fixed makes any failure
// reproducible by re-running the test with the same index.
const roundtripSeedBase = 20240701

// roundtripTestFENs mirrors the FEN set used by the sibling Zobrist
// round-trip test (game/ai/search/zobrist_updates_test.go): a diverse set of
// legal positions exercising the pieces of board state that make/unmake must
// restore exactly: full/partial/no castling rights, en passant availability
// for both colors, imminent promotions for both colors, and a sparse late
// endgame.
var roundtripTestFENs = []string{
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

// maxRoundtripPlies caps the length of each random playout.
const maxRoundtripPlies = 40

// TestMakeUnmakeStateRoundTrip plays bounded, seeded random sequences of
// moves from a diverse set of positions and checks, at every ply, that
// board.Board.MakeMoveWithUndo followed by board.Board.UnmakeMove restores
// the board to exactly the state it was in before the move was made. This is
// checked for every move MakeMoveWithUndo accepts -- legal or not -- because
// unmake must restore state regardless of whether the move turned out to be
// legal. It also checks the MakeNullMove/UnmakeNullMove round trip at every
// position visited.
func TestMakeUnmakeStateRoundTrip(t *testing.T) {
	t.Parallel()
	for idx, fen := range roundtripTestFENs {
		fen := fen
		seed := int64(roundtripSeedBase + idx)
		t.Run(fmt.Sprintf("fen%d", idx), func(t *testing.T) {
			runRoundtripPlayout(t, fen, seed)
		})
	}
}

func runRoundtripPlayout(t *testing.T, fen string, seed int64) {
	t.Helper()

	b := testutil.MustFromFEN(t, fen)
	gen := movegen.NewGenerator()

	rng := rand.New(rand.NewSource(seed))
	player := playerFromSide(b.GetSideToMove())

	var playedMoves []string

	failf := func(format string, args ...interface{}) {
		t.Fatalf("fen=%q seed=%d moves=%v: %s", fen, seed, playedMoves, fmt.Sprintf(format, args...))
	}

	for ply := 0; ply < maxRoundtripPlies; ply++ {
		// Null-move round trip check at the current position, before any
		// real move is made this ply.
		preNull := *b
		nullUndo := b.MakeNullMove()
		b.UnmakeNullMove(nullUndo)
		postNull := *b
		if ok, diff := boardsEqual(&preNull, &postNull); !ok {
			failf("ply %d: null-move round trip did not restore state: %s", ply, diff)
		}

		pseudoMoves := gen.GeneratePseudoLegalMoves(b, player)
		legalMoves := make([]board.Move, 0, pseudoMoves.Count)

		for i := 0; i < pseudoMoves.Count; i++ {
			move := pseudoMoves.Moves[i]

			pre := *b
			undo, err := b.MakeMoveWithUndo(move)
			if err != nil {
				continue
			}

			legal := !gen.IsKingInCheck(b, player)

			b.UnmakeMove(undo)
			post := *b
			if ok, diff := boardsEqual(&pre, &post); !ok {
				failf("ply %d: make/unmake of %s (legal=%v) did not restore state: %s", ply, moveString(move), legal, diff)
			}

			if legal {
				legalMoves = append(legalMoves, move)
			}
		}
		movegen.ReleaseMoveList(pseudoMoves)

		if len(legalMoves) == 0 {
			break
		}

		chosen := legalMoves[rng.Intn(len(legalMoves))]

		pre := *b
		undo, err := b.MakeMoveWithUndo(chosen)
		if err != nil {
			failf("ply %d: chosen move %s unexpectedly failed to make: %v", ply, moveString(chosen), err)
		}
		playedMoves = append(playedMoves, moveString(chosen))

		// Exercise the unmake path for the chosen move too, then re-make it
		// so the playout continues from the resulting position.
		b.UnmakeMove(undo)
		post := *b
		if ok, diff := boardsEqual(&pre, &post); !ok {
			failf("ply %d: make/unmake of chosen move %s did not restore state: %s", ply, moveString(chosen), diff)
		}

		if _, err := b.MakeMoveWithUndo(chosen); err != nil {
			failf("ply %d: chosen move %s unexpectedly failed to re-make: %v", ply, moveString(chosen), err)
		}

		player = opposite(player)
	}
}

// boardsEqual reports whether pre and post represent the same board state
// and, if not, a human-readable description of what differs.
//
// It first tries a full structural comparison via reflect.DeepEqual on the
// dereferenced *board.Board structs. Because pre/post are plain struct
// copies (`snapshot := *b`) taken before/after a make+unmake round trip, this
// is safe even though board.Board has unexported slice fields
// (hashHistory/evalHistory): those slices are only ever grown by append and
// shrunk by re-slicing in board's Push*/Pop* helpers, never mutated
// in-place, so a copy of the slice header taken before the round trip
// compares correctly against the header left after the round trip (same
// length, same backing values in range) even though its backing array may
// have been transiently grown and shrunk in between. reflect.DeepEqual also
// compares unexported fields without issue (unlike Value.Interface(), which
// would panic on them).
//
// If the structural comparison fails, boardsEqual falls back to comparing
// the semantically meaningful state via board's public accessors and
// exported fields, to produce a useful diagnostic without resorting to
// unsafe/reflect tricks to read unexported fields individually. If none of
// those public accessors disagree, the mismatch must be in an unexported
// bookkeeping field not reachable through them (e.g. a leaked
// hash/eval-history stack entry, or the hashUpdater reference) -- itself a
// real bug worth failing on, just not one this fallback can name precisely.
func boardsEqual(pre, post *board.Board) (bool, string) {
	if reflect.DeepEqual(*pre, *post) {
		return true, ""
	}

	var diffs []string

	if preFEN, postFEN := pre.ToFEN(), post.ToFEN(); preFEN != postFEN {
		diffs = append(diffs, fmt.Sprintf("FEN: %q != %q", preFEN, postFEN))
	}
	if pre.GetHash() != post.GetHash() {
		diffs = append(diffs, fmt.Sprintf("hash: %#x != %#x", pre.GetHash(), post.GetHash()))
	}
	if pre.GetCastlingRights() != post.GetCastlingRights() {
		diffs = append(diffs, fmt.Sprintf("castlingRights: %q != %q", pre.GetCastlingRights(), post.GetCastlingRights()))
	}
	preEP, preHasEP := pre.GetEnPassantTarget()
	postEP, postHasEP := post.GetEnPassantTarget()
	if preHasEP != postHasEP || preEP != postEP {
		diffs = append(diffs, fmt.Sprintf("enPassant: (sq=%v has=%v) != (sq=%v has=%v)", preEP, preHasEP, postEP, postHasEP))
	}
	if pre.GetHalfMoveClock() != post.GetHalfMoveClock() {
		diffs = append(diffs, fmt.Sprintf("halfMoveClock: %d != %d", pre.GetHalfMoveClock(), post.GetHalfMoveClock()))
	}
	if pre.GetFullMoveNumber() != post.GetFullMoveNumber() {
		diffs = append(diffs, fmt.Sprintf("fullMoveNumber: %d != %d", pre.GetFullMoveNumber(), post.GetFullMoveNumber()))
	}
	if pre.GetSideToMove() != post.GetSideToMove() {
		diffs = append(diffs, fmt.Sprintf("sideToMove: %q != %q", pre.GetSideToMove(), post.GetSideToMove()))
	}
	if pre.GetMaterialScore() != post.GetMaterialScore() {
		diffs = append(diffs, fmt.Sprintf("materialScore: %d != %d", pre.GetMaterialScore(), post.GetMaterialScore()))
	}
	if pre.GetPSTScore() != post.GetPSTScore() {
		diffs = append(diffs, fmt.Sprintf("pstScore: %d != %d", pre.GetPSTScore(), post.GetPSTScore()))
	}
	if pre.PieceBitboards != post.PieceBitboards {
		diffs = append(diffs, "PieceBitboards differ")
	}
	if pre.Mailbox != post.Mailbox {
		diffs = append(diffs, "Mailbox differ")
	}
	if pre.WhitePieces != post.WhitePieces {
		diffs = append(diffs, fmt.Sprintf("WhitePieces: %#x != %#x", pre.WhitePieces, post.WhitePieces))
	}
	if pre.BlackPieces != post.BlackPieces {
		diffs = append(diffs, fmt.Sprintf("BlackPieces: %#x != %#x", pre.BlackPieces, post.BlackPieces))
	}
	if pre.AllPieces != post.AllPieces {
		diffs = append(diffs, fmt.Sprintf("AllPieces: %#x != %#x", pre.AllPieces, post.AllPieces))
	}

	if len(diffs) == 0 {
		diffs = append(diffs, "no publicly-observable difference found via accessors; mismatch is in an unexported bookkeeping field (e.g. hash/eval history stack depth or hashUpdater reference)")
	}

	return false, strings.Join(diffs, "; ")
}

func playerFromSide(side string) movegen.Player {
	if side == "w" {
		return movegen.White
	}
	return movegen.Black
}

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
