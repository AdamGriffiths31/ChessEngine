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
		Verdict:     SPRTAcceptH1,
		LLR:         2.95,
		LBound:      -2.94,
		UBound:      2.94,
		Wins:        120,
		Losses:      80,
		Draws:       40,
		GamesTotal:  240,
		Pentanomial: [5]int{5, 20, 50, 30, 15},
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

	for _, want := range []string{"H1 accepted", "120-80-40", "[5, 20, 50, 30, 15]", "10+0.1", "2.95"} {
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

func TestSPRTLogResultAddsFastchessTableOnce(t *testing.T) {
	root := t.TempDir()
	mdPath := filepath.Join(root, "tools", "results", "sprt_history.md")
	if err := os.MkdirAll(filepath.Dir(mdPath), 0750); err != nil {
		t.Fatal(err)
	}
	old := "# SPRT Regression Test Results\n\n| Date | Verdict |\n|---|---|\n| 20260818_202050 | old row |\n"
	if err := os.WriteFile(mdPath, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}

	logger := NewSPRTResultsLogger(root)
	for range 2 {
		if err := logger.LogResult(sampleSPRTResult()); err != nil {
			t.Fatalf("LogResult failed: %v", err)
		}
	}

	content, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	if !strings.HasPrefix(got, old) {
		t.Errorf("old history rows were modified:\n%s", got)
	}
	if n := strings.Count(got, fastchessTableHeader); n != 1 {
		t.Errorf("fastchess table header appears %d times, want 1:\n%s", n, got)
	}
	if n := strings.Count(got, "H1 accepted"); n != 2 {
		t.Errorf("got %d result rows, want 2:\n%s", n, got)
	}
}
