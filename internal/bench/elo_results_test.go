package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleEloResult() *EloResult {
	return &EloResult{
		EstimatedElo: 1542.6,
		Anchors:      []int{1400, 1500, 1600},
		Games: []EloGameResult{
			{AnchorElo: 1400, Color: "white", Result: 1.0, Plies: 40, Nodes: 1000, AvgDepth: 10},
			{AnchorElo: 1500, Color: "black", Result: 0.5, Plies: 60, Nodes: 2000, AvgDepth: 12},
			{AnchorElo: 1600, Color: "white", Result: 0.0, Plies: 30, Nodes: 1500, AvgDepth: 11},
		},
		Config:    EloConfig{TimeMs: 60000, IncMs: 1000},
		Duration:  2 * time.Minute,
		Timestamp: "20260630_120000",
	}
}

func TestFormatMarkdownEntry(t *testing.T) {
	logger := NewEloResultsLogger(t.TempDir())
	entry := logger.formatMarkdownEntry(sampleEloResult())

	for _, want := range []string{"1543", "1.5/3", "1400/1500/1600", "60+1"} {
		if !strings.Contains(entry, want) {
			t.Errorf("expected markdown entry to contain %q, got: %s", want, entry)
		}
	}
}

func TestLogResultsWritesBothFiles(t *testing.T) {
	root := t.TempDir()
	logger := NewEloResultsLogger(root)

	if err := logger.LogResults(sampleEloResult()); err != nil {
		t.Fatalf("LogResults failed: %v", err)
	}

	mdContent, err := os.ReadFile(filepath.Join(root, "tools", "results", "elo.md"))
	if err != nil {
		t.Fatalf("failed to read elo.md: %v", err)
	}
	if !strings.Contains(string(mdContent), "1543") {
		t.Errorf("expected elo.md to contain the estimated Elo, got: %s", mdContent)
	}

	jsonContent, err := os.ReadFile(filepath.Join(root, "tools", "results", "elo_results.jsonl"))
	if err != nil {
		t.Fatalf("failed to read elo_results.jsonl: %v", err)
	}
	if !strings.Contains(string(jsonContent), "\"AnchorElo\":1400") {
		t.Errorf("expected elo_results.jsonl to contain game detail, got: %s", jsonContent)
	}
}
