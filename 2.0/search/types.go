package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

type Engine struct {
	Position     *engine.Position
	IsMainEngine bool
	Parent       *EngineHolder
	NodesVisited int
}

type EngineHolder struct {
	Engines            []*Engine
	PvArray            [64]int
	Move               data.Move
	Ctx                context.Context
	CancelSearch       context.CancelFunc
	TranspositionTable *engine.Cache
	NodeCount          uint64
}

func NewEngineHolder(numberOfThreads int) *EngineHolder {
	t := &EngineHolder{}
	t.Ctx, t.CancelSearch = context.WithCancel(context.Background())
	engines := make([]*Engine, numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		engine := NewEngine(t)
		if i == 0 {
			engine.IsMainEngine = true
		}
		engines[i] = engine
	}
	t.Engines = engines
	return t
}

func NewEngine(parent *EngineHolder) *Engine {
	return &Engine{Parent: parent, Position: nil}
}
