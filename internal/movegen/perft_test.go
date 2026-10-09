package movegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

type perftPosition struct {
	Name   string `json:"name"`
	FEN    string `json:"fen"`
	Depths []struct {
		Depth int   `json:"depth"`
		Nodes int64 `json:"nodes"`
	} `json:"depths"`
}

func loadPerftPositions(t *testing.T) []perftPosition {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "perft_tests.json"))
	if err != nil {
		t.Fatalf("reading perft data: %v", err)
	}
	var file struct {
		Positions []perftPosition `json:"positions"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parsing perft data: %v", err)
	}
	return file.Positions
}

// perft counts the leaf nodes reachable in depth plies, making and unmaking
// moves with the same board methods the search uses.
func perft(t *testing.T, b *board.Board, depth int, player Player, gen *Generator) int64 {
	moves := gen.GenerateAllMoves(b, player)
	defer ReleaseMoveList(moves)

	if depth == 1 {
		return int64(moves.Count)
	}

	var nodes int64
	for _, move := range moves.Moves {
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			t.Fatalf("generated move %s%s failed to apply: %v", move.From, move.To, err)
		}
		nodes += perft(t, b, depth-1, opposite(player), gen)
		b.UnmakeMove(undo)
	}
	return nodes
}

// TestPerft_StandardPositions checks move generation against published node
// counts. Set PERFT_MAX_DEPTH to cap the depth; -short caps it at 3.
func TestPerft_StandardPositions(t *testing.T) {
	maxDepth := 5
	if env := os.Getenv("PERFT_MAX_DEPTH"); env != "" {
		if depth, err := strconv.Atoi(env); err == nil {
			maxDepth = depth
		}
	}
	if testing.Short() && maxDepth > 3 {
		maxDepth = 3
	}

	gen := NewGenerator()
	for _, position := range loadPerftPositions(t) {
		t.Run(position.Name, func(t *testing.T) {
			b, err := board.FromFEN(position.FEN)
			if err != nil {
				t.Fatalf("FromFEN(%q): %v", position.FEN, err)
			}
			player := playerFromSide(b.GetSideToMove())

			for _, want := range position.Depths {
				t.Run(fmt.Sprintf("depth_%d", want.Depth), func(t *testing.T) {
					if want.Depth > maxDepth {
						t.Skipf("depth %d exceeds the cap of %d", want.Depth, maxDepth)
					}
					if got := perft(t, b, want.Depth, player, gen); got != want.Nodes {
						t.Errorf("got %d nodes, want %d", got, want.Nodes)
					}
				})
			}
		})
	}
}

func playerFromSide(side string) Player {
	if side == "w" {
		return White
	}
	return Black
}
