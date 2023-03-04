package moveGen

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func GetPvLine(depth int, pos *data.Board) int {
	move := ProbePvMove(pos)
	count := 0
	for move != data.NoMove && count < depth {
		if MoveExists(pos, move) {
			MakeMove(move, pos)
			pos.PvArray[count] = move
			count++
		} else {
			io.PrintBoard(pos)
			fmt.Printf("GetPvLine break %v [%v](depth:%v)\n", move, io.PrintMove(move), depth)
			break
		}
		move = ProbePvMove(pos)
	}

	takeMovesBack(pos)

	return count
}

func StorePvMove(pos *data.Board, move, score, flag, depth int) {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)

	replace := false

	if pos.PvTable.PTable[index].PosistionKey == 0 {
		replace = true
	} else {
		if pos.PvTable.PTable[index].Age < pos.PvTable.CurrentAge || pos.PvTable.PTable[index].Depth <= depth {
			replace = true
		}
	}

	if !replace {
		return
	}

	if score > data.Mate {
		score += pos.Play
	} else if score < -data.Mate {
		score -= pos.Play
	}
	pos.PvTable.PTable[index].Move = move
	pos.PvTable.PTable[index].PosistionKey = pos.PosistionKey
	pos.PvTable.PTable[index].Depth = depth
	pos.PvTable.PTable[index].Score = score
	pos.PvTable.PTable[index].Flag = flag
	pos.PvTable.PTable[index].Age = pos.PvTable.CurrentAge
}

func ProbePvTable(pos *data.Board, move *int, score *int, alpha, beta, depth int) bool {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	if pos.PvTable.PTable[index].PosistionKey == pos.PosistionKey {
		*move = pos.PvTable.PTable[index].Move
		if pos.PvTable.PTable[index].Depth >= depth {
			pos.PvTable.Hit++
			*score = pos.PvTable.PTable[index].Score
			if *score > data.Mate {
				*score -= pos.Play
			} else if *score < -data.Mate {
				*score += pos.Play
			}
			switch pos.PvTable.PTable[index].Flag {
			case data.PVAlpha:
				if *score <= alpha {
					*score = alpha
					return true
				}
			case data.PVBeta:
				if *score >= beta {
					*score = beta
					return true
				}
			case data.PVExact:
				return true
			default:
				panic(fmt.Errorf("ProbePvTable: flag was not found"))
			}
		}
	}
	return false
}

func ProbePvMove(pos *data.Board) int {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	if pos.PvTable.PTable[index].PosistionKey == pos.PosistionKey {
		return pos.PvTable.PTable[index].Move
	}
	return data.NoMove
}

func ClearTable(table *data.PVTable) {
	for i := range table.PTable {
		table.PTable[i].PosistionKey = 0
		table.PTable[i].Move = 0
		table.PTable[i].Depth = 0
		table.PTable[i].Score = 0
		table.PTable[i].Flag = 0
		table.PTable[i].Age = 0
	}
	table.CurrentAge = 0
}

func takeMovesBack(pos *data.Board) {
	for pos.Play != 0 {
		TakeMoveBack(pos)
	}
}
