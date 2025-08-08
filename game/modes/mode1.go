package modes

import (
	"errors"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/ui"
)

type ManualMode struct {
	engine     *game.Engine
	prompter   *ui.Prompter
	parser     *game.MoveParser
	isRunning  bool
}

func NewManualMode() *ManualMode {
	engine := game.NewEngine()
	prompter := ui.NewPrompter()
	parser := game.NewMoveParser(game.White)
	
	return &ManualMode{
		engine:    engine,
		prompter:  prompter,
		parser:    parser,
		isRunning: false,
	}
}

func (mm *ManualMode) Run() error {
	mm.isRunning = true
	mm.prompter.ShowWelcome()
	
	for mm.isRunning {
		state := mm.engine.GetState()
		
		// Show current game state
		mm.prompter.ShowGameState(state)
		
		// Get current player from engine
		currentPlayer := mm.engine.GetCurrentPlayer()
		
		// Update parser's current player
		mm.parser.SetCurrentPlayer(currentPlayer)
		
		// Get move input from user
		input, err := mm.prompter.PromptForMove(currentPlayer)
		if err != nil {
			mm.prompter.ShowError(err)
			continue
		}
		
		// Handle the input
		err = mm.handleInput(input)
		if err != nil {
			mm.prompter.ShowError(err)
		}
	}
	
	mm.prompter.ShowGoodbye()
	return nil
}

func (mm *ManualMode) handleInput(input string) error {
	// Parse the move
	move, err := mm.parser.ParseMove(input, mm.engine.GetState().Board)
	if err != nil {
		return mm.handleSpecialCommand(err.Error())
	}
	
	// Validate and apply the move
	err = mm.engine.ValidateAndMakeMove(move)
	if err != nil {
		return err
	}
	
	// Show validation feedback
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