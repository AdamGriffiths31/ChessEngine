package uci

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
)

// Engine wraps the chess engine with UCI protocol support.
type Engine struct {
	engine      *game.Engine
	aiEngine    search.Engine
	converter   *MoveConverter
	protocol    *ProtocolHandler
	options     map[string]string
	searching   bool
	stopChannel chan struct{}
	output      io.Writer
	debugLogger *log.Logger
	moveNumber  int
	// uciLogEnabled caches slog's Debug-level enablement at construction time
	// so the per-line UCI communication tee (see send/Run) is a near-zero-cost
	// no-op when disabled. slog's level is fixed at startup (see cmd/uci/main.go
	// setupLogging), so this cache never needs to be refreshed.
	uciLogEnabled bool
	bookFile      string
	prober        *book.Prober
	generator     *movegen.Generator
}

// NewUCIEngine creates a new UCI engine wrapper.
func NewUCIEngine() *Engine {
	debugLogger := createDebugLogger()

	minimaxEngine := search.NewMinimaxEngine()
	minimaxEngine.SetTranspositionTableSize(256)

	engine := &Engine{
		engine:        game.NewEngine(),
		aiEngine:      minimaxEngine,
		converter:     NewMoveConverter(),
		protocol:      NewProtocolHandler(),
		options:       make(map[string]string),
		searching:     false,
		stopChannel:   make(chan struct{}),
		debugLogger:   debugLogger,
		uciLogEnabled: slog.Default().Enabled(context.Background(), slog.LevelDebug),
		generator:     movegen.NewGenerator(),
	}

	return engine
}

// SetBookFile configures the opening book path used to build the book prober.
// The prober is constructed once, on the first "go" command that finds a
// non-empty book file (see probeBook); therefore a BookFile set BEFORE the
// first search takes effect, while later changes are ignored. Passing an empty
// path disables the opening book. Resolution is logged via the debug logger.
func (ue *Engine) SetBookFile(path string) {
	ue.bookFile = path
	if path != "" {
		ue.debugLogger.Printf("BOOK: BookFile set to: %s", path)
	} else {
		ue.debugLogger.Printf("BOOK: No book file configured")
	}
}

func createDebugLogger() *log.Logger {
	logLocations := []string{
		filepath.Join(os.TempDir(), "chess"),
	}

	for _, dir := range logLocations {
		if err := os.MkdirAll(dir, 0750); err != nil {
			log.Printf("Failed to create log directory %s: %v", dir, err)
			continue
		}

		logFile := filepath.Join(dir, fmt.Sprintf("uci_debug_%d.log", time.Now().Unix()))
		file, err := os.Create(logFile) // #nosec G304 - log file path is controlled by application
		if err != nil {
			log.Printf("Failed to create log file %s: %v", logFile, err)
			continue
		}

		multiWriter := io.MultiWriter(file, os.Stderr)
		log.Printf("UCI debug logging to: %s", logFile)
		return log.New(multiWriter, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}

	log.Printf("Failed to create debug log file, using stderr only")
	return log.New(os.Stderr, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
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

		response := ue.HandleCommand(line)

		if response != "" {
			ue.send(response)
		}

		cmd := ue.protocol.ParseCommand(line)
		if cmd.Name == "quit" {
			break
		}
	}

	return scanner.Err()
}

// send writes a single line to the UCI output stream, terminated with a
// newline (matching the previous per-call-site fmt.Fprint* behavior exactly),
// and tees it to slog at Debug level (direction=out) when enabled. This is
// the single write choke point for the protocol writer, so every outgoing
// line is captured in one place instead of at each call site.
func (ue *Engine) send(line string) {
	if ue.uciLogEnabled {
		slog.Debug("uci", "dir", "out", "line", line)
	}
	if _, err := fmt.Fprintln(ue.output, line); err != nil {
		ue.debugLogger.Printf("UCI-ERROR: Failed to write response: %v", err)
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
		ue.handleGo(cmd.Args)
		return ""
	case "stop":
		return ue.handleStop()
	case "setoption":
		return ue.handleSetOption(cmd.Args)
	case "ucinewgame":
		return ue.handleNewGame()
	case "quit":
		return ""
	default:
		return ""
	}
}

func (ue *Engine) handleUCI() string {
	response := []string{
		"id name ChessEngine",
		"id author Adam Griffiths",
		"option name Hash type spin default 128 min 1 max 1024",
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
		ue.debugLogger.Printf("POSITION-ERROR: Failed to parse position - fen: %q, error: %v", fen, err)
		return ""
	}

	if err := ue.engine.LoadFromFEN(fen); err != nil {
		ue.debugLogger.Printf("POSITION-ERROR: Failed to load FEN %q: %v", fen, err)
		return ""
	}

	for _, moveStr := range moveList {
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			ue.debugLogger.Printf("POSITION-ERROR: Failed to convert UCI move - move: %s, error: %v", moveStr, err)
			return ""
		}

		err = ue.engine.MakeMove(move)
		if err != nil {
			ue.debugLogger.Printf("POSITION-ERROR: Failed to make move - move: %+v, error: %v", move, err)
			return ""
		}
	}

	return ""
}

