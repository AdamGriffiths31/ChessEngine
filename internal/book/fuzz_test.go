package book

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// fuzzDecodeMoveBoards are fixed, known-good positions decodeMove is
// exercised against. They are built once from hardcoded FEN strings (not
// fuzzed input) so this target isolates decodeMove's own bit-unpacking
// logic rather than re-discovering bugs in board.FromFEN.
func fuzzDecodeMoveBoards(tb testing.TB) []*board.Board {
	tb.Helper()
	fens := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
		"rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
		"8/8/8/8/8/8/1p6/7K b - - 0 1",
		"7k/1P6/8/8/8/8/8/8 w - - 0 1",
		"8/8/8/8/8/8/8/8 w - - 0 1",
	}
	boards := make([]*board.Board, 0, len(fens))
	for _, fen := range fens {
		b, err := board.FromFEN(fen)
		if err != nil {
			tb.Fatalf("fixture FEN %q failed to parse: %v", fen, err)
		}
		boards = append(boards, b)
	}
	return boards
}

// FuzzDecodeMove feeds arbitrary 16-bit Polyglot move encodings to
// decodeMove against a fixed set of known-good board positions. decodeMove
// is unexported, but this fuzz test lives in package book so it can call it
// directly, exercising the exact code path LookupMove reaches. Malformed
// encodings must never panic; where decoding succeeds, the returned move
// must reference in-range squares.
func FuzzDecodeMove(f *testing.F) {
	seeds := []uint16{
		polyglotEncode(4, 0, 7, 0), // white kingside castling
		polyglotEncode(4, 0, 0, 0), // white queenside castling
		polyglotEncode(4, 7, 7, 7), // black kingside castling
		polyglotEncode(4, 0, 4, 1), // ordinary king step
		polyglotEncode(4, 4, 3, 5), // en passant shape
		polyglotEncode(1, 1, 1, 0) | uint16(PromotionQueen<<PromotionShift),
		polyglotEncode(1, 6, 1, 7) | uint16(PromotionQueen<<PromotionShift),
		0,
		0xFFFF,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, encoded uint16) {
		for _, b := range fuzzDecodeMoveBoards(t) {
			pb := NewPolyglotBook()
			move, err := pb.decodeMove(encoded, b)
			if err != nil {
				continue
			}

			if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 {
				t.Fatalf("decodeMove(%d) produced out-of-range From square: %+v", encoded, move.From)
			}
			if move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
				t.Fatalf("decodeMove(%d) produced out-of-range To square: %+v", encoded, move.To)
			}
		}
	})
}
