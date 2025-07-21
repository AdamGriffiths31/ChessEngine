package main

import (
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/game/modes"
)

func main() {
	fmt.Println("Chess Engine")
	fmt.Println("============")
	fmt.Println("\nSelect game mode:")
	fmt.Println("1. Manual Play (Player vs Player)")
	fmt.Println("2. Player vs Computer")
	fmt.Print("\nEnter choice (1 or 2): ")

	var choice int
	fmt.Scanln(&choice)

	var err error
	switch choice {
	case 1:
		manualMode := modes.NewManualMode()
		err = manualMode.Run()
	case 2:
		computerMode := modes.NewComputerMode()
		err = computerMode.Run()
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error running game: %v\n", err)
		os.Exit(1)
	}
}
