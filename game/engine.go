package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

type Player int

const (
	White Player = iota
	Black
)

func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

// GetCurrentPlayer returns the current player based on board's side to move
func (e *Engine) GetCurrentPlayer() Player {
	if e.state.Board.GetSideToMove() == "w" {
		return White
	}
	return Black
}

type GameState struct {
	Board       *board.Board
	MoveCount   int
	GameOver    bool
	Winner      Player
	EnPassant   *board.Square // Track en passant target square
}

type Engine struct {
	state     *GameState
	generator *moves.Generator
	validator *moves.Validator
	logger    *log.Logger
}

func NewEngine() *Engine {
	initialBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Create logger for game engine
	logger := createGameEngineLogger()
	
	engine := &Engine{
		state: &GameState{
			Board:       initialBoard,
			MoveCount:   1,
			GameOver:    false,
			EnPassant:   nil,
		},
		generator: moves.NewGenerator(),
		validator: moves.NewValidator(),
		logger:    logger,
	}
	
	logger.Println("Game engine initialized with starting position")
	return engine
}

// createGameEngineLogger creates a file logger for game engine debugging
func createGameEngineLogger() *log.Logger {
	// Get current working directory to determine project root
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		return log.New(os.Stderr, "[GAME-ENGINE] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	// Create logs directory in project root
	logDir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory %s: %v", logDir, err)
		return log.New(os.Stderr, "[GAME-ENGINE] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	// Create timestamped log file
	logFile := filepath.Join(logDir, fmt.Sprintf("game_engine_%d.log", time.Now().Unix()))
	file, err := os.Create(logFile)
	if err != nil {
		log.Printf("Failed to create log file %s: %v", logFile, err)
		return log.New(os.Stderr, "[GAME-ENGINE] ", log.LstdFlags|log.Lmicroseconds)
	}
	
	log.Printf("Game engine debug logging to: %s", logFile)
	return log.New(file, "[GAME-ENGINE] ", log.LstdFlags|log.Lmicroseconds)
}

func (e *Engine) GetState() *GameState {
	return e.state
}

func (e *Engine) MakeMove(move board.Move) error {
	e.logger.Printf("Attempting to make move: From=%s, To=%s, Piece=%c, Captured=%c", 
		move.From.String(), move.To.String(), move.Piece, move.Captured)
	
	if e.state.GameOver {
		e.logger.Println("Game is over, ignoring move")
		return nil // Ignore moves if game is over
	}
	
	e.logger.Printf("Board state before move: %s", e.state.Board.ToFEN())
	
	// Apply the move to the board
	err := e.state.Board.MakeMove(move)
	if err != nil {
		e.logger.Printf("ERROR: Failed to make move on board: %v", err)
		return err
	}
	
	e.logger.Printf("Successfully applied move to board")
	e.logger.Printf("Board state after move: %s", e.state.Board.ToFEN())
	
	// Synchronize move count with board's full move number
	e.state.MoveCount = int(e.state.Board.GetFullMoveNumber())
	
	e.logger.Printf("Current player after move: %s (from board side-to-move: %s)", 
		e.GetCurrentPlayer().String(), e.state.Board.GetSideToMove())
	e.logger.Printf("Synchronized move count to: %d (matching board full move number)", e.state.MoveCount)
	
	return nil
}


func (e *Engine) Reset() {
	initialBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	e.state = &GameState{
		Board:       initialBoard,
		MoveCount:   1,
		GameOver:    false,
		EnPassant:   nil,
	}
}

func (e *Engine) GetCurrentFEN() string {
	return e.state.Board.ToFEN()
}

func (e *Engine) LoadFromFEN(fen string) error {
	e.logger.Printf("Loading position from FEN: %q", fen)
	
	newBoard, err := board.FromFEN(fen)
	if err != nil {
		e.logger.Printf("ERROR: Failed to parse FEN: %v", err)
		return err
	}
	
	e.state.Board = newBoard
	
	e.logger.Printf("Successfully loaded FEN position")
	e.logger.Printf("New board state: %s", e.state.Board.ToFEN())
	e.logger.Printf("Current player: %s (from board side-to-move: %s)", 
		e.GetCurrentPlayer().String(), e.state.Board.GetSideToMove())
	
	return nil
}

// GetLegalMoves returns all legal moves for the current player
func (e *Engine) GetLegalMoves() *moves.MoveList {
	return e.generator.GenerateAllMoves(e.state.Board, moves.Player(e.GetCurrentPlayer()))
}

// ValidateMove checks if a move is legal for the current player
func (e *Engine) ValidateMove(move board.Move) bool {
	currentPlayer := e.GetCurrentPlayer()
	e.logger.Printf("Validating move: From=%s, To=%s, Piece=%c, Captured=%c for player %s", 
		move.From.String(), move.To.String(), move.Piece, move.Captured, currentPlayer.String())
	
	isValid := e.validator.ValidateMove(e.state.Board, move, moves.Player(currentPlayer))
	
	if isValid {
		e.logger.Printf("Move validation PASSED")
	} else {
		e.logger.Printf("Move validation FAILED")
		
		// Log legal moves for comparison
		legalMoves := e.GetLegalMoves()
		e.logger.Printf("Legal moves available (%d total):", legalMoves.Count)
		for i := 0; i < legalMoves.Count; i++ {
			legalMove := legalMoves.Moves[i]
			e.logger.Printf("  Legal: From=%s, To=%s, Piece=%c", 
				legalMove.From.String(), legalMove.To.String(), legalMove.Piece)
		}
	}
	
	return isValid
}

// ValidateAndMakeMove validates a move and applies it if legal
func (e *Engine) ValidateAndMakeMove(move board.Move) error {
	if !e.ValidateMove(move) {
		return fmt.Errorf("illegal move: %s%s", move.From.String(), move.To.String())
	}
	
	return e.MakeMove(move)
}