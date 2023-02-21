package engine

import (
	"fmt"
	"testing"
)

func TestPerf(t *testing.T) {
	var tests = []struct {
		FEN   string
		Depth int
		want  int64
	}{
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 4, 197281},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 4, 4085603},
		{"n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1", 4, 182838},
		{"4k3/8/8/8/8/8/8/4K2R w K - 0 1", 4, 7059},
		{"4k3/8/8/8/8/8/8/R3K3 w Q - 0 1", 4, 7626},
		{"4k2r/8/8/8/8/8/8/4K3 w k - 0 1", 4, 8290},
		{"r3k2r/8/8/8/8/8/8/R3K1R1 b Qkq - 0 1", 4, 320792},
		{"8/Pk6/8/8/8/8/6Kp/8 w - - 0 1", 4, 8048},
		{"k7/6p1/8/8/8/8/7P/K7 b - - 0 1", 6, 55338},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%d", tt.FEN, tt.Depth)
		t.Run(testname, func(t *testing.T) {
			ans := PerftTest(tt.Depth, tt.FEN)
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}
