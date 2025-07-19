package moves

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestMoveListPool(t *testing.T) {
	// Test basic pool functionality - that we can get and return objects
	ml1 := GetMoveList()
	if ml1 == nil {
		t.Fatal("GetMoveList returned nil")
	}
	
	// Add some data
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	ml1.AddMove(move)
	
	// Return to pool
	ReleaseMoveList(ml1)
	
	// Get another - verify it's clean
	ml2 := GetMoveList()
	if ml2.Count != 0 {
		t.Errorf("Expected clean list from pool, got count %d", ml2.Count)
	}
	
	// Test multiple gets/releases
	lists := make([]*MoveList, 10)
	for i := 0; i < 10; i++ {
		lists[i] = GetMoveList()
		lists[i].AddMove(move)
	}
	
	for _, ml := range lists {
		ReleaseMoveList(ml)
	}
	
	// All should work without panics
	t.Log("Pool test completed successfully")
}

func TestMoveListPoolBasics(t *testing.T) {
	// Test getting from pool
	ml1 := GetMoveList()
	if ml1 == nil {
		t.Fatal("GetMoveList returned nil")
	}
	
	if ml1.Count != 0 {
		t.Errorf("Expected empty list, got count %d", ml1.Count)
	}
	
	if len(ml1.Moves) != 0 {
		t.Errorf("Expected empty moves slice, got length %d", len(ml1.Moves))
	}
	
	// Add some moves
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	ml1.AddMove(move)
	
	if ml1.Count != 1 {
		t.Errorf("Expected count 1, got %d", ml1.Count)
	}
	
	// Release back to pool
	ReleaseMoveList(ml1)
	
	// Get another one - should be clean
	ml2 := GetMoveList()
	if ml2.Count != 0 {
		t.Errorf("Expected clean list from pool, got count %d", ml2.Count)
	}
	
	ReleaseMoveList(ml2)
}

func TestMoveListPoolClear(t *testing.T) {
	ml := GetMoveList()
	
	// Add several moves
	for i := 0; i < 10; i++ {
		move := board.Move{
			From: board.Square{File: i % 8, Rank: 0},
			To:   board.Square{File: i % 8, Rank: 1},
		}
		ml.AddMove(move)
	}
	
	if ml.Count != 10 {
		t.Errorf("Expected 10 moves, got %d", ml.Count)
	}
	
	// Clear should reset everything
	ml.Clear()
	
	if ml.Count != 0 {
		t.Errorf("Expected 0 moves after clear, got %d", ml.Count)
	}
	
	if len(ml.Moves) != 0 {
		t.Errorf("Expected empty slice after clear, got length %d", len(ml.Moves))
	}
	
	// Capacity should be preserved
	if cap(ml.Moves) == 0 {
		t.Error("Expected capacity to be preserved after clear")
	}
	
	ReleaseMoveList(ml)
}

func TestMoveListPoolCapacityLimit(t *testing.T) {
	ml := GetMoveList()
	
	// Add many moves to exceed the pool limit
	for i := 0; i < 600; i++ {
		move := board.Move{
			From: board.Square{File: i % 8, Rank: 0},
			To:   board.Square{File: i % 8, Rank: 1},
		}
		ml.AddMove(move)
	}
	
	// This should not go back to the pool due to capacity limit
	originalCap := cap(ml.Moves)
	ReleaseMoveList(ml)
	
	// Get a new one from pool
	ml2 := GetMoveList()
	
	// Should be a different instance with smaller capacity
	if cap(ml2.Moves) >= originalCap {
		t.Errorf("Expected smaller capacity from pool, got %d vs %d", 
			cap(ml2.Moves), originalCap)
	}
	
	ReleaseMoveList(ml2)
}

func BenchmarkMoveListAllocation(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ml := &MoveList{
				Moves: make([]board.Move, 0, 256),
				Count: 0,
			}
			_ = ml
		}
	})
	
	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ml := GetMoveList()
			ReleaseMoveList(ml)
		}
	})
}

func BenchmarkMoveGeneration(b *testing.B) {
	positions := []struct {
		name string
		fen  string
	}{
		{"Initial", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"Kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -"},
		{"Endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -"},
	}
	
	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			board, _ := board.FromFEN(pos.fen)
			gen := NewGenerator()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				moves := gen.GenerateAllMoves(board, White)
				// Ensure moves is used to prevent optimization
				if moves.Count < 0 {
					b.Fatal("Impossible")
				}
				ReleaseMoveList(moves)
			}
		})
	}
}