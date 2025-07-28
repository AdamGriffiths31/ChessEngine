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
	myColor     *moves.Player // Track which color this engine plays (nil = not determined yet)
}

// NewUCIEngine creates a new UCI engine wrapper
func NewUCIEngine() *UCIEngine {
	// Create debug logger
	debugLogger := createDebugLogger()
	
	engine := &UCIEngine{
		engine:      game.NewEngine(),
		aiEngine:    search.NewMinimaxEngine(),
		converter:   NewMoveConverter(),
		protocol:    NewProtocolHandler(),
		options:     make(map[string]string),
		searching:   false,
		stopChannel: make(chan struct{}),
		debugLogger: debugLogger,
	}
	
	debugLogger.Println("UCI Engine initialized")
	return engine
}

// createDebugLogger creates a file logger for UCI debugging
func createDebugLogger() *log.Logger {
	// Get current working directory to determine project root
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		return log.New(os.Stderr, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	// Create logs directory in project root (assuming we're always run from project root)
	logDir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory %s: %v", logDir, err)
		return log.New(os.Stderr, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	// Create timestamped log file
	logFile := filepath.Join(logDir, fmt.Sprintf("uci_debug_%d.log", time.Now().Unix()))
	file, err := os.Create(logFile)
	if err != nil {
		log.Printf("Failed to create log file %s: %v", logFile, err)
		return log.New(os.Stderr, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	log.Printf("UCI debug logging to: %s", logFile)
	return log.New(file, "[UCI-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
}

// Run starts the UCI engine main loop
func (ue *UCIEngine) Run(input io.Reader, output io.Writer) error {
	ue.output = output
	scanner := bufio.NewScanner(input)
	
	ue.debugLogger.Println("UCI engine started, waiting for commands")
	
	for scanner.Scan() {
		line := scanner.Text()
		ue.debugLogger.Printf("RAW INPUT: %q (length: %d bytes)", line, len(line))
		
		// Log any non-printable characters
		for i, r := range line {
			if r < 32 || r > 126 {
				ue.debugLogger.Printf("WARNING: Non-printable character at position %d: ASCII %d (0x%02X)", i, r, r)
			}
		}
		
		response := ue.HandleCommand(line)
		
		if response != "" {
			ue.debugLogger.Printf("RAW OUTPUT: %q (length: %d bytes)", response, len(response))
			fmt.Fprintln(output, response)
		}
		
		// Check for quit command
		cmd := ue.protocol.ParseCommand(line)
		if cmd.Name == "quit" {
			ue.debugLogger.Println("Received quit command, shutting down")
			break
		}
	}
	
	return scanner.Err()
}

// HandleCommand processes a single UCI command and returns the response
func (ue *UCIEngine) HandleCommand(input string) string {
	ue.debugLogger.Printf("PARSING COMMAND: %q", input)
	
	// Log raw parsing details
	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)
	ue.debugLogger.Printf("After trim/split: %d parts: %v", len(parts), parts)
	
	cmd := ue.protocol.ParseCommand(input)
	ue.debugLogger.Printf("Parsed command: name=%q, args=%v (arg count: %d)", cmd.Name, cmd.Args, len(cmd.Args))
	
	// Validate command name
	if cmd.Name == "" {
		ue.debugLogger.Printf("WARNING: Empty command name from input %q", input)
		return ""
	}
	
	switch cmd.Name {
	case "uci":
		ue.debugLogger.Printf("Processing UCI command")
		return ue.handleUCI()
	case "isready":
		ue.debugLogger.Printf("Processing ISREADY command")
		return ue.handleIsReady()
	case "position":
		ue.debugLogger.Printf("Processing POSITION command with %d args", len(cmd.Args))
		return ue.handlePosition(cmd.Args)
	case "go":
		ue.debugLogger.Printf("Processing GO command with %d args", len(cmd.Args))
		ue.handleGo(cmd.Args) // Run search synchronously
		return ""
	case "stop":
		ue.debugLogger.Printf("Processing STOP command")
		return ue.handleStop()
	case "setoption":
		ue.debugLogger.Printf("Processing SETOPTION command with %d args", len(cmd.Args))
		return ue.handleSetOption(cmd.Args)
	case "ucinewgame":
		ue.debugLogger.Printf("Processing UCINEWGAME command")
		return ue.handleNewGame()
	case "quit":
		ue.debugLogger.Printf("Processing QUIT command")
		return ""
	default:
		ue.debugLogger.Printf("WARNING: Unknown command ignored: %q (args: %v)", cmd.Name, cmd.Args)
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
	ue.debugLogger.Printf("Processing POSITION command with %d args", len(args))
	
	fen, moveList, err := ue.protocol.ParsePosition(args)
	if err != nil {
		ue.debugLogger.Printf("ERROR: Failed to parse position command: %v", err)
		ue.debugLogger.Printf("ERROR: Raw args were: %v", args)
		return "" // Silently ignore invalid position commands
	}
	
	ue.debugLogger.Printf("Parsed position: FEN=%q, %d moves to apply", fen, len(moveList))
	
	// Load the FEN position
	err = ue.engine.LoadFromFEN(fen)
	if err != nil {
		ue.debugLogger.Printf("ERROR: Failed to load FEN position: %v", err)
		return "" // Silently ignore invalid FEN
	}
	
	// Apply the moves
	for _, moveStr := range moveList {
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			ue.debugLogger.Printf("ERROR: Failed to convert UCI move %q: %v", moveStr, err)
			return ""
		}
		
		err = ue.engine.MakeMove(move)
		if err != nil {
			ue.debugLogger.Printf("ERROR: Failed to make move %q: %v", moveStr, err)
			return ""
		}
	}
	
	ue.debugLogger.Printf("Position loaded: %s", ue.engine.GetCurrentFEN())
	
	return ""
}

// handleGo processes the 'go' command and starts searching
func (ue *UCIEngine) handleGo(args []string) {
	if ue.searching {
		ue.debugLogger.Println("Already searching, ignoring go command")
		return // Already searching
	}
	
	params := ue.protocol.ParseGo(args)
	
	// Get current player from game engine state (now synchronized with board)
	player := moves.Player(ue.engine.GetState().CurrentTurn)
	boardSideToMove := ue.engine.GetState().Board.GetSideToMove()
	
	// Determine our color on first 'go' command
	if ue.myColor == nil {
		ue.myColor = &player
		ue.debugLogger.Printf("Determined our color: %v", *ue.myColor)
	}
	
	ue.searching = true
	ue.stopChannel = make(chan struct{})
	
	// Verify synchronization between board and game state
	expectedPlayer := moves.White
	if boardSideToMove == "b" {
		expectedPlayer = moves.Black
	}
	if player != expectedPlayer {
		ue.debugLogger.Printf("WARNING: Player state mismatch! Board=%s, GameState=%v", boardSideToMove, player)
	}
	// Get legal moves before search
	legalMoves := ue.engine.GetLegalMoves()
	ue.debugLogger.Printf("Generated %d legal moves", legalMoves.Count)
	
	// Configure search parameters
	config := ai.SearchConfig{
		MaxDepth:     6, // Default depth
		MaxTime:      5 * time.Second, // Default time
		DebugMode:    false,
		UseAlphaBeta: false,
	}
	
	// Apply search parameters
	if params.Depth > 0 {
		config.MaxDepth = params.Depth
	}
	if params.MoveTime > 0 {
		config.MaxTime = params.MoveTime
	}
	if params.Infinite {
		config.MaxTime = 24 * time.Hour // Very long time for infinite
	}
	
	// Calculate appropriate move time based on time controls
	if params.WTime > 0 || params.BTime > 0 {
		// Use appropriate time for current player
		var timeLeft time.Duration
		if player == moves.White {
			timeLeft = params.WTime
		} else {
			timeLeft = params.BTime
		}
		
		// Simple time management: use 1/20th of remaining time + increment
		config.MaxTime = timeLeft/20 + params.WInc
	}
	
	// Create search context
	ctx, cancel := context.WithTimeout(context.Background(), config.MaxTime)
	defer cancel()
	
	// Create a context that also responds to stop commands
	stopCtx, stopCancel := context.WithCancel(ctx)
	defer stopCancel()
	
	go func() {
		select {
		case <-ue.stopChannel:
			stopCancel()
		case <-ctx.Done():
		}
	}()
	
	// Search for best move
	result := ue.aiEngine.FindBestMove(stopCtx, ue.engine.GetState().Board, player, config)
	
	// Validate that the chosen move is actually legal  
	if !ue.engine.ValidateMove(result.BestMove) {
		ue.debugLogger.Printf("ERROR: AI selected illegal move: %s", ue.converter.ToUCI(result.BestMove))
		return
	}
	
	// Convert result to UCI format
	bestMoveUCI := ue.converter.ToUCI(result.BestMove)
	
	
	// Send the result
	response := ue.protocol.FormatBestMove(bestMoveUCI)
	ue.debugLogger.Printf("SENDING BESTMOVE: %q", response)
	fmt.Fprintf(ue.output, "%s\n", response)
	
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
		return "" // Silently ignore invalid setoption commands
	}
	
	ue.options[name] = value
	return ""
}

// handleNewGame processes the 'ucinewgame' command
func (ue *UCIEngine) handleNewGame() string {
	ue.debugLogger.Println("Processing ucinewgame command - resetting engine")
	ue.engine.Reset()
	ue.myColor = nil // Reset color assignment for new game
	ue.debugLogger.Printf("Engine reset complete. New position: %s", ue.engine.GetCurrentFEN())
	ue.debugLogger.Println("Color assignment reset - will be determined on first 'go' command")
	return ""
}