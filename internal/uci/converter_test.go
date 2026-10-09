package uci

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	startFEN      = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	castlingFEN   = "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
	captureFEN    = "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2"
	enPassantFEN  = "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3"
	whitePromoFEN = "4k3/P7/8/8/8/8/8/4K3 w - - 0 1"
	blackPromoFEN = "4k3/8/8/8/8/8/p7/4K3 b - - 0 1"
)

// sq builds a Square from algebraic notation ("e4"), independently of the
// production parser.
func sq(name string) board.Square {
	return board.Square{File: int(name[0] - 'a'), Rank: int(name[1] - '1')}
}

// mv builds a Move the way the move generator does (no promotion, nothing
// captured), then applies any modifiers.
func mv(from, to string, piece board.Piece, mods ...func(*board.Move)) board.Move {
	m := board.Move{From: sq(from), To: sq(to), Piece: piece, Promotion: board.Empty, Captured: board.Empty}
	for _, mod := range mods {
		mod(&m)
	}
	return m
}

func castling(m *board.Move)  { m.IsCastling = true }
func enPassant(m *board.Move) { m.IsEnPassant = true }

func promotesTo(p board.Piece) func(*board.Move) {
	return func(m *board.Move) { m.Promotion = p }
}

func captures(p board.Piece) func(*board.Move) {
	return func(m *board.Move) { m.Captured, m.IsCapture = p, true }
}

func TestMoveConverter_ToUCI(t *testing.T) {
	converter := NewMoveConverter()

	offBoard := mv("e2", "e4", board.WhitePawn)
	offBoard.From = board.Square{File: 8, Rank: 0}

	tests := []struct {
		name string
		move board.Move
		want string
	}{
		{"pawn move", mv("e2", "e4", board.WhitePawn), "e2e4"},
		{"knight move", mv("b1", "c3", board.WhiteKnight), "b1c3"},
		{"white promotion", mv("a7", "a8", board.WhitePawn, promotesTo(board.WhiteQueen)), "a7a8q"},
		{"black promotion", mv("h2", "h1", board.BlackPawn, promotesTo(board.BlackKnight)), "h2h1n"},
		{"off-board square", offBoard, "0000"},

		// Opening-book castling is stored as king-takes-rook; ToUCI rewrites it
		// to the standard king destination, but only when IsCastling is set.
		{"white kingside castling", mv("e1", "h1", board.WhiteKing, castling), "e1g1"},
		{"white queenside castling", mv("e1", "a1", board.WhiteKing, castling), "e1c1"},
		{"black kingside castling", mv("e8", "h8", board.BlackKing, castling), "e8g8"},
		{"black queenside castling", mv("e8", "a8", board.BlackKing, castling), "e8c8"},

		// Regression: a rook lifted to e1 and moving back to a1 shares its
		// squares with queenside castling. It was rewritten to e1c1, desyncing
		// the recorded game from what was sent to the opponent.
		{"rook move sharing castling squares", mv("e1", "a1", board.WhiteRook), "e1a1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := converter.ToUCI(tt.move); got != tt.want {
				t.Errorf("ToUCI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMoveConverter_FromUCI(t *testing.T) {
	converter := NewMoveConverter()

	tests := []struct {
		name string
		fen  string
		uci  string
		want board.Move
	}{
		{"pawn move", startFEN, "e2e4", mv("e2", "e4", board.WhitePawn)},
		{"knight move", startFEN, "g1f3", mv("g1", "f3", board.WhiteKnight)},
		{"capture", captureFEN, "e4d5", mv("e4", "d5", board.WhitePawn, captures(board.BlackPawn))},
		{"white promotion", whitePromoFEN, "a7a8q", mv("a7", "a8", board.WhitePawn, promotesTo(board.WhiteQueen))},
		{"black promotion", blackPromoFEN, "a2a1n", mv("a2", "a1", board.BlackPawn, promotesTo(board.BlackKnight))},
		{"kingside castling", castlingFEN, "e1g1", mv("e1", "g1", board.WhiteKing, castling)},
		{"en passant", enPassantFEN, "e5d6", mv("e5", "d6", board.WhitePawn, enPassant)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("FromFEN: %v", err)
			}
			got, err := converter.FromUCI(tt.uci, b)
			if err != nil {
				t.Fatalf("FromUCI(%q): %v", tt.uci, err)
			}
			// FromUCI leaves Promotion as the zero value for non-promotions
			// rather than board.Empty; treat the two as equal here.
			if got.Promotion == 0 {
				got.Promotion = board.Empty
			}
			if got != tt.want {
				t.Errorf("FromUCI(%q) =\n  %+v\nwant\n  %+v", tt.uci, got, tt.want)
			}
		})
	}
}

func TestMoveConverter_FromUCIErrors(t *testing.T) {
	converter := NewMoveConverter()
	start, err := board.FromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		uci   string
		board *board.Board
	}{
		{"too short", "e2", start},
		{"too long", "e2e4qq", start},
		{"invalid from square", "z9e4", start},
		{"invalid to square", "e2z9", start},
		{"no piece on from square", "e4e5", start},
		{"nil board", "e2e4", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := converter.FromUCI(tt.uci, tt.board); err == nil {
				t.Errorf("FromUCI(%q) succeeded, want an error", tt.uci)
			}
		})
	}
}

func TestUCISquare(t *testing.T) {
	valid := []struct {
		text   string
		square board.Square
	}{
		{"a1", board.Square{File: 0, Rank: 0}},
		{"e4", board.Square{File: 4, Rank: 3}},
		{"d5", board.Square{File: 3, Rank: 4}},
		{"h8", board.Square{File: 7, Rank: 7}},
	}
	for _, tt := range valid {
		t.Run(tt.text, func(t *testing.T) {
			if got := squareToUCI(tt.square); got != tt.text {
				t.Errorf("squareToUCI(%+v) = %q, want %q", tt.square, got, tt.text)
			}
			got, err := parseUCISquare(tt.text)
			if err != nil {
				t.Fatalf("parseUCISquare(%q): %v", tt.text, err)
			}
			if got != tt.square {
				t.Errorf("parseUCISquare(%q) = %+v, want %+v", tt.text, got, tt.square)
			}
		})
	}

	for _, text := range []string{"", "a", "a11", "z1", "a9"} {
		t.Run("invalid_"+text, func(t *testing.T) {
			if _, err := parseUCISquare(text); err == nil {
				t.Errorf("parseUCISquare(%q) succeeded, want an error", text)
			}
		})
	}
}

func TestParsePromotionPiece(t *testing.T) {
	tests := []struct {
		name string
		char byte
		pawn board.Piece
		want board.Piece
	}{
		{"white queen", 'q', board.WhitePawn, board.WhiteQueen},
		{"white rook", 'r', board.WhitePawn, board.WhiteRook},
		{"white bishop", 'b', board.WhitePawn, board.WhiteBishop},
		{"white knight", 'n', board.WhitePawn, board.WhiteKnight},
		{"black queen", 'q', board.BlackPawn, board.BlackQueen},
		{"black rook", 'r', board.BlackPawn, board.BlackRook},
		{"black bishop", 'b', board.BlackPawn, board.BlackBishop},
		{"black knight", 'n', board.BlackPawn, board.BlackKnight},
		{"unknown letter defaults to queen", 'x', board.WhitePawn, board.WhiteQueen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePromotionPiece(tt.char, tt.pawn); got != tt.want {
				t.Errorf("parsePromotionPiece(%q, %c) = %c, want %c", tt.char, tt.pawn, got, tt.want)
			}
		})
	}
}
