package main

import (
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/game/modes"
)

func main() {
	// Launch Game Mode 1: Manual Play
	manualMode := modes.NewManualMode()
	
	err := manualMode.Run()
	if err != nil {
		fmt.Printf("Error running game: %v\n", err)
		os.Exit(1)
	}
}
