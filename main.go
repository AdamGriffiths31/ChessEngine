// Package main provides the entry point for the chess engine application.
package main

import (
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/game/modes"
)

func showBanner() {
	fmt.Print(`
   ╔═════════════════════════════════════════════════════╗
   ║                                                     ║
   ║   ██████   ██████  █    █  ██████  ██████  ██████   ║
   ║   █        █       █    █  █       █       █        ║
   ║   █  ████  █       ██████  █████   ██████  ██████   ║
   ║   █    █   █       █    █  █            █       █   ║
   ║   ██████   ██████  █    █  ██████  ██████  ██████   ║
   ║                                                     ║
   ║     ♔ ♕ ♖ ♗ ♘ ♙    Chess Engine    ♙ ♘ ♗ ♖ ♕ ♔      ║
   ║                                                     ║
   ╚═════════════════════════════════════════════════════╝
`)
}

func main() {
	showBanner()
	fmt.Println("Select game mode:")
	fmt.Println("1. Benchmark Engine")
	fmt.Println("2. STS Benchmark")
	fmt.Println("3. Player vs Computer")
	fmt.Println("4. Manual Play (Player vs Player)")
	fmt.Print("\nEnter choice (1, 2, 3, or 4): ")

	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	var err error
	switch choice {
	case 1:
		benchmarkMode := modes.NewBenchmarkMode()
		err = benchmarkMode.Run()
	case 2:
		stsMode := modes.NewSTSMode()
		err = stsMode.Run()
	case 3:
		computerMode := modes.NewComputerMode()
		err = computerMode.Run()
	case 4:
		manualMode := modes.NewManualMode()
		err = manualMode.Run()
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error running game: %v\n", err)
		os.Exit(1)
	}
}
