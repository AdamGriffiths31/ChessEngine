// Package main trains an internal/nnue network from datagen output on the
// CPU and writes it quantized.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/nnue"
)

const (
	hidden = nnue.Hidden

	maxBoardPieces = 32
	featureLimit   = math.MaxInt16 / (nnue.QA * (maxBoardPieces + 1.0))
	outputLimit    = math.MaxInt8 / float32(nnue.QB)

	verifySample = 2000

	sideToMove, otherSide = 0, 1

	adamBeta1 = 0.9
	adamBeta2 = 0.999
	adamEps   = 1e-8
)

type options struct {
	data       string
	out        string
	epochs     int
	batch      int
	lr         float64
	wdl        float64
	threads    int
	validFrac  float64
	checkpoint string
	resume     bool
}

type dataset struct {
	features    []uint16
	offsets     []uint32
	blackToMove []bool
	targets     []float32
	fens        []string
}

func (d *dataset) len() int { return len(d.targets) }

func (d *dataset) active(i int) []uint16 {
	return d.features[d.offsets[i]:d.offsets[i+1]]
}

type model struct {
	ftWeights  []float32
	ftBias     []float32
	outWeights []float32
	outBias    []float32
}

func newModel(rng *rand.Rand) *model {
	m := &model{
		ftWeights:  make([]float32, nnue.Inputs*hidden),
		ftBias:     make([]float32, hidden),
		outWeights: make([]float32, 2*hidden),
		outBias:    make([]float32, 1),
	}
	for i := range m.ftWeights {
		m.ftWeights[i] = float32(rng.Float64()*0.2 - 0.1)
	}
	for i := range m.ftBias {
		m.ftBias[i] = 0.1
	}
	limit := 1 / math.Sqrt(2*hidden)
	for i := range m.outWeights {
		m.outWeights[i] = float32((rng.Float64()*2 - 1) * limit)
	}
	return m
}

type paramGroup struct {
	values []float32
	limit  float32
}

func (m *model) groups() []paramGroup {
	return []paramGroup{
		{m.ftWeights, featureLimit},
		{m.ftBias, featureLimit},
		{m.outWeights, outputLimit},
		{m.outBias, featureLimit},
	}
}

func (m *model) params() [][]float32 {
	groups := m.groups()
	params := make([][]float32, len(groups))
	for i, g := range groups {
		params[i] = g.values
	}
	return params
}

func (m *model) column(feature int) []float32 {
	return m.ftWeights[feature*hidden : (feature+1)*hidden]
}

func (m *model) outputWeights(side int) []float32 {
	return m.outWeights[side*hidden : (side+1)*hidden]
}

func newLike(m *model) *model {
	return &model{
		ftWeights:  make([]float32, len(m.ftWeights)),
		ftBias:     make([]float32, len(m.ftBias)),
		outWeights: make([]float32, len(m.outWeights)),
		outBias:    make([]float32, 1),
	}
}

