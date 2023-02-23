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
		//fmt.Printf("----------------------------------------------------------\n")
		//fmt.Printf("Current depth: %v Nodes: %v\n", currentDepth, info.Node)
		bestScore = alphaBeta(-30000, 30000, currentDepth, pos, info)
		if info.Stopped == True {
			break
		}
		GetPvLine(currentDepth, pos)
		bestMove = pos.PvArray[0]
		fmt.Printf("info score cp %d depth %d nodes %v time %d\n", bestScore, currentDepth, info.Node, GetTimeMs()-info.StartTime)
		//fmt.Printf("Count PV: %d\n", pvMoves)
		// for i := 0; i < pvMoves; i++ {
		// 	fmt.Printf("\t%s", PrintMove(pos.PvArray[i]))
		// }
		//fmt.Printf("\n")
		//fmt.Printf("Ordering :%.2f\n", info.FailHighFirst/info.FailHigh)
	}
	fmt.Printf("bestmove %s\n", PrintMove(bestMove))
	//fmt.Printf("----------------------------------------------------------\n")
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
		return quiescence(alpha, beta, pos, info)
	}

	if info.Node&2047 == 0 {
		checkUp(info)
	}

	info.Node++

	if IsRepetition(pos) || pos.FiftyMove >= 100 {
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
	pvMove := ProbePvTable(pos)

	if pvMove != NoMove {
		for i := 0; i < ml.Count; i++ {
			if ml.Moves[i].Move == pvMove {
				ml.Moves[i].Score = 2000000
			}
		}
	}

	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if !MakeMove(ml.Moves[i].Move, pos) {
			continue
		}

		legal++
		score := -alphaBeta(-beta, -alpha, depth-1, pos, info)
		TakeMoveBack(pos)
		if info.Stopped == True {
			return 0
		}
		if score > alpha {
			if score >= beta {
				if legal == 1 {
					info.FailHighFirst++
				}

				if ml.Moves[i].Move&MFLAGCAP == 0 {
					pos.SearchKillers[1][pos.Play] = pos.SearchKillers[0][pos.Play]
					pos.SearchKillers[0][pos.Play] = ml.Moves[i].Move
				}

				info.FailHigh++
				return beta
			}

			alpha = score
			bestMove = ml.Moves[i].Move

			if ml.Moves[i].Move&MFLAGCAP == 0 {
				pos.SearchHistory[pos.Pieces[FromSquare(bestMove)]][pos.Pieces[ToSqaure(bestMove)]] += depth
			}
		}
	}

	if legal == 0 {
		if SquareAttacked(pos.KingSqaure[pos.Side], pos.Side^1, pos) {
			return -29000 + pos.Play
		} else {
			return 0
		}
	}

	if alpha != oldAlpha {
		StorePvMove(pos, bestMove)
	}

	return alpha
}

func quiescence(alpha, beta int, pos *Board, info *SearchInfo) int {
	CheckBoard(pos)

	if info.Node&2047 == 0 {
		checkUp(info)
	}

	info.Node++

	if IsRepetition(pos) || pos.FiftyMove >= 100 {
		return 0
	}

	if pos.Play > MaxDepth-1 {
		return EvalPosistion(pos)
	}

	score := EvalPosistion(pos)

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := &MoveList{}
	GenerateAllCaptures(pos, ml)

	oldAlpha := alpha
	bestMove := NoMove
	pvMove := ProbePvTable(pos)

	if pvMove != NoMove {
		for i := 0; i < ml.Count; i++ {
			if ml.Moves[i].Move == pvMove {
				ml.Moves[i].Score = 2000000
			}
		}
	}

	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if !MakeMove(ml.Moves[i].Move, pos) {
			continue
		}

		score := -quiescence(-beta, -alpha, pos, info)
		TakeMoveBack(pos)
		if info.Stopped == True {
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
		StorePvMove(pos, bestMove)
	}

	return alpha
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

func pickNextMove(moveNum int, ml *MoveList) {
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

func checkUp(info *SearchInfo) {
	if info.TimeSet == True && GetTimeMs() > info.StopTime {
		info.Stopped = True
	}
}
