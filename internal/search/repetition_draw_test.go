package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

func TestRepetitionHistory_DetectsRepeatViaMakeUnmake(t *testing.T) {
	const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b, err := board.FromFEN(startFEN)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	b.SetHashUpdater(engine)
	b.InitializeHashFromPosition(engine.zobrist.HashPosition)

	rootHash := b.GetHash()
	engine.setupRepetitionHistory(rootHash)

	nb1c3 := board.Move{From: sq(1, 0), To: sq(2, 2), Promotion: board.Empty}
	nc3b1 := board.Move{From: sq(2, 2), To: sq(1, 0), Promotion: board.Empty}
	nb8c6 := board.Move{From: sq(1, 7), To: sq(2, 5), Promotion: board.Empty}
	nc6b8 := board.Move{From: sq(2, 5), To: sq(1, 7), Promotion: board.Empty}

	cycle := []board.Move{nb1c3, nb8c6, nc3b1, nc6b8}
	moves := append(append([]board.Move{}, cycle...), cycle...)

	wantRepeatAtPly := map[int]bool{
		1: false, 2: false, 3: false,
		4: true, 5: true, 6: true, 7: true, 8: true,
	}

	undos := make([]board.MoveUndo, 0, len(moves))
	for i, mv := range moves {
		ply := i + 1
		undo, err := b.MakeMoveWithUndo(mv)
		if err != nil {
			t.Fatalf("ply %d: move %+v failed: %v", ply, mv, err)
		}
		undos = append(undos, undo)

		hash := b.GetHash()
		engine.addHistory(hash)

		got := engine.isDrawByRepetition(hash, MaxGamePly)
		want := wantRepeatAtPly[ply]
		if got != want {
			t.Errorf("ply %d: isDrawByRepetition = %v, want %v", ply, got, want)
		}
		if ply == 4 || ply == 8 {
			if hash != rootHash {
				t.Fatalf("ply %d: hash = %d, want it to match root hash %d (shuffle should return to the exact starting position)", ply, hash, rootHash)
			}
		}
	}

	for i := len(moves) - 1; i >= 0; i-- {
		b.UnmakeMove(undos[i])
		engine.removeHistory()
	}
	if engine.zobristHistoryPly != 0 {
		t.Errorf("zobristHistoryPly = %d after fully unwinding, want 0", engine.zobristHistoryPly)
	}
}

func TestFindBestMove_SeedsRepetitionHistoryFromConfig(t *testing.T) {
	const fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	b.SetHashUpdater(engine)
	b.InitializeHashFromPosition(engine.zobrist.HashPosition)
	rootHash := b.GetHash()

	seeded := []uint64{0x1111111111111111, 0x2222222222222222, rootHash}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine.FindBestMove(ctx, b, movegen.White, Config{
		MaxDepth:          3,
		RepetitionHistory: seeded,
	})

	if engine.zobristHistoryPly != uint16(len(seeded)-1) {
		t.Fatalf("zobristHistoryPly = %d after search, want %d (seeded history should survive the search intact)",
			engine.zobristHistoryPly, len(seeded)-1)
	}
	for i, want := range seeded {
		if got := engine.zobristHistory[i]; got != want {
			t.Errorf("zobristHistory[%d] = %d, want %d (seeded value)", i, got, want)
		}
	}
}

func TestFindBestMove_ReturnsDrawScoreOnUnavoidableRepetition(t *testing.T) {
	const fen = "7k/8/8/4n3/4N3/8/8/7K w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, movegen.White, Config{MaxDepth: 6})

	const drawTolerance = 30
	if absScore := result.Score; absScore > drawTolerance || absScore < -drawTolerance {
		t.Errorf("Score = %d, want within +/-%d of eval.DrawScore (%d): bare K+N vs K+N cannot force progress",
			result.Score, drawTolerance, eval.DrawScore)
	}
}

func TestFindBestMove_WinningSideAvoidsRepetition(t *testing.T) {
	const fen = "7k/8/8/4n3/4N3/8/8/3Q3K w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, movegen.White, Config{MaxDepth: 6})

	const winningThreshold = eval.EvaluationScore(500)
	if result.Score <= winningThreshold {
		t.Errorf("Score = %d, want > %d: the extra queen must keep the search from settling for a repetition draw", result.Score, winningThreshold)
	}
	if result.Score == eval.DrawScore {
		t.Error("Score = DrawScore: winning side walked into (or accepted) a repetition instead of playing for the win")
	}
}

func TestIsDrawByRepetition_GameHistoryNeedsThreefold(t *testing.T) {
	const (
		p, q, r, x, y = 0xA1, 0xA2, 0xA3, 0xA4, 0xA5
	)

	tests := []struct {
		name    string
		history []uint64
		path    []uint64
		want    bool
	}{
		{"seen once before the root", []uint64{p, x, r}, []uint64{p}, false},
		{"seen twice before the root", []uint64{p, x, p, y, r}, []uint64{p}, true},
		{"repeated within the search path", []uint64{r}, []uint64{q, x, q}, true},
		{"returns to the root", []uint64{x, r}, []uint64{q, r}, true},
		{"new position", []uint64{p, x, r}, []uint64{q}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewMinimaxEngine()
			engine.seedRepetitionHistory(tt.history)

			var got bool
			for _, hash := range tt.path {
				engine.addHistory(hash)
				got = engine.isDrawByRepetition(hash, MaxGamePly)
			}
			if got != tt.want {
				t.Errorf("isDrawByRepetition after path %v = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
