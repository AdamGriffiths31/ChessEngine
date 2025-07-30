package uci

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestIllegalC4D3Bug(t *testing.T) {
	// This is the exact position where c4d3 was generated as "legal" but rejected by cutechess
	fen := "r3k2r/1pp2ppp/p1n1bn2/4b3/PPQ1p2P/3q1p2/3K4/RNBQ2NR w kq - 0 16"
	
	// Load the position
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	// Generate moves
	generator := moves.NewGenerator()
	defer generator.Release()
	
	moveList := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(moveList)
	
	fmt.Printf("Position: %s\n", fen)
	fmt.Printf("Legal moves found: %d\n", moveList.Count)
	
	// Check if c4d3 is incorrectly generated as legal
	c4d3Found := false
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		if move.From.File == 2 && move.From.Rank == 3 && // c4
		   move.To.File == 3 && move.To.Rank == 2 {     // d3
			c4d3Found = true
			fmt.Printf("❌ ILLEGAL MOVE GENERATED: c4d3 (captures queen but leaves king in check)\n")
			break
		}
	}
	
	if c4d3Found {
		t.Errorf("Illegal move c4d3 was generated as legal!")
	}
	
	// List all moves for debugging
	fmt.Println("All generated moves:")
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
		toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
		fmt.Printf("  %s%s\n", 
			board.SquareToString(fromSquare),
			board.SquareToString(toSquare))
	}
	
	// The king should only be able to move to e1 in this position
	if moveList.Count != 1 {
		t.Errorf("Expected 1 legal move in this position, got %d", moveList.Count)
	}
}