func main() {
	var opts options
	flag.StringVar(&opts.data, "data", "data.txt", "datagen output to train on")
	flag.StringVar(&opts.out, "out", "net.nnue", "quantized network to write")
	flag.IntVar(&opts.epochs, "epochs", 30, "passes over the training data")
	flag.IntVar(&opts.batch, "batch", 16384, "positions per optimizer step")
	flag.Float64Var(&opts.lr, "lr", 0.001, "starting learning rate")
	flag.Float64Var(&opts.wdl, "wdl", 0.3, "weight of the game result in the target; the rest is the search score")
	flag.IntVar(&opts.threads, "threads", runtime.NumCPU(), "worker goroutines")
	flag.Float64Var(&opts.validFrac, "valid", 0.01, "fraction of positions held out for validation")
	flag.StringVar(&opts.checkpoint, "checkpoint", "", "checkpoint written after every epoch; defaults to -out plus .ckpt")
	flag.BoolVar(&opts.resume, "resume", false, "continue from the checkpoint after its last finished epoch")
	flag.Parse()
	if opts.checkpoint == "" {
		opts.checkpoint = opts.out + ".ckpt"
	}

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "nnuetrain:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	start := time.Now()
	data, err := load(opts.data, opts.wdl)
	if err != nil {
		return err
	}
	fmt.Printf("loaded %d positions in %s\n", data.len(), time.Since(start).Round(time.Second))

	rng := rand.New(rand.NewPCG(1, 2))
	order := rng.Perm(data.len())
	validCount := int(float64(data.len()) * opts.validFrac)
	valid, trainOrder := order[:validCount], order[validCount:]
	train := make([]int, len(trainOrder))

	m := newModel(rng)
	adamM, adamV := newLike(m), newLike(m)
	grads := make([]*model, opts.threads)
	for i := range grads {
		grads[i] = newLike(m)
	}

	step, firstEpoch := 0, 1
	if opts.resume {
		ckpt, err := loadCheckpoint(opts.checkpoint)
		switch {
		case os.IsNotExist(err):
			fmt.Printf("no checkpoint at %s; starting fresh\n", opts.checkpoint)
		case err != nil:
			return err
		default:
			if err := ckpt.restore(data.len(), m, adamM, adamV); err != nil {
				return err
			}
			step, firstEpoch = ckpt.Step, ckpt.Epoch+1
			fmt.Printf("resuming after epoch %d from %s\n", ckpt.Epoch, opts.checkpoint)
		}
	}

	for epoch := firstEpoch; epoch <= opts.epochs; epoch++ {
		epochStart := time.Now()
		lr := learningRate(opts.lr, epoch, opts.epochs)
		copy(train, trainOrder)
		epochRNG := rand.New(rand.NewPCG(uint64(epoch), 3))
		epochRNG.Shuffle(len(train), func(i, j int) { train[i], train[j] = train[j], train[i] })

		var trainLoss float64
		for begin := 0; begin < len(train); begin += opts.batch {
			batch := train[begin:min(begin+opts.batch, len(train))]
			trainLoss += accumulateInto(grads, m, data, batch)
			step++
			adamStep(m, grads[0], adamM, adamV, lr, step, len(batch))
		}
		validLoss := evaluateLoss(m, data, valid, opts.threads)
		fmt.Printf("epoch %2d lr %.5f train %.6f valid %.6f (%s)\n",
			epoch, lr, trainLoss/float64(len(train)), validLoss, time.Since(epochStart).Round(time.Second))

		ckpt := checkpoint{Epoch: epoch, Step: step, Positions: data.len(), Hidden: hidden,
			Params: m.params(), AdamM: adamM.params(), AdamV: adamV.params()}
		if err := ckpt.save(opts.checkpoint); err != nil {
			return fmt.Errorf("saving checkpoint: %w", err)
		}
	}

	net := quantize(m)
	f, err := os.Create(opts.out)
	if err != nil {
		return err
	}
	if err := net.Write(f); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", opts.out)
	return verify(m, net, data)
}

func learningRate(base float64, epoch, epochs int) float64 {
	switch {
	case epoch > epochs*9/10:
		return base / 100
	case epoch > epochs*7/10:
		return base / 10
	}
	return base
}

func load(path string, wdl float64) (*dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := &dataset{offsets: []uint32{0}}
	reader := bufio.NewReaderSize(f, 1<<20)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			if line != "" {
				fmt.Println("skipping a cut-off last line")
			}
			return d, nil
		}
		if err != nil {
			return nil, err
		}
		d.add(strings.TrimRight(line, "\r\n"), wdl)
	}
}

func (d *dataset) add(line string, wdl float64) {
	parts := strings.Split(line, " | ")
	if len(parts) != 3 {
		return
	}
	score, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	result, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return
	}
	fields := strings.Fields(parts[0])
	if len(fields) < 2 {
		return
	}
	d.features = appendFeatures(d.features, fields[0])
	d.offsets = append(d.offsets, uint32(len(d.features)))

	blackToMove := fields[1] == "b"
	if blackToMove {
		score, result = -score, 1-result
	}
	target := (1-wdl)*sigmoid(float64(score)/nnue.Scale) + wdl*result
	d.blackToMove = append(d.blackToMove, blackToMove)
	d.targets = append(d.targets, float32(target))
	if len(d.fens) < verifySample {
		d.fens = append(d.fens, parts[0])
	}
}

