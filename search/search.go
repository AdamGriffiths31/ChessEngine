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

var moveListPool = sync.Pool{
	New: func() interface{} {
		return &engine.MoveList{}
	},
}

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
	done := make(chan struct{})

	for _, engine := range h.Engines {
		wg.Add(1)
		fmt.Printf("worker added %v\n", e.Position.PositionKey)
		go func(e *Engine) {
			e.SearchRoot(info)
			wg.Done()
		}(engine)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
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

func (e *Engine) SearchRoot(info *data.SearchInfo) {
	defer recoverFromTimeout()
	info.Stopped = false
	info.ForceStop = false
	bestMove := data.NoMove
	e.ClearForSearch()
	for currentDepth := 1; currentDepth <= info.Depth; currentDepth++ {
		if e.IsMainEngine {
			fmt.Printf("Searching depth %v\n", currentDepth)
		}
		score := e.alphaBeta(-30000, 30000, currentDepth, 0, true, info)
		if info.Stopped {
			break
		}
		if e.IsMainEngine {
			bestMove = e.Parent.TranspositionTable.Probe(e.Position.PositionKey)
			e.Parent.Move.Move = bestMove
			fmt.Printf("info score cp %d depth %d nodes %v time %d pv %v\n", score, currentDepth, info.Node, util.GetTimeMs()-info.StartTime, io.PrintMove(bestMove))
			//fmt.Printf("Ordering: %.2f\n", e.Position.FailHighFirst/e.Position.FailHigh)
		}
	}
	if e.IsMainEngine {
		e.Parent.CancelSearch()
	}
}

func recoverFromTimeout() {
	err := recover()
	if err != nil && err != errTimeout {
		panic(err)
	}
}

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
		e.Position.PositionHistory.ClearPositionHistory()
		e.Position.TakeNullMoveBack(enPas, castle)
		if info.Stopped {
			return 0
		}
		if score >= beta && math.Abs(float64(score)) < data.Mate {
			return beta
		}
	}

	ml := moveListPool.Get().(*engine.MoveList)
	defer moveListPool.Put(ml)
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
		e.Position.PositionHistory.AddPositionHistory(e.Position.PositionKey)
		score = -e.alphaBeta(-beta, -alpha, depthLeft-1, searchHeight+1, true, info)
		e.Position.PositionHistory.ClearPositionHistory()
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

	score := e.Position.Evaluate()

	if !(score > -data.ABInfinite) && !(score < data.ABInfinite) {
		panic(fmt.Errorf("quiescence score error  %v", score))
	}

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := &engine.MoveList{}
	e.Position.GenerateAllCaptures(ml)
	legal := 0
	for i := 0; i < ml.Count; i++ {
		e.PickNextMove(i, ml)
		move := ml.Moves[i].Move
		isAllowed, enPas, CastleRight, fifty := e.Position.MakeMove(move)
		if !isAllowed {
			continue
		}
		e.Position.PositionHistory.AddPositionHistory(e.Position.PositionKey)
		score = -e.quiescence(-beta, -alpha, searchHeight+1, info)
		e.Position.PositionHistory.ClearPositionHistory()
		e.Position.TakeMoveBack(move, enPas, CastleRight, fifty)
		if info.Stopped {
			return 0
		}
		legal++
		if score > alpha {
			if score >= beta {
				if legal == 1 {
					e.Position.FailHighFirst++
				}
				e.Position.FailHigh++
				return beta
			}
			alpha = score
		}
	}
	return alpha
}

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
