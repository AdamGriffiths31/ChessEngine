// Package main generates NNUE training data from fixed-node self-play.
//
// Each line written is "FEN | score | result": the score in centipawns from
// White's side, and the game result from White's side (1.0, 0.5 or 0.0).
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/game"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/nnue"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

const (
	maxGamePlies     = 400
	maxSearchDepth   = 64
	hardNodeFactor   = 8
	adjudicateScore  = 2500
	adjudicatePlies  = 6
	maxRecordedScore = 3000
)

type options struct {
	threads     int
	nodes       int64
	duration    time.Duration
	positions   int64
	randomPlies int
	hashMB      int
	out         string
	eval        string
}

type counters struct {
	games     atomic.Int64
	positions atomic.Int64
	searched  atomic.Int64
	nodes     atomic.Int64
}

type record struct {
	fen   string
	score int
}

func main() {
	var opts options
	flag.IntVar(&opts.threads, "threads", runtime.NumCPU(), "games played in parallel")
	flag.Int64Var(&opts.nodes, "nodes", 5000, "soft node limit per move")
	flag.DurationVar(&opts.duration, "duration", 0, "stop after this long; zero runs until interrupted")
	flag.Int64Var(&opts.positions, "positions", 0, "stop once the output file holds this many positions, counting earlier runs; zero means no cap")
	flag.IntVar(&opts.randomPlies, "random-plies", 8, "random moves played at the start of each game")
	flag.IntVar(&opts.hashMB, "hash", 16, "transposition table size per thread in MB")
	flag.StringVar(&opts.out, "out", "datagen.txt", "output file; a rerun resumes by appending to it")
	flag.StringVar(&opts.eval, "eval", "", `network file to search with; empty uses the built-in network, "hce" the hand-written eval`)
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "datagen:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if _, err := nnue.EvaluatorFor(opts.eval); err != nil {
		return err
	}
	existing, err := trimAndCountLines(opts.out)
	if err != nil {
		return err
	}
	if opts.positions > 0 && existing >= opts.positions {
		fmt.Printf("%s already holds %d positions, at or past -positions %d\n", opts.out, existing, opts.positions)
		return nil
	}
	if existing > 0 {
		fmt.Printf("resuming: %s already holds %d positions\n", opts.out, existing)
	}

	file, err := os.OpenFile(opts.out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriterSize(file, 1<<20)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx, reachedCap := context.WithCancel(ctx)
	defer reachedCap()
	defer stop()
	if opts.duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.duration)
		defer cancel()
	}

	var (
		stats    counters
		writeMu  sync.Mutex
		writeErr error
		wg       sync.WaitGroup
	)
	start := time.Now()
	for worker := range opts.threads {
		wg.Go(func() {
			rng := rand.New(rand.NewPCG(uint64(start.UnixNano()), uint64(worker)))
			engine := search.NewMinimaxEngine()
			engine.SetTranspositionTableSize(opts.hashMB)
			evaluator, _ := nnue.EvaluatorFor(opts.eval)
			engine.SetEvaluator(evaluator)
			for ctx.Err() == nil {
				lines, searched, nodes := playGame(ctx, engine, rng, opts)
				stats.games.Add(1)
				stats.searched.Add(int64(searched))
				stats.nodes.Add(nodes)
				if len(lines) == 0 {
					continue
				}

				writeMu.Lock()
				for _, line := range lines {
					writer.WriteString(line)
				}
				err := writer.Flush()
				if err != nil && writeErr == nil {
					writeErr = err
				}
				writeMu.Unlock()
				if err != nil {
					reachedCap()
					return
				}

				if total := existing + stats.positions.Add(int64(len(lines))); opts.positions > 0 && total >= opts.positions {
					reachedCap()
				}
			}
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			report(&stats, start, existing)
		case <-done:
			report(&stats, start, existing)
			return writeErr
		}
	}
}

