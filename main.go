package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/2.0/search"
	"github.com/AdamGriffiths31/ChessEngine/board"
	consolemode "github.com/AdamGriffiths31/ChessEngine/consoleMode"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
	search2 "github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/uci"
	"github.com/AdamGriffiths31/ChessEngine/util"
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
	//defer util.TimeTrackMilliseconds(time.Now(), "Main")
	p := &engine.Position{}
	p.ParseFen(data.StartFEN)
	e := search.Engine{Position: p, IsMainEngine: true}

	// p1 := &engine.Position{}
	// p1.ParseFen(data.StartFEN)
	// e1 := search.Engine{Position: p1}

	// p2 := &engine.Position{}
	// p2.ParseFen(data.StartFEN)
	// e2 := search.Engine{Position: p2}

	// p3 := &engine.Position{}
	// p3.ParseFen(data.StartFEN)
	// e3 := search.Engine{Position: p3}

	time1 := time.Now()
	//list := []*search.Engine{&e, &e1, &e2, &e3}
	list := []*search.Engine{&e}

	h := search.EngineHolder{Engines: list}
	fmt.Printf("score %v\n", e.Position.Evaluate())
	h.Search(9)
	util.TimeTrackMilliseconds(time1, "new")
	fmt.Printf("\n\n")
	time2 := time.Now()
	// e.Position.PrintMoveList(false)
	// p.MakeMove(9427)

	testOld()
	util.TimeTrackMilliseconds(time2, "old")
	// p.Board.PrintBitboard(p.Board.WhitePieces)
	// p.TakeMoveBack(531363, data.Empty)

	// e.Position.Board.PrintBitboard(e.Position.Board.Pieces)
	// e.Position.PrintMoveList(false)

	// _, enPas, CastleRight := p.MakeMove(123054)
	// p.TakeMoveBack(123054, enPas, CastleRight)
	// e.Position.CheckBitboard()
}

func testOld() {
	table := &data.PVTable{}
	hash := &data.PvHashTable{HashTable: table}
	data.InitPvTable(hash.HashTable)
	pos := data.NewBoardPos()
	board.ParseFEN(data.StartFEN, pos)
	info := &data.SearchInfo{}
	info.WorkerNumber = 0
	info.Depth = 9
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
