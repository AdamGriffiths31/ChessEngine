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
	moveNumber  int           // Track move number for logging
	
	// Enhanced communication logging
	commLogger  *UCICommunicationLogger
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
		commLogger:  NewUCICommunicationLogger(),
	}
	
	return engine
}

// createDebugLogger creates a file logger for UCI debugging
func createDebugLogger() *log.Logger {
	// Try multiple locations for the log file
	logLocations := []string{
		"/tmp",
		".",
		"logs",
		"../logs",
	}
	
	for _, dir := range logLocations {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		
		// Create timestamped log file
		logFile := filepath.Join(dir, fmt.Sprintf("uci_debug_%d.log", time.Now().Unix()))
		file, err := os.Create(logFile)
		if err != nil {
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
		ue.debugLogger.Printf("CMD-PARSE: Empty command name, ignoring")
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
		ue.debugLogger.Printf("POSITION: Parse error: %v", err)
		return "" // Silently ignore invalid position commands
	}
	
	ue.debugLogger.Printf("POSITION: Parsed FEN=%q, %d moves to apply", fen, len(moveList))
	
	// Load the FEN position
	err = ue.engine.LoadFromFEN(fen)
	if err != nil {
		ue.debugLogger.Printf("POSITION: FEN load error: %v", err)
		return "" // Silently ignore invalid FEN
	}
	
	ue.debugLogger.Printf("POSITION: FEN loaded, current=%s", ue.engine.GetCurrentFEN())
	
	// Test move generation immediately after FEN load
	immediateMoves := ue.engine.GetLegalMoves()
	ue.debugLogger.Printf("POSITION: Legal moves after FEN load: %d total", immediateMoves.Count)
	d2d4Found := false
	for i := 0; i < immediateMoves.Count; i++ {
		moveUCI := ue.converter.ToUCI(immediateMoves.Moves[i])
		if i < 8 {
			ue.debugLogger.Printf("POSITION:   %s", moveUCI)
		}
		if moveUCI == "d2d4" {
			d2d4Found = true
		}
	}
	ue.debugLogger.Printf("POSITION: d2d4 found in immediate moves: %v", d2d4Found)
	
	// Apply the moves
	for i, moveStr := range moveList {
		// Log the current board state before applying the move
		beforeFEN := ue.engine.GetCurrentFEN()
		beforeLegalMoves := ue.engine.GetLegalMoves()
		ue.debugLogger.Printf("POSITION: Before applying move %d (%s):", i+1, moveStr)
		ue.debugLogger.Printf("POSITION:   Current FEN: %s", beforeFEN)
		ue.debugLogger.Printf("POSITION:   Legal moves available: %d", beforeLegalMoves.Count)
		
		// Check if the move we're about to apply is in the legal moves
		moveFound := false
		for j := 0; j < beforeLegalMoves.Count; j++ { // Check ALL legal moves, not just first 10
			legalMoveUCI := ue.converter.ToUCI(beforeLegalMoves.Moves[j])
			if legalMoveUCI == moveStr {
				moveFound = true
				ue.debugLogger.Printf("POSITION:   Legal[%d]: %s ← MATCH!", j, legalMoveUCI)
			} else if j < 10 { // Only log first 10 for readability, but check all for validation
				ue.debugLogger.Printf("POSITION:   Legal[%d]: %s", j, legalMoveUCI)
			}
		}
		ue.debugLogger.Printf("POSITION:   Move %s found in legal moves: %v", moveStr, moveFound)
		
		// Convert UCI move to internal format
		ue.debugLogger.Printf("POSITION: Converting UCI move: %s", moveStr)
		move, err := ue.converter.FromUCI(moveStr, ue.engine.GetState().Board)
		if err != nil {
			ue.debugLogger.Printf("POSITION: Move conversion error for %s: %v", moveStr, err)
			return ""
		}
		ue.debugLogger.Printf("POSITION: Converted to internal move: From=%s, To=%s, Piece=%d, Captured=%d", 
			move.From.String(), move.To.String(), move.Piece, move.Captured)
		
		// Apply the move
		ue.debugLogger.Printf("POSITION: Applying move to engine...")
		err = ue.engine.MakeMove(move)
		if err != nil {
			ue.debugLogger.Printf("POSITION: Move application error for %s: %v", moveStr, err)
			ue.debugLogger.Printf("POSITION: Failed move details: From=%s, To=%s, Piece=%d, Captured=%d", 
				move.From.String(), move.To.String(), move.Piece, move.Captured)
			return ""
		}
		
		// Log the new state after applying the move
		afterFEN := ue.engine.GetCurrentFEN()
		ue.debugLogger.Printf("POSITION: Successfully applied %s", moveStr)
		ue.debugLogger.Printf("POSITION:   New FEN: %s", afterFEN)
		ue.debugLogger.Printf("POSITION:   FEN change: %s → %s", beforeFEN, afterFEN)
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
	
	// Get current player from game engine state (now synchronized with board)
	player := moves.Player(ue.engine.GetState().CurrentTurn)
	ue.debugLogger.Printf("GO-PARSE: Current player from engine state: %s", player)
	
	// Determine our color on first 'go' command
	if ue.myColor == nil {
		ue.myColor = &player
	}
	
	ue.searching = true
	ue.stopChannel = make(chan struct{})
	
	// Configure search parameters
	config := ai.SearchConfig{
		MaxDepth:     6, // Default depth
		MaxTime:      5 * time.Second, // Default time
		DebugMode:    false,
		UseAlphaBeta: false,
		// Enable opening book
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
		
		// Improved time management strategy
		
		// Estimate remaining moves (assume 40 moves per side, reduce as game progresses)
		estimatedMovesRemaining := 40 - (ue.moveNumber / 2)
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
		
		config.MaxTime = time.Duration(float64(baseTime)*timeFactor) + safeIncrement
		
		// Ensure we don't use more than 1/3 of remaining time on any single move
		maxSafeTime := timeLeft / 3
		if config.MaxTime > maxSafeTime {
			config.MaxTime = maxSafeTime
		}
		
		// Minimum time: at least 50ms to make a reasonable move
		minTime := 50 * time.Millisecond
		if config.MaxTime < minTime {
			config.MaxTime = minTime
		}
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
	
	// Increment move counter
	ue.moveNumber++
	
	// Log position before search
	ue.debugLogger.Printf("Move %d search starting - Position: %s, Player: %v", 
		ue.moveNumber, ue.engine.GetCurrentFEN(), player)
	
	// Add board synchronization check
	engineBoard := ue.engine.GetState().Board
	ue.debugLogger.Printf("SYNC-CHECK: Engine state validation before search:")
	ue.debugLogger.Printf("SYNC-CHECK:   FEN: %s", ue.engine.GetCurrentFEN())
	ue.debugLogger.Printf("SYNC-CHECK:   Side to move: %s", engineBoard.GetSideToMove())
	ue.debugLogger.Printf("SYNC-CHECK:   Move number: %d", ue.moveNumber)
	ue.debugLogger.Printf("SYNC-CHECK:   Player from game state: %s", player)
	ue.debugLogger.Printf("SYNC-CHECK:   Castling rights: %s", engineBoard.GetCastlingRights())
	enPassant := engineBoard.GetEnPassantTarget()
	if enPassant != nil {
		ue.debugLogger.Printf("SYNC-CHECK:   En passant target: %s", enPassant.String())
	} else {
		ue.debugLogger.Printf("SYNC-CHECK:   En passant target: none")
	}
	
	// Compare what UCI engine sees vs what AI will see  
	ue.debugLogger.Printf("Pre-search comparison:")
	ue.debugLogger.Printf("  Engine FEN: %s", ue.engine.GetCurrentFEN())
	ue.debugLogger.Printf("  Board side to move: %s", engineBoard.GetSideToMove())
	
	// Test: Generate moves using the same board the AI will use
	testLegalMoves := ue.engine.GetLegalMoves()
	ue.debugLogger.Printf("  Pre-search legal moves: %d total", testLegalMoves.Count)
	
	// Log ALL moves to see the full list
	e2e4Found := false
	d2d4Found := false
	for i := 0; i < testLegalMoves.Count; i++ {
		move := testLegalMoves.Moves[i]
		moveUCI := ue.converter.ToUCI(move)
		if i < 12 {  // Show more moves
			ue.debugLogger.Printf("    [%d]: %s (From=%s, To=%s, Piece=%d)", i, moveUCI, move.From.String(), move.To.String(), move.Piece)
		}
		if moveUCI == "e2e4" {
			e2e4Found = true
			ue.debugLogger.Printf("    FOUND e2e4 at index %d: From=%s, To=%s, Piece=%d, Captured=%d", i, move.From.String(), move.To.String(), move.Piece, move.Captured)
		}
		if moveUCI == "d2d4" {
			d2d4Found = true
			ue.debugLogger.Printf("    FOUND d2d4 at index %d: From=%s, To=%s, Piece=%d, Captured=%d", i, move.From.String(), move.To.String(), move.Piece, move.Captured)
		}
	}
	ue.debugLogger.Printf("  Pre-search summary: e2e4=%v, d2d4=%v", e2e4Found, d2d4Found)
	
	// Search for best move
	searchStart := time.Now()
	result := ue.aiEngine.FindBestMove(stopCtx, ue.engine.GetState().Board, player, config)
	searchDuration := time.Since(searchStart)
	
	// Debug: Check if book move was attempted but failed
	if config.UseOpeningBook && !result.Stats.BookMoveUsed && ue.moveNumber <= 10 {
		ue.debugLogger.Printf("Book enabled but no book move found for move %d", ue.moveNumber)
	}
	
	// Log the AI's chosen move before validation
	bestMoveUCI := ue.converter.ToUCI(result.BestMove)
	ue.debugLogger.Printf("AI chose move: %s (From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d)", 
		bestMoveUCI, result.BestMove.From.String(), result.BestMove.To.String(), 
		result.BestMove.Piece, result.BestMove.Captured, result.BestMove.Promotion)
	ue.debugLogger.Printf("Book move used: %v", result.Stats.BookMoveUsed)
	
	// Custom validation: Check if the move exists in legal moves by From/To comparison
	legalMoves := ue.engine.GetLegalMoves()
	isValid := false
	
	for i := 0; i < legalMoves.Count; i++ {
		legal := legalMoves.Moves[i]
		// Compare the essential move components (From/To squares)
		if legal.From.String() == result.BestMove.From.String() && 
		   legal.To.String() == result.BestMove.To.String() {
			isValid = true
			ue.debugLogger.Printf("Move validation: FOUND matching legal move at index %d", i)
			break
		}
	}
	
	ue.debugLogger.Printf("Move validation result: %v (From/To comparison)", isValid)
	
	if !isValid {
		ue.debugLogger.Printf("ILLEGAL-MOVE: AI selected move that failed validation: %s", bestMoveUCI)
		ue.debugLogger.Printf("ILLEGAL-MOVE: Board state: %s", ue.engine.GetCurrentFEN())
		ue.debugLogger.Printf("ILLEGAL-MOVE: Move details - From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d", 
			result.BestMove.From.String(), result.BestMove.To.String(), 
			result.BestMove.Piece, result.BestMove.Captured, result.BestMove.Promotion)
		ue.debugLogger.Printf("ILLEGAL-MOVE: Book move used: %v", result.Stats.BookMoveUsed)
		
		// Get legal moves to compare
		legalMoves := ue.engine.GetLegalMoves()
		ue.debugLogger.Printf("Available legal moves (%d total):", legalMoves.Count)
		
		// Log detailed board state for debugging
		board := ue.engine.GetState().Board
		ue.debugLogger.Printf("Board internal state:")
		ue.debugLogger.Printf("  Side to move: %s", board.GetSideToMove())
		ue.debugLogger.Printf("  Castling rights: %s", board.GetCastlingRights())
		ue.debugLogger.Printf("  En passant: %v", board.GetEnPassantTarget())
		ue.debugLogger.Printf("  Halfmove clock: %d", board.GetHalfMoveClock())
		ue.debugLogger.Printf("  Fullmove number: %d", board.GetFullMoveNumber())
		
		// Look for exact match and similar moves
		exactMatch := false
		targetMoveUCI := bestMoveUCI
		samePieceCount := 0
		
		for i := 0; i < legalMoves.Count; i++ {
			move := legalMoves.Moves[i]
			moveUCI := ue.converter.ToUCI(move)
			
			// Show first 15 moves for better visibility
			if i < 15 {
				matchIndicator := ""
				if moveUCI == targetMoveUCI {
					matchIndicator = " <-- EXACT MATCH!"
					exactMatch = true
				}
				ue.debugLogger.Printf("  [%d] Legal: %s (From=%s, To=%s, Piece=%d, Captured=%d)%s", 
					i, moveUCI, move.From.String(), move.To.String(), move.Piece, move.Captured, matchIndicator)
			}
			
			// Count moves with same piece
			if move.From.String() == result.BestMove.From.String() {
				samePieceCount++
			}
			
			// Detailed comparison for potential matches
			if moveUCI == targetMoveUCI {
				ue.debugLogger.Printf("  MATCH FOUND: Legal move matches UCI notation")
				ue.debugLogger.Printf("  Legal move:  From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d", 
					move.From.String(), move.To.String(), move.Piece, move.Captured, move.Promotion)
				ue.debugLogger.Printf("  AI move:     From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d", 
					result.BestMove.From.String(), result.BestMove.To.String(), 
					result.BestMove.Piece, result.BestMove.Captured, result.BestMove.Promotion)
			}
		}
		
		ue.debugLogger.Printf("Validation analysis: exactMatch=%v, samePieceCount=%d", exactMatch, samePieceCount)
		
		// Log the piece at d2 specifically
		// Need to check what piece is actually at d2
		ue.debugLogger.Printf("Piece at d2: investigating board state...")
		return
	}
	
	// Convert result to UCI format (bestMoveUCI already declared above)
	
	// Send the result
	response := ue.protocol.FormatBestMove(bestMoveUCI)
	
	// CRITICAL LOG: Final bestmove being sent to cutechess-cli
	ue.debugLogger.Printf("=== SENDING FINAL BESTMOVE ===")
	ue.debugLogger.Printf("FINAL-MOVE: Move selected by AI: %s", bestMoveUCI)
	ue.debugLogger.Printf("FINAL-MOVE: Internal move details: From=%s, To=%s, Piece=%d, Captured=%d, Promotion=%d", 
		result.BestMove.From.String(), result.BestMove.To.String(), 
		result.BestMove.Piece, result.BestMove.Captured, result.BestMove.Promotion)
	ue.debugLogger.Printf("FINAL-MOVE: Current position: %s", ue.engine.GetCurrentFEN())
	ue.debugLogger.Printf("FINAL-MOVE: Move validated against legal moves: %v", isValid)
	ue.debugLogger.Printf("FINAL-MOVE: Formatted response: %s", response)
	ue.debugLogger.Printf("================================")
	
	ue.debugLogger.Printf("UCI-OUT: %s", response)
	if ue.output != nil {
		fmt.Fprintf(ue.output, "%s\n", response)
	}
	
	// Log concise move summary with book move indicator
	bookIndicator := ""
	if result.Stats.BookMoveUsed {
		bookIndicator = " [BOOK]"
	}
	ue.debugLogger.Printf("Move %d: %s%s, Time used: %v / %v (%.1f%%), Validation: OK", 
		ue.moveNumber, bestMoveUCI, bookIndicator, searchDuration, config.MaxTime, float64(searchDuration)/float64(config.MaxTime)*100)
	
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
	ue.debugLogger.Println("NEWGAME: Resetting engine...")
	ue.engine.Reset()
	ue.myColor = nil // Reset color assignment for new game
	ue.moveNumber = 0 // Reset move counter for new game
	
	// Test move generation immediately after reset
	resetMoves := ue.engine.GetLegalMoves()
	ue.debugLogger.Printf("NEWGAME: After reset, %d legal moves available", resetMoves.Count)
	d2d4Found := false
	for i := 0; i < resetMoves.Count; i++ {
		moveUCI := ue.converter.ToUCI(resetMoves.Moves[i])
		if i < 8 {
			ue.debugLogger.Printf("NEWGAME:   %s", moveUCI)
		}
		if moveUCI == "d2d4" {
			d2d4Found = true
		}
	}
	ue.debugLogger.Printf("NEWGAME: d2d4 found after reset: %v", d2d4Found)
	ue.debugLogger.Printf("NEWGAME: Reset position: %s", ue.engine.GetCurrentFEN())
	ue.debugLogger.Println("NEWGAME: New game started - opening book enabled")
	return ""
}