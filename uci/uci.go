package uci

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	movegen "github.com/AdamGriffiths31/ChessEngine/moveGen"
	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

func Uci(pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) {
	info.GameMode = data.UCIMode
	fmt.Println("id name MyGoEngine")
	fmt.Println("id author Adam")
	fmt.Println("uciok")
	fmt.Printf("engine book: %v\n", data.EngineSettings.UseBook)
	data.EngineSettings.UseBook = false
	board.ParseFEN(data.StartFEN, pos)

	reader := bufio.NewReader(os.Stdin)
	inputCh := make(chan string)

	go func() {
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				panic(fmt.Errorf("UCI  Mode reader loop: %v", err))
			}

			input = strings.TrimSpace(input)
			inputCh <- input
		}
	}()

	for input := range inputCh {

		text := strings.TrimSpace(input)
		fmt.Printf("debug: text %v\n", text)
		if text == "uci" {
			fmt.Println("id name MyGoEngine")
			fmt.Println("id author Me")
			fmt.Println("uciok")
		} else if text == "isready" {
			fmt.Println("readyok")
		} else if strings.HasPrefix(text, "position") {
			parsePosition(text, pos)
		} else if text == "ucinewgame" {
			movegen.ClearTable(table.HashTable)
			board.ParseFEN(data.StartFEN, pos)
		} else if strings.HasPrefix(text, "go") {
			parseGo(text, info, pos, table)
		} else if strings.HasPrefix(text, "setoption") {
			parseOption(text, info, pos)
		} else if text == "print" {
			io.PrintBoard(pos)
		} else if text == "stop" {
			info.ForceStop = true
			fmt.Printf("debug: ForceStop %v\n", info.ForceStop)
		} else if text == "run" {
			data.EngineSettings.UseBook = false
			board.ParseFEN(data.StartFEN, pos)
			parseGo("go infinite", info, pos, table)
		} else if text == "quit" {
			info.Quit = data.True
			break
		}
	}
}

func parseOption(line string, info *data.SearchInfo, pos *data.Board) {
	tokens := strings.Split(line, " ")

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "book":
			parseBook(tokens[i+1], info, pos)
		}
	}
}

func parseGo(line string, info *data.SearchInfo, pos *data.Board, table *data.PvHashTable) {
	tokens := strings.Split(line, " ")
	info.MoveTime = -1
	info.MovesToGo = 30
	info.Depth = -1
	info.Time = -1

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "binc":
			parseInc(tokens[i+1], data.Black, pos, info)
		case "winc":
			parseInc(tokens[i+1], data.White, pos, info)
		case "wtime":
			parseTime(tokens[i+1], data.White, pos, info)
		case "btime":
			parseTime(tokens[i+1], data.Black, pos, info)
		case "movestogo":
			parseMovesToGo(tokens[i+1], pos, info)
		case "movetime":
			parseMoveTime(tokens[i+1], pos, info)
		case "depth":
			parseDepth(tokens[i+1], info)
		}
	}

	info.StartTime = util.GetTimeMs()

	if info.MoveTime != -1 {
		info.TimeSet = data.True
		info.MovesToGo = 1
		info.StopTime = info.StartTime + int64(info.MoveTime)
	} else if info.Time != -1 {
		info.TimeSet = data.True
		info.MovesToGo = 30
		time := info.Time / info.MovesToGo
		time -= 50
		info.Time = time
		info.StopTime = info.StartTime + int64(time) + int64(info.Inc)
	}

	if info.Depth == -1 || info.Depth > data.MaxDepth {
		info.Depth = data.MaxDepth
	}

	fmt.Printf("time:%d start:%d stop:%d depth:%d timeset:%v\n", info.Time, info.StartTime, info.StopTime, info.Depth, info.TimeSet)

	go search.SearchPosition(pos, info, table)
}

func parseBook(line string, info *data.SearchInfo, pos *data.Board) {
	switch line {
	case "true":
		data.EngineSettings.UseBook = true
		fmt.Printf("book turned on\n")
	case "false":
		data.EngineSettings.UseBook = false
		fmt.Printf("book turned off\n")
	default:
		fmt.Printf("Unknown book command expected true / false ")
	}
}

func parseInc(token string, side int, pos *data.Board, info *data.SearchInfo) {
	inc, _ := strconv.Atoi(token)
	if pos.Side == side {
		info.Inc = inc
	}
}

func parseTime(token string, side int, pos *data.Board, info *data.SearchInfo) {
	time, _ := strconv.Atoi(token)
	if pos.Side == side {
		info.Time = time
	}
}

func parseMovesToGo(token string, pos *data.Board, info *data.SearchInfo) {
	movesToGo, _ := strconv.Atoi(token)
	info.MovesToGo = movesToGo
}

func parseMoveTime(token string, pos *data.Board, info *data.SearchInfo) {
	moveTime, _ := strconv.Atoi(token)
	info.MoveTime = moveTime
}

func parseDepth(token string, info *data.SearchInfo) {
	depth, _ := strconv.Atoi(token)
	info.Depth = depth
}

func parsePosition(lineIn string, pos *data.Board) {
	parts := strings.Split(lineIn, " ")

	if len(parts) < 2 {
		panic(fmt.Errorf("UCI parsePosition: unexpected length %v", lineIn))
	}

	if parts[1] == "startpos" {
		fmt.Printf("startpos called\n\n")
		board.ParseFEN(data.StartFEN, pos)
	}

	if parts[1] == "fen" {
		fen := strings.Join(parts[2:], " ")
		board.ParseFEN(fen, pos)
	}

	startIndex := 0
	for i, v := range parts {
		if v == "moves" {
			startIndex = i
		}
	}

	if startIndex != 0 {
		for i := startIndex + 1; i < len(parts); i++ {
			move := movegen.ParseMove([]byte(parts[i]), pos)
			if move == data.NoMove {
				io.PrintBoard(pos)
				fmt.Printf("UCI move error: Parsing UCI (%v) (%v) %v - %v\n", parts[i], lineIn, move, io.PrintMove(move))
			}
			movegen.MakeMove(move, pos)
			pos.Play = 0
		}

	}
}
