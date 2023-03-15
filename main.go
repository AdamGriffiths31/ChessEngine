package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/2.0/search"
	"github.com/AdamGriffiths31/ChessEngine/board"
	consolemode "github.com/AdamGriffiths31/ChessEngine/consoleMode"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
	search2 "github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/uci"
	"github.com/AdamGriffiths31/ChessEngine/xboard"
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
	// test := engine.Position{}
	// test.ParseFen(data.StartFEN)
	// test2 := engine.Position{}
	// test2.ParseFen(data.StartFEN)
	// test3 := engine.Position{}
	// test3.ParseFen(data.StartFEN)
	// test4 := engine.Position{}
	// test4.ParseFen(data.StartFEN)
	// h := search.NewEngineHolder(4)
	// h.Engines[0].Position = &test
	// h.Engines[1].Position = &test2
	// h.Engines[2].Position = &test3
	// h.Engines[3].Position = &test4
	// h.ClearForSearch()

	// time1 := time.Now()
	// h.Search(15)
	//util.TimeTrackMilliseconds(time1, "new")
	//fmt.Printf("\n\nstored %v (%v)\n", h.TranspositionTable.Stored, h.NodeCount)
	search.RunBenchmark()
	//time2 := time.Now()

	//testOld()
	//util.TimeTrackMilliseconds(time2, "old")

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
			// search.RunBenchmark()
		}

		if input == "quit" {
			break
		}
	}
}
