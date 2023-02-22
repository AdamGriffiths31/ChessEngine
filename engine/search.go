package engine

import (
	"fmt"
)

func IsRepetition(pos *Board) bool {
	for i := pos.HistoryPlay - pos.FiftyMove; i < pos.HistoryPlay-1; i++ {
		if pos.PosistionKey == pos.History[i].PosistionKey {
			return true
		}
	}
	return false
}

func SearchPosistion(pos *Board, info *SearchInfo) {
	bestScore := 30000
	bestMove := NoMove
	clearForSearch(pos, info)

	for currentDepth := 1; currentDepth < info.Depth; currentDepth++ {
		fmt.Printf("Depth: %v\n", currentDepth)
		bestScore = alphaBeta(-30000, 30000, currentDepth, pos, info)
		pvMoves := GetPvLine(currentDepth, pos)
		bestMove = pos.PvArray[0]
		fmt.Printf("Depth %v score: %v move: %v nodes %v\n", currentDepth, bestScore, PrintMove(bestMove), info.Node)
		fmt.Printf("Count PV: %d\n", pvMoves)
		for i := 0; i < pvMoves; i++ {
			fmt.Printf("\t%s", PrintMove(pos.PvArray[i]))
		}
		fmt.Printf("\n")
		fmt.Printf("Ordering :%.2f\n", info.FailHighFirst/info.FailHigh)
	}
}

func alphaBeta(alpha, beta, depth int, pos *Board, info *SearchInfo) int {
	CheckBoard(pos)

	if beta < alpha {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depth < 0 {
		panic(fmt.Errorf("alphaBeta depth %v", depth))
	}

	if depth == 0 {
		info.Node++
		return EvalPosistion(pos)
	}

	info.Node++

	if IsRepetition(pos) || pos.FiftyMove >= 100 {
		fmt.Printf("FiftyMove: %v or rep \n", pos.FiftyMove)
		return 0
	}

	if pos.Play > MaxDepth-1 {
		return EvalPosistion(pos)
	}

	ml := &MoveList{}
	GenerateAllMoves(pos, ml)

	legal := 0
	oldAlpha := alpha
	bestMove := NoMove

	for i := 0; i < ml.Count; i++ {
		if !MakeMove(ml.Moves[i].Move, pos) {
			continue
		}

		legal++
		score := -alphaBeta(-beta, -alpha, depth-1, pos, info)
		TakeMoveBack(pos)

		if !MoveExists(pos, ml.Moves[i].Move) {
			PrintBoard(pos)
			panic(fmt.Errorf("alphaBeta move error %v", PrintMove(ml.Moves[i].Move)))
		}
		CheckBoard(pos)

		if score > alpha {
			if score >= beta {
				if legal == 1 {
					info.FailHighFirst++
				}

				info.FailHigh++
				return beta
			}

			alpha = score
			bestMove = ml.Moves[i].Move
		}
	}

	if legal == 0 {
		fmt.Printf("\n\nLegal = 0\n\n")
		if SquareAttacked(pos.KingSqaure[pos.Side], pos.Side^1, pos) {
			fmt.Printf("\nMate?\n")
			return pos.Play //TODO Check this
		} else {
			return 0
		}
	}

	if alpha != oldAlpha {
		fmt.Printf("store score %v", alpha)
		StorePvMove(pos, bestMove)
	}

	return alpha
}

func quiescence(alpha, beta int, pos *Board, info *SearchInfo) int {
	return 0
}

func clearForSearch(pos *Board, info *SearchInfo) {
	for i := 0; i < 13; i++ {
		for j := 0; j < 120; j++ {
			pos.SearchHistory[i][j] = 0
		}
	}

	for i := 0; i < 2; i++ {
		for j := 0; j < MaxDepth; j++ {
			pos.SearchHistory[i][j] = 0
		}
	}

	clearTable(pos.PvTable)
	pos.Play = 0

	info.StartTime = GetTimeMs()
	info.Stopped = 0
	info.Node = 0
	info.FailHighFirst = 0
	info.FailHigh = 0
}

func checkUp() {
}
