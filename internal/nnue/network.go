// Package nnue evaluates positions with a small quantized neural network:
// 768 piece-square inputs per side, one shared hidden layer per side, and a
// single output from both hidden layers, side to move first.
package nnue

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

const (
	// Inputs is one input per (piece, square) from one side's view.
	Inputs = 768
	// Hidden is the width of each side's hidden layer.
	Hidden = 128

	// QA scales the hidden layer: activations are clamped to [0, QA].
	QA = 255
	// QB scales the output weights.
	QB = 256
	// Scale converts the network's output to centipawns.
	Scale = 400

	magic = "GCNN"
)

// Network holds quantized weights.
type Network struct {
	FeatureWeights [Inputs * Hidden]int16
	FeatureBias    [Hidden]int16
	OutputWeights  [2 * Hidden]int16
	OutputBias     int32
}

// FeatureIndex returns the input for piece (0-11, the board's bitboard
// index: white P R N B Q K, then black) on square (a1 = 0) seen from
// perspective (0 white, 1 black). Black sees the board flipped, with the
// colours swapped.
func FeatureIndex(piece, square, perspective int) int {
	if perspective == 1 {
		piece = (piece + 6) % 12
		square ^= 56
	}
	return piece*64 + square
}

// Load reads a network written by Write.
func Load(path string) (*Network, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

// Read decodes a network written by Write.
func Read(r io.Reader) (*Network, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("nnue: reading header: %w", err)
	}
	if string(header[:4]) != magic {
		return nil, errors.New("nnue: not a network file")
	}
	if hidden := binary.LittleEndian.Uint32(header[4:]); hidden != Hidden {
		return nil, fmt.Errorf("nnue: network has %d hidden units, engine expects %d", hidden, Hidden)
	}
	n := &Network{}
	if err := binary.Read(r, binary.LittleEndian, n); err != nil {
		return nil, fmt.Errorf("nnue: reading weights: %w", err)
	}
	for i, w := range n.OutputWeights {
		if w < -math.MaxInt8 || w > math.MaxInt8 {
			return nil, fmt.Errorf("nnue: output weight %d is %d, outside ±%d, which the int16 SIMD multiply needs", i, w, math.MaxInt8)
		}
	}
	return n, nil
}

// Write encodes the network.
func (n *Network) Write(w io.Writer) error {
	header := make([]byte, 8)
	copy(header, magic)
	binary.LittleEndian.PutUint32(header[4:], Hidden)
	if _, err := w.Write(header); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, n)
}

// Evaluator scores positions with a network. It keeps state between calls,
// so one Evaluator serves one search at a time.
type Evaluator struct {
	net    *Network
	acc    [2][Hidden]int16
	pieces [12]board.Bitboard
}

// NewEvaluator returns an evaluator for net.
func NewEvaluator(net *Network) *Evaluator {
	e := &Evaluator{net: net}
	e.acc[0] = net.FeatureBias
	e.acc[1] = net.FeatureBias
	return e
}

// NewDefaultEvaluator returns an evaluator for the built-in network.
func NewDefaultEvaluator() *Evaluator {
	return NewEvaluator(Default())
}

// Evaluate returns the score from White's side, as eval.Evaluator requires.
func (e *Evaluator) Evaluate(b *board.Board) eval.EvaluationScore {
	e.update(&b.PieceBitboards)

	us, them := 0, 1
	whiteToMove := b.GetSideToMove() == "w"
	if !whiteToMove {
		us, them = 1, 0
	}
	score := e.output(&e.acc[us], &e.acc[them])
	if !whiteToMove {
		score = -score
	}
	return eval.EvaluationScore(score)
}

const refreshAbove = 24

func (e *Evaluator) update(pieces *[12]board.Bitboard) {
	changed := 0
	for piece := range pieces {
		changed += (pieces[piece] ^ e.pieces[piece]).PopCount()
	}
	if changed > refreshAbove {
		e.refresh(pieces)
		return
	}
	for piece := range pieces {
		for removed := e.pieces[piece] &^ pieces[piece]; removed != 0; {
			var square int
			square, removed = removed.PopLSB()
			e.subFeature(piece, square)
		}
		for added := pieces[piece] &^ e.pieces[piece]; added != 0; {
			var square int
			square, added = added.PopLSB()
			e.addFeature(piece, square)
		}
		e.pieces[piece] = pieces[piece]
	}
}

func (e *Evaluator) refresh(pieces *[12]board.Bitboard) {
	e.acc[0] = e.net.FeatureBias
	e.acc[1] = e.net.FeatureBias
	for piece, bb := range pieces {
		for bb != 0 {
			var square int
			square, bb = bb.PopLSB()
			e.addFeature(piece, square)
		}
	}
	e.pieces = *pieces
}

func (e *Evaluator) addFeature(piece, square int) {
	addColumn(&e.acc[0], e.column(FeatureIndex(piece, square, 0)))
	addColumn(&e.acc[1], e.column(FeatureIndex(piece, square, 1)))
}

func (e *Evaluator) subFeature(piece, square int) {
	subColumn(&e.acc[0], e.column(FeatureIndex(piece, square, 0)))
	subColumn(&e.acc[1], e.column(FeatureIndex(piece, square, 1)))
}

func (e *Evaluator) output(us, them *[Hidden]int16) int {
	w := &e.net.OutputWeights
	sum := screluDot(us, them, (*[Hidden]int16)(w[:Hidden]), (*[Hidden]int16)(w[Hidden:]))/QA + int64(e.net.OutputBias)
	return int(sum * Scale / (QA * QB))
}

func (e *Evaluator) column(feature int) *[Hidden]int16 {
	return (*[Hidden]int16)(e.net.FeatureWeights[feature*Hidden : (feature+1)*Hidden])
}
