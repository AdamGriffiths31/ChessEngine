// Package modes provides different game modes for the chess engine.
package modes

import (
	"errors"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/ui"
)

// ManualMode implements human vs human game mode
type ManualMode struct {
	engine    *game.Engine
	prompter  *ui.Prompter
	parser    *game.MoveParser
	isRunning bool
}

// NewManualMode creates a new manual game mode
func NewManualMode() *ManualMode {
	engine := game.NewEngine()
	prompter := ui.NewPrompter()
	parser := game.NewMoveParser(true)

	return &ManualMode{
		engine:    engine,
		prompter:  prompter,
		parser:    parser,
		isRunning: false,
	}
}

// Run starts the manual game mode
func (mm *ManualMode) Run() error {
	mm.isRunning = true
	mm.prompter.ShowWelcome()

	for mm.isRunning {
		state := mm.engine.GetState()

		mm.prompter.ShowGameState(state)

		currentPlayer := mm.engine.GetCurrentPlayer()

		mm.parser.SetCurrentPlayer(currentPlayer == game.White)

		input, err := mm.prompter.PromptForMove(currentPlayer)
		if err != nil {
			mm.prompter.ShowError(err)
			continue
		}

		err = mm.handleInput(input)
		if err != nil {
			mm.prompter.ShowError(err)
		}
	}

	mm.prompter.ShowGoodbye()
	return nil
}

func (mm *ManualMode) handleInput(input string) error {
	move, err := mm.parser.ParseMove(input, mm.engine.GetState().Board)
	if err != nil {
		return mm.handleSpecialCommand(err.Error())
	}

	err = mm.engine.ValidateAndMakeMove(move)
	if err != nil {
		return err
	}

	mm.prompter.ShowMoveValidated()

	return nil
}

func (mm *ManualMode) handleSpecialCommand(command string) error {
	switch command {
	case "QUIT":
		if mm.prompter.ConfirmQuit() {
			mm.isRunning = false
		}
		return nil

	case "RESET":
		if mm.prompter.ConfirmReset() {
			mm.engine.Reset()
			mm.prompter.ShowMessage("Game reset!")
		}
		return nil

	case "FEN":
		fen := mm.engine.GetCurrentFEN()
		mm.prompter.ShowFEN(fen)
		return nil

	case "MOVES":
		moveList := mm.engine.GetLegalMoves()
		defer moves.ReleaseMoveList(moveList)
		playerName := mm.engine.GetCurrentPlayer().String()
		mm.prompter.ShowMoves(moveList, playerName)
		return nil

	default:
		return errors.New(command)
	}
}
