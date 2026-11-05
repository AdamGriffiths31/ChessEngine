package nnue

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

func randomNetwork(seed uint64) *Network {
	rng := rand.New(rand.NewPCG(seed, seed))
	n := &Network{}
	for i := range n.FeatureWeights {
		n.FeatureWeights[i] = int16(rng.IntN(201) - 100)
	}
	for i := range n.FeatureBias {
		n.FeatureBias[i] = int16(rng.IntN(201) - 50)
	}
	for i := range n.OutputWeights {
		n.OutputWeights[i] = int16(rng.IntN(255) - 127)
	}
	n.OutputBias = int32(rng.IntN(20001) - 10000)
	return n
}

func TestNetwork_WriteThenReadRoundTrips(t *testing.T) {
	want := randomNetwork(1)
	var buf bytes.Buffer
	if err := want.Write(&buf); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if *got != *want {
		t.Error("network read back differs from the one written")
	}
}

func TestRead_RejectsOutputWeightsTheSIMDMultiplyCannotHold(t *testing.T) {
	n := randomNetwork(3)
	n.OutputWeights[17] = 128
	var buf bytes.Buffer
	if err := n.Write(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(&buf); err == nil {
		t.Error("Read accepted an output weight of 128, outside ±127")
	}
}

func TestEvaluate_ColourMirrorNegatesScore(t *testing.T) {
	tests := []struct{ fen, mirrored string }{
		{"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1", "rnbqkbnr/pppp1ppp/8/4p3/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"r1bq1rk1/pp2bppp/2n1pn2/3p4/3P4/2NBPN2/PP3PPP/R2QK2R w KQ - 0 9", "r2qk2r/pp3ppp/2nbpn2/3p4/3P4/2N1PN2/PP2BPPP/R1BQ1RK1 b kq - 0 9"},
		{"8/5k2/8/3K4/8/8/4P3/8 w - - 0 1", "8/4p3/8/8/3k4/8/5K2/8 b - - 0 1"},
	}
	e := NewEvaluator(randomNetwork(2))
	for _, tc := range tests {
		score := e.Evaluate(testutil.MustFromFEN(t, tc.fen))
		mirrored := e.Evaluate(testutil.MustFromFEN(t, tc.mirrored))
		if score != -mirrored {
			t.Errorf("%s scored %d but its mirror scored %d; want opposite scores", tc.fen, score, mirrored)
		}
	}
}

func TestDefault_EmbeddedNetworkLoads(t *testing.T) {
	score := NewEvaluator(Default()).Evaluate(testutil.MustFromFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"))
	if score < -100 || score > 100 {
		t.Errorf("embedded network scores the start position %d, want within ±100", score)
	}
}

func TestEvaluate_IncrementalUpdatesMatchAFreshEvaluator(t *testing.T) {
	fens := []string{
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	}
	net := randomNetwork(4)
	incremental := NewEvaluator(net)
	gen := movegen.NewGenerator()
	for seed, fen := range fens {
		rng := rand.New(rand.NewPCG(uint64(seed), 7))
		b := testutil.MustFromFEN(t, fen)
		for ply := 0; ply < 80; ply++ {
			player := movegen.White
			if b.GetSideToMove() == "b" {
				player = movegen.Black
			}
			legal := gen.GenerateAllMoves(b, player)
			if legal.Count == 0 {
				movegen.ReleaseMoveList(legal)
				break
			}
			for i := 0; i < legal.Count; i++ {
				undo, err := b.MakeMoveWithUndo(legal.Moves[i])
				if err != nil {
					t.Fatal(err)
				}
				if got, want := incremental.Evaluate(b), NewEvaluator(net).Evaluate(b); got != want {
					t.Fatalf("%s ply %d: incremental eval %d, fresh eval %d", fen, ply, got, want)
				}
				b.UnmakeMove(undo)
			}
			move := legal.Moves[rng.IntN(legal.Count)]
			movegen.ReleaseMoveList(legal)
			if _, err := b.MakeMoveWithUndo(move); err != nil {
				t.Fatal(err)
			}
		}
	}
}
