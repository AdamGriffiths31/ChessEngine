package game

import (
	"fmt"

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

type GameState struct {
	Board       *board.Board
	CurrentTurn Player
	MoveCount   int
	GameOver    bool
	Winner      Player
	EnPassant   *board.Square // Track en passant target square
}

type Engine struct {
	state     *GameState
	generator *moves.Generator
	validator *moves.Validator
}

func NewEngine() *Engine {
	initialBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	return &Engine{
		state: &GameState{
			Board:       initialBoard,
			CurrentTurn: White,
			MoveCount:   1,
			GameOver:    false,
			EnPassant:   nil,
		},
		generator: moves.NewGenerator(),
		validator: moves.NewValidator(),
	}
}

func (e *Engine) GetState() *GameState {
	return e.state
}

func (e *Engine) MakeMove(move board.Move) error {
	if e.state.GameOver {
		return nil // Ignore moves if game is over
	}
	
	// Apply the move to the board
	err := e.state.Board.MakeMove(move)
	if err != nil {
		return err
	}
	
	// Switch turns
	e.switchTurn()
	
	return nil
}

func (e *Engine) switchTurn() {
	if e.state.CurrentTurn == White {
		e.state.CurrentTurn = Black
	} else {
		e.state.CurrentTurn = White
		e.state.MoveCount++
	}
}

func (e *Engine) Reset() {
	initialBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	e.state = &GameState{
		Board:       initialBoard,
		CurrentTurn: White,
		MoveCount:   1,
		GameOver:    false,
		EnPassant:   nil,
	}
}

func (e *Engine) GetCurrentFEN() string {
	return e.state.Board.ToFEN()
}

func (e *Engine) LoadFromFEN(fen string) error {
	newBoard, err := board.FromFEN(fen)
	if err != nil {
		return err
	}
	
	e.state.Board = newBoard
	return nil
}

// GetLegalMoves returns all legal moves for the current player
func (e *Engine) GetLegalMoves() *moves.MoveList {
	return e.generator.GenerateAllMoves(e.state.Board, moves.Player(e.state.CurrentTurn))
}

// ValidateMove checks if a move is legal for the current player
func (e *Engine) ValidateMove(move board.Move) bool {
	return e.validator.ValidateMove(e.state.Board, move, moves.Player(e.state.CurrentTurn))
}

// ValidateAndMakeMove validates a move and applies it if legal
func (e *Engine) ValidateAndMakeMove(move board.Move) error {
	if !e.ValidateMove(move) {
		return fmt.Errorf("illegal move: %s%s", move.From.String(), move.To.String())
	}
	
	return e.MakeMove(move)
}