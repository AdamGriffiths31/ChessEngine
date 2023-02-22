package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func main() {
	// p := &engine.PVTable{}
	search := &engine.SearchInfo{}
	search.Depth = 4
	pvTable := &engine.PVTable{}
	b := &engine.Board{PvTable: pvTable}

	engine.ParseFEN(engine.StartFEN, b)

	engine.CheckBoard(b)

	reader := bufio.NewReader(os.Stdin)
	engine.PrintBoard(b)
	engine.InitPvTable(b.PvTable)
	for {
		fmt.Printf("Please enter a move:")
		text, _ := reader.ReadString('\n')
		fmt.Println("You entered:", text)

		move := engine.ParseMove([]byte(text), b, search)
		if move != engine.NoMove {
			fmt.Printf("Storing: %v for %v\n", engine.PrintMove(move), b.PosistionKey)
			engine.StorePvMove(b, move)
			engine.MakeMove(move, b)
			engine.PrintBoard(b)
		}

	}

}
