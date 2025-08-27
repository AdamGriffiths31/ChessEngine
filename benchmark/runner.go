package benchmark

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Runner executes benchmark matches using cutechess-cli.
type Runner struct {
	engineManager *EngineManager
	rootPath      string
	resultsDir    string
}

// NewRunner creates a new benchmark runner.
func NewRunner(rootPath string) *Runner {
	return &Runner{
		engineManager: NewEngineManager(rootPath),
		rootPath:      rootPath,
		resultsDir:    filepath.Join(rootPath, "tools", "results"),
	}
}

// LoadEngines loads the engine configurations.
func (r *Runner) LoadEngines() error {
	return r.engineManager.LoadEngines()
}

// GetEngineManager returns the engine manager.
func (r *Runner) GetEngineManager() *EngineManager {
	return r.engineManager
}

// RunBenchmark executes a benchmark with the given settings.
func (r *Runner) RunBenchmark(settings Settings) (*Result, error) {
	if err := r.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	if err := r.buildChessEngine(); err != nil {
		return nil, fmt.Errorf("failed to build ChessEngine: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	pgnFile := filepath.Join(r.resultsDir, fmt.Sprintf("benchmark_%s.pgn", timestamp))

	chessEngine, err := r.engineManager.GetChessEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to get ChessEngine config: %w", err)
	}

	cmd, err := r.buildCutechessCommand(*chessEngine, settings.Opponent, settings, pgnFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build cutechess command: %w", err)
	}

	startTime := time.Now()
	output, err := r.runCutechessCommand(cmd, timestamp)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("cutechess-cli failed: %w", err)
	}

	result, err := r.analyzeResults(pgnFile, settings, timestamp, duration, output)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze results: %w", err)
	}

	if result.IllegalMoves {
		return result, fmt.Errorf("benchmark stopped due to illegal moves detected")
	}

	return result, nil
}

// ensureDirectories creates necessary directories.
func (r *Runner) ensureDirectories() error {
	return os.MkdirAll(r.resultsDir, 0750)
}

// buildChessEngine builds the ChessEngine UCI binary.
func (r *Runner) buildChessEngine() error {
	cmd := exec.Command("go", "build", "-o", "tools/bin/uci", "cmd/uci/main.go")
	cmd.Dir = r.rootPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}

	return nil
}

// buildCutechessCommand constructs the cutechess-cli command.
func (r *Runner) buildCutechessCommand(chessEngine, opponent Engine, settings Settings, pgnFile string) (*exec.Cmd, error) {
	cutechessPath := filepath.Join(r.rootPath, "tools", "engines", "cutechess-cli")

	args := []string{
		"-engine", fmt.Sprintf("cmd=%s", chessEngine.Command), "name=ChessEngine", "proto=uci", "option.Hash=512",
		"-engine", fmt.Sprintf("cmd=%s", opponent.Command), fmt.Sprintf("name=%s", opponent.Name), "proto=uci",
	}

	// Add opponent options
	opponentOptions := r.engineManager.FormatEngineOptions(&opponent)
	if len(opponentOptions) > 0 {
		args = append(args, opponentOptions...)
	}

	// Add time control
	if strings.HasPrefix(settings.TimeControl.Value, "st=") {
		args = append(args, fmt.Sprintf("-%s", settings.TimeControl.Value))
	} else {
		args = append(args, "-each", fmt.Sprintf("tc=%s", settings.TimeControl.Value))
	}

	// Add remaining parameters
	args = append(args,
		"-games", strconv.Itoa(settings.GameCount),
		"-concurrency", "1",
		"-ratinginterval", "1",
		"-outcomeinterval", "1",
		"-event", "ChessEngine Benchmark",
		"-site", "Local Testing",
		"-pgnout", pgnFile,
	)

	cmd := exec.Command(cutechessPath, args...) // #nosec G204
	cmd.Dir = r.rootPath

	return cmd, nil
}

// runCutechessCommand executes the cutechess-cli command.
func (r *Runner) runCutechessCommand(cmd *exec.Cmd, timestamp string) (string, error) {
	outputFile := filepath.Join(r.resultsDir, fmt.Sprintf("cutechess_output_%s.log", timestamp))

	output, err := cmd.CombinedOutput()

	// Save output to file for analysis
	if writeErr := os.WriteFile(outputFile, output, 0600); writeErr != nil {
		// Continue even if we can't write the log file
		fmt.Printf("Warning: Failed to write cutechess output to %s: %v\n", outputFile, writeErr)
	}

	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// analyzeResults parses the PGN file and cutechess output to generate results.
func (r *Runner) analyzeResults(pgnFile string, settings Settings, timestamp string, duration time.Duration, output string) (*Result, error) {
	result := &Result{
		Settings:  settings,
		Duration:  duration,
		PGNFile:   pgnFile,
		Timestamp: timestamp,
	}

	// Check for illegal moves first
	if r.checkForIllegalMoves(output, pgnFile) {
		result.IllegalMoves = true
		return result, nil
	}

	// Parse PGN file for results
	if err := r.parsePGNResults(pgnFile, result); err != nil {
		return nil, err
	}

	return result, nil
}

// checkForIllegalMoves checks output and PGN for illegal move indicators.
func (r *Runner) checkForIllegalMoves(output, pgnFile string) bool {
	// Check cutechess output for illegal move messages
	if strings.Contains(output, "illegal move") || strings.Contains(output, "makes an illegal move") {
		return true
	}

	// Check PGN file content if it exists
	// #nosec G304
	if data, err := os.ReadFile(pgnFile); err == nil {
		if strings.Contains(string(data), "illegal move") {
			return true
		}
	}

	return false
}

// parsePGNResults analyzes the PGN file to extract game results.
func (r *Runner) parsePGNResults(pgnFile string, result *Result) error {
	data, err := os.ReadFile(pgnFile) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read PGN file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var isChessEngineWhite, isChessEngineBlack bool

	// Regular expressions for parsing
	whiteRegex := regexp.MustCompile(`^\[White "(.*)"\]`)
	blackRegex := regexp.MustCompile(`^\[Black "(.*)"\]`)
	resultRegex := regexp.MustCompile(`^\[Result "(.*)"\]`)

	for _, line := range lines {
		if matches := whiteRegex.FindStringSubmatch(line); len(matches) > 1 {
			isChessEngineWhite = matches[1] == "ChessEngine"
			isChessEngineBlack = false
		} else if matches := blackRegex.FindStringSubmatch(line); len(matches) > 1 {
			isChessEngineBlack = matches[1] == "ChessEngine"
			if !isChessEngineWhite {
				isChessEngineWhite = false
			}
		} else if matches := resultRegex.FindStringSubmatch(line); len(matches) > 1 {
			gameResult := matches[1]
			result.TotalGames++

			if isChessEngineWhite {
				switch gameResult {
				case "1-0":
					result.Wins++
				case "0-1":
					result.Losses++
				case "1/2-1/2":
					result.Draws++
				}
			} else if isChessEngineBlack {
				switch gameResult {
				case "0-1":
					result.Wins++
				case "1-0":
					result.Losses++
				case "1/2-1/2":
					result.Draws++
				}
			}

			// Reset flags after processing result
			isChessEngineWhite = false
			isChessEngineBlack = false
		}
	}

	// Calculate score percentage
	if result.TotalGames > 0 {
		points := result.Wins*2 + result.Draws
		maxPoints := result.TotalGames * 2
		result.Score = (points * 100) / maxPoints
	}

	return nil
}
