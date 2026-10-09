package book

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// polyglotEncode builds a raw Polyglot move from 0-indexed squares.
func polyglotEncode(fromFile, fromRank, toFile, toRank int) uint16 {
	from := fromRank*8 + fromFile
	to := toRank*8 + toFile
	return uint16(from<<FromSquareShift) | uint16(to)
}

func queenPromotion(encoded uint16) uint16 {
	return encoded | uint16(PromotionQueen<<PromotionShift)
}

// Polyglot writes castling as the king moving onto its own rook (e1-h1), so
// decodeMove must remap it to the king's real destination with IsCastling set.
func TestDecodeMove(t *testing.T) {
	const startPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR %s KQkq - 0 1"
	tests := []struct {
		name        string
		fen         string
		encoded     uint16
		wantTo      board.Square
		wantCastle  bool
		wantPromoTo board.Piece
	}{
		{"white kingside castling", fmt.Sprintf(startPos, "w"), polyglotEncode(4, 0, 7, 0), board.Square{File: 6, Rank: 0}, true, 0},
		{"white queenside castling", fmt.Sprintf(startPos, "w"), polyglotEncode(4, 0, 0, 0), board.Square{File: 2, Rank: 0}, true, 0},
		{"black kingside castling", fmt.Sprintf(startPos, "b"), polyglotEncode(4, 7, 7, 7), board.Square{File: 6, Rank: 7}, true, 0},
		{"ordinary king step is not castling", fmt.Sprintf(startPos, "w"), polyglotEncode(4, 0, 4, 1), board.Square{File: 4, Rank: 1}, false, 0},
		{"white promotion gets a white piece", "7k/1P6/8/8/8/8/8/8 w - - 0 1", queenPromotion(polyglotEncode(1, 6, 1, 7)), board.Square{File: 1, Rank: 7}, false, board.WhiteQueen},
		{"black promotion gets a black piece", "8/8/8/8/8/8/1p6/7K b - - 0 1", queenPromotion(polyglotEncode(1, 1, 1, 0)), board.Square{File: 1, Rank: 0}, false, board.BlackQueen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			move, err := NewPolyglotBook().decodeMove(tt.encoded, b)
			if err != nil {
				t.Fatalf("decodeMove: %v", err)
			}

			if move.To != tt.wantTo {
				t.Errorf("To = %v, want %v", move.To, tt.wantTo)
			}
			if move.IsCastling != tt.wantCastle {
				t.Errorf("IsCastling = %v, want %v", move.IsCastling, tt.wantCastle)
			}
			if tt.wantCastle && move.IsCapture {
				t.Error("castling must not be a capture (the rook is the king's own)")
			}
			if tt.wantPromoTo != 0 && move.Promotion != tt.wantPromoTo {
				t.Errorf("Promotion = %c, want %c", move.Promotion, tt.wantPromoTo)
			}
		})
	}
}

// A decoded move must apply cleanly to a real board: the castling rook moves
// too, and en passant removes the captured pawn.
func TestDecodedMovesApplyToBoard(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		encoded uint16
		want    map[[2]int]board.Piece // [rank, file] -> piece after the move
	}{
		{
			name:    "white kingside castling",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			encoded: polyglotEncode(4, 0, 7, 0),
			want: map[[2]int]board.Piece{
				{0, 6}: board.WhiteKing, {0, 5}: board.WhiteRook, {0, 4}: board.Empty, {0, 7}: board.Empty,
			},
		},
		{
			name:    "en passant",
			fen:     "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
			encoded: polyglotEncode(4, 4, 3, 5),
			want: map[[2]int]board.Piece{
				{5, 3}: board.WhitePawn, {4, 4}: board.Empty, {4, 3}: board.Empty, // d6, e5, captured d5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			move, err := NewPolyglotBook().decodeMove(tt.encoded, b)
			if err != nil {
				t.Fatalf("decodeMove: %v", err)
			}
			if err := b.MakeMove(move); err != nil {
				t.Fatalf("MakeMove: %v", err)
			}
			for pos, piece := range tt.want {
				if got := b.GetPiece(pos[0], pos[1]); got != piece {
					t.Errorf("square (rank %d, file %d) = %c, want %c", pos[0], pos[1], got, piece)
				}
			}
		})
	}
}

func TestDecodeMoveErrors(t *testing.T) {
	start := startBoard(t)
	tests := []struct {
		name    string
		encoded uint16
	}{
		{"no piece on the from square", polyglotEncode(4, 3, 4, 4)},
		{"invalid promotion code", polyglotEncode(4, 1, 4, 3) | uint16(7<<PromotionShift)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPolyglotBook().decodeMove(tt.encoded, start); err == nil {
				t.Error("decodeMove succeeded, want an error")
			}
		})
	}
}
