// Package ui provides interactive command-line prompts for the benchmark launcher.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompter handles user interaction and console output for the benchmark launcher
type Prompter struct {
	scanner *bufio.Scanner
}

// NewPrompter creates a new Prompter instance for handling user input
func NewPrompter() *Prompter {
	return &Prompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// ShowError displays an error message to the user
func (p *Prompter) ShowError(err error) {
	fmt.Printf("Error: %s\n", err.Error())
	fmt.Println()
}

// PromptForNumber prompts the user to enter a number
func (p *Prompter) PromptForNumber(prompt string, minVal, maxVal int) (int, error) {
	for {
		fmt.Printf("%s (%d-%d): ", prompt, minVal, maxVal)

		if !p.scanner.Scan() {
			return 0, fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(p.scanner.Text())

		var number int
		_, err := fmt.Sscanf(input, "%d", &number)
		if err != nil || number < minVal || number > maxVal {
			fmt.Printf("Invalid input. Please enter a number between %d and %d.\n", minVal, maxVal)
			continue
		}

		return number, nil
	}
}

// PromptForConfirmation prompts the user for a yes/no confirmation
func (p *Prompter) PromptForConfirmation(prompt string, defaultYes bool) (bool, error) {
	var suffix string
	if defaultYes {
		suffix = " (Y/n): "
	} else {
		suffix = " (y/N): "
	}

	fmt.Print(prompt + suffix)

	if !p.scanner.Scan() {
		return false, fmt.Errorf("failed to read input")
	}

	input := strings.ToLower(strings.TrimSpace(p.scanner.Text()))

	if input == "" {
		return defaultYes, nil
	}

	return input == "y" || input == "yes", nil
}
