package uci

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

type UCI struct {
	engineHolder *search.EngineHolder
}

func NewUCI() *UCI {
	return &UCI{
		search.NewEngineHolder(6),
	}
}

func (uci *UCI) UCIMode() {
	search.InitPolyBook(uci.engineHolder)
	var game engine.Game = engine.ParseFen(data.StartFEN)
	info := data.SearchInfo{}
	uci.printUCIok()

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
			uci.printUCIok()
		} else if text == "isready" {
			fmt.Println("readyok")
		} else if text == "ucinewgame" {
			game = engine.ParseFen(data.StartFEN)
		} else if strings.HasPrefix(text, "setoption") {
			uci.parseOption(text)
		} else if strings.HasPrefix(text, "position") {
			uci.parsePosition(text, game)
		} else if strings.HasPrefix(text, "go") {
			uci.parseGo(text, game, &info)
		} else if text == "stop" {
			info.ForceStop = true
			fmt.Printf("debug: ForceStop %v\n", info.ForceStop)
		} else if text == "run" {
			uci.engineHolder.UseBook = false
			uci.parseGo("go infinite", game, &info)
		} else if text == "test" {
			uci.parsePosition("position startpos moves d2d4 d7d5 c1f4 g8f6 b1c3 c8f5 e2e3 e7e6 f1d3 f8b4 g1e2 e8g8 e1g1 b8c6 d3f5 e6f5 f4g5 b4e7 g5f6 e7f6 d1d3 c6e7 f2f3 c7c6 e3e4 f5e4 f3e4 d8b6 b2b3 a8d8 e4e5 f6g5 d3g3 g5d2 c3a4 b6b5 g3f3 e7g6 a1d1 d2b4 f3e3 b5a5 g1h1 f8e8 e3f3 e8e7 d1a1 d8f8 a2a3 b4d2 f3h3 f7f6 e5e6 f8e8 e2g3 b7b6 g3f5 e7e6 h3g3 e8c8 g3g4 c8e8 g4g3", game)
			game.Position().Board.PrintBoard()
			uci.parseGo("go wtime 93687 btime 51739 winc 5000 binc 5000", game, &info)
		} else if text == "quit" {
			info.Quit = data.True
			break
		}
	}

}

func (uci *UCI) printUCIok() {
	fmt.Println("id name MyGoEngine")
	fmt.Println("id author Adam")
	fmt.Println("uciok")
	fmt.Printf("option name OwnBook type check default %t\n", uci.engineHolder.UseBook)
}

func (uci *UCI) parseOption(line string) {
	tokens := strings.Split(line, " ")

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "book":
			uci.parseBook(tokens[i+1])
		}
	}
}

func (uci *UCI) parseBook(line string) {
	switch line {
	case "true":
		uci.engineHolder.UseBook = true
		fmt.Printf("book turned on\n")
	case "false":
		uci.engineHolder.UseBook = false
		fmt.Printf("book turned off\n")
	default:
		fmt.Printf("Unknown book command expected true / false ")
	}
}

func (uci *UCI) parseGo(line string, game engine.Game, info *data.SearchInfo) {
	tokens := strings.Split(line, " ")
	info.MoveTime = -1
	info.MovesToGo = 30
	info.Depth = -1
	info.Time = -1

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "binc":
			uci.parseInc(tokens[i+1], data.Black, game, info)
		case "winc":
			uci.parseInc(tokens[i+1], data.White, game, info)
		case "wtime":
			uci.parseTime(tokens[i+1], data.White, game, info)
		case "btime":
			uci.parseTime(tokens[i+1], data.Black, game, info)
		case "movestogo":
			uci.parseMovesToGo(tokens[i+1], info)
		case "movetime":
			uci.parseMoveTime(tokens[i+1], info)
		case "depth":
			uci.parseDepth(tokens[i+1], info)
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

	for _, eng := range uci.engineHolder.Engines {
		eng.Position = game.Position().Copy()
	}

	uci.engineHolder.Ctx, uci.engineHolder.CancelSearch = context.WithCancel(context.Background())

	go uci.engineHolder.Search(info)
}

func (uci *UCI) parseInc(token string, side int, game engine.Game, info *data.SearchInfo) {
	inc, _ := strconv.Atoi(token)
	if game.Position().Side == side {
		info.Inc = inc
	}
}

func (uci *UCI) parseTime(token string, side int, game engine.Game, info *data.SearchInfo) {
	time, _ := strconv.Atoi(token)
	if game.Position().Side == side {
		info.Time = time
	}
}

func (uci *UCI) parseMovesToGo(token string, info *data.SearchInfo) {
	movesToGo, _ := strconv.Atoi(token)
	info.MovesToGo = movesToGo
}

func (uci *UCI) parseMoveTime(token string, info *data.SearchInfo) {
	moveTime, _ := strconv.Atoi(token)
	info.MoveTime = moveTime
}

func (uci *UCI) parseDepth(token string, info *data.SearchInfo) {
	depth, _ := strconv.Atoi(token)
	info.Depth = depth
}

func (uci *UCI) parsePosition(lineIn string, game engine.Game) {
	parts := strings.Split(lineIn, " ")

	if len(parts) < 2 {
		panic(fmt.Errorf("UCI parsePosition: unexpected length %v", lineIn))
	}

	if parts[1] == "startpos" {
		fmt.Printf("startpos called\n\n")
		game.Position().ParseFen(data.StartFEN)
	}

	if parts[1] == "fen" {
		fen := strings.Join(parts[2:], " ")
		game.Position().ParseFen(fen)

	}

	startIndex := 0
	for i, v := range parts {
		if v == "moves" {
			startIndex = i
		}
	}

	if startIndex != 0 {
		for i := startIndex + 1; i < len(parts); i++ {
			move := game.Position().ParseMove([]byte(parts[i]))
			if move == data.NoMove {
				fmt.Printf("UCI move error: Parsing UCI (%v) (%v) %v - %v\n", parts[i], lineIn, move, io.PrintMove(move))
			}
			game.Position().MakeMove(move)
			game.Position().Positions[game.Position().PositionKey]++
			game.Position().Play = 0
			game.Position().PositionHistory.RemovePositionHistory()
		}
	}
}
