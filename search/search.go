package search

import (
	"fmt"
	"math"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/evaluate"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
	polyglot "github.com/AdamGriffiths31/ChessEngine/polyGlot"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

// SearchPosition starts the iterative deepening alpha beta search
func SearchPosition(pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) {
	bestMove := data.NoMove
	clearForSearch(pos, info, table)
	if data.EngineSettings.UseBook {
		bestMove = polyglot.GetBookMove(pos)
	}

	if bestMove == data.NoMove {
		createWorkers(pos, info, table)
	} else {
		printSearchResult(pos, info, bestMove)
	}
}

func createWorkers(pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) {
	fmt.Printf("Creating workers\n")
	var wg sync.WaitGroup
	var workerSlice []*data.SearchWorker
	for i := 0; i < info.WorkerNumber; i++ {
		workerSlice = append(workerSlice, setupWorker(i, pos, info, table))
	}
	for i := 0; i < info.WorkerNumber; i++ {
		wg.Add(1)
		go func(number int) {
			iterativeDeepen(workerSlice[number], &wg)
		}(i)
	}

	wg.Wait()
}

func setupWorker(number int, pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) *data.SearchWorker {
	posCopy := board.Clone(pos)
	return &data.SearchWorker{Pos: posCopy, Info: info, Hash: table, Number: number}
}

// iterativeDeepen the search performed by the worker
func iterativeDeepen(worker *data.SearchWorker, wg *sync.WaitGroup) {
	defer wg.Done()
	worker.BestMove = data.NoMove

	for currentDepth := 1; currentDepth < worker.Info.Depth+1; currentDepth++ {
		bestScore := alphaBeta(-data.ABInfinite, data.ABInfinite, currentDepth, worker.Pos, worker.Info, true, worker.Hash)
		if worker.Info.Stopped {
			break
		}
		if worker.Number == 0 {
			pvMoves := moveGen.GetPvLine(currentDepth, worker.Pos, worker.Hash)
			worker.BestMove = worker.Pos.PvArray[0]
			printPVData(worker.Info, currentDepth, bestScore)
			printPVLine(worker.Pos, worker.Info, pvMoves)
		}
	}
	if worker.Number == 0 {
		printSearchResult(worker.Pos, worker.Info, worker.BestMove)
	}
}

// printPVData prints the principle variation data
func printPVData(info *data.SearchInfo, currentDepth, bestScore int) {
	if info.GameMode == data.UCIMode {
		fmt.Printf("info score cp %d depth %d nodes %v time %d ", bestScore, currentDepth, info.Node, util.GetTimeMs()-info.StartTime)
	} else if info.GameMode == data.XboardMode && info.PostThinking {
		fmt.Printf("%d %d %d %v\n", currentDepth, bestScore, (util.GetTimeMs()-info.StartTime)/10, info.Node)
	} else if info.PostThinking {
		fmt.Printf("depth: %d score: %d time:%d  nodes:%v\n", currentDepth, bestScore, (util.GetTimeMs()-info.StartTime)/10, info.Node)
	}
}

// printPVLine prints the principle variation line data
func printPVLine(pos *data.Board, info *data.SearchInfo, pvMoves int) {
	if info.GameMode == data.UCIMode || info.PostThinking {
		fmt.Printf("pv")
		for _, move := range pos.PvArray[:pvMoves] {
			fmt.Printf(" %s", io.PrintMove(move))
		}
		fmt.Printf("\n")
		fmt.Printf("Ordering: %.2f\n", info.FailHighFirst/info.FailHigh)
	}
}

// printSearchResult prints the result of the search
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

// alphaBeta generates the score / best move for a position using quiescence, transposition tables
// and null move pruning
func alphaBeta(alpha, beta, depth int, pos *data.Board, info *data.SearchInfo, doNull bool, table *data.PvHashTable) int {
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

	checkUp(info)

	info.Node++

	if isRepetitionOrFiftyMove(pos) {
		return 0
	}

	if pos.Play > data.MaxDepth-1 {
		return evaluate.EvalPosition(pos)
	}

	// limited value in checking the position if in check
	inCheck := attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos)
	if inCheck {
		depth++
	}

	// check principle variation for a transposition score to save repeating searches
	score := -data.ABInfinite
	pvMove := data.NoMove
	if moveGen.ProbePvTable(pos, &pvMove, &score, alpha, beta, depth, table) {
		table.HashTable.Cut++
		return score
	}

	// null move check reduces the search by trying a 'null' move, then seeing if the score
	// of the subtree search is still high enough to cause a beta cutoff. Nodes are saved by
	//reducing the depth of the subtree under the null move.
	if doNull && !inCheck && pos.Play != 0 && pos.BigPiece[pos.Side] > 1 && depth >= 4 {
		moveGen.MakeNullMove(pos)
		score = -alphaBeta(-beta, -beta+1, depth-4, pos, info, false, table)
		moveGen.TakeBackNullMove(pos)
		if info.Stopped {
			return 0
		}
		if score >= beta && math.Abs(float64(score)) < data.Mate {
			info.NullCut++
			return beta
		}
	}

	ml := &data.MoveList{}
	moveGen.GenerateAllMoves(pos, ml)

	legal := 0
	oldAlpha := alpha
	bestMove := data.NoMove
	score = -data.ABInfinite
	bestScore := -data.ABInfinite

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
			score = -alphaBeta(-beta, -alpha, depth-1, pos, info, true, table)
			moveGen.TakeMoveBack(pos)
			if info.Stopped {
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
						moveGen.StorePvMove(pos, bestMove, beta, data.PVBeta, depth, table)
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

	// if no legal moves then the position must be either mate or stalemate
	if legal == 0 {
		if attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos) {
			return -data.ABInfinite + pos.Play
		} else {
			return 0
		}
	}
	if !(alpha >= oldAlpha) {
		panic(fmt.Errorf("alphaBeta alpha %v oldAlpha %v", score, oldAlpha))
	}
	if alpha != oldAlpha {
		moveGen.StorePvMove(pos, bestMove, bestScore, data.PVExact, depth, table)
	} else {
		moveGen.StorePvMove(pos, bestMove, alpha, data.PVAlpha, depth, table)
	}
	return alpha
}

