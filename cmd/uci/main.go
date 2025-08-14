// Package main implements the ChessEngine UCI protocol interface.
package main

import (
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/uci"
)

func main() {
	engine := uci.NewUCIEngine()
	if err := engine.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "UCI engine failed: %v\n", err)
		os.Exit(1)
	}
}
