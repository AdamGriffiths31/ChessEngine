package engine

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func TestMakeMoveEnPasIsSet(t *testing.T) {
	game := ParseFen(data.StartFEN)
	game.position.MakeMove(531363) //e2e4

	if game.position.EnPassant != data.E3 {
		t.Errorf("Expected en passant square to be E3 but was %v", io.SquareString(game.position.EnPassant))
	}
}

func TestMakeMoveInvalidMoveReturnsFalse(t *testing.T) {
	game := ParseFen("4k3/4q3/8/8/8/8/4Q3/4K3 b - - 0 1")
	valid, _, _, _ := game.position.MakeMove(5333)

	if valid {
		t.Errorf("Expected invalid but got %v", valid)
	}
}

func TestMakeMoveCastlePieceUpdates(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	game.position.MakeMove(16780697)

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.F1]) != data.WR {
		t.Errorf("Expected WR but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.F1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.G1]) != data.WK {
		t.Errorf("Expected WK but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.G1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.H1]) != data.Empty {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.H1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.E1]) != data.Empty {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.E1]))
	}

	if game.Position().CastlePermission != 0 {
		t.Errorf("Expected 0 but got %v", game.Position().CastlePermission)
	}
}

func TestTakeBackCastlePieceUpdates(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	_, enPas, castle, fifty := game.position.MakeMove(16780697)
	game.Position().TakeMoveBack(16780697, enPas, castle, fifty)

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.H1]) != data.WR {
		t.Errorf("Expected WR but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.H1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.E1]) != data.WK {
		t.Errorf("Expected WR but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.E1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.G1]) != data.Empty {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.E1]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.F1]) != data.Empty {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.F1]))
	}

	if game.Position().CastlePermission != 1 {
		t.Errorf("Expected 1 but got %v", game.Position().CastlePermission)
	}
}

func TestMakeMovePromotion(t *testing.T) {
	game := ParseFen("1b6/P7/7k/8/8/8/8/6K1 w - - 0 0")
	game.position.MakeMove(3305041)
	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.A7]) != data.Empty {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.A7]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.B8]) != data.WB {
		t.Errorf("Expected empty but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.B8]))
	}

	if game.Position().Board.CountBits(game.position.Board.BlackPieces) != 1 {
		t.Errorf("Expected 1 but got %v", game.Position().Board.CountBits(game.position.Board.BlackPieces))
	}

	if game.Position().Board.CountBits(game.position.Board.WhitePieces) != 2 {
		t.Errorf("Expected 2 but got %v", game.Position().Board.CountBits(game.position.Board.WhitePieces))
	}
}

func TestTakeBackPromotion(t *testing.T) {
	game := ParseFen("1b6/P7/7k/8/8/8/8/6K1 w - - 0 0")
	_, enPas, castle, fifty := game.position.MakeMove(3305041)
	game.Position().TakeMoveBack(3305041, enPas, castle, fifty)

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.A7]) != data.WP {
		t.Errorf("Expected WP but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.A7]))
	}

	if game.Position().Board.PieceAt(data.Square120ToSquare64[data.B8]) != data.BB {
		t.Errorf("Expected BB but got %v", game.Position().Board.PieceAt(data.Square120ToSquare64[data.B8]))
	}

	if game.Position().Board.CountBits(game.position.Board.BlackPieces) != 2 {
		t.Errorf("Expected 1 but got %v", game.Position().Board.CountBits(game.position.Board.BlackPieces))
	}

	if game.Position().Board.CountBits(game.position.Board.WhitePieces) != 2 {
		t.Errorf("Expected 2 but got %v", game.Position().Board.CountBits(game.position.Board.WhitePieces))
	}
}

func TestMakeNullMove(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	preMovePlay := game.Position().Play
	game.Position().MakeNullMove()

	if preMovePlay+1 != game.Position().Play {
		t.Errorf("Expected %v but got %v", preMovePlay+1, game.Position().Play)
	}

	if game.Position().Side != data.Black {
		t.Errorf("Expected %v but got %v", data.Black, game.Position().Side)
	}
}

func TestTakeBackNullMove(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	preMoveHash := game.Position().PositionKey
	preMovePlay := game.Position().Play
	_, enPas, castle := game.Position().MakeNullMove()
	game.Position().TakeNullMoveBack(enPas, castle)

	if preMoveHash != game.Position().PositionKey {
		t.Errorf("Expected %v but got %v", preMoveHash, game.Position().PositionKey)
	}

	if preMovePlay != game.Position().Play {
		t.Errorf("Expected %v but got %v", preMovePlay, game.Position().Play)
	}
}
