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
			uci.parsePosition("position startpos moves d2d4 d7d5 g1f3 c8f5 c1f4 e7e6 e2e3 f8d6 b1c3 g8e7 f1d3 d6f4 e3f4 d8d6 d3f5 e7f5 d1d2 d6b6 c3a4 b6c6 a4c5 b8a6 c5d3 e8c8 e1g1 f7f6 f1e1 h8e8 e1e2 f5d6 a1e1 c6a4 a2a3 h7h5 c2c3 d6e4 d2e3 d8d6 f3d2 e4d2 e2d2 g7g6 e3g3 e8g8 g3h4 g8f8 d2e2 a4d7 h4h3 f6f5 h3f3 f8g8 f3g3 d7e8 e2e5 e8f7 g3e3 g8e8 d3c5 a6c5 d4c5 d6c6 e5d5 e6d5 e3e8 f7e8 e1e8 c8d7 e8b8 c6e6 f2f3 b7b6 c5b6 a7b6 g1f2 c7c5 b8a8 d5d4 c3d4 c5d4 a8a4 e6d6 f2e2 d4d3 e2d2 h5h4 a4c4 d7e6 c4b4 e6d5 b4b3 d5c5 b3d3 d6e6 d3c3 c5d5 c3c7 d5d4 a3a4 d4d5 d2d3 e6e1 c7d7 d5c5 d7c7 c5d5", game)
			game.Position().Board.PrintBoard()
			uci.parseGo("go wtime 45195 btime 58866 winc 5000 binc 5000", game, &info)
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
		}
	}
}
