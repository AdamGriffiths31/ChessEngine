package uci

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Engine wraps the chess engine with UCI protocol support.
type Engine struct {
	engine      *game.Engine
	aiEngine    ai.Engine
	converter   *MoveConverter
	protocol    *ProtocolHandler
	options     map[string]string
	searching   bool
	stopChannel chan struct{}
	output      io.Writer
	debugLogger *log.Logger
	moveNumber  int
	commLogger  *CommunicationLogger
}

// NewUCIEngine creates a new UCI engine wrapper.
func NewUCIEngine() *Engine {
	debugLogger := createDebugLogger()

	minimaxEngine := search.NewMinimaxEngine()
	minimaxEngine.SetTranspositionTableSize(256)

	engine := &Engine{
		engine:      game.NewEngine(),
		aiEngine:    minimaxEngine,
		converter:   NewMoveConverter(),
		protocol:    NewProtocolHandler(),
		options:     make(map[string]string),
		searching:   false,
		stopChannel: make(chan struct{}),
		debugLogger: debugLogger,
		commLogger:  NewUCICommunicationLogger(),
	}

	return engine
}

func createDebugLogger() *log.Logger {
	logLocations := []string{
		"/tmp/chess",
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
	if ue.commLogger != nil {
		ue.output = ue.commLogger.WrapWriter(output)
	}

	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()

		if ue.commLogger != nil {
			ue.commLogger.LogIncoming(line)
		}

		response := ue.HandleCommand(line)

		if response != "" {
			if _, err := fmt.Fprintln(ue.output, response); err != nil {
				ue.debugLogger.Printf("UCI-ERROR: Failed to write response: %v", err)
			}
		}

		cmd := ue.protocol.ParseCommand(line)
		if cmd.Name == "quit" {
			if ue.commLogger != nil {
				ue.commLogger.LogGameTermination("UCI quit command")
			}
			break
		}
	}

	if ue.commLogger != nil {
		ue.commLogger.Close()
	}

	return scanner.Err()
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
		"option name Threads type spin default 1 min 1 max 32",
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

	ue.engine.Reset()

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

	// Log final position for debugging evaluation swings
	ue.debugLogger.Printf("POSITION: %s", ue.engine.GetCurrentFEN())

	return ""
}

