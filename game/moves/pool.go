package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"sync"
)

// MoveListPool manages a pool of reusable MoveList objects to reduce allocation overhead.
// Uses sync.Pool for thread-safe object pooling with automatic garbage collection integration.
type MoveListPool struct {
	pool sync.Pool
}

// Global pool instance
var globalMoveListPool = &MoveListPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &MoveList{
				Moves: make([]board.Move, 0, PoolPreAllocCapacity),
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
		// This should never happen if the pool is used correctly
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

	// Only pool lists with reasonable capacity to avoid memory bloat
	if cap(ml.Moves) <= MaxMoveListCapacity {
		ml.Clear()
		globalMoveListPool.pool.Put(ml)
	}
}