func (ue *Engine) handleGo(args []string) {
	if ue.searching {
		return
	}

	params := ue.protocol.ParseGo(args)
	player := movegen.Player(ue.engine.GetCurrentPlayer())

	ue.searching = true
	ue.stopChannel = make(chan struct{})

	config := ue.buildSearchConfig(params)
	config.MaxTime = ue.searchTimeout(params, player)

	ctx, cancel := context.WithTimeout(context.Background(), config.MaxTime)
	defer cancel()

	stopCtx, stopCancel := context.WithCancel(ctx)
	defer stopCancel()

	go func() {
		select {
		case <-ue.stopChannel:
			stopCancel()
		case <-ctx.Done():
		}
	}()

	ue.moveNumber++

	searchFEN := ue.engine.GetCurrentFEN()
	searchStart := time.Now()
	result := ue.runSearch(stopCtx, config, player)
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

	var ttStatsStr string
	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		ttStats := minimaxEngine.TTStats()
		ttStatsStr = fmt.Sprintf("TT: %d hits, %d misses, %d collisions, %.1f%% hit rate | 2B: %d uses, %.1f%% rate",
			ttStats.Hits, ttStats.Misses, ttStats.Collisions, ttStats.HitRate, ttStats.SecondBucketUse, ttStats.SecondBucketRate)
	}

	var moveOrderPct float64
	if result.Stats.TotalCutoffs > 0 {
		moveOrderPct = float64(result.Stats.FirstMoveCutoffs) / float64(result.Stats.TotalCutoffs) * 100
	}

	var ttHitRate float64
	if result.Stats.TTProbes > 0 {
		ttHitRate = float64(result.Stats.TTHits) / float64(result.Stats.TTProbes) * 100
	}

	var ebf float64
	ebfCount := 0
	for d := 0; d < len(result.Stats.NodesByDepth)-1 && d < result.Stats.Depth; d++ {
		if result.Stats.NodesByDepth[d] > 0 && result.Stats.NodesByDepth[d+1] > 0 {
			ebf += float64(result.Stats.NodesByDepth[d+1]) / float64(result.Stats.NodesByDepth[d])
			ebfCount++
		}
	}
	if ebfCount > 0 {
		ebf /= float64(ebfCount)
	}

	cutoff1st := result.Stats.CutoffsByMoveIndex[0]
	cutoff2nd := result.Stats.CutoffsByMoveIndex[1]
	cutoff3rd := result.Stats.CutoffsByMoveIndex[2]

	totalNodes := result.Stats.PVNodes + result.Stats.CutNodes + result.Stats.AllNodes
	var pvNodePct, cutNodePct, allNodePct float64
	if totalNodes > 0 {
		pvNodePct = float64(result.Stats.PVNodes) / float64(totalNodes) * 100
		cutNodePct = float64(result.Stats.CutNodes) / float64(totalNodes) * 100
		allNodePct = float64(result.Stats.AllNodes) / float64(totalNodes) * 100
	}

	pvLog := "none"
	if pvString != "" {
		pvLog = pvString
	}
	ue.debugLogger.Printf("Move %d: %s | Score: %d | Depth: %d | Nodes: %d | Q: %d | NM: %d/%d | LMR: %d | TTC: %d | DP: %d | RZ: %d/%d | MO: %.0f%% | TTHit: %.0f%% | EBF: %.2f | Cutoffs: [%d,%d,%d] | NodeTypes: PV=%.0f%% Cut=%.0f%% All=%.0f%% | Time: %.3fs | Book: %t | PV: %s | %s | FEN: %s",
		ue.moveNumber, bestMoveUCI, result.Score, result.Stats.Depth, result.Stats.NodesSearched, result.Stats.QNodes, result.Stats.NullCutoffs, result.Stats.NullMoves, result.Stats.LMRReductions, result.Stats.TTCutoffs, result.Stats.DeltaPruned, result.Stats.RazoringCutoffs, result.Stats.RazoringAttempts, moveOrderPct, ttHitRate, ebf, cutoff1st, cutoff2nd, cutoff3rd, pvNodePct, cutNodePct, allNodePct, searchDuration.Seconds(), result.Stats.BookMoveUsed, pvLog, ttStatsStr, searchFEN)

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
	ue.searching = false
}

