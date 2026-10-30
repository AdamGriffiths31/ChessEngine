package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SPRTResultsLogger logs SPRT run results to markdown and JSONL files.
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
	if err := rl.ensureFastchessTable(); err != nil {
		return fmt.Errorf("failed to prepare sprt_history.md: %w", err)
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

const fastchessTableHeader = `## fastchess runs (pentanomial, normalized Elo)

| Date | Verdict | LLR | Bounds | W-L-D | Ptnml(0-2) | Games | TC | nElo0/nElo1 |
|------|---------|-----|--------|-------|------------|-------|-----|-------------|
`

func (rl *SPRTResultsLogger) ensureFastchessTable() error {
	if err := os.MkdirAll(filepath.Dir(rl.mdFile), 0750); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}
	existing, err := os.ReadFile(rl.mdFile)
	if os.IsNotExist(err) {
		content := `# SPRT Regression Test Results

This file tracks SPRT (Sequential Probability Ratio Test) runs comparing
a baseline and candidate ChessEngine build.

` + fastchessTableHeader
		return os.WriteFile(rl.mdFile, []byte(content), 0600)
	}
	if err != nil {
		return err
	}
	if strings.Contains(string(existing), fastchessTableHeader) {
		return nil
	}
	f, err := os.OpenFile(rl.mdFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString("\n" + fastchessTableHeader)
	return err
}

func (rl *SPRTResultsLogger) formatMarkdownEntry(result *SPRTResult) string {
	p := result.Pentanomial
	return fmt.Sprintf("| %s | %s | %.2f | [%.2f, %.2f] | %d-%d-%d | [%d, %d, %d, %d, %d] | %d | %s | %g/%g |",
		result.Timestamp,
		result.Verdict.String(),
		result.LLR,
		result.LBound, result.UBound,
		result.Wins, result.Losses, result.Draws,
		p[0], p[1], p[2], p[3], p[4],
		result.GamesTotal,
		result.Config.TimeControl,
		result.Config.Elo0, result.Config.Elo1,
	)
}
