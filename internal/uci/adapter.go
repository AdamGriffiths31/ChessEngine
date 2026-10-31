package uci

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
	"github.com/AdamGriffiths31/ChessEngine/internal/game"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

const (
	// DefaultMaxDepth is the search depth cap used when a "go" command does
	// not specify a depth; effectively unlimited, so time is the real bound.
	DefaultMaxDepth = 999
	// DefaultMoveTime is the per-move time budget used when a "go" command
	// specifies neither a move time nor clock times.
	DefaultMoveTime = 5 * time.Second

	emptyOptionValue = "<empty>"
)

// Engine wraps the chess engine with UCI protocol support.
type Engine struct {
	engine           *game.Engine
	aiEngine         search.Engine
	converter        *MoveConverter
	protocol         *ProtocolHandler
	options          map[string]string
	cancelSearch     context.CancelFunc
	searchDone       chan struct{}
	outputMu         sync.Mutex
	output           io.Writer
	moveNumber       int
	session          string
	gameNumber       int
	prevMoveFromBook bool
	uciLogEnabled    bool
	bookFile         string
	prober           *book.Prober
	generator        *movegen.Generator
}

// NewUCIEngine creates a new UCI engine wrapper.
func NewUCIEngine() *Engine {
	minimaxEngine := search.NewMinimaxEngine()
	minimaxEngine.SetTranspositionTableSize(256)

	engine := &Engine{
		engine:        game.NewEngine(),
		aiEngine:      minimaxEngine,
		converter:     NewMoveConverter(),
		protocol:      NewProtocolHandler(),
		options:       make(map[string]string),
		session:       fmt.Sprintf("%d-%d", time.Now().Unix(), os.Getpid()),
		uciLogEnabled: slog.Default().Enabled(context.Background(), slog.LevelDebug),
		generator:     movegen.NewGenerator(),
	}

	return engine
}

// SetBookFile configures the opening book path used to build the book prober.
// The prober is constructed once, on the first "go" command that finds a
// non-empty book file after a change of path (see probeBook). Passing an empty
// path disables the opening book. The choice is logged at Info level.
func (ue *Engine) SetBookFile(path string) {
	ue.bookFile = path
	ue.prober = nil
	if path != "" {
		slog.Info("book file set", "path", path)
	} else {
		slog.Info("no book file configured")
	}
}

// Run starts the UCI engine main loop.
func (ue *Engine) Run(input io.Reader, output io.Writer) error {
	if input == nil || output == nil {
		return fmt.Errorf("input and output cannot be nil")
	}

	ue.output = output

	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()

		if ue.uciLogEnabled {
			slog.Debug("uci", "dir", "in", "line", line)
		}

		cmd := ue.protocol.ParseCommand(line)
		switch cmd.Name {
		case "go":
			ue.startSearch(cmd.Args)
			continue
		case "quit":
			ue.handleStop()
			ue.waitForSearch()
			return scanner.Err()
		case "position", "setoption", "ucinewgame", "bench":
			ue.waitForSearch()
		}

		if response := ue.HandleCommand(line); response != "" {
			ue.send(response)
		}
	}

	ue.waitForSearch()
	return scanner.Err()
}

func (ue *Engine) startSearch(args []string) {
	ue.waitForSearch()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	ue.cancelSearch = cancel
	ue.searchDone = done

	go func() {
		defer close(done)
		defer cancel()
		ue.handleGo(ctx, args)
	}()
}

func (ue *Engine) waitForSearch() {
	if ue.searchDone != nil {
		<-ue.searchDone
		ue.searchDone = nil
		ue.cancelSearch = nil
	}
}

func (ue *Engine) send(line string) {
	if ue.uciLogEnabled {
		slog.Debug("uci", "dir", "out", "line", line)
	}
	ue.outputMu.Lock()
	defer ue.outputMu.Unlock()
	if _, err := fmt.Fprintln(ue.output, line); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

// HandleCommand processes a single UCI command and returns the response.
func (ue *Engine) HandleCommand(input string) string {
	cmd := ue.protocol.ParseCommand(input)

	if cmd.Name == "" {
		return ""
	}

	switch cmd.Name {
	case "uci":
		return ue.handleUCI()
	case "isready":
		return ue.handleIsReady()
	case "position":
		return ue.handlePosition(cmd.Args)
	case "go":
		ue.handleGo(context.Background(), cmd.Args)
		return ""
	case "stop":
		return ue.handleStop()
	case "setoption":
		return ue.handleSetOption(cmd.Args)
	case "ucinewgame":
		return ue.handleNewGame()
	case "bench":
		return ue.handleBench(cmd.Args)
	case "quit":
		return ""
	default:
		return ""
	}
}

func (ue *Engine) handleBench(args []string) string {
	report, err := BenchReport(args)
	if err != nil {
		slog.Error("bench failed", "error", err)
		return ""
	}
	return report
}

// BenchReport runs search.RunBench at the depth given in args[0] (default
// search.DefaultBenchDepth) and formats the result. The final line,
// "Bench: <nodes> nodes <nps> nps", is the shape OpenBench-style tooling parses.
func BenchReport(args []string) (string, error) {
	depth := search.DefaultBenchDepth
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &depth); err != nil || depth < 1 {
			return "", fmt.Errorf("invalid bench depth %q", args[0])
		}
	}
	result, err := search.RunBench(depth)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Depth: %d\nPositions: %d\nTime (ms): %d\nBench: %d nodes %d nps",
		depth, len(search.BenchFENs), result.Elapsed.Milliseconds(), result.Nodes, result.NodesPerSecond()), nil
}

