package search

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

var errTimeout = errors.New("Search timeout")

func (h *EngineHolder) Search(info *data.SearchInfo) {
	e := h.Engines[0]
	e.IsMainEngine = true
	if h.UseBook {
		bestMove := GetBookMove(e.Position)
		if bestMove != data.NoMove {
			fmt.Printf("bestmove %s\n", io.PrintMove(bestMove))
			return
		}
		fmt.Printf("No book move found for %v\n", e.Position.Side)
	}
	h.ClearForSearch()

	var wg sync.WaitGroup

	for _, engine := range h.Engines {
		wg.Add(1)
		fmt.Printf("worker added %v\n", e.Position.PositionKey)
		go func(e *Engine) {
			e.SearchRoot(info)
			wg.Done()
		}(engine)
	}

	wg.Wait()

	fmt.Printf("bestmove %v \n", io.PrintMove(h.Move.Move))

}

func (e *EngineHolder) ClearForSearch() {
	e.TranspositionTable.CurrentAge++
}

func (e *Engine) ClearForSearch() {
	e.Position.FiftyMove = 0
	e.Position.CurrentScore = 0
	e.Position.PositionHistory.ClearPositionHistory()

	for i := 0; i < 13; i++ {
		for j := 0; j < 120; j++ {
			e.Position.MoveHistory.History[i][j] = 0
		}
	}

	for i := 0; i < 2; i++ {
		for j := 0; j < data.MaxDepth; j++ {
			e.Position.MoveHistory.Killers[i][j] = 0
		}
	}
}

// SearchRoot start the search from the root position
func (e *Engine) SearchRoot(searchInfo *data.SearchInfo) {
	defer recoverFromTimeout()

	searchInfo.Stopped = false
	searchInfo.ForceStop = false
	window := 50
	e.ClearForSearch()
	alpha, beta := e.getInitialAlphaBeta()

	for depth := 1; depth <= searchInfo.Depth; depth++ {
		if e.IsMainEngine {
			fmt.Printf("History = %v\n", e.Position.PositionHistory.Count)
		}
		score := e.alphaBeta(alpha, beta, depth, 0, true, searchInfo)
		if searchInfo.Stopped {
			break
		}
		e.Position.PositionHistory.ClearPositionHistory()
		if score <= alpha || score >= beta {
			alpha, beta = e.getInitialAlphaBeta()
			continue
		}

		alpha = score - window
		beta = score + window

		if e.IsMainEngine {
			e.printSearchInfo(score, depth, searchInfo.Node, searchInfo.StartTime)
		}
	}

	if e.IsMainEngine {
		e.Parent.CancelSearch()
	}
}

// getInitialAlphaBeta sets the initial alpha and beta values
func (e *Engine) getInitialAlphaBeta() (alpha, beta int) {
	alpha = -data.ABInfinite
	beta = data.ABInfinite
	return
}

// printSearchInfo prints the search info
func (e *Engine) printSearchInfo(score, depth int, nodes int64, startTime int64) {
	bestMove := e.Parent.TranspositionTable.Probe(e.Position.PositionKey)
	e.Parent.Move.Move = bestMove
	fmt.Printf("info score cp %d depth %d nodes %v time %d pv %v\n", score, depth, nodes, util.GetTimeMs()-startTime, io.PrintMove(bestMove))
	//fmt.Printf("Ordering: %.2f\n", e.Position.FailHighFirst/e.Position.FailHigh)
}

// recoverFromTimeout if the search times out, recover from the panic
func recoverFromTimeout() {
	err := recover()
	if err != nil && err != errTimeout {
		panic(err)
	}
}

