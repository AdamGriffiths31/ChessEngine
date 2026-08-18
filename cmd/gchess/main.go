// Package main provides the entry point for the chess engine application.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/cmd/gchess/modes"
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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	showBanner()
	fmt.Println("Select mode:")
	fmt.Println("1. STS Benchmark")
	fmt.Println("2. Elo Benchmark")
	fmt.Print("\nEnter choice (1-2): ")

	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	var err error
	switch choice {
	case 1:
		stsMode := modes.NewSTSMode()
		err = stsMode.Run()
	case 2:
		eloMode := modes.NewEloBenchmarkMode()
		err = eloMode.Run()
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error running game: %v\n", err)
		os.Exit(1)
	}
}
