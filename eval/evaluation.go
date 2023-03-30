package eval

import (
	custom "github.com/AdamGriffiths31/ChessEngine/eval/custom"
	pesto "github.com/AdamGriffiths31/ChessEngine/eval/pesto"
)

func Get(key string) func() interface{} {
	return func() interface{} {
		switch key {
		case "custom":
			return custom.NewEvaluationService()
		case "pesto":
			return pesto.NewEvaluationService()
		}

		panic("unknown evaluator")
	}
}