func appendFeatures(features []uint16, placement string) []uint16 {
	rank, file := 7, 0
	for i := 0; i < len(placement); i++ {
		c := placement[i]
		switch {
		case c == '/':
			rank, file = rank-1, 0
		case c >= '1' && c <= '8':
			file += int(c - '0')
		default:
			if piece := board.PieceToBitboardIndex(board.Piece(c)); piece >= 0 {
				features = append(features, uint16(nnue.FeatureIndex(piece, rank*8+file, 0)))
			}
			file++
		}
	}
	return features
}

var blackView = func() (view [nnue.Inputs]uint16) {
	for f := range view {
		view[f] = uint16(nnue.FeatureIndex(f/64, f%64, 1))
	}
	return view
}()

type activations struct {
	acc [2][hidden]float32
}

func sideColumns(f uint16, blackToMove bool) (us, them int) {
	us, them = int(f), int(blackView[f])
	if blackToMove {
		return them, us
	}
	return us, them
}

func (m *model) forward(active []uint16, blackToMove bool, a *activations) float32 {
	a.acc[sideToMove] = [hidden]float32(m.ftBias)
	a.acc[otherSide] = [hidden]float32(m.ftBias)
	for _, f := range active {
		us, them := sideColumns(f, blackToMove)
		addInto(a.acc[sideToMove][:], m.column(us))
		addInto(a.acc[otherSide][:], m.column(them))
	}
	out := m.outBias[0]
	for side := range a.acc {
		weights := m.outputWeights(side)
		for h, x := range a.acc[side] {
			c := clamp01(x)
			out += c * c * weights[h]
		}
	}
	return out
}

func parallelSum(positions []int, workers int, fn func(worker int, share []int) float64) float64 {
	sums := make([]float64, workers)
	var wg sync.WaitGroup
	chunk := (len(positions) + workers - 1) / workers
	for w := range workers {
		share := positions[min(w*chunk, len(positions)):min((w+1)*chunk, len(positions))]
		wg.Go(func() { sums[w] = fn(w, share) })
	}
	wg.Wait()
	var total float64
	for _, s := range sums {
		total += s
	}
	return total
}

func accumulateInto(grads []*model, m *model, data *dataset, batch []int) float64 {
	loss := parallelSum(batch, len(grads), func(w int, share []int) float64 {
		g := grads[w]
		for _, p := range g.params() {
			clear(p)
		}
		var a activations
		var sum float64
		for _, i := range share {
			sum += backward(m, g, data.active(i), data.blackToMove[i], data.targets[i], &a)
		}
		return sum
	})

	total := grads[0].params()
	for _, g := range grads[1:] {
		for k, p := range g.params() {
			addInto(total[k], p)
		}
	}
	return loss
}

func backward(m, g *model, active []uint16, blackToMove bool, target float32, a *activations) float64 {
	out := m.forward(active, blackToMove, a)
	pred := float32(sigmoid(float64(out)))
	diff := pred - target
	grad := 2 * diff * pred * (1 - pred)

	var dAcc [2][hidden]float32
	g.outBias[0] += grad
	for side := range a.acc {
		weights, gradWeights := m.outputWeights(side), g.outputWeights(side)
		for h, x := range a.acc[side] {
			c := clamp01(x)
			gradWeights[h] += grad * c * c
			if x > 0 && x < 1 {
				dAcc[side][h] = grad * weights[h] * 2 * x
			}
		}
	}
	addInto(g.ftBias, dAcc[sideToMove][:])
	addInto(g.ftBias, dAcc[otherSide][:])
	for _, f := range active {
		us, them := sideColumns(f, blackToMove)
		addInto(g.column(us), dAcc[sideToMove][:])
		addInto(g.column(them), dAcc[otherSide][:])
	}
	return float64(diff * diff)
}

