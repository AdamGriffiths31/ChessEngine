package movegen

import (
	"fmt"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

var pvSize = 0x100000 * 2

func InitPvTable(table *data.PVTable) {
	table.NumberEntries = pvSize / int(unsafe.Sizeof(data.PVEntry{}))
	table.NumberEntries -= 2
	table.PTable = make([]data.PVEntry, table.NumberEntries)
	fmt.Printf("InitPvTable completed with: %d\n", table.NumberEntries)
}

func GetPvLine(depth int, pos *data.Board) int {
	move := ProbePvTable(pos)
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
		move = ProbePvTable(pos)
	}

	takeMovesBack(count, pos)
	return count
}

func StorePvMove(pos *data.Board, move int) {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	pos.PvTable.PTable[index].Move = move
	pos.PvTable.PTable[index].PosistionKey = pos.PosistionKey
}

func ProbePvTable(pos *data.Board) int {
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
	}
}

func takeMovesBack(count int, pos *data.Board) {
	for i := 0; i < count; i++ {
		TakeMoveBack(pos)
	}
}
