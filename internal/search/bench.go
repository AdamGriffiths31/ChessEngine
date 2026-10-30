package search

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// DefaultBenchDepth is the fixed depth the bench command searches each
// position to. Changing it changes every bench signature, so only do so
// deliberately and note it in the commit message.
const DefaultBenchDepth = 9

const benchTTSizeMB = 16

// BenchFENs is the fixed position set for the bench signature. Never reorder
// or edit entries casually: the total node count is compared across commits.
var BenchFENs = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
	"rnbqkb1r/pp2pppp/3p1n2/2p5/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 0 4",
	"r1bq1rk1/ppp2ppp/2np1n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQ1RK1 w - - 0 8",
	"1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
	"r1b2rk1/2q1b1pp/p2ppn2/1p6/3QP3/1BN1B3/PPP3PP/R4RK1 w - - 0 1",
	"2r3k1/pp3ppp/2n1p3/3p4/3P4/2P1PN2/P4PPP/2R3K1 b - - 0 1",
	"r2q1rk1/pb1nbppp/1p2pn2/2pp4/2PP4/1PN1PN2/PB2BPPP/R2Q1RK1 w - - 0 10",
	"3r2k1/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1",
	"8/8/4k3/8/3nK3/8/4B3/8 w - - 0 1",
	"8/5pk1/6p1/8/3R4/6P1/5PK1/3r4 w - - 0 1",
	"8/8/1p1k4/1P6/2K5/8/8/8 w - - 0 1",
	"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
}

// BenchResult is the outcome of RunBench.
type BenchResult struct {
	Nodes   int64
	Elapsed time.Duration
}

// NodesPerSecond returns the bench speed, or 0 if no time elapsed.
func (r BenchResult) NodesPerSecond() int64 {
	if r.Elapsed <= 0 {
		return 0
	}
	return int64(float64(r.Nodes) / r.Elapsed.Seconds())
}

// RunBench searches every BenchFENs position to depth with a fresh engine,
// clearing search state between positions, and returns the total node count.
// The node count is deterministic for a given build and depth, so it serves
// as a search signature: any change that alters it changes search behaviour.
func RunBench(depth int) (BenchResult, error) {
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(benchTTSizeMB)
	config := Config{MaxDepth: depth}

	var total BenchResult
	for _, fen := range BenchFENs {
		b, err := board.FromFEN(fen)
		if err != nil {
			return BenchResult{}, fmt.Errorf("bench FEN %q: %w", fen, err)
		}
		player := movegen.White
		if b.GetSideToMove() == "b" {
			player = movegen.Black
		}

		engine.ClearSearchState()
		start := time.Now()
		result := engine.FindBestMove(context.Background(), b, player, config)
		total.Elapsed += time.Since(start)
		total.Nodes += result.Stats.NodesSearched
	}
	return total, nil
}
