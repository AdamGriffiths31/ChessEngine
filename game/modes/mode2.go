package modes

import (
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/ui"
)

// ComputerMode implements Player vs Computer game mode
type ComputerMode struct {
	engine     *game.Engine
	prompter   *ui.Prompter
	parser     *game.MoveParser
	computer   *ai.ComputerPlayer
	humanColor game.Player
	isRunning  bool
}

// NewComputerMode creates a new Player vs Computer mode
func NewComputerMode() *ComputerMode {
	engine := game.NewEngine()
	prompter := ui.NewPrompter()
	parser := game.NewMoveParser(game.White)

	// Create computer player with minimax engine
	aiEngine := search.NewMinimaxEngine()
	config := ai.SearchConfig{
		MaxDepth:            4,
		MaxTime:             3 * time.Second,
		UseOpeningBook:      true,
		BookFiles:           []string{"game/openings/testdata/performance.bin"},
		BookSelectMode:      ai.BookSelectWeightedRandom,
		BookWeightThreshold: 1,
	}
	computer := ai.NewComputerPlayer("Computer", aiEngine, config)

	return &ComputerMode{
		engine:     engine,
		prompter:   prompter,
		parser:     parser,
		computer:   computer,
		humanColor: game.White, // Human plays white by default
		isRunning:  false,
	}
}

// Run starts the Player vs Computer game
func (cm *ComputerMode) Run() error {
	cm.isRunning = true
	cm.prompter.ShowWelcomeComputer()

	// Ask player to choose color
	cm.selectPlayerColor()

	// Ask for difficulty level
	cm.selectDifficulty()

	// Ask for debug mode
	cm.selectDebugMode()

	for cm.isRunning {
		state := cm.engine.GetState()

		// Show current game state
		cm.prompter.ShowGameState(state)

		// Check game over
		if state.GameOver {
			cm.handleGameOver(state)
			break
		}

		// Handle current player's turn
		if state.CurrentTurn == cm.humanColor {
			cm.handleHumanTurn()
		} else {
			cm.handleComputerTurn()
		}
	}

	cm.prompter.ShowGoodbye()
	return nil
}

// selectPlayerColor asks the player to choose their color
func (cm *ComputerMode) selectPlayerColor() {
	for {
		color, err := cm.prompter.PromptForColorChoice()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
		cm.humanColor = color
		fmt.Printf("You selected: %s\n\n", color)
		break
	}
}

// selectDifficulty asks the player to choose difficulty
func (cm *ComputerMode) selectDifficulty() {
	for {
		difficulty, err := cm.prompter.PromptForDifficulty()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
		cm.computer.SetDifficulty(difficulty)
		fmt.Printf("Difficulty set to: %s\n\n", cm.computer.GetDifficulty())
		break
	}
}

// selectDebugMode asks the player to enable/disable debug mode
func (cm *ComputerMode) selectDebugMode() {
	for {
		debugEnabled, err := cm.prompter.PromptForDebugMode()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
		cm.computer.SetDebugMode(debugEnabled)
		if debugEnabled {
			fmt.Printf("Debug mode: Enabled\n\n")
		} else {
			fmt.Printf("Debug mode: Disabled\n\n")
		}
		break
	}
}

// handleHumanTurn processes the human player's move
func (cm *ComputerMode) handleHumanTurn() {
	state := cm.engine.GetState()
	cm.parser.SetCurrentPlayer(state.CurrentTurn)

	input, err := cm.prompter.PromptForMove(state.CurrentTurn)
	if err != nil {
		cm.prompter.ShowError(err)
		return
	}

	// Parse and validate move
	move, err := cm.parser.ParseMove(input, state.Board)
	if err != nil {
		cm.handleSpecialCommand(err.Error())
		return
	}

	// Apply the move
	err = cm.engine.ValidateAndMakeMove(move)
	if err != nil {
		cm.prompter.ShowError(err)
		return
	}

	cm.prompter.ShowMoveValidated()
}

// handleComputerTurn processes the computer's move
func (cm *ComputerMode) handleComputerTurn() {
	state := cm.engine.GetState()

	fmt.Println("Computer is thinking...")

	// Convert game.Player to moves.Player
	var movesPlayer moves.Player
	if state.CurrentTurn == game.White {
		movesPlayer = moves.White
	} else {
		movesPlayer = moves.Black
	}

	// Get computer's move with statistics
	result, err := cm.computer.GetMoveWithStats(state.Board, movesPlayer, 3*time.Second)
	if err != nil {
		cm.prompter.ShowError(err)
		return
	}

	// Apply the move
	err = cm.engine.MakeMove(result.BestMove)
	if err != nil {
		cm.prompter.ShowError(err)
		return
	}

	moveString := fmt.Sprintf("%s%s", result.BestMove.From.String(), result.BestMove.To.String())

	// Show debug information if debug mode is enabled
	if cm.computer.IsDebugMode() {
		cm.prompter.ShowSearchStats(moveString, result.Stats, result.Score, movesPlayer)
	} else {
		fmt.Printf("Computer plays: %s\n", moveString)
	}
}

// handleSpecialCommand handles special commands like quit, reset, etc.
func (cm *ComputerMode) handleSpecialCommand(command string) {
	// Similar to mode1 implementation
	switch command {
	case "QUIT":
		if cm.prompter.ConfirmQuit() {
			cm.isRunning = false
		}
	case "RESET":
		if cm.prompter.ConfirmReset() {
			cm.engine.Reset()
			cm.selectPlayerColor()
			cm.selectDifficulty()
			cm.selectDebugMode()
		}
		// ... other commands
	}
}

// handleGameOver handles the end of the game
func (cm *ComputerMode) handleGameOver(state *game.GameState) {
	if state.Winner == cm.humanColor {
		fmt.Println("Congratulations! You won!")
	} else {
		fmt.Println("Computer wins! Better luck next time.")
	}
}