// buildSearchConfig assembles the search configuration for a "go" command from
// the parsed parameters and the configured opening book. MaxTime is set to the
// default budget here; handleGo overrides it with searchTimeout. The opening
// book is driven entirely by the BookFile UCI option (ue.bookFile); the old
// executable-relative filesystem walk has been removed.
func (ue *Engine) buildSearchConfig(params SearchParams) search.SearchConfig {
	var bookFiles []string
	if ue.bookFile != "" {
		bookFiles = []string{ue.bookFile}
	}

	config := search.SearchConfig{
		MaxDepth:       DefaultMaxDepth,
		MaxTime:        DefaultMoveTime,
		DebugMode:      false,
		UseOpeningBook: len(bookFiles) > 0,
		BookFiles:      bookFiles,
	}

	if params.Depth > 0 {
		config.MaxDepth = params.Depth
	}

	return config
}

// searchTimeout determines the time budget for a "go" command. Precedence
// mirrors the original handleGo logic: default move time, then an explicit
// movetime (minus a safety margin), then infinite, then clock-based management
// when wtime/btime are supplied.
func (ue *Engine) searchTimeout(params SearchParams, player movegen.Player) time.Duration {
	maxTime := DefaultMoveTime

	if params.MoveTime > 0 {
		safetyMargin := 100 * time.Millisecond
		maxTime = params.MoveTime - safetyMargin
	}
	if params.Infinite {
		maxTime = 24 * time.Hour
	}
	if params.WTime > 0 || params.BTime > 0 {
		maxTime = ue.calculateMoveTime(params, player, ue.moveNumber)
	}

	return maxTime
}

// runSearch produces a search result for the current position. It first
// consults the opening book; on a miss it runs the engine search guarded by a
// panic recovery that plays a fallback move instead of forfeiting the game.
func (ue *Engine) runSearch(ctx context.Context, config search.SearchConfig, player movegen.Player) search.SearchResult {
	if bookMove, ok := ue.probeBook(config); ok {
		return search.SearchResult{
			BestMove: bookMove,
			Score:    0,
			Stats:    search.SearchStats{BookMoveUsed: true},
		}
	}

	return ue.searchWithFallback(ctx, config, player)
}

// searchWithFallback runs the engine search and recovers from any panic. A UCI
// engine that dies mid-search would forfeit, so on panic we log the context and
// play the first legal move rather than re-panicking. If no legal move exists,
// we emit the UCI null move (bestmove 0000) via an out-of-range move that the
// converter renders as "0000".
func (ue *Engine) searchWithFallback(ctx context.Context, config search.SearchConfig, player movegen.Player) (result search.SearchResult) {
	defer func() {
		if r := recover(); r != nil {
			ue.debugLogger.Printf("PANIC CAUGHT: Search panicked for move %d: %v", ue.moveNumber, r)
			ue.debugLogger.Printf("PANIC CONTEXT: Position=%s, Player=%s",
				ue.engine.GetCurrentFEN(), player)
			result = ue.fallbackResult(player)
		}
	}()

	result = ue.aiEngine.FindBestMove(ctx, ue.engine.GetState().Board, player, config)
	return result
}