// alphaBeta performs the alpha beta search
func (e *Engine) alphaBeta(alpha, beta, depthLeft, searchHeight int, nullAllowed bool, info *data.SearchInfo) int {
	e.Position.CheckBitboard()
	if depthLeft < 0 {
		panic(fmt.Errorf("alphaBeta depth was  %v", depthLeft))
	}
	if beta < alpha {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depthLeft <= 0 {
		return e.quiescence(alpha, beta, searchHeight, info)
	}

	e.Checkup(info)

	e.NodesVisited++

	if e.isRepetitionOrFiftyMove() {
		return 0
	}

	if searchHeight > data.MaxDepth-1 {
		return e.Position.Evaluate()
	}

	inCheck := e.Position.IsKingAttacked(e.Position.Side ^ 1)
	if inCheck {
		depthLeft++
	}

	score := -data.ABInfinite
	pvMove := data.NoMove
	if e.Parent.TranspositionTable.Get(e.Position.PositionKey, e.Position.Play, &pvMove, &score, alpha, beta, depthLeft) {
		e.Parent.TranspositionTable.Cut++
		return score
	}

	doNullMove := nullAllowed && !inCheck && e.Position.Play != 0 && depthLeft >= 4 && !e.Position.IsEndGame()
	if doNullMove {
		_, enPas, castle := e.Position.MakeNullMove()
		e.Position.PositionHistory.AddPositionHistory(e.Position.PositionKey)
		score = -e.alphaBeta(-beta, -beta+1, depthLeft-4, searchHeight+1, false, info)
		e.Position.PositionHistory.RemovePositionHistory()
		e.Position.TakeNullMoveBack(enPas, castle)
		if info.Stopped {
			return 0
		}
		if score >= beta && math.Abs(float64(score)) < data.Mate {
			return beta
		}
	}

	ml := &engine.MoveList{}
	e.Position.GenerateAllMoves(ml)

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
		e.PickNextMove(i, ml)
		isAllowed, enPas, CastleRight, fifty := e.Position.MakeMove(ml.Moves[i].Move)
		if !isAllowed {
			continue
		}
		legal++
		score = -e.alphaBeta(-beta, -alpha, depthLeft-1, searchHeight+1, true, info)
		e.Position.TakeMoveBack(ml.Moves[i].Move, enPas, CastleRight, fifty)
		if info.Stopped {
			return 0
		}
		if score > bestScore {
			bestScore = score
			bestMove = ml.Moves[i].Move
			if score > alpha {
				if score >= beta {
					if legal == 1 {
						e.Position.FailHighFirst++
					}
					if ml.Moves[i].Move&data.MFLAGCAP == 0 {
						e.Position.MoveHistory.Killers[1][e.Position.Play] = e.Position.MoveHistory.Killers[0][e.Position.Play]
						e.Position.MoveHistory.Killers[0][e.Position.Play] = ml.Moves[i].Move
					}
					e.Position.FailHigh++
					e.Parent.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, beta, data.PVBeta, depthLeft)

					return beta
				}
				alpha = score

				if ml.Moves[i].Move&data.MFLAGCAP == 0 {
					e.Position.MoveHistory.History[e.Position.Board.PieceAt(data.Square120ToSquare64[data.FromSquare(bestMove)])][data.ToSquare(bestMove)] += e.Position.Play
				}
			}
		}

	}
	if legal == 0 {
		if e.Position.IsKingAttacked(e.Position.Side ^ 1) {
			return -data.ABInfinite + e.Position.Play
		} else {
			return 0
		}
	}
	if !(alpha >= oldAlpha) {
		panic(fmt.Errorf("alphaBeta alpha %v oldAlpha %v", score, oldAlpha))
	}
	if alpha != oldAlpha {
		e.Parent.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, bestScore, data.PVExact, depthLeft)
	} else {
		e.Parent.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, alpha, data.PVAlpha, depthLeft)
	}
	return alpha
}

// quiescence is the quiescence search function.
func (e *Engine) quiescence(alpha, beta, searchHeight int, info *data.SearchInfo) int {
	e.Position.CheckBitboard()

	if e.isRepetitionOrFiftyMove() {
		return 0
	}

	e.Checkup(info)

	e.NodesVisited++

	if searchHeight > data.MaxDepth-1 {
		return e.Position.Evaluate()
	}

	score := -data.ABInfinite
	pvMove := data.NoMove
	if e.Parent.TranspositionTable.Get(e.Position.PositionKey, e.Position.Play, &pvMove, &score, alpha, beta, 0) {
		e.Parent.TranspositionTable.Cut++
		return score
	}

	score = e.Position.Evaluate()

	if !(score > -data.ABInfinite) && !(score < data.ABInfinite) {
		panic(fmt.Errorf("quiescence score error  %v", score))
	}

	if score >= beta {
		return beta
	}

	bigDelta := 1000 //Queen Value

	if score < alpha-bigDelta {
		return alpha
	}

	if alpha < score {
		alpha = score
	}

	flag := data.PVAlpha
	bestMove := data.NoMove
	bestScore := -data.ABInfinite
	ml := &engine.MoveList{}
	e.Position.GenerateAllCaptures(ml)
	for i := 0; i < ml.Count; i++ {
		e.PickNextMove(i, ml)
		move := ml.Moves[i].Move
		isAllowed, enPas, CastleRight, fifty := e.Position.MakeMove(move)
		if !isAllowed {
			continue
		}
		score = -e.quiescence(-beta, -alpha, searchHeight+1, info)
		e.Position.TakeMoveBack(move, enPas, CastleRight, fifty)
		if info.Stopped {
			return 0
		}
		if score > bestScore {
			bestScore = score
			if score > alpha {
				alpha = score
				bestMove = move
				flag = data.PVExact
			}
		}
		if alpha >= beta {
			flag = data.PVBeta
			break
		}
	}

	e.Parent.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, bestScore, flag, 0)

	return bestScore
}

// PickNextMove picks the next move to be searched
func (e *Engine) PickNextMove(moveNum int, ml *engine.MoveList) {
	bestScore := 0
	bestNum := moveNum
	for i := moveNum; i < ml.Count; i++ {
		if ml.Moves[i].Score > bestScore {
			bestScore, bestNum = ml.Moves[i].Score, i
		}
	}

	holder := ml.Moves[moveNum]
	ml.Moves[moveNum] = ml.Moves[bestNum]
	ml.Moves[bestNum] = holder
}

// isRepetitionOrFiftyMove checks if the position is a repetition or a fifty move draw
func (e *Engine) isRepetitionOrFiftyMove() bool {
	if e.Position.FiftyMove >= 50 {
		return true
	}

	for i := e.Position.PositionHistory.Count - 2; i >= 0; i -= 2 {
		var candidate = e.Position.PositionHistory.History[i]
		if e.Position.PositionKey == candidate {
			return true
		}
	}

	previouslySeen := e.Position.Positions[e.Position.PositionKey]

	return previouslySeen >= 2
}

// Checkup checks if the search should be stopped
func (e *Engine) Checkup(info *data.SearchInfo) {
	if (e.NodesVisited % 2048) == 0 {
		if (info.TimeSet == data.True && util.GetTimeMs() > info.StopTime) || info.ForceStop {
			info.Stopped = true
		}
		select {
		case <-e.Parent.Ctx.Done():
			fmt.Printf("Ending early (%v)\n", e.IsMainEngine)
			panic(errTimeout)
		default:
		}
	}
}
