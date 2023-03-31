package eval

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func TestKnightMobility(t *testing.T) {
	game := engine.ParseFen("7k/4pppp/3R1R2/2b3R1/4N3/2R3R1/3R1RPP/K7 w - - 0 1")
	e := NewEvaluationService()
	e.SetupEvaluate(game.Position())
	eval := e.EvaluateMobilityKnights(game.Position(), data.White)
	if eval != e.KnightMobility[6] {
		t.Errorf("Expected %v but got %v", e.KnightMobility[1], eval)
	}
}

func TestBishopMobility(t *testing.T) {
	game := engine.ParseFen("r6k/p7/1p6/3P4/8/5B2/6P1/K7 w - - 0 0")
	e := NewEvaluationService()
	e.SetupEvaluate(game.Position())
	eval := e.EvaluateMobilityBishops(game.Position(), data.White)
	if eval != e.BishopMobility[6] {
		t.Errorf("Expected %v but got %v", e.BishopMobility[1], eval)
	}
}
