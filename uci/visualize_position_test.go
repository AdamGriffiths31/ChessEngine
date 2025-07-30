package uci

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestVisualizePosition(t *testing.T) {
	// The position where cutechess-cli rejects c4d3
	fen := "r3k2r/1pp2ppp/p1n1bn2/4b3/PPQ1p2P/3q1p2/3K4/RNBQ2NR w kq - 0 16"
	
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	fmt.Printf("Position: %s\n", fen)
	fmt.Printf("Visual representation:\n")
	
	// Print the board from White's perspective (rank 8 to rank 1)  
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	
	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			switch piece {
			case board.WhiteKing:
				fmt.Printf("♔ ")
			case board.WhiteQueen:
				fmt.Printf("♕ ")
			case board.WhiteRook:
				fmt.Printf("♖ ")
			case board.WhiteBishop:
				fmt.Printf("♗ ")
			case board.WhiteKnight:
				fmt.Printf("♘ ")
			case board.WhitePawn:
				fmt.Printf("♙ ")
			case board.BlackKing:
				fmt.Printf("♚ ")
			case board.BlackQueen:
				fmt.Printf("♛ ")
			case board.BlackRook:
				fmt.Printf("♜ ")
			case board.BlackBishop:
				fmt.Printf("♝ ")
			case board.BlackKnight:
				fmt.Printf("♞ ")
			case board.BlackPawn:
				fmt.Printf("♟ ")
			default:
				fmt.Printf(". ")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("  ")
	for _, file := range files {
		fmt.Printf("%s ", file)
	}
	fmt.Printf("\n\n")
	
	// Show the key pieces
	fmt.Printf("Key pieces:\n")
	fmt.Printf("  White King (♔) on d2 (file=%d, rank=%d)\n", 3, 1)
	fmt.Printf("  White Queen (♕) on c4 (file=%d, rank=%d)\n", 2, 3)  
	fmt.Printf("  Black Queen (♛) on d3 (file=%d, rank=%d)\n", 3, 2)
	
	fmt.Printf("\nMove c4d3 means:\n")
	fmt.Printf("  White Queen moves from c4 to d3\n")
	fmt.Printf("  This captures the Black Queen on d3\n")
	fmt.Printf("  This should resolve the check on the White King\n")
}

// Helper to convert piece enum to symbol
func pieceToSymbol(piece board.Piece) string {
	switch piece {
	case board.WhiteKing:
		return "♔"
	case board.WhiteQueen:
		return "♕"
	case board.WhiteRook:
		return "♖"
	case board.WhiteBishop:
		return "♗"
	case board.WhiteKnight:
		return "♘"
	case board.WhitePawn:
		return "♙"
	case board.BlackKing:
		return "♚"
	case board.BlackQueen:
		return "♛"
	case board.BlackRook:
		return "♜"
	case board.BlackBishop:
		return "♝"
	case board.BlackKnight:
		return "♞"
	case board.BlackPawn:
		return "♟"
	default:
		return "."
	}
}