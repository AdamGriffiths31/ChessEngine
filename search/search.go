package search

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
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
	bestMove := data.NoMove
	clearForSearch(pos, info)
	if data.EngineSettings.UseBook {
		bestMove = polyglot.GetBookMove(pos)
	}

	if bestMove == data.NoMove {
		for currentDepth := 1; currentDepth < info.Depth+1; currentDepth++ {
			bestScore := alphaBeta(-30000, 30000, currentDepth, pos, info, true)
			if info.Stopped == data.True {
				break
			}
			pvMoves := moveGen.GetPvLine(currentDepth, pos)
			bestMove = pos.PvArray[0]
			printPVData(info, currentDepth, bestScore)
			printPVLine(pos, info, pvMoves)
			printSearchResult(pos, info, bestMove)
		}
	}
}

func printPVData(info *data.SearchInfo, currentDepth, bestScore int) {
	if info.GameMode == data.UCIMode {
		fmt.Printf("info score cp %d depth %d nodes %v time %d ", bestScore, currentDepth, info.Node, util.GetTimeMs()-info.StartTime)
	} else if info.GameMode == data.XboardMode && info.PostThinking {
		fmt.Printf("%d %d %d %v\n", currentDepth, bestScore, (util.GetTimeMs()-info.StartTime)/10, info.Node)
	} else if info.PostThinking {
		fmt.Printf("depth: %d score: %d time:%d  nodes:%v\n", currentDepth, bestScore, (util.GetTimeMs()-info.StartTime)/10, info.Node)
	}
}

func printPVLine(pos *data.Board, info *data.SearchInfo, pvMoves int) {
	if info.GameMode == data.UCIMode || info.PostThinking {
		fmt.Printf("pv")
		for i := 0; i < pvMoves; i++ {
			fmt.Printf(" %s", io.PrintMove(pos.PvArray[i]))
		}
		fmt.Printf("\n")
		fmt.Printf("Ordering: %.2f\n", info.FailHighFirst/info.FailHigh)
	}
}

func printSearchResult(pos *data.Board, info *data.SearchInfo, bestMove int) {
	if info.GameMode == data.UCIMode {
		fmt.Printf("bestmove %s\n", io.PrintMove(bestMove))
	} else if info.GameMode == data.XboardMode {
		fmt.Printf("move %s\n", io.PrintMove(bestMove))
		moveGen.MakeMove(bestMove, pos)
	} else {
		fmt.Printf("Engine plays %s\n", io.PrintMove(bestMove))
		moveGen.MakeMove(bestMove, pos)
		io.PrintBoard(pos)
	}
}

func alphaBeta(alpha, beta, depth int, pos *data.Board, info *data.SearchInfo, doNull bool) int {
	board.CheckBoard(pos)

	if depth < 0 {
		panic(fmt.Errorf("alphaBeta depth was  %v", depth))
	}
	if beta < alpha {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depth <= 0 {
		return quiescence(alpha, beta, pos, info)
	}

	if info.Node&2047 == 0 {
		checkUp(info)
	}

	info.Node++

	if (IsRepetition(pos) || pos.FiftyMove >= 100) && pos.Play != 0 {
		return 0
	}

	if pos.Play > data.MaxDepth-1 {
		return evaluate.EvalPosistion(pos)
	}

	inCheck := attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos)
	if inCheck {
		depth++
	}

	score := -30000
	pvMove := data.NoMove
	if moveGen.ProbePvTable(pos, &pvMove, &score, alpha, beta, depth) {
		pos.PvTable.Cut++
		return score
	}

	if doNull && !inCheck && pos.Play != 0 && pos.BigPiece[pos.Side] > 1 && depth >= 4 {
		moveGen.MakeNullMove(pos)
		score = -alphaBeta(-beta, -beta+1, depth-4, pos, info, false)
		moveGen.TakeBackNullMove(pos)
		if info.Stopped == data.True {
			return 0
		}
		if score >= beta && score < data.Mate {
			info.NullCut++
			return beta
		}
	}

	ml := &data.MoveList{}
	moveGen.GenerateAllMoves(pos, ml)

	legal := 0
	oldAlpha := alpha
	bestMove := data.NoMove
	score = -30000
	bestScore := -30000

	if pvMove != data.NoMove {
		for i := 0; i < ml.Count; i++ {
			if ml.Moves[i].Move == pvMove {
				ml.Moves[i].Score = 2000000
				break
			}
		}
	}

	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if moveGen.MakeMove(ml.Moves[i].Move, pos) {
			legal++
			score = -alphaBeta(-beta, -alpha, depth-1, pos, info, true)
			moveGen.TakeMoveBack(pos)
			if info.Stopped == data.True {
				return 0
			}
			if score > bestScore {
				bestScore = score
				bestMove = ml.Moves[i].Move
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
						moveGen.StorePvMove(pos, bestMove, beta, data.PVBeta, depth)
						return beta
					}

					alpha = score

					if ml.Moves[i].Move&data.MFLAGCAP == 0 {
						pos.SearchHistory[pos.Pieces[data.FromSquare(bestMove)]][data.ToSquare(bestMove)] += depth
					}
				}
			}
		}
	}

	if legal == 0 {
		if attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos) {
			return -30000 + pos.Play
		} else {
			return 0
		}
	}
	if !(alpha >= oldAlpha) {
		panic(fmt.Errorf("alphaBeta alpha %v oldAlpha %v", score, oldAlpha))
	}
	if alpha != oldAlpha {
		moveGen.StorePvMove(pos, bestMove, bestScore, data.PVExact, depth)
	} else {
		moveGen.StorePvMove(pos, bestMove, alpha, data.PVAlpha, depth)
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

	if !(score > -30000) && !(score < 30000) {
		panic(fmt.Errorf("quiescence score error  %v", score))
	}
	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := &data.MoveList{}
	moveGen.GenerateAllCaptures(pos, ml)
	legal := 0
	for i := 0; i < ml.Count; i++ {
		pickNextMove(i, ml)
		if moveGen.MakeMove(ml.Moves[i].Move, pos) {
			score = -quiescence(-beta, -alpha, pos, info)
			moveGen.TakeMoveBack(pos)
			legal++
			if info.Stopped == data.True {
				return 0
			}
			if score > alpha {
				if score >= beta {
					if legal == 1 {
						info.FailHighFirst++
					}
					info.FailHigh++
					return beta
				}
				alpha = score
			}
		}
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
			pos.SearchKillers[i][j] = 0
		}
	}

	pos.Play = 0

	pos.PvTable.Cut = 0
	pos.PvTable.Hit = 0

	info.StartTime = util.GetTimeMs()
	info.Stopped = 0
	info.Node = 0
	info.FailHighFirst = 0
	info.FailHigh = 0
	info.NullCut = 0
	info.Cut = 0
}

func pickNextMove(moveNum int, ml *data.MoveList) {
	bestScore := 0
	bestNum := moveNum
	for i := moveNum; i < ml.Count; i++ {
		if ml.Moves[i].Score > bestScore {
			bestScore = ml.Moves[i].Score
			bestNum = i
		}
	}
	if moveNum < 0 || moveNum > ml.Count {
		panic(fmt.Errorf("pickNextMove: moveNum %v", moveNum))
	}

	if bestNum < 0 || bestNum > ml.Count {
		panic(fmt.Errorf("pickNextMove: bestNum %v", bestNum))
	}

	if bestNum < moveNum {
		panic(fmt.Errorf("pickNextMove: bestNum %v moveNum %v", bestNum, moveNum))
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
