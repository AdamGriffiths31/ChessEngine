package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SPRTResultsLogger logs SPRT run results to markdown and JSONL files,
// structurally identical to EloResultsLogger (see elo_results.go).
type SPRTResultsLogger struct {
	rootPath string
	mdFile   string
	jsonFile string
}

// NewSPRTResultsLogger creates a new SPRT results logger.
func NewSPRTResultsLogger(rootPath string) *SPRTResultsLogger {
	return &SPRTResultsLogger{
		rootPath: rootPath,
		mdFile:   filepath.Join(rootPath, "tools", "results", "sprt_history.md"),
		jsonFile: filepath.Join(rootPath, "tools", "results", "sprt_results.jsonl"),
	}
}

// LogResult appends result to sprt_history.md (a markdown summary table)
// and sprt_results.jsonl (one JSON object per run with full detail).
func (rl *SPRTResultsLogger) LogResult(result *SPRTResult) error {
	if err := rl.ensureMarkdownFile(); err != nil {
		return fmt.Errorf("failed to create sprt_history.md: %w", err)
	}

	mdFile, err := os.OpenFile(rl.mdFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open sprt_history.md: %w", err)
	}
	defer func() { _ = mdFile.Close() }()
	if _, err := mdFile.WriteString(rl.formatMarkdownEntry(result) + "\n"); err != nil {
		return fmt.Errorf("failed to write to sprt_history.md: %w", err)
	}

	jsonLine, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal SPRT result: %w", err)
	}
	jsonFile, err := os.OpenFile(rl.jsonFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open sprt_results.jsonl: %w", err)
	}
	defer func() { _ = jsonFile.Close() }()
	if _, err := jsonFile.Write(append(jsonLine, '\n')); err != nil {
		return fmt.Errorf("failed to write to sprt_results.jsonl: %w", err)
	}

	return nil
}

func (rl *SPRTResultsLogger) ensureMarkdownFile() error {
	if err := os.MkdirAll(filepath.Dir(rl.mdFile), 0750); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}
	if _, err := os.Stat(rl.mdFile); os.IsNotExist(err) {
		content := `# SPRT Regression Test Results

This file tracks SPRT (Sequential Probability Ratio Test) runs comparing
a baseline and candidate ChessEngine build.

| Date | Verdict | LLR | Bounds | W-L-D | Games | TC | Elo0/Elo1 |
|------|---------|-----|--------|-------|-------|-----|-----------|
`
		if err := os.WriteFile(rl.mdFile, []byte(content), 0600); err != nil {
			return err
		}
	}
	return nil
}

func (rl *SPRTResultsLogger) formatMarkdownEntry(result *SPRTResult) string {
	return fmt.Sprintf("| %s | %s | %.2f | [%.2f, %.2f] | %d-%d-%d | %d | %s | %g/%g |",
		result.Timestamp,
		result.Verdict.String(),
		result.LLR,
		result.LBound, result.UBound,
		result.Wins, result.Losses, result.Draws,
		result.GamesTotal,
		result.Config.TimeControl,
		result.Config.Elo0, result.Config.Elo1,
	)
}
