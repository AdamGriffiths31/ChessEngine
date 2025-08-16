package ui

import (
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// RenderBoard renders a chess board to a string representation
func RenderBoard(b *board.Board) string {
	if b == nil {
		return "ERROR: Board is nil"
	}

	var lines []string

	lines = append(lines, "  a b c d e f g h")

	for displayRank := 0; displayRank < 8; displayRank++ {
		rankNumber := 8 - displayRank
		boardRank := rankNumber - 1 // Convert chess rank to array index (rank 1 = index 0)
		line := fmt.Sprintf("%d", rankNumber)

		for file := 0; file < 8; file++ {
			piece := b.GetPiece(boardRank, file)
			line += fmt.Sprintf(" %c", piece)
		}

		line += fmt.Sprintf(" %d", rankNumber)
		lines = append(lines, line)
	}

	lines = append(lines, "  a b c d e f g h")

	return strings.Join(lines, "\n")
}

// RenderBoardFromFEN renders a chess board from FEN notation to string
func RenderBoardFromFEN(fen string) string {
	b, err := board.FromFEN(fen)
	if err != nil {
		if strings.Contains(err.Error(), "must have exactly 8 ranks") {
			return "ERROR: Invalid FEN - too many ranks"
		}
		return fmt.Sprintf("ERROR: %s", err.Error())
	}

	return RenderBoard(b)
}