func (ue *Engine) handleUCI() string {
	response := []string{
		"id name ChessEngine",
		"id author Adam Griffiths",
		"option name Hash type spin default 128 min 1 max 1024",
		"option name Move Overhead type spin default 50 min 0 max 5000",
		"option name BookFile type string default " + emptyOptionValue,
		"uciok",
	}

	return strings.Join(response, "\n")
}

func (ue *Engine) handleIsReady() string {
	return ue.protocol.FormatReadyOK()
}

func (ue *Engine) handlePosition(args []string) string {
	fen, moveList, err := ue.protocol.ParsePosition(args)
	if err != nil {
		slog.Error("position: parse failed", "fen", fen, "error", err)
		return ""
	}

	if err := ue.engine.LoadFromFEN(fen); err != nil {
		slog.Error("position: FEN load failed", "fen", fen, "error", err)
		return ""
	}

	for _, moveStr := range moveList {
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			slog.Error("position: move conversion failed", "move", moveStr, "error", err)
			return ""
		}

		err = ue.engine.MakeMove(move)
		if err != nil {
			slog.Error("position: make move failed", "move", moveStr, "error", err)
			return ""
		}
	}

	return ""
}

func (ue *Engine) handleGo(ctx context.Context, args []string) {
	params := ue.protocol.ParseGo(args)
	player := ue.engine.GetCurrentPlayer()

	config := ue.buildSearchConfig(params)
	budget := ue.searchTimeout(params, player)
	config.MaxTime = budget.Hard
	config.SoftTime = budget.Soft

	ctx, cancel := context.WithTimeout(ctx, config.MaxTime)
	defer cancel()

	ue.moveNumber++

	searchFEN := ue.engine.GetCurrentFEN()
	searchStart := time.Now()
	result, bookOutcome := ue.runSearch(ctx, config, player)
	searchDuration := time.Since(searchStart)

	bestMoveUCI := ue.converter.ToUCI(result.BestMove)

	pvString := ""
	if len(result.Stats.PrincipalVariation) > 0 {
		pvMoves := make([]string, len(result.Stats.PrincipalVariation))
		for i, move := range result.Stats.PrincipalVariation {
			pvMoves[i] = ue.converter.ToUCI(move)
		}
		pvString = strings.Join(pvMoves, " ")
	}

	infoMessage := ue.protocol.FormatInfo(result.Stats.Depth, int(result.Score), result.Stats.NodesSearched, searchDuration, pvString)
	ue.send(infoMessage)

	formattedBestMove := ue.protocol.FormatBestMove(bestMoveUCI)

	ue.logMove(moveLog{
		fen:     searchFEN,
		player:  player,
		params:  params,
		budget:  config.MaxTime,
		soft:    config.SoftTime,
		elapsed: searchDuration,
		move:    bestMoveUCI,
		result:  result,
		book:    bookOutcome,
	})

	if ue.moveNumber%10 == 0 {
		if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
			ttStats := minimaxEngine.TTStats()
			if ttStats.Hits > 0 || ttStats.Misses > 0 {
				ue.send(fmt.Sprintf("info string TT: %d hits, %d misses, %d collisions, %.1f%% hit rate | Second bucket: %d uses, %.1f%% rate",
					ttStats.Hits, ttStats.Misses, ttStats.Collisions, ttStats.HitRate, ttStats.SecondBucketUse, ttStats.SecondBucketRate))
			}
		}
	}

	ue.send(formattedBestMove)
}

func (ue *Engine) buildSearchConfig(params SearchParams) search.Config {
	var bookFiles []string
	if ue.bookFile != "" {
		bookFiles = []string{ue.bookFile}
	}

	config := search.Config{
		MaxDepth:          DefaultMaxDepth,
		MaxTime:           DefaultMoveTime,
		UseOpeningBook:    len(bookFiles) > 0,
		BookFiles:         bookFiles,
		RepetitionHistory: ue.engine.HashHistory(),
	}

	if params.Depth > 0 {
		config.MaxDepth = params.Depth
	}

	return config
}

