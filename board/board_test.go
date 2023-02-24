package board_test

import (
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
)

func TestMirror(t *testing.T) {
	// Open the file
	startTime := time.Now()

	file, err := os.Open("mirror.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pos := data.NewBoardPos()
	scanner := bufio.NewScanner(file)
	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		board.ParseFEN(line, pos)
		counter++
		eval1 := evaluate.EvalPosistion(pos)
		board.MirrorBoard(pos)
		eval2 := evaluate.EvalPosistion(pos)

		if eval1 != eval2 {
			t.Errorf("got %d, want %d for %v", eval2, eval1, line)
		}
	}
	duration := time.Since(startTime)
	fmt.Printf("\n\nTime elapsed:%d for %d", duration, counter)

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
