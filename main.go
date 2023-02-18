package main

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func main() {

	engine.SetSquares()

	for index := 0; index < 120; index++ {
		if index%10 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%5d", engine.Sqaure120ToSquare64[index])
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")

	for index := 0; index < 64; index++ {
		if index%8 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%5d", engine.Sqaure64ToSquare120[index])
	}
}
