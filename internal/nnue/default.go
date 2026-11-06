package nnue

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

//go:embed default.nnue
var defaultNetFile []byte

// Default returns the network built into the binary.
var Default = sync.OnceValue(func() *Network {
	net, err := Read(bytes.NewReader(defaultNetFile))
	if err != nil {
		panic("nnue: embedded default network: " + err.Error())
	}
	return net
})

// HandWritten selects the hand-written eval in EvaluatorFor.
const HandWritten = "hce"

// EvaluatorFor returns the built-in network for an empty spec, the
// hand-written eval for HandWritten, and otherwise the network file at spec.
func EvaluatorFor(spec string) (eval.Evaluator, error) {
	switch spec {
	case "":
		return NewDefaultEvaluator(), nil
	case HandWritten:
		return eval.NewEvaluator(), nil
	}
	net, err := Load(spec)
	if err != nil {
		return nil, err
	}
	return NewEvaluator(net), nil
}
