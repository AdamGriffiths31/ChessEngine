package bench

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GenerateReport regenerates the STS history table from the JSONL
// result files under resultsDir.
func GenerateReport(resultsDir string) (string, error) {
	stsTable, err := GenerateSTSHistoryTable(resultsDir)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# ChessEngine Benchmark History (regenerated)\n\n")
	fmt.Fprintf(&b, "Regenerated from %s at %s.\n\n", resultsDir, time.Now().Format("2006-01-02 15:04"))
	b.WriteString(stsTable)
	return b.String(), nil
}

// GenerateSTSHistoryTable builds a markdown table (one row per file) from
// the trailing summary line of every sts_*.jsonl file under resultsDir.
func GenerateSTSHistoryTable(resultsDir string) (string, error) {
	pattern := filepath.Join(resultsDir, "sts_*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob %s: %w", pattern, err)
	}
	sort.Strings(files)

	var b strings.Builder
	b.WriteString("## STS Benchmark Results History\n\n")
	b.WriteString("This table tracks the STS (Strategic Test Suite) performance of ChessEngine over time.\n")
	b.WriteString("Regenerated from tools/results/sts_*.jsonl (JSONL-era runs only; historical rows predating\n")
	b.WriteString("per-run JSONL output are not reproduced here).\n\n")
	b.WriteString("| Date | Commit | EPD File | Positions | Score | Max | Percent | STS Rating | Correct | Depth | Timeout | Avg Time | Total Time | NPS | Avg Depth | Notes |\n")
	b.WriteString("|------|--------|----------|-----------|-------|-----|---------|------------|---------|-------|---------|----------|------------|-----|-----------|-------|\n")

	for _, f := range files {
		summary, err := readSTSSummary(resultsDir, filepath.Base(f))
		if err != nil {
			slog.Warn("bench report: skipping STS results file", "path", f, "error", err)
			continue
		}
		b.WriteString(formatSTSHistoryRow(summary))
		b.WriteString("\n")
	}

	return b.String(), nil
}

func readSTSSummary(resultsDir, name string) (STSSummaryRecord, error) {
	root, err := os.OpenRoot(resultsDir)
	if err != nil {
		return STSSummaryRecord{}, fmt.Errorf("failed to open %s: %w", resultsDir, err)
	}
	defer func() { _ = root.Close() }()

	data, err := root.ReadFile(name)
	if err != nil {
		return STSSummaryRecord{}, fmt.Errorf("failed to read %s: %w", name, err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var summary STSSummaryRecord
		if err := json.Unmarshal([]byte(line), &summary); err != nil {
			return STSSummaryRecord{}, fmt.Errorf("failed to parse last line of %s: %w", name, err)
		}
		if summary.Type != "summary" {
			return STSSummaryRecord{}, fmt.Errorf("last line of %s is not a summary record (type=%q)", name, summary.Type)
		}
		return summary, nil
	}
	return STSSummaryRecord{}, fmt.Errorf("%s has no data lines", name)
}

func formatSTSHistoryRow(s STSSummaryRecord) string {
	avgTime := (time.Duration(s.AvgTimeMS) * time.Millisecond).Round(time.Millisecond)
	totalTime := (time.Duration(s.TotalTimeMS) * time.Millisecond).Round(time.Second)

	return fmt.Sprintf("| %s | %s | %s | %d | %d | %d | %.0f%% | %d | %d | %d | %ds | %v | %v | %s | %.1f | %s |",
		s.Timestamp,
		s.GitCommit,
		s.FileDescription,
		s.PositionCount,
		s.TotalScore,
		s.MaxScore,
		s.ScorePercent,
		s.STSRating,
		s.CorrectMoves,
		s.Depth,
		s.TimeoutSeconds,
		avgTime,
		totalTime,
		formatNodeCount(s.NodesPerSecond),
		s.AverageDepth,
		s.Notes,
	)
}

func formatNodeCount(nodes float64) string {
	switch {
	case nodes >= 1_000_000:
		return fmt.Sprintf("%.1fM", nodes/1_000_000)
	case nodes >= 1_000:
		return fmt.Sprintf("%.1fK", nodes/1_000)
	default:
		return fmt.Sprintf("%.0f", nodes)
	}
}
