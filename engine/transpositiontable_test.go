package engine

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func TestProbeTT(t *testing.T) {
	var tt = TranspositionTable
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	move := game.Position().ParseMove([]byte("e1g1"))
	tt.Store(game.position.PositionKey, game.position.Play, move, 0, data.PVExact, 0)
	if tt.Probe(game.position.PositionKey) != move {
		t.Errorf("Expected %v but got %v", move, tt.Probe(game.position.PositionKey))
	}
}

func TestProbeTTOverwrite(t *testing.T) {
	var tt = TranspositionTable
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	move := game.Position().ParseMove([]byte("e1g1"))
	move2 := game.Position().ParseMove([]byte("e1g3"))
	tt.Store(game.position.PositionKey, game.position.Play, move, 0, data.PVExact, 0)
	tt.Store(game.position.PositionKey, game.position.Play, move2, 0, data.PVExact, 0)
	if tt.Probe(game.position.PositionKey) != move2 {
		t.Errorf("Expected %v but got %v", move2, tt.Probe(game.position.PositionKey))
	}
}

func TestProbeTTEmpty(t *testing.T) {
	var tt = TranspositionTable
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	if tt.Probe(game.position.PositionKey) != data.NoMove {
		t.Errorf("Expected %v but got %v", data.NoMove, tt.Probe(game.position.PositionKey))
	}
}

func TestProbeTTWrongKey(t *testing.T) {
	var tt = TranspositionTable
	game := ParseFen("4k3/8/8/8/8/8/5PPP/4K2R w K - 0 1")
	game2 := ParseFen("4k3/8/8/8/8/8/5BPP/4K2R w K - 0 2")
	move := game.Position().ParseMove([]byte("e1g1"))
	tt.Store(game.position.PositionKey, game.position.Play, move, 0, data.PVExact, 0)
	if tt.Probe(game2.position.PositionKey) != data.NoMove {
		t.Errorf("Expected %v but got %v", data.NoMove, tt.Probe(game2.position.PositionKey))
	}
}
