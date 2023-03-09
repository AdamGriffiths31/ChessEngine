package search

import (
	"fmt"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
)

func createWorkers(pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) {
	fmt.Printf("Creating workers\n")
	result := make(chan data.Move)
	var wg sync.WaitGroup
	var workerSlice []*data.SearchWorker
	for i := 0; i < info.WorkerNumber; i++ {
		workerSlice = append(workerSlice, setupWorker(i, pos, info, table))
	}
	for i := 0; i < info.WorkerNumber; i++ {
		wg.Add(1)
		go func(number int) {
			iterativeDeepen(workerSlice[number], result, &wg)
		}(i)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	bestResult := <-result
	for r := range result {
		fmt.Printf("%v\n", io.PrintMove(r.Move))
		if r.Score > bestResult.Score && r.Depth >= workerSlice[0].Depth {
			bestResult = r
		}
	}
	printSearchResult(pos, info, bestResult.Move)
}

// setupWorker clones the Board for the worker
func setupWorker(number int, pos *data.Board, info *data.SearchInfo, table *data.PvHashTable) *data.SearchWorker {
	posCopy := board.Clone(pos)
	return &data.SearchWorker{Pos: posCopy, Info: info, Hash: table, Number: number}
}

// iterativeDeepen the search performed by the worker
func iterativeDeepen(worker *data.SearchWorker, result chan data.Move, wg *sync.WaitGroup) {
	defer wg.Done()
	worker.BestMove = data.NoMove
	worker.BestScore = -data.Infinite
	for currentDepth := 1; currentDepth < worker.Info.Depth+1; currentDepth++ {
		worker.BestScore = alphaBeta(-data.ABInfinite, data.ABInfinite, currentDepth, worker.Pos, worker.Info, true, worker.Hash)
		if worker.Info.Stopped {
			break
		}

		moveGen.GetPvLine(currentDepth, worker.Pos, worker.Hash)
		worker.BestMove = worker.Pos.PvArray[0]
		worker.Depth = currentDepth
	}
	result <- data.Move{Score: worker.BestScore, Move: worker.BestMove, Depth: worker.Depth}
}
