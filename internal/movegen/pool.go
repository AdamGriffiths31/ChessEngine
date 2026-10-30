package movegen

import (
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	preAllocCapacity  = 128
	maxPooledCapacity = 512
)

// MoveListPool manages a pool of reusable MoveList objects to reduce allocation overhead.
type MoveListPool struct {
	pool sync.Pool
}

var globalMoveListPool = &MoveListPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &MoveList{
				Moves: make([]board.Move, 0, preAllocCapacity),
				Count: 0,
			}
		},
	},
}

// GetMoveList retrieves a clean MoveList from the pool for reuse.
// Always returns a list with Count=0 and empty Moves slice.
// Must be paired with ReleaseMoveList() when done to return it to the pool.
func GetMoveList() *MoveList {
	item := globalMoveListPool.pool.Get()
	ml, ok := item.(*MoveList)
	if !ok {
		return &MoveList{}
	}
	ml.Clear()
	return ml
}

// ReleaseMoveList returns a MoveList to the pool for reuse.
// Lists with excessive capacity are discarded to prevent memory bloat.
// Safe to call with nil - will be ignored.
func ReleaseMoveList(ml *MoveList) {
	if ml == nil {
		return
	}

	if cap(ml.Moves) <= maxPooledCapacity {
		ml.Clear()
		globalMoveListPool.pool.Put(ml)
	}
}
