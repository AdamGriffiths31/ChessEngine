package moveGen

import (
	"fmt"
	"math/rand"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func GetPvLine(depth int, pos *data.Board, table *data.PvHashTable) int {
	move := ProbePvMove(pos, table)
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
		move = ProbePvMove(pos, table)
	}

	takeMovesBack(pos)

	return count
}

func StorePvMove(pos *data.Board, move, score, flag, depth int, table *data.PvHashTable) {
	index := pos.PositionKey % uint64(table.HashTable.NumberEntries)

	replace := false

	if table.HashTable.PTable[index].SMPKey == 0 {
		replace = true
	} else {
		if table.HashTable.PTable[index].Age < table.HashTable.CurrentAge {
			replace = true
		} else if extractDepth(table.HashTable.PTable[index].SMPData) <= uint64(depth) {
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

	smpData := foldData(uint64(score), uint64(depth), uint64(flag), move)
	smpKey := pos.PositionKey ^ smpData

	table.HashTable.PTable[index].Age = table.HashTable.CurrentAge
	table.HashTable.PTable[index].SMPData = smpData
	table.HashTable.PTable[index].SMPKey = smpKey
}

func ProbePvTable(pos *data.Board, move *int, score *int, alpha, beta, depth int, table *data.PvHashTable) bool {
	index := pos.PositionKey % uint64(table.HashTable.NumberEntries)
	testKey := pos.PositionKey ^ table.HashTable.PTable[index].SMPData
	if testKey == table.HashTable.PTable[index].SMPKey {
		*move = extractMove(table.HashTable.PTable[index].SMPData)
		if int(extractDepth(table.HashTable.PTable[index].SMPData)) >= depth {
			table.HashTable.Hit++
			*score = int(extractScore(table.HashTable.PTable[index].SMPData))
			if *score > data.Mate {
				*score -= pos.Play
			} else if *score < -data.Mate {
				*score += pos.Play
			}
			switch extractFlag(table.HashTable.PTable[index].SMPData) {
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

func ProbePvMove(pos *data.Board, table *data.PvHashTable) int {
	index := pos.PositionKey % uint64(table.HashTable.NumberEntries)
	testKey := pos.PositionKey ^ table.HashTable.PTable[index].SMPData
	if testKey == table.HashTable.PTable[index].SMPKey {
		return extractMove(table.HashTable.PTable[index].SMPData)
	}
	return data.NoMove
}

func ClearTable(table *data.PVTable) {
	for i := range table.PTable {
		table.PTable[i].Age = 0
		table.PTable[i].SMPData = 0
		table.PTable[i].SMPKey = 0
	}
	table.CurrentAge = 0
}

func takeMovesBack(pos *data.Board) {
	for pos.Play != 0 {
		TakeMoveBack(pos)
	}
}

func extractScore(value uint64) uint64 {
	return value&0xFFFF - data.Infinite
}

func extractDepth(value uint64) uint64 {
	return (value >> 16) & 0x3F
}

func extractFlag(value uint64) uint64 {
	return (value >> 23) & 0x3
}

func extractMove(value uint64) int {
	return int(value >> 25)
}

func foldData(score, depth, flag uint64, move int) uint64 {
	return (score + data.Infinite) | (depth << 16) | (flag << 23) | (uint64(move) << 25)
}

// func verifySMPData(entry data.PVEntry) {
// 	value := foldData(uint64(entry.Score), uint64(entry.Depth), uint64(entry.Flag), entry.Move)
// 	key := entry.PositionKey ^ value
// 	if value != entry.SMPData {
// 		panic(fmt.Errorf("verifySMPData: value was not the same as entry %v - %v", value, entry.SMPData))
// 	}
// 	if key != entry.SMPKey {
// 		panic(fmt.Errorf("verifySMPData: key was not the same as entry %v - %v", key, entry.SMPKey))
// 	}

// 	move := extractMove(value)
// 	depth := extractDepth(value)
// 	flag := extractFlag(value)
// 	score := extractScore(value)

// 	if move != entry.Move || depth != uint64(entry.Depth) || flag != uint64(entry.Flag) || score != uint64(entry.Score) {
// 		fmt.Printf("verifySMPData extract error\nMove %v-%v\nDepth %v-%v\nFlag %v-%v\nScore %v-%v", move, entry.Move, depth, uint64(entry.Depth), flag, uint64(entry.Flag), score, uint64(entry.Score))
// 	}
// }

func DataCheck(move int) {
	depth := rand.Uint64() % data.MaxDepth
	flag := rand.Uint64() % 3
	score := rand.Uint64() % data.ABInfinite

	data := foldData(score, depth, flag, move)
	fmt.Printf("original: move %s d:%d fl:%d sc:%d data: %v \n", io.PrintMove(move), depth, flag, score, data)
	fmt.Printf("check: move %s d:%d fl:%d sc:%d\n", io.PrintMove(extractMove(data)), extractDepth(data), extractFlag(data), extractScore(data))
}

func TempHashTest(fen string) {
	pos := &data.Board{}
	board.ParseFEN(fen, pos)
	ml := data.MoveList{}
	GenerateAllMoves(pos, &ml)
	for moveNum := 0; moveNum < ml.Count; moveNum++ {
		if !MakeMove(ml.Moves[moveNum].Move, pos) {
			continue
		}
		TakeMoveBack(pos)
		DataCheck(ml.Moves[moveNum].Move)
	}
}
