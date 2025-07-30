package uci

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestReproduceGameSequence(t *testing.T) {
	// The exact game sequence from the PGN that led to the illegal move
	gameMoves := []string{
		"c2c4", "e7e5", "a2a3", "b8c6", "b2b3", "g8f6", "d2d3", "d7d5",
		"e2e3", "c8e6", "f2f3", "a7a6", "g2g3", "f8d6", "h2h3", "d8d7",
		"a3a4", "e5e4", "b3b4", "d6g3", "e1d2", "g3e5", "d3d4", "e4f3",
		"e3e4", "d5e4", "h3h4", "d7d4", "f1d3", "d4d3", // Added the missing move!
	}
	
	// Start from the initial position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to load starting FEN: %v", err)
	}
	
	fmt.Printf("Replaying game moves to reach the position...\n")
	
	// Apply each move in sequence
	moveExecutor := &moves.MoveExecutor{}
	player := moves.White
	
	for i, moveStr := range gameMoves {
		fmt.Printf("Move %d: %s (Player: %v)\n", i+1, moveStr, player)
		
		// Parse the move string (e.g., "c2c4" -> from c2 to c4)
		if len(moveStr) != 4 {
			t.Fatalf("Invalid move format: %s", moveStr)
		}
		
		fromFile := int(moveStr[0] - 'a')
		fromRank := int(moveStr[1] - '1')
		toFile := int(moveStr[2] - 'a') 
		toRank := int(moveStr[3] - '1')
		
		// Get the piece at the from square
		piece := b.GetPiece(fromRank, fromFile)
		captured := b.GetPiece(toRank, toFile)
		
		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       piece,
			Captured:    captured,
			Promotion:   board.Empty,
			IsCapture:   captured != board.Empty,
			IsCastling:  false,
			IsEnPassant: false,
		}
		
		// Make the move
		moveExecutor.MakeMove(b, move, func(b *board.Board, move board.Move) {
			// Empty callback - just apply the move
		})
		
		// Switch players
		if player == moves.White {
			player = moves.Black
		} else {
			player = moves.White
		}
		
		fmt.Printf("  Position after move: %s\n", b.ToFEN())
	}
	
	fmt.Printf("\nFinal position reached: %s\n", b.ToFEN())
	fmt.Printf("Expected position:      r3k2r/1pp2ppp/p1n1bn2/4b3/PPQ1p2P/3q1p2/3K4/RNBQ2NR w kq - 0 16\n")
	
	// Check if white king is in check
	kingSquare := board.FileRankToSquare(3, 1) // d2
	inCheck := b.IsSquareAttackedByColor(kingSquare, board.BitboardBlack)
	fmt.Printf("White king in check: %v\n", inCheck)
	
	// Now test if c4d3 should be legal
	generator := moves.NewGenerator()
	defer generator.Release()
	
	moveList := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(moveList)
	
	fmt.Printf("Legal moves: %d\n", moveList.Count)
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
		toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
		fmt.Printf("  %s%s\n", 
			board.SquareToString(fromSquare),
			board.SquareToString(toSquare))
	}
}