func (ue *Engine) searchTimeout(params SearchParams, player movegen.Player) timeBudget {
	overhead := ue.moveOverhead()
	budget := timeBudget{Hard: DefaultMoveTime}

	switch {
	case params.Infinite:
		budget.Hard = 24 * time.Hour
	case params.MoveTime > 0:
		budget.Hard = max(params.MoveTime-overhead, time.Millisecond)
	case params.WTime > 0 || params.BTime > 0:
		timeLeft, increment := params.WTime, params.WInc
		if player == movegen.Black {
			timeLeft, increment = params.BTime, params.BInc
		}
		budget = allocateTime(timeLeft, increment, params.MovesToGo, overhead)
	}

	return budget
}

func (ue *Engine) moveOverhead() time.Duration {
	var ms int
	if v, ok := ue.options["Move Overhead"]; ok {
		if _, err := fmt.Sscanf(v, "%d", &ms); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return DefaultMoveOverhead
}

func (ue *Engine) runSearch(ctx context.Context, config search.Config, player movegen.Player) (search.Result, bookOutcome) {
	bookMove, outcome := ue.probeBook(config)
	ue.prevMoveFromBook = outcome.Used
	if outcome.Used {
		return search.Result{
			BestMove: bookMove,
			Score:    0,
			Stats:    search.Stats{BookMoveUsed: true, StopReason: search.StopReasonBook},
		}, outcome
	}

	return ue.searchWithFallback(ctx, config, player), outcome
}

func (ue *Engine) searchWithFallback(ctx context.Context, config search.Config, player movegen.Player) (result search.Result) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("search panicked; playing fallback move",
				"move_number", ue.moveNumber, "panic", fmt.Sprint(r),
				"fen", ue.engine.GetCurrentFEN(), "player", player.String())
			result = ue.fallbackResult(player)
		}
	}()

	result = ue.aiEngine.FindBestMove(ctx, ue.engine.GetState().Board, player, config)
	return result
}

func (ue *Engine) fallbackResult(player movegen.Player) search.Result {
	moves := ue.generator.GenerateAllMoves(ue.engine.GetState().Board, player)
	defer movegen.ReleaseMoveList(moves)

	if moves.Count > 0 {
		return search.Result{
			BestMove: moves.Moves[0],
			Score:    0,
			Stats:    search.Stats{},
		}
	}

	return search.Result{
		BestMove: board.Move{
			From: board.Square{File: -1, Rank: -1},
			To:   board.Square{File: -1, Rank: -1},
		},
		Score: 0,
		Stats: search.Stats{},
	}
}

func (ue *Engine) probeBook(config search.Config) (board.Move, bookOutcome) {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		return board.Move{}, bookOutcome{Reason: BookNotConfigured}
	}
	if ue.prober == nil {
		ue.prober = book.NewProber(config.BookFiles)
	}
	move, ok := ue.prober.Probe(ue.engine.GetState().Board)
	if !ok {
		if ue.prevMoveFromBook {
			return board.Move{}, bookOutcome{Reason: BookLeft}
		}
		return board.Move{}, bookOutcome{Reason: BookNoEntry}
	}
	return *move, bookOutcome{Used: true, Reason: BookHit}
}

func (ue *Engine) handleStop() string {
	if ue.cancelSearch != nil {
		ue.cancelSearch()
	}
	return ""
}

func (ue *Engine) handleSetOption(args []string) string {
	name, value, err := ue.protocol.ParseSetOption(args)
	if err != nil {
		slog.Warn("invalid setoption", "args", args, "error", err)
		return ""
	}

	ue.options[name] = value

	switch name {
	case "Hash":
		var hashSizeMB int
		if _, err := fmt.Sscanf(value, "%d", &hashSizeMB); err == nil && hashSizeMB > 0 {
			if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
				minimaxEngine.SetTranspositionTableSize(hashSizeMB)
			}
		}
	case "BookFile":
		if value == emptyOptionValue {
			value = ""
		}
		ue.SetBookFile(value)
	}

	return ""
}

func (ue *Engine) handleNewGame() string {
	ue.engine.Reset()
	ue.moveNumber = 0
	ue.gameNumber++
	ue.prevMoveFromBook = false

	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		minimaxEngine.ClearSearchState()
	}

	hashSize := 128
	if hashStr, exists := ue.options["Hash"]; exists {
		if _, err := fmt.Sscanf(hashStr, "%d", &hashSize); err != nil || hashSize < 1 {
			hashSize = 128
		}
	}

	slog.Info("new game", "session", ue.session, "game", ue.gameNumber, "hash_mb", hashSize)

	return ""
}
