package book

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// polyglotEncode builds a raw Polyglot move encoding from 0-indexed
// from/to squares (file+rank*8), matching decodeMove's own unpacking.
func polyglotEncode(fromFile, fromRank, toFile, toRank int) uint16 {
	from := fromRank*8 + fromFile
	to := toRank*8 + toFile
	return uint16(from<<FromSquareShift) | uint16(to)
}

func TestDecodeMoveWhiteKingsideCastling(t *testing.T) {
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// Polyglot encodes castling as king-from to the ROOK's own square:
	// White kingside is e1 (file 4, rank 0) -> h1 (file 7, rank 0).
	encoded := polyglotEncode(4, 0, 7, 0)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if !move.IsCastling {
		t.Error("expected IsCastling to be true for a Polyglot-encoded castling move")
	}
	if move.To.File != 6 || move.To.Rank != 0 {
		t.Errorf("expected king destination g1 (file 6, rank 0), got file %d rank %d", move.To.File, move.To.Rank)
	}
	if move.IsCapture {
		t.Error("expected IsCapture to be false for a castling move (own rook is not a capture)")
	}
}

func TestDecodeMoveWhiteQueensideCastling(t *testing.T) {
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// White queenside: e1 (file 4, rank 0) -> a1 (file 0, rank 0).
	encoded := polyglotEncode(4, 0, 0, 0)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if !move.IsCastling {
		t.Error("expected IsCastling to be true for a Polyglot-encoded castling move")
	}
	if move.To.File != 2 || move.To.Rank != 0 {
		t.Errorf("expected king destination c1 (file 2, rank 0), got file %d rank %d", move.To.File, move.To.Rank)
	}
}

func TestDecodeMoveBlackKingsideCastling(t *testing.T) {
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// Black kingside: e8 (file 4, rank 7) -> h8 (file 7, rank 7).
	encoded := polyglotEncode(4, 7, 7, 7)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if !move.IsCastling {
		t.Error("expected IsCastling to be true for a Polyglot-encoded castling move")
	}
	if move.To.File != 6 || move.To.Rank != 7 {
		t.Errorf("expected king destination g8 (file 6, rank 7), got file %d rank %d", move.To.File, move.To.Rank)
	}
}

func TestDecodeMoveOrdinaryKingMoveIsNotCastling(t *testing.T) {
	// King still on e1 (starting position); decoding a hypothetical e1-e2
	// step (a perfectly ordinary, non-castling move) must not be
	// misdetected as castling.
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// e1 (file 4, rank 0) -> e2 (file 4, rank 1): ordinary king step.
	encoded := polyglotEncode(4, 0, 4, 1)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if move.IsCastling {
		t.Error("expected an ordinary king move not to be flagged as castling")
	}
	if move.To.File != 4 || move.To.Rank != 1 {
		t.Errorf("expected destination e2 (file 4, rank 1) unchanged, got file %d rank %d", move.To.File, move.To.Rank)
	}
}

func TestDecodeMoveThenApplyEnPassantCaptureOnBoard(t *testing.T) {
	// Polyglot en passant moves are plain diagonal pawn moves in standard
	// notation (no special encoding quirk like castling has), and
	// board.Board.MakeMove auto-detects en passant on its own (board/moves.go)
	// even when the caller doesn't set IsEnPassant. This test exists to
	// document and pin down that end-to-end behavior stays correct.
	b, err := board.FromFEN("rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	encoded := polyglotEncode(4, 4, 3, 5)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if err := b.MakeMove(move); err != nil {
		t.Fatalf("MakeMove failed: %v", err)
	}

	if piece := b.GetPiece(5, 3); piece != board.WhitePawn { // d6
		t.Errorf("expected white pawn on d6, got %c", piece)
	}
	if piece := b.GetPiece(4, 4); piece != board.Empty { // e5
		t.Errorf("expected e5 to be empty after the capturing pawn moved, got %c", piece)
	}
	if piece := b.GetPiece(4, 3); piece != board.Empty { // d5 - the captured pawn
		t.Errorf("expected the captured black pawn on d5 to be removed, got %c", piece)
	}
}

func TestDecodeMoveBlackPromotionUsesBlackPiece(t *testing.T) {
	// Black pawn on b2 promoting to a queen on b1. decodeMove must
	// produce a BLACK queen, not a White one.
	b, err := board.FromFEN("8/8/8/8/8/8/1p6/7K b - - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// b2 (file 1, rank 1) -> b1 (file 1, rank 0), promotion=Queen.
	encoded := polyglotEncode(1, 1, 1, 0) | uint16(PromotionQueen<<PromotionShift)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if move.Promotion != board.BlackQueen {
		t.Errorf("expected BlackQueen promotion for a black pawn, got %c", move.Promotion)
	}
}

func TestDecodeMoveWhitePromotionUsesWhitePiece(t *testing.T) {
	b, err := board.FromFEN("7k/1P6/8/8/8/8/8/8 w - - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	pb := NewPolyglotBook()
	// b7 (file 1, rank 6) -> b8 (file 1, rank 7), promotion=Queen.
	encoded := polyglotEncode(1, 6, 1, 7) | uint16(PromotionQueen<<PromotionShift)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if move.Promotion != board.WhiteQueen {
		t.Errorf("expected WhiteQueen promotion for a white pawn, got %c", move.Promotion)
	}
}

func TestDecodeMoveThenApplyWhiteKingsideCastlingOnBoard(t *testing.T) {
	// End-to-end: decode a book castling move and actually apply it to a
	// real board.Board, verifying the king and rook land on the correct
	// squares (the exact scenario that was silently corrupting games).
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("failed to build board: %v", err)
	}

	// Clear the squares between king and rook so kingside castling is legal
	// (this board isn't checking legality here, but MakeMove's castling
	// handler unconditionally relocates the rook, so squares must be sane).
	// f1 and g1 are already empty in the starting position.

	pb := NewPolyglotBook()
	encoded := polyglotEncode(4, 0, 7, 0)

	move, err := pb.decodeMove(encoded, b)
	if err != nil {
		t.Fatalf("decodeMove failed: %v", err)
	}

	if err := b.MakeMove(move); err != nil {
		t.Fatalf("MakeMove failed: %v", err)
	}

	if piece := b.GetPiece(0, 6); piece != board.WhiteKing { // g1
		t.Errorf("expected white king on g1, got %c", piece)
	}
	if piece := b.GetPiece(0, 5); piece != board.WhiteRook { // f1
		t.Errorf("expected white rook on f1, got %c", piece)
	}
	if piece := b.GetPiece(0, 4); piece != board.Empty { // e1
		t.Errorf("expected e1 to be empty after castling, got %c", piece)
	}
	if piece := b.GetPiece(0, 7); piece != board.Empty { // h1
		t.Errorf("expected h1 to be empty after castling (rook moved away), got %c", piece)
	}
}
