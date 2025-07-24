package main

import (
	"fmt"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

func main() {
	fmt.Println("=== Chess Engine Debug Runner ===")
	fmt.Println("Testing the missing pawn bug fix...")
	
	// Test the bug reproduction
	fmt.Println("\n1. Running bug reproduction test:")
	board.DebugReproduceBug()
	
	// Test basic debug functions
	fmt.Println("\n2. Running basic debug tests:")
	board.RunBasicDebugTests()
	
	fmt.Println("\n=== Debug Runner Complete ===")
}