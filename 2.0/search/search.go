package search

import (
	"fmt"
	"math"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func (h *EngineHolder) Search(depth int) {
	var wg sync.WaitGroup
	for i := 0; i < len(h.Engines); i++ {
		wg.Add(1)
		fmt.Printf("worker added %v\n", h.Engines[i].Position.PositionKey)
		go h.Engines[i].ParallelSearch(&wg, depth)
	}
	wg.Wait()
}

func (e *Engine) ParallelSearch(wg *sync.WaitGroup, depth int) {
	e.ClearForSearch()
	e.SearchRoot(wg, depth)
}

func (e *Engine) ClearForSearch() {
	e.TranspositionTable = engine.TranspositionTable
}

func (e *Engine) SearchRoot(wg *sync.WaitGroup, depth int) {
	defer wg.Done()
	fmt.Printf("worker searching\n")
	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		score := e.alphaBeta(-30000, 30000, currentDepth, 0, true)
		if e.IsMainEngine {
			fmt.Printf("Score:%v Depth %v Nodes:%v hit: %v cut:%v\n", score, currentDepth, e.Nodes, e.TranspositionTable.Hit, e.TranspositionTable.Cut)
			fmt.Printf("Ordering: %.2f\n", e.Position.FailHighFirst/e.Position.FailHigh)
		} else {
			fmt.Printf("\tSecond runner : depth %v nodes:%v\n", currentDepth, e.Nodes)
		}
	}
	if e.IsMainEngine {
		e.Parent.CancelSearch()
		fmt.Printf("best move %v \n", io.PrintMove(e.TranspositionTable.Probe(e.Position.PositionKey)))
	}
}

func (e *Engine) alphaBeta(alpha, beta, depthLeft, searchHeight int, nullAllowed bool) int {
	e.Position.CheckBitboard()
	if depthLeft < 0 {
		panic(fmt.Errorf("alphaBeta depth was  %v", depthLeft))
	}
	if beta < alpha {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depthLeft <= 0 {
		//return e.Position.Evaluate()
		return e.quiescence(alpha, beta, searchHeight)
	}
	//TODO checkUp
	e.Nodes++
	//TODO isRepetitionOrFiftyMove

	inCheck := e.Position.IsKingAttacked(e.Position.Side ^ 1)
	if inCheck {
		depthLeft++
	}

	score := -data.ABInfinite
	pvMove := data.NoMove
	if e.TranspositionTable.Get(e.Position.PositionKey, e.Position.Play, &pvMove, &score, alpha, beta, depthLeft) {
		e.TranspositionTable.Cut++
		return score
	}

	//TODO NULL Move
	doNullMove := nullAllowed && !inCheck && e.Position.Play != 0 && depthLeft >= 4 && !e.Position.IsEndGame()
	if doNullMove {
		_, enPas, castle := e.Position.MakeNullMove()
		score = -e.alphaBeta(-beta, -beta+1, depthLeft-4, searchHeight+1, false)
		e.Position.TakeNullMoveBack(enPas, castle)
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
		//fmt.Printf("%v %v\n", io.PrintMove(move), ml.Moves[i].Score)
		isAllowed, enPas, CastleRight := e.Position.MakeMove(ml.Moves[i].Move)
		if !isAllowed {
			continue
		}
		legal++
		score = -e.alphaBeta(-beta, -alpha, depthLeft-1, searchHeight+1, true)
		e.Position.TakeMoveBack(ml.Moves[i].Move, enPas, CastleRight)
		if score > bestScore {
			bestScore = score
			bestMove = ml.Moves[i].Move
			if score > alpha {
				if score >= beta {
					if legal == 1 {
						e.Position.FailHighFirst++
					}
					if ml.Moves[i].Move&data.MFLAGCAP == 0 {
						//e.SearchHistory.Killers[1][searchHeight] = e.SearchHistory.Killers[0][searchHeight]
						//e.SearchHistory.Killers[0][searchHeight] = move
					}
					e.Position.FailHigh++
					e.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, beta, data.PVBeta, depthLeft)
					return beta
				}
				alpha = score

				if ml.Moves[i].Move&data.MFLAGCAP == 0 {
					//e.SearchHistory.History[e.Position.Board.PieceAt(data.FromSquare(bestMove))][data.ToSquare(bestMove)] += searchHeight
				}
			}
		}

	}
	if legal == 0 {
		if e.Position.IsKingAttacked(e.Position.Side) {
			return -data.ABInfinite + searchHeight
		} else {
			return 0
		}
	}
	if !(alpha >= oldAlpha) {
		panic(fmt.Errorf("alphaBeta alpha %v oldAlpha %v", score, oldAlpha))
	}
	if alpha != oldAlpha {
		//fmt.Printf("storing %v (%v) for depth %v\n", io.PrintMove(bestMove), bestScore, depthLeft)
		e.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, bestScore, data.PVExact, depthLeft)
	} else {
		e.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, alpha, data.PVAlpha, depthLeft)
	}
	return alpha
}
func (e *Engine) quiescence(alpha, beta, searchHeight int) int {
	e.Position.CheckBitboard()
	//TODO isRepetitionOrFiftyMove
	e.Nodes++

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
		isAllowed, enPas, CastleRight := e.Position.MakeMove(move)
		if !isAllowed {
			continue
		}
		score = -e.quiescence(-beta, -alpha, searchHeight+1)
		e.Position.TakeMoveBack(move, enPas, CastleRight)
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