func adamStep(m, g, adamM, adamV *model, lr float64, step, batchSize int) {
	gs, ms, vs := g.params(), adamM.params(), adamV.params()

	scale := float32(1 / float64(batchSize))
	correction1 := 1 - math.Pow(adamBeta1, float64(step))
	correction2 := 1 - math.Pow(adamBeta2, float64(step))
	stepSize := float32(lr * math.Sqrt(correction2) / correction1)
	for k, group := range m.groups() {
		p, gk, mk, vk, limit := group.values, gs[k], ms[k], vs[k], group.limit
		for i := range p {
			gi := gk[i] * scale
			mk[i] = adamBeta1*mk[i] + (1-adamBeta1)*gi
			vk[i] = adamBeta2*vk[i] + (1-adamBeta2)*gi*gi
			p[i] -= stepSize * mk[i] / (float32(math.Sqrt(float64(vk[i]))) + adamEps)
			p[i] = max(-limit, min(limit, p[i]))
		}
	}
}

func evaluateLoss(m *model, data *dataset, positions []int, workers int) float64 {
	loss := parallelSum(positions, workers, func(_ int, share []int) float64 {
		var a activations
		var sum float64
		for _, i := range share {
			pred := float32(sigmoid(float64(m.forward(data.active(i), data.blackToMove[i], &a))))
			d := pred - data.targets[i]
			sum += float64(d * d)
		}
		return sum
	})
	return loss / float64(max(len(positions), 1))
}

func quantize(m *model) *nnue.Network {
	n := &nnue.Network{}
	for i, w := range m.ftWeights {
		n.FeatureWeights[i] = int16(math.Round(float64(w) * nnue.QA))
	}
	for i, b := range m.ftBias {
		n.FeatureBias[i] = int16(math.Round(float64(b) * nnue.QA))
	}
	for i, w := range m.outWeights {
		n.OutputWeights[i] = int16(math.Round(float64(w) * nnue.QB))
	}
	n.OutputBias = int32(math.Round(float64(m.outBias[0]) * nnue.QA * nnue.QB))
	return n
}

func verify(m *model, net *nnue.Network, data *dataset) error {
	evaluator := nnue.NewEvaluator(net)
	rounded := dequantize(net)
	var a activations
	var sumDiff, maxDiff, sumImpl float64
	for i, fen := range data.fens {
		b, err := board.FromFEN(fen)
		if err != nil {
			return fmt.Errorf("verify: %w", err)
		}
		blackToMove := data.blackToMove[i]
		floatCP := whiteCP(m.forward(data.active(i), blackToMove, &a), blackToMove)
		roundedCP := whiteCP(rounded.forward(data.active(i), blackToMove, &a), blackToMove)
		engineCP := float64(evaluator.Evaluate(b))
		diff := math.Abs(floatCP - engineCP)
		sumDiff += diff
		maxDiff = max(maxDiff, diff)
		sumImpl += math.Abs(roundedCP - engineCP)
	}
	count := float64(len(data.fens))
	fmt.Printf("verify: %d positions, mean |float - engine| %.2f cp (max %.2f), mean |dequantized float - engine| %.2f cp\n",
		len(data.fens), sumDiff/count, maxDiff, sumImpl/count)
	return nil
}

func whiteCP(out float32, blackToMove bool) float64 {
	cp := float64(out) * nnue.Scale
	if blackToMove {
		return -cp
	}
	return cp
}

func dequantize(n *nnue.Network) *model {
	m := &model{
		ftWeights:  make([]float32, len(n.FeatureWeights)),
		ftBias:     make([]float32, len(n.FeatureBias)),
		outWeights: make([]float32, len(n.OutputWeights)),
		outBias:    []float32{float32(n.OutputBias) / (nnue.QA * nnue.QB)},
	}
	for i, w := range n.FeatureWeights {
		m.ftWeights[i] = float32(w) / nnue.QA
	}
	for i, b := range n.FeatureBias {
		m.ftBias[i] = float32(b) / nnue.QA
	}
	for i, w := range n.OutputWeights {
		m.outWeights[i] = float32(w) / nnue.QB
	}
	return m
}

func addInto(dst, src []float32) {
	for i := range dst {
		dst[i] += src[i]
	}
}

func clamp01(x float32) float32 {
	return max(0, min(1, x))
}

func sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}
