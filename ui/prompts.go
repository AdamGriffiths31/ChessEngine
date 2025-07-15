package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

type Prompter struct {
	scanner *bufio.Scanner
}

func NewPrompter() *Prompter {
	return &Prompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (p *Prompter) ShowWelcome() {
	fmt.Println("Chess Engine - Game Mode 1: Manual Play")
	fmt.Println("========================================")
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

func (p *Prompter) ShowGameState(state *game.GameState) {
	fmt.Printf("Current turn: %s\n", state.CurrentTurn)
	fmt.Printf("Move: %d\n", state.MoveCount)
	fmt.Println()
	fmt.Println(RenderBoard(state.Board))
	fmt.Println()
}

func (p *Prompter) PromptForMove(currentPlayer game.Player) (string, error) {
	fmt.Printf("Enter move for %s (or 'quit', 'reset', 'fen', 'moves'): ", currentPlayer)
	
	if !p.scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}
	
	return strings.TrimSpace(p.scanner.Text()), nil
}

func (p *Prompter) ShowError(err error) {
	fmt.Printf("Error: %s\n", err.Error())
	fmt.Println()
}

func (p *Prompter) ShowMessage(message string) {
	fmt.Println(message)
	fmt.Println()
}

func (p *Prompter) ShowFEN(fen string) {
	fmt.Printf("Current FEN: %s\n", fen)
	fmt.Println()
}

func (p *Prompter) ConfirmQuit() bool {
	fmt.Print("Are you sure you want to quit? (y/N): ")
	
	if !p.scanner.Scan() {
		return false
	}
	
	response := strings.ToLower(strings.TrimSpace(p.scanner.Text()))
	return response == "y" || response == "yes"
}

func (p *Prompter) ConfirmReset() bool {
	fmt.Print("Are you sure you want to reset the game? (y/N): ")
	
	if !p.scanner.Scan() {
		return false
	}
	
	response := strings.ToLower(strings.TrimSpace(p.scanner.Text()))
	return response == "y" || response == "yes"
}

func (p *Prompter) ShowGoodbye() {
	fmt.Println("Thanks for playing!")
}

func (p *Prompter) ShowMoves(moveList *moves.MoveList, playerName string) {
	displayer := NewMovesDisplayer()
	displayer.ShowMoves(moveList, playerName)
}

func (p *Prompter) ShowMoveValidated() {
	fmt.Println("Move validated ✓")
	fmt.Println()
}