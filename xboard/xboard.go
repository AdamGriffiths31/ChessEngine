package xboard

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

func Xboard(pos *data.Board, info *data.SearchInfo) {
	var (
		depth      = -1
		movestogo  = [2]int{30, 30}
		time       = -1
		inc        = 0
		engineSide = data.Both
		timeLeft   int
		sec        int
		mps        int
		move       = data.NoMove
	)
	fmt.Printf("Xboard starting...\n")
	info.GameMode = data.XboardMode
	info.PostThinking = true
	engineSide = data.Black
	board.ParseFEN(data.StartFEN, pos)
	depth = -1
	time = -1

	reader := bufio.NewReader(os.Stdin)
	inputCh := make(chan string)

	go func() {
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				panic(fmt.Errorf("Xboard Mode reader loop: %v", err))
			}

			input = strings.TrimSpace(input)
			inputCh <- input
		}
	}()

	for {
		select {
		case input := <-inputCh:

			input = strings.TrimSpace(input)

			if input == "quit" {
				break
			}

			if input == "force" {
				engineSide = data.Both
				continue
			}

			if input == "protover" {
				printOptions()
				continue
			}

			if strings.HasPrefix(input, "sd") {
				parts := strings.Split(input, " ")
				inputDepth, _ := strconv.Atoi(parts[1])
				depth = inputDepth
			}

			if strings.HasPrefix(input, "level") {
				parts := strings.Split(input, " ")
				fmt.Printf("Level command %v %v %v %v\n", parts[0], parts[1], parts[2], parts[3])
				if len(parts) >= 4 {
					mps, _ = strconv.Atoi(parts[1])
					timeLeft, _ = strconv.Atoi(parts[2])
					inc, _ = strconv.Atoi(parts[3])
				}

				timeLeft *= 6000
				timeLeft += sec * 1000
				movestogo[0], movestogo[1] = 30, 30
				if mps != 0 {
					movestogo[0], movestogo[1] = mps, mps
				}
				time = -1
				continue
			}

			if input == "ping" {
				fmt.Printf("pong")
				continue
			}

			if input == "new" {
				engineSide = data.Black
				board.ParseFEN(data.StartFEN, pos)
				depth = -1
				continue
			}

			if strings.HasPrefix(input, "setboard") {
				parts := strings.Split(input, " ")

				engineSide = data.Both
				board.ParseFEN(parts[1], pos)
				continue
			}

			if strings.HasPrefix(input, "time") {
				parts := strings.Split(input, " ")
				inputTime, _ := strconv.Atoi(parts[1])
				inputTime *= 10
				time = inputTime
			}

			if input == "go" {
				engineSide = pos.Side
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

			if len(input) >= 3 {
				move = moveGen.ParseMove([]byte(input), pos)
				if move != data.NoMove {
					moveGen.MakeMove(move, pos)
				}
				continue
			}
		default:
			if pos.Side == engineSide && !CheckResult(pos) {
				fmt.Printf("Search\n")
				info.StartTime = util.GetTimeMs()
				info.Depth = depth
				if time != -1 {
					info.TimeSet = data.True
					time /= movestogo[pos.Side]
					info.StopTime = info.StartTime + int64(time) + int64(inc)
				}
				if depth == -1 || depth > data.MaxDepth {
					info.Depth = data.MaxDepth
				}

				fmt.Printf("time:%d start:%d stop:%d depth:%d timeset:%d movestogo:%d mps:%d inc:%d\n", time, info.StartTime, info.StopTime, info.Depth, info.TimeSet, movestogo[pos.Side], mps, inc)
				search.SearchPosistion(pos, info)

				if mps != 0 {
					movestogo[pos.Side^1]--
					if movestogo[pos.Side^1] < 1 {
						movestogo[pos.Side^1] = mps
					}
				}
			}
		}
	}
}

func CheckResult(pos *data.Board) bool {
	if pos.FiftyMove > 100 {
		fmt.Printf("1/2-1/2 {fifty move rule (claimed by GoChessEngine)}\n")
		return true
	}
	if isThreeFoldRepetition(pos) {
		fmt.Printf("1/2-1/2 {3-fold repetition (claimed by GoChessEngine)}\n")
		return true
	}
	if drawMaterial(pos) {
		fmt.Printf("1/2-1/2 {insufficient material (claimed by GoChessEngine)}\n")
		return true
	}

	ml := &data.MoveList{}
	moveGen.GenerateAllMoves(pos, ml)
	legal := 0

	for i := 0; i < ml.Count; i++ {
		if moveGen.MakeMove(ml.Moves[i].Move, pos) {
			legal++
			moveGen.TakeMoveBack(pos)
			break
		}
	}

	if legal != 0 {
		return false
	}
	inCheck := attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos)

	if inCheck {
		if pos.Side == data.White {
			fmt.Printf("0-1 {black mates (claimed by GoChessEngine)}\n")
			return true
		} else {
			fmt.Printf("1-0 {white mates (claimed by GoChessEngine)}\n")
			return true
		}
	} else {
		fmt.Printf("1/2-1/2 {stalemate (claimed by GoChessEngine)}\n")
		return true
	}
}

func drawMaterial(pos *data.Board) bool {
	if pos.PieceNumber[data.WP] > 0 || pos.PieceNumber[data.BP] > 0 {
		return false
	}
	if pos.PieceNumber[data.WQ] > 0 || pos.PieceNumber[data.BQ] > 0 || pos.PieceNumber[data.WR] > 0 || pos.PieceNumber[data.BR] > 0 {
		return false
	}
	if pos.PieceNumber[data.WB] > 1 || pos.PieceNumber[data.BB] > 1 {
		return false
	}
	if pos.PieceNumber[data.WN] > 0 && pos.PieceNumber[data.WB] > 0 {
		return false
	}
	if pos.PieceNumber[data.BN] > 0 && pos.PieceNumber[data.BB] > 0 {
		return false
	}
	return true
}

func isThreeFoldRepetition(pos *data.Board) bool {
	rep := 0
	for i := 0; i < pos.HistoryPlay && rep < 2; i++ {
		if pos.History[i].PosistionKey == pos.PosistionKey {
			rep++
		}
	}
	return rep == 2
}

func printOptions() {
	fmt.Printf("feature ping=1 setboard=1 colors=0 usermove=1\n")
	fmt.Printf("feature done=1\n")
}
