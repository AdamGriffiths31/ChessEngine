package consolemode

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/util"
	"github.com/AdamGriffiths31/ChessEngine/xboard"
)

func ConsoleMode(pos *data.Board, info *data.SearchInfo) {
	fmt.Printf("Console mode started\nType help for commands\n\n")
	info.GameMode = data.ConsoleMode
	info.PostThinking = true

	board.ParseFEN(data.StartFEN, pos)
	io.PrintBoard(pos)
	depth := data.MaxDepth
	moveTime := 3000
	move := data.NoMove
	engineSide := data.Black
	movestogo := [2]int{30, 30}

	reader := bufio.NewReader(os.Stdin)

	inputCh := make(chan string)

	go func() {
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				panic(fmt.Errorf("console Mode reader loop: %v", err))
			}

			input = strings.TrimSpace(input)
			inputCh <- input
		}
	}()

	for {
		select {
		case input := <-inputCh:
			if pos.Side == engineSide && !xboard.CheckResult(pos) {
				info.StartTime = util.GetTimeMs()
				info.Depth = depth
				if moveTime != 0 {
					info.TimeSet = data.True
					info.StopTime = info.StartTime + int64(moveTime)
				}

				search.SearchPosistion(pos, info)
			}

			if input == "quit" {
				break
			}

			if input == "force" {
				engineSide = data.Both
				continue
			}

			if input == "sd" {
				inputDepth, _ := strconv.Atoi(input)
				depth = inputDepth
				continue
			}

			if strings.HasPrefix(input, "usermove") {
				parts := strings.Split(input, " ")
				movestogo[pos.Side]--
				move = moveGen.ParseMove([]byte(parts[1]), pos)
				if move == data.NoMove {
					continue
				}
				moveGen.MakeMove(move, pos)
				pos.Play = 0
			}

			if input == "print" {
				io.PrintBoard(pos)
				continue
			}

			if input == "new" {
				moveGen.ClearTable(pos.PvTable)
				engineSide = data.Black
				board.ParseFEN(data.StartFEN, pos)
				continue
			}

			if strings.HasPrefix(input, "setboard") {
				parts := strings.Split(input, " ")

				engineSide = data.Both
				board.ParseFEN(parts[1], pos)
				continue
			}

			if input == "help" {
				fmt.Printf("Commands:\n")
				fmt.Printf("quit - quit game\n")
				fmt.Printf("force - computer will not think\n")
				fmt.Printf("print - show board\n")
				fmt.Printf("new - start new game\n")
				fmt.Printf("setboard x - set position to fen x\n")
				continue
			}

			if len(input) >= 3 {
				move = moveGen.ParseMove([]byte(input), pos)
				if move == data.NoMove {
					fmt.Printf("Unknown command -- type help for commands")
					continue
				}
				moveGen.MakeMove(move, pos)
				pos.Play = 0
				fmt.Printf("User has played %s\n", input)
			}
		default:
			if pos.Side == engineSide && !xboard.CheckResult(pos) {
				info.StartTime = util.GetTimeMs()
				info.Depth = depth
				if moveTime != 0 {
					info.TimeSet = data.True
					info.StopTime = info.StartTime + int64(moveTime)
				}

				search.SearchPosistion(pos, info)
			}
		}
	}
}
