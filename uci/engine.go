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

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// UCIEngine wraps the chess engine with UCI protocol support
type UCIEngine struct {
	engine      *game.Engine
	aiEngine    ai.Engine
	converter   *MoveConverter
	protocol    *ProtocolHandler
	options     map[string]string
	searching   bool
	stopChannel chan struct{}
	output      io.Writer
	debugLogger *log.Logger
	moveNumber  int // Track move number for logging

	// Enhanced communication logging
	commLogger *UCICommunicationLogger
}

// NewUCIEngine creates a new UCI engine wrapper
func NewUCIEngine() *UCIEngine {
	// Create debug logger
	debugLogger := createDebugLogger()

	// Create minimax engine with transposition table enabled by default for UCI
	minimaxEngine := search.NewMinimaxEngine()
	minimaxEngine.SetTranspositionTableSize(256) // Default 256MB TT for UCI mode

	engine := &UCIEngine{
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

// createDebugLogger creates a file logger for UCI debugging
func createDebugLogger() *log.Logger {
	// Try multiple locations for the log file
	logLocations := []string{
		"/tmp/chess",
	}

	for _, dir := range logLocations {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create log directory %s: %v", dir, err)
			continue
		}

		// Create timestamped log file
		logFile := filepath.Join(dir, fmt.Sprintf("uci_debug_%d.log", time.Now().Unix()))
		file, err := os.Create(logFile)
		if err != nil {
			log.Printf("Failed to create log file %s: %v", logFile, err)
			continue
		}

		// Log to both file and stderr for visibility
		multiWriter := io.MultiWriter(file, os.Stderr)
		log.Printf("UCI debug logging to: %s", logFile)
		return log.New(multiWriter, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}

	// Fallback to stderr only
	log.Printf("Failed to create debug log file, using stderr only")
	return log.New(os.Stderr, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
}

// Run starts the UCI engine main loop
func (ue *UCIEngine) Run(input io.Reader, output io.Writer) error {
	// Set up output and enhanced logging
	ue.output = output
	if ue.commLogger != nil {
		ue.output = ue.commLogger.WrapWriter(output)
	}

	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()

		// Log all incoming UCI commands with timestamp
		ue.debugLogger.Printf("UCI-IN: [%s] %s", time.Now().Format("15:04:05.000"), line)

		// Enhanced communication logging
		if ue.commLogger != nil {
			ue.commLogger.LogIncoming(line)
		}

		response := ue.HandleCommand(line)

		if response != "" {
			// Log all outgoing UCI responses with timestamp
			ue.debugLogger.Printf("UCI-OUT: [%s] %s", time.Now().Format("15:04:05.000"), response)
			fmt.Fprintln(ue.output, response)
		} else {
			// Log when no response is given
			ue.debugLogger.Printf("UCI-OUT: [%s] (no response)", time.Now().Format("15:04:05.000"))
		}

		// Check for quit command
		cmd := ue.protocol.ParseCommand(line)
		if cmd.Name == "quit" {
			if ue.commLogger != nil {
				ue.commLogger.LogGameTermination("UCI quit command")
			}
			break
		}
	}

	// Clean up
	if ue.commLogger != nil {
		ue.commLogger.Close()
	}

	return scanner.Err()
}

// HandleCommand processes a single UCI command and returns the response
func (ue *UCIEngine) HandleCommand(input string) string {
	cmd := ue.protocol.ParseCommand(input)

	// Log command parsing details
	ue.debugLogger.Printf("CMD-PARSE: Command='%s', Args=%v", cmd.Name, cmd.Args)

	// Validate command name
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
		ue.handleGo(cmd.Args) // Run search synchronously
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
		return "" // Ignore unknown commands
	}
}

// handleUCI responds to the 'uci' command
func (ue *UCIEngine) handleUCI() string {
	response := ue.protocol.FormatUCIResponse("ChessEngine", "Adam Griffiths")

	// Add engine options
	options := []string{
		ue.protocol.FormatOption("Hash", "spin", "128"),
		ue.protocol.FormatOption("Threads", "spin", "1"),
		ue.protocol.FormatOption("MaxDepth", "spin", "6"),
	}

	return response + "\n" + strings.Join(options, "\n")
}

// handleIsReady responds to the 'isready' command
func (ue *UCIEngine) handleIsReady() string {
	return ue.protocol.FormatReadyOK()
}

// handlePosition processes the 'position' command
func (ue *UCIEngine) handlePosition(args []string) string {
	ue.debugLogger.Printf("POSITION: Received args=%v", args)

	fen, moveList, err := ue.protocol.ParsePosition(args)
	if err != nil {
		ue.debugLogger.Fatalf("POSITION: Failed to parse position - fen: %q, error: %v", fen, err)
	}

	ue.engine.Reset()

	// Apply the moves
	for _, moveStr := range moveList {
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			ue.debugLogger.Fatalf("POSITION: Failed to convert UCI move - move: %s, board: %s, error: %v",
				moveStr, ue.engine.GetCurrentFEN(), err)
		}

		err = ue.engine.MakeMove(move)
		if err != nil {
			ue.debugLogger.Fatalf("POSITION: Failed to make move - move: %+v, error: %v", move, err)
		}
	}

	ue.debugLogger.Printf("POSITION: Final position loaded: %s", ue.engine.GetCurrentFEN())
	return ""
}

