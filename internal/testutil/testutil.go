// Package testutil provides shared helpers for tests across the engine.
package testutil

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// MustFromFEN parses fen into a Board and fails the test on error.
func MustFromFEN(t *testing.T, fen string) *board.Board {
	t.Helper()
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(%q): %v", fen, err)
	}
	return b
}
