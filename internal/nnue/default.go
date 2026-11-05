package nnue

import (
	"bytes"
	_ "embed"
	"sync"
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
