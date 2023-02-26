package moveGen

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func GetPvLine(depth int, pos *data.Board) int {
	move := ProbePvTableSingle(pos)
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
		move = ProbePvTableSingle(pos)
	}

	takeMovesBack(count, pos)
	return count
}

func StorePvMove(pos *data.Board, move, score, flag, depth int) {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)

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
}

func ProbePvTable(pos *data.Board, move *int, score *int, alpha, beta, depth int) int {
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
				*score = alpha
				return data.True
			case data.PVBeta:
				*score = beta
				return data.True
			case data.PVExact:
				return data.True
			}
		}
	}
	return data.False
}

func ProbePvTableSingle(pos *data.Board) int {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	if pos.PvTable.PTable[index].PosistionKey == pos.PosistionKey {
		return pos.PvTable.PTable[index].Move
	}
	return data.NoMove
}

func showTable(table *data.PVTable) {
	for i := range table.PTable {
		fmt.Printf("%v %v %v\n", i, table.PTable[i].PosistionKey, table.PTable[i].Move)
	}
}

func ClearTable(table *data.PVTable) {
	for i := range table.PTable {
		table.PTable[i].PosistionKey = 0
		table.PTable[i].Move = 0
		table.PTable[i].Depth = 0
		table.PTable[i].Score = 0
		table.PTable[i].Flag = 0
	}
}

func takeMovesBack(count int, pos *data.Board) {
	for i := 0; i < count; i++ {
		TakeMoveBack(pos)
	}
}
