package uci

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
}

// NewUCIEngine creates a new UCI engine wrapper
func NewUCIEngine() *UCIEngine {
	return &UCIEngine{
		engine:      game.NewEngine(),
		aiEngine:    search.NewMinimaxEngine(),
		converter:   NewMoveConverter(),
		protocol:    NewProtocolHandler(),
		options:     make(map[string]string),
		searching:   false,
		stopChannel: make(chan struct{}),
	}
}

// Run starts the UCI engine main loop
func (ue *UCIEngine) Run(input io.Reader, output io.Writer) error {
	ue.output = output
	scanner := bufio.NewScanner(input)
	
	for scanner.Scan() {
		line := scanner.Text()
		response := ue.HandleCommand(line)
		
		if response != "" {
			fmt.Fprintln(output, response)
		}
		
		// Check for quit command
		cmd := ue.protocol.ParseCommand(line)
		if cmd.Name == "quit" {
			break
		}
	}
	
	return scanner.Err()
}

// HandleCommand processes a single UCI command and returns the response
func (ue *UCIEngine) HandleCommand(input string) string {
	
	cmd := ue.protocol.ParseCommand(input)
	
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
	fen, moveList, err := ue.protocol.ParsePosition(args)
	if err != nil {
		return "" // Silently ignore invalid position commands
	}
	
	// Load the FEN position
	err = ue.engine.LoadFromFEN(fen)
	if err != nil {
		return "" // Silently ignore invalid FEN
	}
	
	// Apply the moves
	for _, moveStr := range moveList {
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			return "" // Silently ignore invalid moves
		}
		
		err = ue.engine.MakeMove(move)
		if err != nil {
			return "" // Silently ignore illegal moves
		}
	}
	
	return ""
}

// handleGo processes the 'go' command and starts searching
func (ue *UCIEngine) handleGo(args []string) {
	if ue.searching {
		return // Already searching
	}
	
	ue.searching = true
	ue.stopChannel = make(chan struct{})
	
	params := ue.protocol.ParseGo(args)
	
	// Convert current turn to moves.Player (read from board, not engine state)
	var player moves.Player
	boardSideToMove := ue.engine.GetState().Board.GetSideToMove()
	if boardSideToMove == "w" {
		player = moves.White
	} else {
		player = moves.Black
	}
	
	
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
	
	// Convert result to UCI format
	bestMoveUCI := ue.converter.ToUCI(result.BestMove)
	
	// Send the result
	response := ue.protocol.FormatBestMove(bestMoveUCI)
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
	ue.engine.Reset()
	return ""
}