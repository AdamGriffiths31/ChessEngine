package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	consolemode "github.com/AdamGriffiths31/ChessEngine/consoleMode"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/uci"
	"github.com/AdamGriffiths31/ChessEngine/xboard"
)

func main() {

	p := engine.Position{}
	p.ParseFen(data.StartFEN)
	fmt.Printf("en pas %v\n", io.SquareString(p.EnPassant))

	//p.MakeMove(4892)
	//p.MakeMove(6985)

	p.Board.PrintBitboard(p.Board.WhitePieces)

	p.PrintMoveList()

	//p.Board.PrintBitboard(p.Board.WhitePieces)
	// p.TakeMoveBack(531363, data.Empty)

}

func old() {
	polyglot.InitPolyBook()
	table := &data.PVTable{}
	hash := &data.PvHashTable{HashTable: table}
	data.InitPvTable(hash.HashTable)
	pos := data.NewBoardPos()
	info := &data.SearchInfo{}
	info.WorkerNumber = 0
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("PVTable: %v entries (%v)\n", hash.HashTable.NumberEntries, hash.HashTable.CurrentAge)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("main reader loop: %v", err))
		}
		input = strings.TrimSpace(input)

		if input == "uci" {
			uci.Uci(pos, info, hash)
			continue
		}

		if input == "xboard" {
			xboard.Xboard(pos, info, hash)
			continue
		}

		if input == "console" {
			consolemode.ConsoleMode(pos, info, hash)
		}

		if input == "b" {
			search.RunBenchmark()
		}

		if input == "quit" {
			break
		}
	}
}
