package eval

import (
	pesto "github.com/AdamGriffiths31/ChessEngine/eval/pesto"
)

func Get(key string) func() interface{} {
	return func() interface{} {
		switch key {
		case "pesto":
			return pesto.NewEvaluationService()
		}
		panic("unknown evaluator")
	}
}