// fallbackResult selects the first legal move in the current position so the
// engine can respond after a search panic. If there are no legal moves it
// returns an out-of-range move that the converter renders as the UCI null move
// "0000".
func (ue *Engine) fallbackResult(player movegen.Player) search.SearchResult {
	moves := ue.generator.GenerateAllMoves(ue.engine.GetState().Board, player)
	defer movegen.ReleaseMoveList(moves)

	if moves.Count > 0 {
		return search.SearchResult{
			BestMove: moves.Moves[0],
			Score:    0,
			Stats:    search.SearchStats{},
		}
	}

	return search.SearchResult{
		BestMove: board.Move{
			From: board.Square{File: -1, Rank: -1},
			To:   board.Square{File: -1, Rank: -1},
		},
		Score: 0,
		Stats: search.SearchStats{},
	}
}

// probeBook consults the opening book for the current position before a
// search is started. The book.Prober is created lazily on first use with a
// non-empty file list and reused afterward, so the book is loaded from disk
// at most once per Engine lifetime (matching the previous engine-side
// lazy-init behavior); a load failure is logged once by the Prober itself
// and every subsequent call cheaply reports a miss.
func (ue *Engine) probeBook(config search.SearchConfig) (board.Move, bool) {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		return board.Move{}, false
	}
	if ue.prober == nil {
		ue.prober = book.NewProber(config.BookFiles)
	}
	move, ok := ue.prober.Probe(ue.engine.GetState().Board)
	if !ok {
		return board.Move{}, false
	}
	return *move, true
}

func (ue *Engine) handleStop() string {
	if ue.searching {
		close(ue.stopChannel)
	}
	return ""
}

func (ue *Engine) handleSetOption(args []string) string {
	name, value, err := ue.protocol.ParseSetOption(args)
	if err != nil {
		ue.debugLogger.Printf("invalid option: %v", args)
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
		// Applied only if set before the first "go"; see SetBookFile/probeBook.
		ue.SetBookFile(value)
	}

	return ""
}

func (ue *Engine) handleNewGame() string {
	ue.debugLogger.Println("=== NEW GAME ===")
	ue.engine.Reset()
	ue.moveNumber = 0

	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		minimaxEngine.ClearSearchState()
	}

	hashSize := 128
	if hashStr, exists := ue.options["Hash"]; exists {
		if _, err := fmt.Sscanf(hashStr, "%d", &hashSize); err != nil || hashSize < 1 {
			hashSize = 128
		}
	}

	ue.debugLogger.Printf("ENGINE CONFIG: Hash=%dMB", hashSize)

	return ""
}

func (ue *Engine) calculateMoveTime(params SearchParams, player movegen.Player, moveNumber int) time.Duration {
	var timeLeft time.Duration
	var increment time.Duration
	if player == movegen.White {
		timeLeft = params.WTime
		increment = params.WInc
	} else {
		timeLeft = params.BTime
		increment = params.BInc
	}

	estimatedMovesRemaining := 40 - (moveNumber / 2)
	if estimatedMovesRemaining < 10 {
		estimatedMovesRemaining = 10
	}

	baseTime := timeLeft / time.Duration(estimatedMovesRemaining)
	safeIncrement := increment * 9 / 10

	var timeFactor float64
	if timeLeft > 60*time.Second {
		timeFactor = 1.5
	} else if timeLeft > 30*time.Second {
		timeFactor = 1.2
	} else if timeLeft > 10*time.Second {
		timeFactor = 1.0
	} else {
		timeFactor = 0.7
	}

	maxTime := time.Duration(float64(baseTime)*timeFactor) + safeIncrement

	var maxSafeTime time.Duration
	if increment > 0 {
		incrementPortion := increment * 9 / 10
		baseTimePortion := timeLeft / 10
		maxSafeTime = incrementPortion + baseTimePortion

		if timeLeft < 5*time.Second {
			maxSafeTime = increment * 9 / 10
		}
	} else {
		maxSafeTime = timeLeft / 3
	}

	if maxTime > maxSafeTime {
		maxTime = maxSafeTime
	}

	minTime := 50 * time.Millisecond
	if maxTime < minTime {
		maxTime = minTime
	}

	return maxTime
}
