package search

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
)

func (e *Engine) AlphaBetatest(alpha, beta, depthLeft, searchHeight int, nullAllowed bool, alpha1, beta1, depth int, pos *data.Board, info *data.SearchInfo, doNull bool, table *data.PvHashTable) int {
	e.Position.CheckBitboard()
	if depthLeft < 0 || depth < 0 {
		panic(fmt.Errorf("alphaBeta depth was  %v %v", depthLeft, depth))
	}
	if beta < alpha || beta1 < alpha1 {
		panic(fmt.Errorf("alphaBeta beta %v < alpha %v", beta, alpha))
	}

	if depthLeft <= 0 {
		return e.Position.Evaluate()
		return e.quiescence(alpha, beta, searchHeight)
	}
	//TODO checkUp
	e.Parent.NodeCount++
	//TODO isRepetitionOrFiftyMove
	inCheck2 := attack.SquareAttacked(pos.KingSquare[pos.Side], pos.Side^1, pos)
	inCheck := e.Position.IsKingAttacked(e.Position.Side ^ 1)
	if inCheck {
		depthLeft++
		depth++
	}
	if inCheck != inCheck2 {
		e.Position.Board.PrintBoard()
		io.PrintBoard(pos)
		e.Position.IsKingAttacked(e.Position.Side)
		fmt.Printf("new %v old %v\n", e.Position.Side, pos.Side)
		panic("err")
	}

	score := -data.ABInfinite
	pvMove := data.NoMove
	if e.Parent.TranspositionTable.Get(e.Position.PositionKey, e.Position.Play, &pvMove, &score, alpha, beta, depthLeft) {
		e.Parent.TranspositionTable.Cut++
		return score
	}

	//score2 := -data.ABInfinite
	pvMove2 := data.NoMove
	if moveGen.ProbePvTable(pos, &pvMove2, &score, alpha1, beta1, depth, table) {
		table.HashTable.Cut++
		return score
	}

	if pvMove != pvMove2 {
		fmt.Printf("new %v - old %v\n", io.PrintMove(pvMove), io.PrintMove(pvMove2))
		panic("err")

	}

	//TODO NULL Move
	// doNullMove := nullAllowed && !inCheck && e.Position.Play != 0 && depthLeft >= 4 && !e.Position.IsEndGame()
	// if doNullMove {
	// 	_, enPas, castle := e.Position.MakeNullMove()
	// 	score = -e.alphaBetatest(-beta, -beta+1, depthLeft-4, searchHeight+1, false, -beta1, -alpha1, depth-1, pos, info, true, table)
	// 	e.Position.TakeNullMoveBack(enPas, castle)
	// 	if score >= beta && math.Abs(float64(score)) < data.Mate {
	// 		return beta
	// 	}
	// }

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
		moveGen.MakeMove(ml.Moves[i].Move, pos)
		legal++
		score = -e.AlphaBetatest(-beta, -alpha, depthLeft-1, searchHeight+1, true, -beta1, -alpha1, depth-1, pos, info, true, table)
		moveGen.TakeMoveBack(pos)
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
					fmt.Printf("Storing %v for depth %v\n", io.PrintMove(bestMove), depth)
					moveGen.StorePvMove(pos, bestMove, beta, data.PVBeta, depth, table)
					fmt.Printf("Storing %v for depth %v\n\n", io.PrintMove(bestMove), depthLeft)
					e.Parent.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, beta, data.PVBeta, depthLeft)
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
		//moveGen.StorePvMove(pos, bestMove, bestScore, data.PVExact, depth, table)
		//e.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, bestScore, data.PVExact, depthLeft)
	} else {
		//moveGen.StorePvMove(pos, bestMove, alpha, data.PVAlpha, depth, table)
		//e.TranspositionTable.Store(e.Position.PositionKey, e.Position.Play, bestMove, alpha, data.PVAlpha, depthLeft)
	}
	return alpha
}