// quiescence the purpose of this search is to only evaluate "quiet" positions, or
// positions where there are no winning tactical moves to be made. This search is
// needed to avoid the horizon effect.
// TODO Delta Pruning
func quiescence(alpha, beta int, pos *data.Board, info *data.SearchInfo) int {
	board.CheckBoard(pos)

	checkUp(info)

	info.Node++

	if isRepetitionOrFiftyMove(pos) {
		return 0
	}

	if pos.Play > data.MaxDepth-1 {
		return evaluate.EvalPosition(pos)
	}

	score := evaluate.EvalPosition(pos)

	if !(score > -data.ABInfinite) && !(score < data.ABInfinite) {
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
			if info.Stopped {
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

// clearForSearch resets data used in search
func clearForSearch(pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) {
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

	table.HashTable.Cut = 0
	table.HashTable.Hit = 0
	table.HashTable.CurrentAge++

	info.StartTime = util.GetTimeMs()
	info.Stopped = false
	info.Node = 0
	info.FailHighFirst = 0
	info.FailHigh = 0
	info.NullCut = 0
	info.Cut = 0
}

// pickNextMove improves the order the moves are searched in
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

// checkUp determines if the engine needs to stop searching and return its findings
func checkUp(info *data.SearchInfo) {
	if info.Node&2047 == 0 {
		if info.TimeSet == data.True && util.GetTimeMs() > info.StopTime {
			info.Stopped = true
		}
	}
}

// isRepetitionOrFiftyMove evaluate if its a repetition or a fifty move draw
func isRepetitionOrFiftyMove(pos *data.Board) bool {
	return isRepetition(pos) || pos.FiftyMove >= 100 && pos.Play > 0
}

// isRepetition check if the position has been seen before
// TODO Convert this to var repetitionTable = make(map[uint64]int) in data
func isRepetition(pos *data.Board) bool {
	for i := pos.HistoryPlay - pos.FiftyMove; i < pos.HistoryPlay-1; i++ {
		if pos.PositionKey == pos.History[i].PositionKey {
			return true
		}
	}
	return false
}
