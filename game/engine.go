// Package game provides the chess game engine with move validation and game state management.
package game

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

// Player represents which side is to move in the game.
type Player int

// Player constants representing the two sides in chess.
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

// State holds the current state of a chess game.
type State struct {
	Board     *board.Board
	MoveCount int
	GameOver  bool
	Winner    Player
	IsDraw    bool
	EnPassant *board.Square
}

// Engine manages the chess game state and move execution.
type Engine struct {
	state       *State
	generator   *moves.Generator
	validator   *moves.Validator
	hashHistory []uint64
}

// NewEngine creates a new chess engine with starting position.
func NewEngine() *Engine {
	initialBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		panic(fmt.Sprintf("failed to create starting position: %v", err))
	}

	return &Engine{
		state: &State{
			Board:     initialBoard,
			MoveCount: 1,
			GameOver:  false,
			EnPassant: nil,
		},
		generator:   moves.NewGenerator(),
		validator:   moves.NewValidator(),
		hashHistory: []uint64{openings.GetPolyglotHash().HashPosition(initialBoard)},
	}
}

// GetCurrentPlayer returns the current player based on board's side to move
func (e *Engine) GetCurrentPlayer() Player {
	if e.state.Board.GetSideToMove() == "w" {
		return White
	}
	return Black
}

// GetState returns the current game state.
func (e *Engine) GetState() *State {
	return e.state
}

// MakeMove applies a move to the game board.
func (e *Engine) MakeMove(move board.Move) error {
	// board.Board.MakeMove only falls back to looking up the moving piece
	// when move.Piece == board.Empty; a zero-value (unset) Piece field is
	// used as-is and silently corrupts the board. Mirror the defensive
	// lookup that board.Board.MakeMoveWithUndo already performs so callers
	// that only supply From/To/Promotion (as game.Engine's own tests do)
	// behave correctly.
	if move.Piece == board.Empty || move.Piece == 0 {
		move.Piece = e.state.Board.GetPiece(move.From.Rank, move.From.File)
	}

	err := e.state.Board.MakeMove(move)
	if err != nil {
		return fmt.Errorf("failed to make move %s%s: %w", move.From.String(), move.To.String(), err)
	}

	e.state.MoveCount = e.state.Board.GetFullMoveNumber()
	e.updateGameOverState()
	return nil
}

// updateGameOverState checks for checkmate, stalemate, the 50-move rule,
// and threefold repetition, populating State.GameOver/Winner/IsDraw.
func (e *Engine) updateGameOverState() {
	hash := openings.GetPolyglotHash().HashPosition(e.state.Board)
	e.hashHistory = append(e.hashHistory, hash)

	currentPlayer := e.GetCurrentPlayer()
	legalMoves := e.generator.GenerateAllMoves(e.state.Board, moves.Player(currentPlayer))
	legalCount := legalMoves.Count
	moves.ReleaseMoveList(legalMoves)

	switch {
	case legalCount == 0:
		e.state.GameOver = true
		if e.generator.IsKingInCheck(e.state.Board, moves.Player(currentPlayer)) {
			e.state.Winner = opponentOf(currentPlayer)
		} else {
			e.state.IsDraw = true
		}
	case e.state.Board.GetHalfMoveClock() >= 100:
		e.state.GameOver = true
		e.state.IsDraw = true
	case e.occursThreeTimes(hash):
		e.state.GameOver = true
		e.state.IsDraw = true
	}
}

// occursThreeTimes reports whether hash has appeared at least three times
// in the game's position history (threefold repetition).
func (e *Engine) occursThreeTimes(hash uint64) bool {
	count := 0
	for _, h := range e.hashHistory {
		if h == hash {
			count++
		}
	}
	return count >= 3
}

// opponentOf returns the other player.
func opponentOf(p Player) Player {
	if p == White {
		return Black
	}
	return White
}

// Reset returns the engine to the starting position.
func (e *Engine) Reset() {
	initialBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		panic(fmt.Sprintf("failed to create starting position: %v", err))
	}

	e.state = &State{
		Board:     initialBoard,
		MoveCount: 1,
		GameOver:  false,
		EnPassant: nil,
	}
	e.hashHistory = []uint64{openings.GetPolyglotHash().HashPosition(initialBoard)}
}

// GetCurrentFEN returns the current position in FEN notation.
func (e *Engine) GetCurrentFEN() string {
	return e.state.Board.ToFEN()
}

// LoadFromFEN loads a position from FEN notation.
func (e *Engine) LoadFromFEN(fen string) error {
	newBoard, err := board.FromFEN(fen)
	if err != nil {
		return fmt.Errorf("failed to parse FEN %q: %w", fen, err)
	}

	e.state.Board = newBoard
	e.state.GameOver = false
	e.state.IsDraw = false
	e.hashHistory = []uint64{openings.GetPolyglotHash().HashPosition(newBoard)}
	return nil
}

// GetLegalMoves returns all legal moves for the current player.
// IMPORTANT: Caller must call moves.ReleaseMoveList() when done with the returned MoveList.
func (e *Engine) GetLegalMoves() *moves.MoveList {
	return e.generator.GenerateAllMoves(e.state.Board, moves.Player(e.GetCurrentPlayer()))
}

// ValidateMove checks if a move is legal for the current player.
func (e *Engine) ValidateMove(move board.Move) bool {
	currentPlayer := e.GetCurrentPlayer()
	return e.validator.ValidateMove(e.state.Board, move, moves.Player(currentPlayer))
}

// ValidateAndMakeMove validates a move and applies it if legal.
func (e *Engine) ValidateAndMakeMove(move board.Move) error {
	if !e.ValidateMove(move) {
		return fmt.Errorf("illegal move: %s%s", move.From.String(), move.To.String())
	}

	return e.MakeMove(move)
}
