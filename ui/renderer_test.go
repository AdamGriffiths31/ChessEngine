package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type TestCase struct {
	Name     string `json:"name"`
	FEN      string `json:"fen"`
	Expected string `json:"expected"`
}

func TestRenderBoardFromFEN_GoldenFile(t *testing.T) {
	goldenFile := filepath.Join("testdata", "golden_boards.json")
	data, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	var testCases []TestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		t.Fatalf("Failed to unmarshal golden file: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := RenderBoardFromFEN(tc.FEN)
			if result != tc.Expected {
				t.Errorf("Test case %s failed.\nExpected:\n%s\n\nGot:\n%s", tc.Name, tc.Expected, result)
			}
		})
	}
}

func TestRenderBoard_NilBoard(t *testing.T) {
	result := RenderBoard(nil)
	expected := "ERROR: Board is nil"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}