package modes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AdamGriffiths31/ChessEngine/benchmark"
	"github.com/AdamGriffiths31/ChessEngine/ui"
)

// BenchmarkMode implements chess engine benchmarking functionality
type BenchmarkMode struct {
	prompter *ui.Prompter
	runner   *benchmark.Runner
	logger   *benchmark.ResultsLogger
	rootPath string
}

// NewBenchmarkMode creates a new benchmark mode
func NewBenchmarkMode() *BenchmarkMode {
	// Get the root path of the project
	rootPath, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		rootPath = "."
	}

	prompter := ui.NewPrompter()
	runner := benchmark.NewRunner(rootPath)
	logger := benchmark.NewResultsLogger(rootPath)

	return &BenchmarkMode{
		prompter: prompter,
		runner:   runner,
		logger:   logger,
		rootPath: rootPath,
	}
}

// Run starts the benchmark mode
func (bm *BenchmarkMode) Run() error {
	bm.prompter.ShowBenchmarkWelcome()

	// Load engine configurations
	if err := bm.runner.LoadEngines(); err != nil {
		return fmt.Errorf("failed to load engine configurations: %w", err)
	}

	// Check dependencies
	if err := bm.checkDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Get benchmark configuration from user
	settings, err := bm.getBenchmarkSettings()
	if err != nil {
		return fmt.Errorf("failed to get benchmark settings: %w", err)
	}

	// Show configuration summary and confirm
	bm.prompter.ShowBenchmarkSummary(
		settings.OpponentName,
		settings.TimeControl.Description,
		settings.GameCount,
		settings.Notes,
		settings.RecordData,
	)

	proceed, err := bm.prompter.PromptForConfirmation("Proceed with benchmark?", false)
	if err != nil {
		return err
	}

	if !proceed {
		fmt.Println("Benchmark cancelled.")
		return nil
	}

	// Run the benchmark
	bm.prompter.ShowBenchmarkProgress("Starting benchmark...")
	result, err := bm.runner.RunBenchmark(settings)
	if err != nil {
		bm.prompter.ShowError(err)
		if result != nil && result.IllegalMoves {
			bm.showIllegalMoveInfo(result)
		}
		return err
	}

	// Display and log results
	bm.logger.DisplayResults(result)

	if err := bm.logger.LogResults(result); err != nil {
		bm.prompter.ShowError(fmt.Errorf("failed to log results: %w", err))
	}

	fmt.Println("\nBenchmark completed successfully!")
	return nil
}

// getBenchmarkSettings prompts the user for all benchmark configuration
func (bm *BenchmarkMode) getBenchmarkSettings() (benchmark.Settings, error) {
	var settings benchmark.Settings

	// Select opponent
	opponent, err := bm.selectOpponent()
	if err != nil {
		return settings, err
	}
	settings.OpponentName = opponent.Name
	settings.Opponent = opponent

	// Select time control
	timeControl, err := bm.selectTimeControl()
	if err != nil {
		return settings, err
	}
	settings.TimeControl = timeControl

	// Select game count
	gameCount, err := bm.selectGameCount()
	if err != nil {
		return settings, err
	}
	settings.GameCount = gameCount

	// Get notes
	notes, err := bm.prompter.PromptForText("Enter notes about this benchmark", "-")
	if err != nil {
		return settings, err
	}
	settings.Notes = notes

	// Ask about recording data
	recordData, err := bm.prompter.PromptForConfirmation("Save results to benchmark history?", true)
	if err != nil {
		return settings, err
	}
	settings.RecordData = recordData

	return settings, nil
}

// selectOpponent prompts the user to select an opponent engine
func (bm *BenchmarkMode) selectOpponent() (benchmark.Engine, error) {
	engines := bm.runner.GetEngineManager().GetOpponentEngines()
	if len(engines) == 0 {
		return benchmark.Engine{}, fmt.Errorf("no opponent engines available")
	}

	options := make([]string, len(engines))
	for i, engine := range engines {
		options[i] = engine.Name
	}

	choice, err := bm.prompter.PromptForChoice("Select opponent engine:", options)
	if err != nil {
		return benchmark.Engine{}, err
	}

	return engines[choice], nil
}

// selectTimeControl prompts the user to select a time control
func (bm *BenchmarkMode) selectTimeControl() (benchmark.TimeControl, error) {
	timeControls := benchmark.GetAvailableTimeControls()
	options := make([]string, len(timeControls)+1)

	for i, tc := range timeControls {
		options[i] = tc.Description
	}
	options[len(timeControls)] = "Custom"

	choice, err := bm.prompter.PromptForChoice("Select time control:", options)
	if err != nil {
		return benchmark.TimeControl{}, err
	}

	if choice == len(timeControls) {
		// Custom time control
		customValue, err := bm.prompter.PromptForText("Enter custom time control (e.g., 5+3, 120, st=30)", "")
		if err != nil {
			return benchmark.TimeControl{}, err
		}

		return benchmark.TimeControl{
			Name:        "Custom",
			Value:       customValue,
			Description: fmt.Sprintf("Custom (%s)", customValue),
		}, nil
	}

	return timeControls[choice], nil
}

// selectGameCount prompts the user to select the number of games
func (bm *BenchmarkMode) selectGameCount() (int, error) {
	gameCounts := benchmark.GetAvailableGameCounts()
	options := make([]string, len(gameCounts)+1)

	for i, gc := range gameCounts {
		options[i] = fmt.Sprintf("%s (%d games)", gc.Name, gc.Count)
	}
	options[len(gameCounts)] = "Custom"

	choice, err := bm.prompter.PromptForChoice("Select number of games:", options)
	if err != nil {
		return 0, err
	}

	if choice == len(gameCounts) {
		// Custom game count
		return bm.prompter.PromptForNumber("Enter number of games", 1, 1000)
	}

	return gameCounts[choice].Count, nil
}

// checkDependencies verifies that required dependencies are available
func (bm *BenchmarkMode) checkDependencies() error {
	cutechessPath := filepath.Join(bm.rootPath, "tools", "engines", "cutechess-cli")
	if _, err := os.Stat(cutechessPath); os.IsNotExist(err) {
		return fmt.Errorf("cutechess-cli not found at %s", cutechessPath)
	}

	// Check if cutechess-cli is executable
	info, err := os.Stat(cutechessPath)
	if err != nil {
		return fmt.Errorf("failed to stat cutechess-cli: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("cutechess-cli is not executable at %s", cutechessPath)
	}

	return nil
}

// showIllegalMoveInfo displays information about illegal moves detected
func (bm *BenchmarkMode) showIllegalMoveInfo(result *benchmark.Result) {
	fmt.Println("\n🚨 ILLEGAL MOVE DETECTED! 🚨")
	fmt.Println("================================")
	fmt.Println("The benchmark was stopped because ChessEngine generated an illegal move.")
	fmt.Println()
	fmt.Println("Debug information:")
	fmt.Printf("  - PGN file: %s\n", result.PGNFile)
	fmt.Printf("  - Timestamp: %s\n", result.Timestamp)
	fmt.Println()
	fmt.Println("Please check the logs and PGN file to investigate the issue.")
	fmt.Println("Look for recent UCI debug logs in /tmp/ and game engine logs.")
}
