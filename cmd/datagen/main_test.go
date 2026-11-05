package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrimAndCountLines_TrimsCutOffTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	whole := "8/8/8/8/8/8/8/K6k w - - 0 1 | 0 | 0.5\n8/8/8/8/8/8/8/K6k b - - 0 1 | 0 | 0.5\n"
	if err := os.WriteFile(path, []byte(whole+"8/8/8/8/8/8/8/K6k w - - 0 1 | 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	lines, err := trimAndCountLines(path)
	if err != nil {
		t.Fatal(err)
	}
	if lines != 2 {
		t.Errorf("lines = %d, want 2", lines)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != whole {
		t.Errorf("file after trim = %q, want only the complete lines", got)
	}
}