// handleGo processes the 'go' command and starts searching
func (ue *UCIEngine) handleGo(args []string) {
	if ue.searching {
		ue.debugLogger.Printf("GO-CMD: Ignoring go command - already searching")
		return // Already searching
	}

	ue.debugLogger.Printf("GO-CMD: Starting search with args: %v", args)

	params := ue.protocol.ParseGo(args)
	ue.debugLogger.Printf("GO-PARSE: Parsed parameters - Depth=%d, MoveTime=%v, WTime=%v, BTime=%v, WInc=%v, BInc=%v, Infinite=%v",
		params.Depth, params.MoveTime, params.WTime, params.BTime, params.WInc, params.BInc, params.Infinite)

	// Get current player from game engine
	player := moves.Player(ue.engine.GetCurrentPlayer())
	ue.debugLogger.Printf("GO-PARSE: Current player from engine state: %s", player)

	ue.searching = true
	ue.stopChannel = make(chan struct{})

	// Configure search parameters
	config := ai.SearchConfig{
		MaxDepth:            6,               // Default depth
		MaxTime:             5 * time.Second, // Default time
		DebugMode:           false,
		UseAlphaBeta:        false,
		UseOpeningBook:      true,
		BookFiles:           []string{"/home/adam/Documents/git/ChessEngine/game/openings/testdata/performance.bin"},
		BookSelectMode:      ai.BookSelectWeightedRandom,
		BookWeightThreshold: 1,
	}

	// Apply search parameters
	if params.Depth > 0 {
		config.MaxDepth = params.Depth
	}
	if params.MoveTime > 0 {
		// Leave a 100ms safety margin to avoid time forfeits
		safetyMargin := 100 * time.Millisecond
		config.MaxTime = params.MoveTime - safetyMargin
	}
	if params.Infinite {
		config.MaxTime = 24 * time.Hour // Very long time for infinite
	}

	// Calculate appropriate move time based on time controls
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

	ue.debugLogger.Printf("Move %d search starting - Position: %s, Player: %v",
		ue.moveNumber, ue.engine.GetCurrentFEN(), player)

	// Search for best move
	searchStart := time.Now()
	result := ue.aiEngine.FindBestMove(stopCtx, ue.engine.GetState().Board, player, config)
	searchDuration := time.Since(searchStart)

	if config.UseOpeningBook && result.Stats.BookMoveUsed {
		ue.debugLogger.Printf("Book move used for move %d", ue.moveNumber)
	}

	bestMoveUCI := ue.converter.ToUCI(result.BestMove)
	formattedBestMove := ue.protocol.FormatBestMove(bestMoveUCI)

	// Get transposition table statistics if available
	var ttStatsStr string
	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		hits, misses, collisions, hitRate := minimaxEngine.GetTranspositionTableStats()
		ttStatsStr = fmt.Sprintf("TT: %d hits, %d misses, %d collisions, %.1f%% hit rate", 
			hits, misses, collisions, hitRate)
	}

	ue.debugLogger.Printf("AI chose move: %s (%s) (From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d score=%d depth=%d nodes=%d book=%t time=%.3fs/%.3fs) %s",
		bestMoveUCI, formattedBestMove, result.BestMove.From.String(), result.BestMove.To.String(),
		result.BestMove.Piece, result.BestMove.Captured, result.BestMove.Promotion, result.Score, result.Stats.Depth, result.Stats.NodesSearched, result.Stats.BookMoveUsed, searchDuration.Seconds(), config.MaxTime.Seconds(), ttStatsStr)

	// Print TT stats to UCI output as info string (visible to GUI)
	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		hits, misses, collisions, hitRate := minimaxEngine.GetTranspositionTableStats()
		if hits > 0 || misses > 0 {
			fmt.Fprintf(ue.output, "info string TT: %d hits, %d misses, %d collisions, %.1f%% hit rate\n", 
				hits, misses, collisions, hitRate)
		}
	}

	fmt.Fprintf(ue.output, "%s\n", formattedBestMove)
	ue.searching = false
}

