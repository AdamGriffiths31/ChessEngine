package engine

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func TestParseMove(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	move := game.Position().ParseMove([]byte("e1g1"))

	if move == data.NoMove {
		t.Errorf("Expected move but got %v", move)
	}
}

func TestIllegalParseMove(t *testing.T) {
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	move := game.Position().ParseMove([]byte("e1g2"))

	if move != data.NoMove {
		t.Errorf("Expected no move but got %v", move)
	}
}

func TestPromotionParseMove(t *testing.T) {
	game := ParseFen("8/P7/8/8/8/8/3k4/7K w - - 0 1")
	move := game.Position().ParseMove([]byte("a7a8q"))

	if move == data.NoMove {
		t.Errorf("Expected move but got %v", move)
	}
}

func TestIllegalPromotionParseMove(t *testing.T) {
	game := ParseFen("8/P7/8/8/8/8/3k4/7K w - - 0 1")
	move := game.Position().ParseMove([]byte("a7a8z"))

	if move != data.NoMove {
		t.Errorf("Expected no move but got %v", move)
	}
}

func TestEnPasParseMove(t *testing.T) {
	game := ParseFen("rnbqkbnr/pppp1pp1/7p/4pP2/8/8/PPPPP1PP/RNBQKBNR w KQkq e6 0 3")
	move := game.Position().ParseMove([]byte("f5e6"))

	if move == data.NoMove {
		t.Errorf("Expected move but got %v", move)
	}
}

func TestCastleParseMove(t *testing.T){
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	move := game.Position().ParseMove([]byte("e1g1"))

	if move == data.NoMove {
		t.Errorf("Expected move but got %v", move)
	}
}
