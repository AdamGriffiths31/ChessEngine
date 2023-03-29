package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

type Engine struct {
	Position     *engine.Position
	IsMainEngine bool
	Parent       *EngineHolder
	NodesVisited int
	evaluator    IUpdatableEvaluator
}

type EngineHolder struct {
	Engines            []*Engine
	PvArray            [64]int
	Move               data.Move
	Ctx                context.Context
	CancelSearch       context.CancelFunc
	TranspositionTable *engine.Cache
	NodeCount          uint64
	UseBook            bool
	EvalBuilder        func() interface{}
}

type IEvaluator interface {
	Evaluate(p *engine.Position) int
}

type IUpdatableEvaluator interface {
	Evaluate(p *engine.Position) int
}

type EvaluatorAdapter struct {
	evaluator IEvaluator
}

func (e *EvaluatorAdapter) Evaluate(p *engine.Position) int {
	return e.evaluator.Evaluate(p)
}

func NewEngineHolder(numberOfThreads int, evalBuilder func() interface{}) *EngineHolder {
	t := &EngineHolder{EvalBuilder: evalBuilder}
	t.Ctx, t.CancelSearch = context.WithCancel(context.Background())
	engines := make([]*Engine, numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		engine := NewEngine(t)
		if i == 0 {
			engine.IsMainEngine = true
		}
		engines[i] = engine
		engines[i].evaluator = t.buildEvaluator()
	}
	t.Engines = engines
	t.TranspositionTable = engine.NewCache()
	return t
}

func NewEngine(parent *EngineHolder) *Engine {
	return &Engine{Parent: parent, Position: nil}
}

func (h *EngineHolder) buildEvaluator() IUpdatableEvaluator {
	var evaluationService = h.EvalBuilder()
	if ue, ok := evaluationService.(IUpdatableEvaluator); ok {
		return ue
	}
	if e, ok := evaluationService.(IEvaluator); ok {
		return &EvaluatorAdapter{evaluator: e}
	}

	panic("unknown evaluator")
}
