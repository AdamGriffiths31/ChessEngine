package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Prompter handles user interaction and console output for the chess game
type Prompter struct {
	scanner *bufio.Scanner
}

// NewPrompter creates a new Prompter instance for handling user input
func NewPrompter() *Prompter {
	return &Prompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// ShowWelcome displays the welcome message and command instructions
func (p *Prompter) ShowWelcome() {
	fmt.Println("Chess Engine - Manual Play")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  - Enter moves in coordinate notation (e.g., e2e4)")
	fmt.Println("  - 'o-o' or '0-0' for kingside castling")
	fmt.Println("  - 'o-o-o' or '0-0-0' for queenside castling")
	fmt.Println("  - 'moves' to show all legal moves")
	fmt.Println("  - 'quit' or 'exit' to quit the game")
	fmt.Println("  - 'reset' to start a new game")
	fmt.Println("  - 'fen' to display current FEN string")
	fmt.Println()
}

// ShowWelcomeComputer displays welcome message for computer vs human games
func (p *Prompter) ShowWelcomeComputer() {
	fmt.Println("Chess Engine - Player vs Computer")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  - Enter moves in coordinate notation (e.g., e2e4)")
	fmt.Println("  - 'o-o' or '0-0' for kingside castling")
	fmt.Println("  - 'o-o-o' or '0-0-0' for queenside castling")
	fmt.Println("  - 'moves' to show all legal moves")
	fmt.Println("  - 'quit' or 'exit' to quit the game")
	fmt.Println("  - 'reset' to start a new game")
	fmt.Println("  - 'fen' to display current FEN string")
	fmt.Println()
}

// ShowGameState displays the current game state including board and status
func (p *Prompter) ShowGameState(state *game.State) {
	// Get current turn from board's side to move
	var currentTurn string
	if state.Board.GetSideToMove() == "w" {
		currentTurn = "White"
	} else {
		currentTurn = "Black"
	}
	fmt.Printf("Current turn: %s\n", currentTurn)
	fmt.Printf("Move: %d\n", state.MoveCount)
	fmt.Println()
	fmt.Println(RenderBoard(state.Board))
	fmt.Println()
}

// PromptForMove prompts the user to enter a move and returns the input
func (p *Prompter) PromptForMove(currentPlayer game.Player) (string, error) {
	fmt.Printf("Enter move for %s (or 'quit', 'reset', 'fen', 'moves'): ", currentPlayer)

	if !p.scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	return strings.TrimSpace(p.scanner.Text()), nil
}

// ShowError displays an error message to the user
func (p *Prompter) ShowError(err error) {
	fmt.Printf("Error: %s\n", err.Error())
	fmt.Println()
}

// ShowMessage displays a message to the user
func (p *Prompter) ShowMessage(message string) {
	fmt.Println(message)
	fmt.Println()
}

// ShowFEN displays the current position in FEN notation
func (p *Prompter) ShowFEN(fen string) {
	fmt.Printf("Current FEN: %s\n", fen)
	fmt.Println()
}

// ConfirmQuit asks the user to confirm if they want to quit the game
func (p *Prompter) ConfirmQuit() bool {
	fmt.Print("Are you sure you want to quit? (y/N): ")

	if !p.scanner.Scan() {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(p.scanner.Text()))
	return response == "y" || response == "yes"
}

// ConfirmReset asks the user to confirm if they want to reset the game
func (p *Prompter) ConfirmReset() bool {
	fmt.Print("Are you sure you want to reset the game? (y/N): ")

	if !p.scanner.Scan() {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(p.scanner.Text()))
	return response == "y" || response == "yes"
}

// ShowGoodbye displays the goodbye message when exiting
func (p *Prompter) ShowGoodbye() {
	fmt.Println("Thanks for playing!")
}

// ShowMoves displays the available legal moves for debugging
func (p *Prompter) ShowMoves(moveList *moves.MoveList, playerName string) {
	displayer := NewMovesDisplayer()
	displayer.ShowMoves(moveList, playerName)
}

// ShowMoveValidated confirms that a move was successfully validated
func (p *Prompter) ShowMoveValidated() {
	fmt.Println("Move validated ✓")
	fmt.Println()
}

// PromptForChoice prompts the user to select from a list of options
func (p *Prompter) PromptForChoice(prompt string, options []string) (int, error) {
	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	fmt.Print("Enter choice (1-" + fmt.Sprintf("%d", len(options)) + "): ")

	if !p.scanner.Scan() {
		return -1, fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(p.scanner.Text())

	// Parse the choice
	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(options) {
		return -1, fmt.Errorf("invalid choice, please enter a number between 1 and %d", len(options))
	}

	return choice - 1, nil // Convert to 0-indexed
}

// PromptForColorChoice prompts the user to select their color
func (p *Prompter) PromptForColorChoice() (game.Player, error) {
	options := []string{"White (play first)", "Black (play second)"}
	choice, err := p.PromptForChoice("\nChoose your color:", options)
	if err != nil {
		return game.White, err
	}

	if choice == 0 {
		return game.White, nil
	}
	return game.Black, nil
}

// PromptForDifficulty prompts the user to select difficulty level
func (p *Prompter) PromptForDifficulty() (string, error) {
	options := []string{"Easy", "Medium", "Hard"}
	choice, err := p.PromptForChoice("\nChoose difficulty:", options)
	if err != nil {
		return "", err
	}

	difficulties := []string{"easy", "medium", "hard"}
	return difficulties[choice], nil
}

// PromptForDebugMode prompts the user to enable/disable debug mode
func (p *Prompter) PromptForDebugMode() (bool, error) {
	options := []string{"Disabled", "Enabled"}
	choice, err := p.PromptForChoice("\nDebug mode (show search statistics):", options)
	if err != nil {
		return false, err
	}

	return choice == 1, nil
}

// ShowSearchStats displays debug information about the search
func (p *Prompter) ShowSearchStats(move string, stats ai.SearchStats, score ai.EvaluationScore, player moves.Player) {
	fmt.Printf("Computer plays: %s\n", move)
	fmt.Printf("   Search depth: %d\n", stats.Depth)
	fmt.Printf("   Nodes searched: %d\n", stats.NodesSearched)
	fmt.Printf("   Search time: %v\n", stats.Time)

	// Convert score to White's perspective for consistent display
	displayScore := score
	if player == moves.Black {
		displayScore = -score
	}

	fmt.Printf("   Position evaluation: %d centipawns (from White's perspective)\n", int(displayScore))

	// Show what the score means
	if displayScore > 0 {
		fmt.Printf("   Assessment: White is better by %d centipawns\n", int(displayScore))
	} else if displayScore < 0 {
		fmt.Printf("   Assessment: Black is better by %d centipawns\n", int(-displayScore))
	} else {
		fmt.Printf("   Assessment: Position is equal\n")
	}

	// Show efficiency metrics
	nodesPerSecond := float64(stats.NodesSearched) / stats.Time.Seconds()
	fmt.Printf("   Search efficiency: %.0f nodes/second\n", nodesPerSecond)

	// Show debug information if available
	if len(stats.DebugInfo) > 0 {
		fmt.Println("   Debug info:")
		for _, info := range stats.DebugInfo {
			fmt.Printf("     %s\n", info)
		}
	}

	fmt.Println()
}
