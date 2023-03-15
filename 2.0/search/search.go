package search

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

var errTimeout = errors.New("Search timeout")

func (h *EngineHolder) Search(depth int) {
	h.ClearForSearch()

	var wg sync.WaitGroup
	done := make(chan struct{})

	for _, engine := range h.Engines {
		wg.Add(1)
		fmt.Printf("worker added %v\n", engine.Position.PositionKey)
		go func(e *Engine) {
			defer wg.Done()
			e.SearchRoot(depth)
		}(engine)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
}

func (e *EngineHolder) ClearForSearch() {
	e.TranspositionTable = engine.TranspositionTable
}

func (e *Engine) SearchRoot(depth int) {
	defer recoverFromTimeout()
	fmt.Printf("worker searching\n")
	timeNow := util.GetTimeMs()
	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		score := e.alphaBeta(-30000, 30000, currentDepth, 0, true)
		if e.IsMainEngine {
			fmt.Printf("Score:%v Depth %v Nodes:%v time %v hit: %v cut:%v\n", score, currentDepth, e.Parent.NodeCount, util.GetTimeMs()-timeNow, e.Parent.TranspositionTable.Hit, e.Parent.TranspositionTable.Cut)
			fmt.Printf("Ordering: %.2f\n", e.Position.FailHighFirst/e.Position.FailHigh)
		}
	}

	if e.IsMainEngine {
		e.Parent.CancelSearch()
		fmt.Printf("best move %v \n", io.PrintMove(e.Parent.TranspositionTable.Probe(e.Position.PositionKey)))
	} else {
		fmt.Printf("Other thread ends\n")
	}
}

func recoverFromTimeout() {
	err := recover()
	if err != nil && err != errTimeout {
		panic(err)
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
		return e.quiescence(alpha, beta, searchHeight)
	}

	e.Checkup()

	e.NodesVisited++

	if e.isRepetitionOrFiftyMove() {
		return 0
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
		if e.Position.IsKingAttacked(e.Position.Side) {
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
func (e *Engine) quiescence(alpha, beta, searchHeight int) int {
	e.Position.CheckBitboard()

	if e.isRepetitionOrFiftyMove() {
		return 0
	}

	e.NodesVisited++

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

	holder := ml.Moves[moveNum]
	ml.Moves[moveNum] = ml.Moves[bestNum]
	ml.Moves[bestNum] = holder
}

func (e *Engine) isRepetitionOrFiftyMove() bool {
	return false
}

func (e *Engine) Checkup() {
	if (e.NodesVisited % 2048) == 0 {
		select {
		case <-e.Parent.Ctx.Done():
			fmt.Printf("Ending early (%v)\n", e.IsMainEngine)
			panic(errTimeout)
		default:
		}
	}
}
