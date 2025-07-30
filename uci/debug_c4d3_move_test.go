package uci

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestDebugC4D3Move(t *testing.T) {
	// This is the exact position where c4d3 was generated as "legal" but rejected by cutechess
	fen := "r3k2r/1pp2ppp/p1n1bn2/4b3/PPQ1p2P/3q1p2/3K4/RNBQ2NR w kq - 0 16"
	
	// Load the position
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	fmt.Printf("Original position: %s\n", fen)
	
	// Create the move c4d3 manually
	c4d3 := board.Move{
		From:        board.Square{File: 2, Rank: 3}, // c4
		To:          board.Square{File: 3, Rank: 2}, // d3
		Piece:       board.WhiteQueen,
		Captured:    board.BlackQueen,
		Promotion:   board.Empty,
		IsCapture:   true,
		IsCastling:  false,
		IsEnPassant: false,
	}
	
	// Test the make/unmake process manually
	generator := moves.NewGenerator()
	defer generator.Release()
	
	moveExecutor := &moves.MoveExecutor{}
	
	fmt.Printf("Before move: White king at d2, Black queen at d3\n")
	
	// Check if white king is initially in check
	kingSquare := board.FileRankToSquare(3, 1) // d2
	initialCheck := b.IsSquareAttackedByColor(kingSquare, board.BitboardBlack)
	fmt.Printf("Initial check status: %v\n", initialCheck)
	
	// Make the move
	fmt.Printf("Making move c4d3...\n")
	history := moveExecutor.MakeMove(b, c4d3, func(b *board.Board, move board.Move) {
		// Empty callback to test what happens
	})
	
	// Check if king is in check after the move
	afterMoveCheck := b.IsSquareAttackedByColor(kingSquare, board.BitboardBlack)
	fmt.Printf("After c4d3: King in check = %v\n", afterMoveCheck)
	
	// Show the position after move
	fmt.Printf("Position after c4d3: %s\n", b.ToFEN())
	
	// Unmake the move
	moveExecutor.UnmakeMove(b, history)
	fmt.Printf("After unmake: %s\n", b.ToFEN())
	
	// The key test: c4d3 should be illegal because it leaves the king in check
	if !afterMoveCheck {
		t.Errorf("c4d3 should leave the king in check, but it doesn't!")
	}
}