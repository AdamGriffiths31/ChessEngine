package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/AdamGriffiths31/ChessEngine/2.0/search"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	search2 "github.com/AdamGriffiths31/ChessEngine/search"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	search.RunBenchmark()
	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	input, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		panic(fmt.Errorf("main reader loop: %v", err))
	// 	}
	// 	input = strings.TrimSpace(input)

	// 	if input == "uci" {
	// 		uci := uci.NewUCI()
	// 		uci.UCIMode()
	// 		continue
	// 	}

	// 	if input == "b" {
	// 		search.RunBenchmark()
	// 	}

	// 	if input == "quit" {
	// 		break
	// 	}
	// }
}

func testOld() {
	table := &data.PVTable{}
	hash := &data.PvHashTable{HashTable: table}
	data.InitPvTable(hash.HashTable)
	pos := data.NewBoardPos()
	board.ParseFEN(data.StartFEN, pos)
	info := &data.SearchInfo{}
	info.WorkerNumber = 0
	info.Depth = 10
	info.PostThinking = false
	info.GameMode = data.ConsoleMode
	search2.SearchPosition(pos, info, hash)

	fmt.Printf("score %v\n", evaluate.EvalPosition(pos))

}

func old() {
	// polyglot.InitPolyBook()
	// table := &data.PVTable{}
	// hash := &data.PvHashTable{HashTable: table}
	// data.InitPvTable(hash.HashTable)
	// pos := data.NewBoardPos()
	// info := &data.SearchInfo{}
	// info.WorkerNumber = 0
	// reader := bufio.NewReader(os.Stdin)
	// fmt.Printf("PVTable: %v entries (%v)\n", hash.HashTable.NumberEntries, hash.HashTable.CurrentAge)

	// for {
	// 	input, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		panic(fmt.Errorf("main reader loop: %v", err))
	// 	}
	// 	input = strings.TrimSpace(input)

	// 	if input == "uci" {
	// 		uci.Uci(pos, info, hash)
	// 		continue
	// 	}

	// 	if input == "xboard" {
	// 		xboard.Xboard(pos, info, hash)
	// 		continue
	// 	}

	// 	if input == "console" {
	// 		consolemode.ConsoleMode(pos, info, hash)
	// 	}

	// 	if input == "b" {
	// 		// search.RunBenchmark()
	// 	}

	// 	if input == "quit" {
	// 		break
	// 	}
	// }
}
