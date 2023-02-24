package search

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	"github.com/AdamGriffiths31/ChessEngine/io"
	movegen "github.com/AdamGriffiths31/ChessEngine/moveGen"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

func IsRepetition(pos *data.Board) bool {
	for i := pos.HistoryPlay - pos.FiftyMove; i < pos.HistoryPlay-1; i++ {
		if pos.PosistionKey == pos.History[i].PosistionKey {
			return true
		}
	}
	return false
}

func SearchPosistion(pos *data.Board, info *data.SearchInfo) {
	bestScore := 30000
	bestMove := data.NoMove
	clearForSearch(pos, info)

	for currentDepth := 1; currentDepth < info.Depth; currentDepth++ {
		//fmt.Printf("----------------------------------------------------------\n")
		//fmt.Printf("Current depth: %v Nodes: %v\n", currentDepth, info.Node)
		bestScore = alphaBeta(-30000, 30000, currentDepth, pos, info)
		if info.Stopped == data.True {
			break
		}
		movegen.GetPvLine(currentDepth, pos)
		bestMove = pos.PvArray[0]
		fmt.Printf("info score cp %d depth %d nodes %v time %d\n", bestScore, currentDepth, info.Node, util.GetTimeMs()-info.StartTime)
		//fmt.Printf("Count PV: %d\n", pvMoves)
		// for i := 0; i < pvMoves; i++ {
		// 	fmt.Printf("\t%s", PrintMove(pos.PvArray[i]))
		// }
		//fmt.Printf("\n")
		//fmt.Printf("Ordering :%.2f\n", info.FailHighFirst/info.FailHigh)
	}
	fmt.Printf("bestmove %s\n", io.PrintMove(bestMove))
	//fmt.Printf("----------------------------------------------------------\n")
}

func alphaBeta(alpha, beta, depth int, pos *data.Board, info *data.SearchInfo) int {
	board.CheckBoard(pos)

	if beta < alpha {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depth < 0 {
		panic(fmt.Errorf("alphaBeta depth %v", depth))
	}

	if depth == 0 {
		return quiescence(alpha, beta, pos, info)
	}

	if info.Node&2047 == 0 {
		checkUp(info)
	}

	info.Node++

	if IsRepetition(pos) || pos.FiftyMove >= 100 {
		return 0
	}

	if pos.Play > data.MaxDepth-1 {
		return evaluate.EvalPosistion(pos)
	}

	ml := &data.MoveList{}
	movegen.GenerateAllMoves(pos, ml)

	legal := 0
	oldAlpha := alpha
	bestMove := data.NoMove
	pvMove := movegen.ProbePvTable(pos)

	if pvMove != data.NoMove {
		for i := 0; i < ml.Count; i++ {
			if ml.Moves[i].Move == pvMove {
				ml.Moves[i].Score = 2000000
			}
		}
	}

	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if !movegen.MakeMove(ml.Moves[i].Move, pos) {
			continue
		}

		legal++
		score := -alphaBeta(-beta, -alpha, depth-1, pos, info)
		movegen.TakeMoveBack(pos)
		if info.Stopped == data.True {
			return 0
		}
		if score > alpha {
			if score >= beta {
				if legal == 1 {
					info.FailHighFirst++
				}

				if ml.Moves[i].Move&data.MFLAGCAP == 0 {
					pos.SearchKillers[1][pos.Play] = pos.SearchKillers[0][pos.Play]
					pos.SearchKillers[0][pos.Play] = ml.Moves[i].Move
				}

				info.FailHigh++
				return beta
			}

			alpha = score
			bestMove = ml.Moves[i].Move

			if ml.Moves[i].Move&data.MFLAGCAP == 0 {
				pos.SearchHistory[pos.Pieces[data.FromSquare(bestMove)]][pos.Pieces[data.ToSqaure(bestMove)]] += depth
			}
		}
	}

	if legal == 0 {
		if attack.SquareAttacked(pos.KingSqaure[pos.Side], pos.Side^1, pos) {
			return -29000 + pos.Play
		} else {
			return 0
		}
	}

	if alpha != oldAlpha {
		movegen.StorePvMove(pos, bestMove)
	}

	return alpha
}

func quiescence(alpha, beta int, pos *data.Board, info *data.SearchInfo) int {
	board.CheckBoard(pos)

	if info.Node&2047 == 0 {
		checkUp(info)
	}

	info.Node++

	if IsRepetition(pos) || pos.FiftyMove >= 100 {
		return 0
	}

	if pos.Play > data.MaxDepth-1 {
		return evaluate.EvalPosistion(pos)
	}

	score := evaluate.EvalPosistion(pos)

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := &data.MoveList{}
	movegen.GenerateAllCaptures(pos, ml)

	oldAlpha := alpha
	bestMove := data.NoMove
	pvMove := movegen.ProbePvTable(pos)

	if pvMove != data.NoMove {
		for i := 0; i < ml.Count; i++ {
			if ml.Moves[i].Move == pvMove {
				ml.Moves[i].Score = 2000000
			}
		}
	}

	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if !movegen.MakeMove(ml.Moves[i].Move, pos) {
			continue
		}

		score := -quiescence(-beta, -alpha, pos, info)
		movegen.TakeMoveBack(pos)
		if info.Stopped == data.True {
			return 0
		}
		if score > alpha {
			if score >= beta {
				return beta
			}

			alpha = score
			bestMove = ml.Moves[i].Move
		}
	}

	if alpha != oldAlpha {
		movegen.StorePvMove(pos, bestMove)
	}

	return alpha
}

func clearForSearch(pos *data.Board, info *data.SearchInfo) {
	for i := 0; i < 13; i++ {
		for j := 0; j < 120; j++ {
			pos.SearchHistory[i][j] = 0
		}
	}

	for i := 0; i < 2; i++ {
		for j := 0; j < data.MaxDepth; j++ {
			pos.SearchHistory[i][j] = 0
		}
	}

	movegen.ClearTable(pos.PvTable)
	pos.Play = 0

	info.StartTime = util.GetTimeMs()
	info.Stopped = 0
	info.Node = 0
	info.FailHighFirst = 0
	info.FailHigh = 0
}

func pickNextMove(moveNum int, ml *data.MoveList) {
	bestScore := 0
	bestNum := 0
	for i := moveNum; i < ml.Count; i++ {
		if ml.Moves[i].Score > bestScore {
			bestScore = ml.Moves[i].Score
			bestNum = i
		}
	}
	holder := ml.Moves[moveNum]
	ml.Moves[moveNum] = ml.Moves[bestNum]
	ml.Moves[bestNum] = holder
}

func checkUp(info *data.SearchInfo) {
	if info.TimeSet == data.True && util.GetTimeMs() > info.StopTime {
		info.Stopped = data.True
	}
}
