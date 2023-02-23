package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Uci() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("id name MyGoEngine")
	fmt.Println("id author Adam")
	fmt.Println("uciok")

	pvTable := &PVTable{}
	pos := &Board{PvTable: pvTable}
	info := &SearchInfo{}
	InitPvTable(pos.PvTable)
	ParseFEN(StartFEN, pos)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("UCI reader loop: %v", err))
		}

		text = strings.TrimSpace(text)

		if text == "uci" {
			fmt.Println("id name MyGoEngine")
			fmt.Println("id author Me")
			fmt.Println("uciok")
		} else if text == "isready" {
			fmt.Println("readyok")
		} else if strings.HasPrefix(text, "position") {
			parsePosition(text, pos)
		} else if text == "ucinewgame" {
			ParseFEN(StartFEN, pos)
		} else if strings.HasPrefix(text, "go") {
			parseGo(text, info, pos)
		} else if text == "quit" {
			info.Quit = True
			break
		}

		if info.Quit == True {
			break
		}
	}
}

func parseGo(line string, info *SearchInfo, pos *Board) {
	tokens := strings.Split(line, " ")
	info.MoveTime = -1
	info.MovesToGo = 30
	info.Depth = -1
	info.Time = -1

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "binc":
			parseInc(tokens[i+1], Black, pos, info)
		case "winc":
			parseInc(tokens[i+1], White, pos, info)
		case "wtime":
			parseTime(tokens[i+1], White, pos, info)
		case "btime":
			parseTime(tokens[i+1], Black, pos, info)
		case "movestogo":
			parseMovesToGo(tokens[i+1], pos, info)
		case "movetime":
			parseMoveTime(tokens[i+1], pos, info)
		case "depth":
			parseDepth(tokens[i+1], info)
		}
	}

	info.StartTime = GetTimeMs()

	if info.MoveTime != -1 {
		info.TimeSet = True
		info.MovesToGo = 1
		info.StopTime = info.StartTime + int64(info.MoveTime)
	} else if info.Time != -1 {
		info.TimeSet = True
		info.MovesToGo = 30
		time := info.Time / info.MovesToGo
		time -= 50
		info.Time = time
		info.StopTime = info.StartTime + int64(time) + int64(info.Inc)
	}

	if info.Depth == -1 || info.Depth > MaxDepth {
		info.Depth = MaxDepth
	}

	fmt.Printf("time:%d start:%d stop:%d depth:%d timeset:%v\n", info.Time, info.StartTime, info.StopTime, info.Depth, info.TimeSet)

	SearchPosistion(pos, info)
}

func parseInc(token string, side int, pos *Board, info *SearchInfo) {
	inc, _ := strconv.Atoi(token)
	if pos.Side == side {
		info.Inc = inc
	}
}

func parseTime(token string, side int, pos *Board, info *SearchInfo) {
	time, _ := strconv.Atoi(token)
	if pos.Side == side {
		info.Time = time
	}
}

func parseMovesToGo(token string, pos *Board, info *SearchInfo) {
	movesToGo, _ := strconv.Atoi(token)
	info.MovesToGo = movesToGo
}

func parseMoveTime(token string, pos *Board, info *SearchInfo) {
	moveTime, _ := strconv.Atoi(token)
	info.MoveTime = moveTime
}

func parseDepth(token string, info *SearchInfo) {
	depth, _ := strconv.Atoi(token)
	info.Depth = depth
}

func parsePosition(lineIn string, pos *Board) {
	parts := strings.Split(lineIn, " ")

	if len(parts) < 2 {
		panic(fmt.Errorf("UCI parsePosition: unexpected length %v", lineIn))
	}

	if parts[1] == "startpos" {
		ParseFEN(StartFEN, pos)
	}

	if parts[1] == "fen" {
		fen := strings.Join(parts[2:], " ")
		ParseFEN(fen, pos)
	}

	startIndex := 0
	for i, v := range parts {
		if v == "moves" {
			startIndex = i
		}
	}

	if startIndex != 0 {
		for i := startIndex + 1; i < len(parts); i++ {
			move := ParseMove([]byte(parts[i]), pos)
			MakeMove(move, pos)
			pos.Play = 0
		}

	}
}
