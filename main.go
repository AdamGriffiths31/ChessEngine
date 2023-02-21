package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func main() {

	b := &engine.Board{}
	engine.ParseFEN("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N w - - 0 1", b)

	engine.CheckBoard(b)

	reader := bufio.NewReader(os.Stdin)

	for {
		engine.PrintBoard(b)
		fmt.Printf("Please enter a move:")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text) // remove leading/trailing white space
		fmt.Println("You entered:", text)
		// if text == "t" {
		// 	fmt.Printf("Take Back\n")
		// 	engine.TakeMoveBack(b)
		// } else {
		move := engine.ParseMove([]byte(text), b)
		if move != engine.NoMove {
			fmt.Printf("Making move\n")
			engine.MakeMove(move, b)
			//}
		}
	}

}
