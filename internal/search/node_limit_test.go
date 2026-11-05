package search

import (
	"context"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/testutil"
)

const middlegameFEN = "r1bq1rk1/pp2bppp/2n1pn2/3p4/3P4/2NBPN2/PP3PPP/R2QK2R w KQ - 0 9"

func TestFindBestMove_MaxNodesStopsTheSearch(t *testing.T) {
	const maxNodes = 20000
	m := NewMinimaxEngine()
	m.SetTranspositionTableSize(16)
	res := m.FindBestMove(context.Background(), testutil.MustFromFEN(t, middlegameFEN), movegen.White, Config{MaxDepth: 64, MaxNodes: maxNodes})

	if res.Stats.NodesSearched > maxNodes+1024 {
		t.Errorf("searched %d nodes, want the search stopped near %d", res.Stats.NodesSearched, maxNodes)
	}
	if res.Stats.StopReason != StopReasonNodes {
		t.Errorf("stop reason = %q, want %q", res.Stats.StopReason, StopReasonNodes)
	}
	if !res.BestMove.IsValid() {
		t.Error("no move returned")
	}
}

func TestFindBestMove_SoftNodesFinishesTheIteration(t *testing.T) {
	const softNodes = 5000
	m := NewMinimaxEngine()
	m.SetTranspositionTableSize(16)
	res := m.FindBestMove(context.Background(), testutil.MustFromFEN(t, middlegameFEN), movegen.White, Config{MaxDepth: 64, SoftNodes: softNodes})

	if res.Stats.NodesSearched < softNodes {
		t.Errorf("searched %d nodes, want at least the soft limit %d", res.Stats.NodesSearched, softNodes)
	}
	if res.Stats.StopReason != StopReasonNodes {
		t.Errorf("stop reason = %q, want %q", res.Stats.StopReason, StopReasonNodes)
	}
}