func trimAndCountLines(path string) (int64, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var (
		lines       int64
		offset      int64
		lastNewline int64 = -1
		buf               = make([]byte, 1<<20)
	)
	for {
		n, err := file.Read(buf)
		for i, c := range buf[:n] {
			if c == '\n' {
				lines++
				lastNewline = offset + int64(i)
			}
		}
		offset += int64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	if keep := lastNewline + 1; keep < offset {
		fmt.Printf("trimming a cut-off last line (%d bytes) from %s\n", offset-keep, path)
		if err := file.Truncate(keep); err != nil {
			return 0, err
		}
	}
	return lines, nil
}

func report(stats *counters, start time.Time, existing int64) {
	elapsed := time.Since(start).Seconds()
	positions := stats.positions.Load()
	fmt.Printf("%6.0fs games=%d searched=%d kept=%d total=%d kept/s=%.0f searched/s=%.0f nodes/search=%.0f nps=%.0f\n",
		elapsed, stats.games.Load(), stats.searched.Load(), positions, existing+positions,
		float64(positions)/elapsed, float64(stats.searched.Load())/elapsed,
		float64(stats.nodes.Load())/float64(max(stats.searched.Load(), 1)), float64(stats.nodes.Load())/elapsed)
}

func playGame(ctx context.Context, engine *search.MinimaxEngine, rng *rand.Rand, opts options) (lines []string, searched int, nodes int64) {
	engine.ClearSearchState()
	generator := movegen.NewGenerator()
	g := game.NewEngine()
	if !playRandomOpening(g, generator, rng, opts.randomPlies) {
		return nil, 0, 0
	}

	config := search.Config{MaxDepth: maxSearchDepth, SoftNodes: opts.nodes, MaxNodes: opts.nodes * hardNodeFactor}
	var (
		records   []record
		result    string
		decisive  int
		lastScore int
	)
	for ply := 0; ply < maxGamePlies && !g.GetState().GameOver; ply++ {
		b, player := g.GetState().Board, g.GetCurrentPlayer()
		config.RepetitionHistory = g.HashHistory()
		res := engine.FindBestMove(ctx, b, player, config)
		if ctx.Err() != nil {
			return nil, searched, nodes
		}
		searched++
		nodes += res.Stats.NodesSearched

		whiteScore := int(res.Score)
		if player == movegen.Black {
			whiteScore = -whiteScore
		}
		if keepPosition(b, generator, player, res) {
			records = append(records, record{fen: b.ToFEN(), score: whiteScore})
		}

		if abs(whiteScore) >= adjudicateScore && sameSign(whiteScore, lastScore) {
			decisive++
		} else {
			decisive = 0
		}
		lastScore = whiteScore
		if decisive >= adjudicatePlies {
			result = winFor(whiteScore > 0)
			break
		}

		if err := g.MakeMove(res.BestMove); err != nil {
			panic(fmt.Sprintf("datagen: search returned an unplayable move at %s: %v", b.ToFEN(), err))
		}
	}
	if result == "" {
		result = gameResult(g.GetState())
	}

	lines = make([]string, len(records))
	for i, r := range records {
		lines[i] = fmt.Sprintf("%s | %d | %s\n", r.fen, r.score, result)
	}
	return lines, searched, nodes
}

func playRandomOpening(g *game.Engine, generator *movegen.Generator, rng *rand.Rand, plies int) bool {
	for range plies {
		moves := generator.GenerateAllMoves(g.GetState().Board, g.GetCurrentPlayer())
		if moves.Count == 0 {
			movegen.ReleaseMoveList(moves)
			return false
		}
		move := moves.Moves[rng.IntN(moves.Count)]
		movegen.ReleaseMoveList(moves)
		if err := g.MakeMove(move); err != nil || g.GetState().GameOver {
			return false
		}
	}
	return true
}

func keepPosition(b *board.Board, generator *movegen.Generator, player movegen.Player, res search.Result) bool {
	return !generator.IsKingInCheck(b, player) &&
		!res.BestMove.IsCapture && !res.BestMove.HasPromotion() &&
		abs(int(res.Score)) < maxRecordedScore
}

func gameResult(state *game.State) string {
	if state.GameOver && !state.IsDraw {
		return winFor(state.Winner == movegen.White)
	}
	return "0.5"
}

func winFor(white bool) string {
	if white {
		return "1.0"
	}
	return "0.0"
}

func sameSign(a, b int) bool {
	return (a > 0 && b > 0) || (a < 0 && b < 0)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