func (ue *Engine) handleGo(args []string) {
	if ue.searching {
		return
	}

	params := ue.protocol.ParseGo(args)
	player := moves.Player(ue.engine.GetCurrentPlayer())

	ue.searching = true
	ue.stopChannel = make(chan struct{})

	threadCount := 1
	if threadStr, exists := ue.options["Threads"]; exists {
		if _, err := fmt.Sscanf(threadStr, "%d", &threadCount); err != nil || threadCount < 1 || threadCount > 32 {
			threadCount = 1
		}
	}

	// Get absolute path to opening book relative to executable
	execPath, err := os.Executable()
	if err != nil {
		ue.debugLogger.Printf("BOOK-ERROR: Failed to get executable path: %v", err)
		execPath = ""
	}

	var bookFiles []string
	if execPath != "" {
		// Assuming executable is in tools/bin/ and book is at game/openings/testdata/performance.bin
		execDir := filepath.Dir(execPath)
		projectRoot := filepath.Join(execDir, "..", "..")
		bookPath := filepath.Join(projectRoot, "game", "openings", "testdata", "performance.bin")
		if absPath, err := filepath.Abs(bookPath); err == nil {
			bookPath = absPath
		}

		if _, err := os.Stat(bookPath); err == nil {
			bookFiles = []string{bookPath}
			ue.debugLogger.Printf("BOOK: Found opening book at: %s", bookPath)
		} else {
			ue.debugLogger.Printf("BOOK-ERROR: Opening book not found at: %s", bookPath)
		}
	}

	config := ai.SearchConfig{
		MaxDepth:            999,
		MaxTime:             5 * time.Second,
		DebugMode:           false,
		UseOpeningBook:      len(bookFiles) > 0,
		BookFiles:           bookFiles,
		BookSelectMode:      ai.BookSelectWeightedRandom,
		BookWeightThreshold: 1,
		LMRMinDepth:         3,
		LMRMinMoves:         4,
		NumThreads:          threadCount,
	}

	if params.Depth > 0 {
		config.MaxDepth = params.Depth
	}
	if params.MoveTime > 0 {
		safetyMargin := 100 * time.Millisecond
		config.MaxTime = params.MoveTime - safetyMargin
	}
	if params.Infinite {
		config.MaxTime = 24 * time.Hour
	}

	if params.WTime > 0 || params.BTime > 0 {
		config.MaxTime = ue.calculateMoveTime(params, player, ue.moveNumber)
	}

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

	searchStart := time.Now()
	var result ai.SearchResult
	var searchDuration time.Duration

	func() {
		defer func() {
			if r := recover(); r != nil {
				ue.debugLogger.Printf("PANIC CAUGHT: Search panicked for move %d: %v", ue.moveNumber, r)
				ue.debugLogger.Printf("PANIC CONTEXT: Position=%s, Player=%s, ThreadCount=%d",
					ue.engine.GetCurrentFEN(), player, config.NumThreads)

				result = ai.SearchResult{
					BestMove: board.Move{
						From: board.Square{File: -1, Rank: -1},
						To:   board.Square{File: -1, Rank: -1},
					},
					Score: -ai.MateScore,
					Stats: ai.SearchStats{},
				}
				searchDuration = time.Since(searchStart)
				panic(r)
			}
		}()

		result = ue.aiEngine.FindBestMove(stopCtx, ue.engine.GetState().Board, player, config)
		searchDuration = time.Since(searchStart)
	}()

	bestMoveUCI := ue.converter.ToUCI(result.BestMove)
	formattedBestMove := ue.protocol.FormatBestMove(bestMoveUCI)

	var ttStatsStr string
	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		hits, misses, collisions, hitRate := minimaxEngine.GetTranspositionTableStats()
		ttStatsStr = fmt.Sprintf("TT: %d hits, %d misses, %d collisions, %.1f%% hit rate",
			hits, misses, collisions, hitRate)
	}

	// Calculate move ordering percentage
	var moveOrderPct float64
	if result.Stats.TotalCutoffs > 0 {
		moveOrderPct = float64(result.Stats.FirstMoveCutoffs) / float64(result.Stats.TotalCutoffs) * 100
	}

	ue.debugLogger.Printf("Move %d: %s | Score: %d | Depth: %d | Nodes: %d | Q: %d | NM: %d/%d | LMR: %d | TTC: %d | DP: %d | RZ: %d/%d | MO: %.0f%% | Time: %.3fs | Book: %t | %s",
		ue.moveNumber, bestMoveUCI, result.Score, result.Stats.Depth, result.Stats.NodesSearched, result.Stats.QNodes, result.Stats.NullCutoffs, result.Stats.NullMoves, result.Stats.LMRReductions, result.Stats.TTCutoffs, result.Stats.DeltaPruned, result.Stats.RazoringCutoffs, result.Stats.RazoringAttempts, moveOrderPct, searchDuration.Seconds(), result.Stats.BookMoveUsed, ttStatsStr)

	if ue.moveNumber%10 == 0 {
		if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
			hits, misses, collisions, hitRate := minimaxEngine.GetTranspositionTableStats()
			if hits > 0 || misses > 0 {
				if _, err := fmt.Fprintf(ue.output, "info string TT: %d hits, %d misses, %d collisions, %.1f%% hit rate\n",
					hits, misses, collisions, hitRate); err != nil {
					ue.debugLogger.Printf("UCI-ERROR: Failed to write TT stats: %v", err)
				}
			}
		}
	}

	if _, err := fmt.Fprintf(ue.output, "%s\n", formattedBestMove); err != nil {
		ue.debugLogger.Printf("UCI-ERROR: Failed to write best move: %v", err)
	}
	ue.searching = false
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
	case "Threads":
		var threadCount int
		if _, err := fmt.Sscanf(value, "%d", &threadCount); err != nil || threadCount <= 0 || threadCount > 32 {
			ue.debugLogger.Printf("Invalid thread count: %s (must be 1-32)", value)
		}
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

	return ""
}

func (ue *Engine) calculateMoveTime(params SearchParams, player moves.Player, moveNumber int) time.Duration {
	var timeLeft time.Duration
	var increment time.Duration
	if player == moves.White {
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