// handleStop processes the 'stop' command
func (ue *UCIEngine) handleStop() string {
	if ue.searching {
		close(ue.stopChannel)
	}
	return ""
}

// handleSetOption processes the 'setoption' command
func (ue *UCIEngine) handleSetOption(args []string) string {
	name, value, err := ue.protocol.ParseSetOption(args)
	if err != nil {
		ue.debugLogger.Printf("invalid option: %v", args)
		return ""
	}

	ue.options[name] = value

	// Handle specific options
	switch name {
	case "Hash":
		// Parse hash table size in MB
		var hashSizeMB int
		if _, err := fmt.Sscanf(value, "%d", &hashSizeMB); err == nil && hashSizeMB > 0 {
			if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
				minimaxEngine.SetTranspositionTableSize(hashSizeMB)
				ue.debugLogger.Printf("Set transposition table size to %d MB", hashSizeMB)
			}
		}
	}

	return ""
}

// handleNewGame processes the 'ucinewgame' command
func (ue *UCIEngine) handleNewGame() string {
	ue.debugLogger.Println("NEWGAME: Resetting engine...")
	ue.engine.Reset()
	ue.moveNumber = 0

	// Clear transposition table for new game
	if minimaxEngine, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		minimaxEngine.ClearSearchState()
		ue.debugLogger.Println("NEWGAME: Cleared transposition table")
	}

	return ""
}

// calculateMoveTime calculates the appropriate time allocation for a move based on time controls
func (ue *UCIEngine) calculateMoveTime(params SearchParams, player moves.Player, moveNumber int) time.Duration {
	// Use appropriate time for current player
	var timeLeft time.Duration
	var increment time.Duration
	if player == moves.White {
		timeLeft = params.WTime
		increment = params.WInc
	} else {
		timeLeft = params.BTime
		increment = params.BInc
	}

	// Estimate remaining moves (assume 40 moves per side, reduce as game progresses)
	estimatedMovesRemaining := 40 - (moveNumber / 2)
	if estimatedMovesRemaining < 10 {
		estimatedMovesRemaining = 10 // Minimum for safety
	}

	// Base time allocation: divide remaining time by estimated moves
	baseTime := timeLeft / time.Duration(estimatedMovesRemaining)

	// Add most of the increment (keep small safety margin)
	safeIncrement := increment * 9 / 10

	// Use larger portion of time when we have plenty, be more conservative when low
	var timeFactor float64
	if timeLeft > 60*time.Second {
		timeFactor = 1.5 // Use 150% of average when we have time
	} else if timeLeft > 30*time.Second {
		timeFactor = 1.2 // Use 120% of average
	} else if timeLeft > 10*time.Second {
		timeFactor = 1.0 // Use exactly average allocation
	} else {
		timeFactor = 0.7 // Be conservative when very low on time
	}

	maxTime := time.Duration(float64(baseTime)*timeFactor) + safeIncrement

	// With increment, we can be more aggressive since we get time back
	// Without increment, be conservative to avoid time trouble
	var maxSafeTime time.Duration
	if increment > 0 {
		// With increment: use up to 90% of increment + reasonable portion of base time
		incrementPortion := increment * 9 / 10
		baseTimePortion := timeLeft / 10 // Only use 10% of base time as safety
		maxSafeTime = incrementPortion + baseTimePortion

		if timeLeft < 5*time.Second {
			maxSafeTime = increment * 8 / 10
		}
	} else {
		// No increment: use conservative 1/3 of remaining time
		maxSafeTime = timeLeft / 3
	}

	if maxTime > maxSafeTime {
		maxTime = maxSafeTime
	}

	// Minimum time: at least 50ms to make a reasonable move
	minTime := 50 * time.Millisecond
	if maxTime < minTime {
		maxTime = minTime
	}

	return maxTime
}
