// Package modes provides different game modes for the chess engine.
package modes

import (
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
	"github.com/AdamGriffiths31/ChessEngine/internal/ui"
)

// EloBenchmarkMode plays calibrated games against Stockfish at several
// UCI_Elo levels and estimates ChessEngine's own Elo rating.
type EloBenchmarkMode struct {
	prompter *ui.Prompter
	runner   *bench.EloRunner
	logger   *bench.EloResultsLogger
}

// NewEloBenchmarkMode creates a new Elo benchmark mode.
func NewEloBenchmarkMode() *EloBenchmarkMode {
	rootPath, err := os.Getwd()
	if err != nil {
		rootPath = "."
	}

	return &EloBenchmarkMode{
		prompter: ui.NewPrompter(),
		runner:   bench.NewEloRunner(rootPath),
		logger:   bench.NewEloResultsLogger(rootPath),
	}
}

// Run starts the Elo benchmark mode.
func (em *EloBenchmarkMode) Run() error {
	fmt.Println("Chess Engine - Elo Benchmark")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Plays calibrated games against Stockfish at several UCI_Elo")
	fmt.Println("levels and estimates ChessEngine's own Elo rating.")
	fmt.Println()

	if err := em.runner.LoadEngines(); err != nil {
		return fmt.Errorf("failed to load engine configurations: %w", err)
	}

	config, err := em.getConfig()
	if err != nil {
		return fmt.Errorf("failed to get benchmark settings: %w", err)
	}

	fmt.Println()
	fmt.Printf("Anchor center: %d, count: %d, step: %d\n", config.AnchorCenter, config.AnchorCount, config.AnchorStep)
	fmt.Printf("Games per level: %d, Time control: %d+%d\n", config.GamesPerLevel, config.TimeMs/1000, config.IncMs/1000)
	fmt.Printf("Total games: %d\n\n", config.AnchorCount*config.GamesPerLevel)

	proceed, err := em.prompter.PromptForConfirmation("Proceed with Elo benchmark?", false)
	if err != nil {
		return err
	}
	if !proceed {
		fmt.Println("Elo benchmark cancelled.")
		return nil
	}

	fmt.Println("\n>>> Starting Elo benchmark...")
	result, err := em.runner.Run(config)
	if err != nil {
		em.prompter.ShowError(err)
		return err
	}

	fmt.Println(em.logger.FormatResults(result))

	if err := em.logger.LogResults(result); err != nil {
		em.prompter.ShowError(fmt.Errorf("failed to log results: %w", err))
	}

	fmt.Println("\nElo benchmark completed successfully!")
	return nil
}

func (em *EloBenchmarkMode) getConfig() (bench.EloConfig, error) {
	anchorCenter, err := em.prompter.PromptForNumber("Anchor center Elo", 600, 3000)
	if err != nil {
		return bench.EloConfig{}, err
	}

	anchorCount, err := em.prompter.PromptForNumber("Number of anchor levels", 1, 9)
	if err != nil {
		return bench.EloConfig{}, err
	}

	gamesPerLevel, err := em.prompter.PromptForNumber("Games per anchor level", 1, 20)
	if err != nil {
		return bench.EloConfig{}, err
	}

	baseMinutes, err := em.prompter.PromptForNumber("Base time per side in minutes", 1, 30)
	if err != nil {
		return bench.EloConfig{}, err
	}

	incSeconds, err := em.prompter.PromptForNumber("Increment in seconds", 0, 30)
	if err != nil {
		return bench.EloConfig{}, err
	}

	return bench.EloConfig{
		AnchorCenter:  anchorCenter,
		AnchorCount:   anchorCount,
		AnchorStep:    100,
		GamesPerLevel: gamesPerLevel,
		TimeMs:        baseMinutes * 60000,
		IncMs:         incSeconds * 1000,
		MaxPlies:      300,
	}, nil
}
