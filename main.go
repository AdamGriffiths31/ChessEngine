package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	consolemode "github.com/AdamGriffiths31/ChessEngine/consoleMode"
	"github.com/AdamGriffiths31/ChessEngine/data"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
	"github.com/AdamGriffiths31/ChessEngine/uci"
	"github.com/AdamGriffiths31/ChessEngine/xboard"
)

func main() {
	polyglot.InitPolyBook()
	pos := data.NewBoardPos()
	info := &data.SearchInfo{}

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("main reader loop: %v", err))
		}
		input = strings.TrimSpace(input)

		if input == "uci" {
			uci.Uci(pos, info)
			continue
		}

		if input == "xboard" {
			xboard.Xboard(pos, info)
			continue
		}

		if input == "console" {
			consolemode.ConsoleMode(pos, info)
		}
	}
}
