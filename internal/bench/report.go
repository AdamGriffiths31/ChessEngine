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

// GenerateReport regenerates the STS and Elo history tables from the JSONL
// result files under resultsDir (tools/results by default), replacing the
// hand-maintained sts_history.md/elo.md tables with data derived straight
// from tools/results/sts_*.jsonl and tools/results/elo_results.jsonl.
//
// Only JSONL-era runs are reproduced: historical rows that predate per-run
// JSONL output are not present in the source files and so cannot appear
// here.
func GenerateReport(resultsDir string) (string, error) {
	stsTable, err := GenerateSTSHistoryTable(resultsDir)
	if err != nil {
		return "", err
	}
	eloTable, err := GenerateEloHistoryTable(resultsDir)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# ChessEngine Benchmark History (regenerated)\n\n")
	fmt.Fprintf(&b, "Regenerated from %s at %s.\n\n", resultsDir, time.Now().Format("2006-01-02 15:04"))
	b.WriteString(stsTable)
	b.WriteString("\n")
	b.WriteString(eloTable)
	return b.String(), nil
}

// GenerateSTSHistoryTable builds a markdown table (one row per file) from
// the trailing summary line of every tools/results/sts_*.jsonl file under
// resultsDir, in the same column layout as the hand-maintained
// tools/results/sts_history.md.
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
		summary, err := readSTSSummary(f)
		if err != nil {
			slog.Warn("bench report: skipping STS results file", "path", f, "error", err)
			continue
		}
		b.WriteString(formatSTSHistoryRow(summary))
		b.WriteString("\n")
	}

	return b.String(), nil
}

// readSTSSummary reads path (a sts_<timestamp>.jsonl file) and returns its
// trailing summary record - the last non-empty line, which the STS runner
// always writes after every per-position record.
func readSTSSummary(path string) (STSSummaryRecord, error) {
	data, err := os.ReadFile(path) // #nosec G304 - path comes from filepath.Glob over a fixed results directory
	if err != nil {
		return STSSummaryRecord{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var summary STSSummaryRecord
		if err := json.Unmarshal([]byte(line), &summary); err != nil {
			return STSSummaryRecord{}, fmt.Errorf("failed to parse last line of %s: %w", path, err)
		}
		if summary.Type != "summary" {
			return STSSummaryRecord{}, fmt.Errorf("last line of %s is not a summary record (type=%q)", path, summary.Type)
		}
		return summary, nil
	}
	return STSSummaryRecord{}, fmt.Errorf("%s has no data lines", path)
}

// formatSTSHistoryRow renders one markdown table row from an STS summary
// record, matching the column layout of tools/results/sts_history.md.
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

// GenerateEloHistoryTable builds a markdown table (one row per recorded
// run) from tools/results/elo_results.jsonl, in the same column layout as
// the hand-maintained tools/results/elo.md.
func GenerateEloHistoryTable(resultsDir string) (string, error) {
	path := filepath.Join(resultsDir, "elo_results.jsonl")

	var b strings.Builder
	b.WriteString("## Elo Benchmark Results History\n\n")
	b.WriteString("This table tracks ChessEngine's estimated playing strength over time.\n")
	b.WriteString("Regenerated from tools/results/elo_results.jsonl.\n\n")
	b.WriteString("| Date | Elo | Score | Anchors | TC | Games | Avg Depth | Nodes |\n")
	b.WriteString("|------|-----|-------|---------|-------|-------|-----------|-------|\n")

	data, err := os.ReadFile(path) // #nosec G304 - fixed filename under the results directory
	if err != nil {
		if os.IsNotExist(err) {
			return b.String(), nil
		}
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}

	logger := NewEloResultsLogger("")
	for lineNum, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var result EloResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			slog.Warn("bench report: skipping unparseable elo run", "path", path, "line", lineNum+1, "error", err)
			continue
		}
		b.WriteString(logger.formatMarkdownEntry(&result))
		b.WriteString("\n")
	}

	return b.String(), nil
}

// formatNodeCount renders a nodes-per-second figure compactly (e.g.
// "1.4M", "76.9K"), matching the style used elsewhere for node counts.
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
