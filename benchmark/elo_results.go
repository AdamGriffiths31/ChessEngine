package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EloResultsLogger handles logging Elo benchmark results to markdown and
// JSONL files.
type EloResultsLogger struct {
	rootPath string
	mdFile   string
	jsonFile string
}

// NewEloResultsLogger creates a new Elo results logger.
func NewEloResultsLogger(rootPath string) *EloResultsLogger {
	return &EloResultsLogger{
		rootPath: rootPath,
		mdFile:   filepath.Join(rootPath, "elo.md"),
		jsonFile: filepath.Join(rootPath, "elo_results.jsonl"),
	}
}

// LogResults appends result to elo.md (a markdown summary table) and
// elo_results.jsonl (one JSON object per run with full per-game detail).
func (rl *EloResultsLogger) LogResults(result *EloResult) error {
	if err := rl.ensureMarkdownFile(); err != nil {
		return fmt.Errorf("failed to create elo.md: %w", err)
	}

	mdFile, err := os.OpenFile(rl.mdFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open elo.md: %w", err)
	}
	defer func() { _ = mdFile.Close() }()
	if _, err := mdFile.WriteString(rl.formatMarkdownEntry(result) + "\n"); err != nil {
		return fmt.Errorf("failed to write to elo.md: %w", err)
	}

	jsonLine, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal elo result: %w", err)
	}
	jsonFile, err := os.OpenFile(rl.jsonFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open elo_results.jsonl: %w", err)
	}
	defer func() { _ = jsonFile.Close() }()
	if _, err := jsonFile.Write(append(jsonLine, '\n')); err != nil {
		return fmt.Errorf("failed to write to elo_results.jsonl: %w", err)
	}

	return nil
}

func (rl *EloResultsLogger) ensureMarkdownFile() error {
	if _, err := os.Stat(rl.mdFile); os.IsNotExist(err) {
		content := `# Elo Benchmark Results

This file tracks ChessEngine's estimated playing strength over time.

| Date | Elo | Score | Anchors | TC | Games | Avg Depth | Nodes |
|------|-----|-------|---------|-------|-------|-----------|-------|
`
		if err := os.WriteFile(rl.mdFile, []byte(content), 0600); err != nil {
			return err
		}
	}
	return nil
}

func (rl *EloResultsLogger) formatMarkdownEntry(result *EloResult) string {
	totalScore := 0.0
	var totalNodes uint64
	totalDepth, searchedGames := 0, 0
	for _, g := range result.Games {
		totalScore += g.Result
		totalNodes += g.Nodes
		if g.AvgDepth > 0 {
			totalDepth += g.AvgDepth
			searchedGames++
		}
	}
	avgDepth := 0.0
	if searchedGames > 0 {
		avgDepth = float64(totalDepth) / float64(searchedGames)
	}

	anchorStrs := make([]string, len(result.Anchors))
	for i, a := range result.Anchors {
		anchorStrs[i] = fmt.Sprintf("%d", a)
	}

	tc := fmt.Sprintf("%d+%d", result.Config.TimeMs/1000, result.Config.IncMs/1000)

	return fmt.Sprintf("| %s | %d | %s/%d | %s | %s | %d | %.1f | %d |",
		result.Timestamp,
		int(result.EstimatedElo+0.5),
		formatScore(totalScore), len(result.Games),
		strings.Join(anchorStrs, "/"),
		tc,
		len(result.Games),
		avgDepth,
		totalNodes,
	)
}

func formatScore(score float64) string {
	if score == float64(int(score)) {
		return fmt.Sprintf("%d", int(score))
	}
	return fmt.Sprintf("%.1f", score)
}

// DisplayResults prints the Elo benchmark result to the console.
func (rl *EloResultsLogger) DisplayResults(result *EloResult) {
	totalScore := 0.0
	for _, g := range result.Games {
		totalScore += g.Result
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Elo Benchmark Results")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Estimated Elo: %d\n", int(result.EstimatedElo+0.5))
	fmt.Printf("Score: %s/%d\n", formatScore(totalScore), len(result.Games))
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Second))
	fmt.Printf("Results logged to: %s and %s\n", rl.mdFile, rl.jsonFile)
}
