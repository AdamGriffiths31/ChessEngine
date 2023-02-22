package engine

import (
	"fmt"
	"unsafe"
)

var pvSize = 0x100000 * 2

func InitPvTable(table *PVTable) {
	table.NumberEntries = pvSize / int(unsafe.Sizeof(PVEntry{}))
	table.NumberEntries -= 2
	table.PTable = make([]PVEntry, table.NumberEntries)
	fmt.Printf("InitPvTable completed with: %d\n", table.NumberEntries)
}

func GetPvLine(depth int, pos *Board) int {
	//PrintBoard(pos)
	move := ProbePvTable(pos)
	count := 0
	for move != NoMove && count < depth {
		fmt.Printf("Move:%v\n", PrintMove(move))
		if MoveExists(pos, move) {
			MakeMove(move, pos)
			fmt.Printf("Move added %v at %v\n", PrintMove(move), count)
			pos.PvArray[count] = move
			count++
		} else {
			PrintBoard(pos)
			fmt.Printf("GetPvLine break %v [%v](depth:%v)\n", move, PrintMove(move), depth)
			break
		}
		move = ProbePvTable(pos)
	}
	for pos.Play != 0 {
		TakeMoveBack(pos)
	}
	return count
}

func StorePvMove(pos *Board, move int) {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	fmt.Printf("key: %v move: %v index:%v\n", pos.PosistionKey, PrintMove(move), index)
	pos.PvTable.PTable[index].Move = move
	pos.PvTable.PTable[index].PosistionKey = pos.PosistionKey
}

func ProbePvTable(pos *Board) int {
	index := pos.PosistionKey % uint64(pos.PvTable.NumberEntries)
	fmt.Printf("Retrival for %v pos: %v index: %v", PrintMove(pos.PvTable.PTable[index].Move), pos.PvTable.PTable[index].PosistionKey, index)
	if pos.PvTable.PTable[index].PosistionKey == pos.PosistionKey {
		return pos.PvTable.PTable[index].Move
	}
	return NoMove
}

func showTable(table *PVTable) {
	for i := range table.PTable {
		fmt.Printf("%v %v %v\n", i, table.PTable[i].PosistionKey, table.PTable[i].Move)
	}
}

func clearTable(table *PVTable) {
	for i := range table.PTable {
		table.PTable[i].PosistionKey = 0
		table.PTable[i].Move = 0
	}
}
