package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleSPRTResult() *SPRTResult {
	return &SPRTResult{
		Verdict:    SPRTAcceptH1,
		LLR:        2.95,
		LBound:     -2.94,
		UBound:     2.94,
		Wins:       120,
		Losses:     80,
		Draws:      40,
		GamesTotal: 240,
		Config: SPRTConfig{
			BasePath:    "/bin/base-uci",
			DevPath:     "/bin/dev-uci",
			Elo0:        0,
			Elo1:        5,
			Alpha:       0.05,
			Beta:        0.05,
			TimeControl: "10+0.1",
		},
		Timestamp: "20260818_120000",
		Duration:  10 * time.Minute,
	}
}

func TestSPRTFormatMarkdownEntry(t *testing.T) {
	logger := NewSPRTResultsLogger(t.TempDir())
	entry := logger.formatMarkdownEntry(sampleSPRTResult())

	for _, want := range []string{"H1 accepted", "120-80-40", "10+0.1", "2.95"} {
		if !strings.Contains(entry, want) {
			t.Errorf("expected markdown entry to contain %q, got: %s", want, entry)
		}
	}
}

func TestSPRTLogResultWritesBothFiles(t *testing.T) {
	root := t.TempDir()
	logger := NewSPRTResultsLogger(root)

	if err := logger.LogResult(sampleSPRTResult()); err != nil {
		t.Fatalf("LogResult failed: %v", err)
	}

	mdContent, err := os.ReadFile(filepath.Join(root, "tools", "results", "sprt_history.md"))
	if err != nil {
		t.Fatalf("failed to read sprt_history.md: %v", err)
	}
	if !strings.Contains(string(mdContent), "H1 accepted") {
		t.Errorf("expected sprt_history.md to contain the verdict, got: %s", mdContent)
	}

	jsonContent, err := os.ReadFile(filepath.Join(root, "tools", "results", "sprt_results.jsonl"))
	if err != nil {
		t.Fatalf("failed to read sprt_results.jsonl: %v", err)
	}
	if !strings.Contains(string(jsonContent), "\"Wins\":120") {
		t.Errorf("expected sprt_results.jsonl to contain game detail, got: %s", jsonContent)
	}
}
