package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ResultsLogger handles logging benchmark results to markdown files.
type ResultsLogger struct {
	rootPath string
	logFile  string
}

// NewResultsLogger creates a new results logger.
func NewResultsLogger(rootPath string) *ResultsLogger {
	return &ResultsLogger{
		rootPath: rootPath,
		logFile:  filepath.Join(rootPath, "history.md"),
	}
}

// LogResults writes benchmark results to the markdown history file.
func (rl *ResultsLogger) LogResults(result *Result) error {
	if !result.Settings.RecordData {
		return nil
	}

	if err := rl.ensureHistoryFile(); err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}

	entry := rl.formatResultEntry(result)

	file, err := os.OpenFile(rl.logFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("failed to write to history file: %w", err)
	}

	return nil
}

// ensureHistoryFile creates the markdown history file if it doesn't exist.
func (rl *ResultsLogger) ensureHistoryFile() error {
	if _, err := os.Stat(rl.logFile); os.IsNotExist(err) {
		content := `# ChessEngine Benchmark History

This file tracks the performance of ChessEngine against various opponents over time.

## Results Summary

| Date | Opponent | Time Control | Games | Wins | Losses | Draws | Score | Notes |
|------|----------|--------------|-------|------|--------|-------|-------|-------|
`
		if err := os.WriteFile(rl.logFile, []byte(content), 0600); err != nil {
			return err
		}
	}

	return nil
}

// formatResultEntry formats a benchmark result as a markdown table entry.
func (rl *ResultsLogger) formatResultEntry(result *Result) string {
	// Extract date and time from timestamp (format: YYYYMMDD_HHMMSS)
	timestamp := result.Timestamp
	if len(timestamp) >= 15 {
		datePart := timestamp[:8]
		timePart := timestamp[9:15]

		date := fmt.Sprintf("%s-%s-%s", datePart[:4], datePart[4:6], datePart[6:8])
		time := fmt.Sprintf("%s:%s", timePart[:2], timePart[2:4])

		return fmt.Sprintf("| %s %s | %s | %s | %d | %d | %d | %d | %d%% | %s |",
			date, time,
			result.Settings.OpponentName,
			result.Settings.TimeControl.Description,
			result.TotalGames,
			result.Wins,
			result.Losses,
			result.Draws,
			result.Score,
			result.Settings.Notes,
		)
	}

	return fmt.Sprintf("| %s | %s | %s | %d | %d | %d | %d | %d%% | %s |",
		result.Timestamp,
		result.Settings.OpponentName,
		result.Settings.TimeControl.Description,
		result.TotalGames,
		result.Wins,
		result.Losses,
		result.Draws,
		result.Score,
		result.Settings.Notes,
	)
}

// DisplayResults prints benchmark results to the console.
func (rl *ResultsLogger) DisplayResults(result *Result) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Benchmark Results")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Opponent: %s\n", result.Settings.OpponentName)
	fmt.Printf("Time Control: %s\n", result.Settings.TimeControl.Description)
	fmt.Printf("Total Games: %d\n", result.TotalGames)
	fmt.Printf("Wins: %d\n", result.Wins)
	fmt.Printf("Losses: %d\n", result.Losses)
	fmt.Printf("Draws: %d\n", result.Draws)
	fmt.Printf("Score: %d%%\n", result.Score)
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Second))
	fmt.Printf("PGN File: %s\n", result.PGNFile)

	if result.Settings.RecordData {
		fmt.Printf("Results logged to: %s\n", rl.logFile)
	} else {
		fmt.Println("Results not logged to file (as requested)")
	}
}
