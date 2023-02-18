package game

import (
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/fen"
	"github.com/AdamGriffiths31/ChessEngine/pieces"
)

func View(pos string) {
	grid := fen.Grid(pos)
	var s strings.Builder
	s.WriteString(board.BuildTopBorder())

	for currentRow := board.FirstRow; currentRow <= board.LastRow; currentRow++ {
		row := pieces.DisplayRow(grid[currentRow])
		s.WriteString((fmt.Sprintf(" %d ", board.LastRow-currentRow)) + board.Vertical)
		for _, value := range row {
			s.WriteString(fmt.Sprintf(" %s %s", value.Display(), board.Vertical))
		}
		s.WriteRune('\n')
		if currentRow != board.LastRow {
			s.WriteString(board.BuildMiddleBorder())
		}
	}
	s.WriteString(board.BuildBottomBorder())
	s.WriteString(board.BuildBottomLabels())
	fmt.Println(s.String())
}
