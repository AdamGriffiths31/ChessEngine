package search

import (
	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

type Engine struct {
	Position           *engine.Position
	TranspositionTable *engine.Cache
	SearchHistory      MoveHistory
	Nodes              uint64
	IsMainEngine       bool
}

type EngineHolder struct {
	Engines []*Engine
	PvArray [64]int
	Move    data.Move
}

func (r *EngineHolder) ResetHistory() {
	for i := 0; i < len(r.Engines); i++ {
		r.Engines[i].SearchHistory = MoveHistory{}
	}
}
