package main

import (
	"os"

	"github.com/AdamGriffiths31/ChessEngine/uci"
)

func main() {
	engine := uci.NewUCIEngine()
	engine.Run(os.Stdin, os.Stdout)
}

