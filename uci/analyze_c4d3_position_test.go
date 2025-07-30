package uci

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestAnalyzeC4D3Position(t *testing.T) {
	// This is the exact position where c4d3 was generated
	fen := "r3k2r/1pp2ppp/p1n1bn2/4b3/PPQ1p2P/3q1p2/3K4/RNBQ2NR w kq - 0 16"
	
	// Load the position
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	fmt.Printf("Position: %s\n", fen)
	fmt.Printf("Analysis:\n")
	
	// Check what pieces are where
	whiteKingSquare := board.FileRankToSquare(3, 1) // d2
	whiteQueenSquare := board.FileRankToSquare(2, 3) // c4
	blackQueenSquare := board.FileRankToSquare(3, 2) // d3
	
	whiteKingPiece := b.GetPieceOnSquare(whiteKingSquare)
	whiteQueenPiece := b.GetPieceOnSquare(whiteQueenSquare)
	blackQueenPiece := b.GetPieceOnSquare(blackQueenSquare)
	
	fmt.Printf("  White King on d2: %d\n", whiteKingPiece)
	fmt.Printf("  White Queen on c4: %d\n", whiteQueenPiece)
	fmt.Printf("  Black Queen on d3: %d\n", blackQueenPiece)
	
	// Check if White King is in check initially
	isInCheck := b.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
	fmt.Printf("  White King in check: %v\n", isInCheck)
	
	// Find what's attacking the king
	if isInCheck {
		fmt.Printf("  King is in check - legal moves must either:\n")
		fmt.Printf("    1. Move the king out of check\n")
		fmt.Printf("    2. Block the check\n")
		fmt.Printf("    3. Capture the attacking piece\n")
	}
	
	// Generate legal moves to see what the engine thinks
	generator := moves.NewGenerator()
	defer generator.Release()
	
	moveList := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(moveList)
	
	fmt.Printf("  Legal moves according to engine: %d\n", moveList.Count)
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
		toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
		fmt.Printf("    %s->%s\n", 
			board.SquareToString(fromSquare),
			board.SquareToString(toSquare))
	}
	
	// Test if c4d3 is actually legal by chess rules
	// If king is in check from d3, then capturing on d3 should be legal
	fmt.Printf("\nChess rules analysis:\n")
	if isInCheck {
		// Check if the black queen on d3 is what's giving check
		fmt.Printf("  Since king is in check, c4d3 (capturing the checking piece) should be LEGAL\n")
		fmt.Printf("  cutechess-cli might be wrong or there's a different issue\n")
	}
}