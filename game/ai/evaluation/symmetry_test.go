package evaluation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// mirrorFEN returns the color-flipped FEN: ranks reversed, piece case
// swapped, side to move flipped, castling rights swapped between colors,
// en-passant rank mirrored (3<->6). Halfmove/fullmove kept as-is.
func mirrorFEN(fen string) string {
	fields := strings.Fields(fen)

	ranks := strings.Split(fields[0], "/")
	mirrored := make([]string, len(ranks))
	for i, rank := range ranks {
		mirrored[len(ranks)-1-i] = swapCase(rank)
	}

	side := "w"
	if fields[1] == "w" {
		side = "b"
	}

	castling := fields[2]
	if castling != "-" {
		castling = swapCase(castling)
		// Normalize to conventional KQkq order
		ordered := ""
		for _, c := range []string{"K", "Q", "k", "q"} {
			if strings.Contains(castling, c) {
				ordered += c
			}
		}
		castling = ordered
	}

	ep := fields[3]
	if ep != "-" {
		ep = string(ep[0]) + string('1'+('8'-ep[1]))
	}

	out := []string{strings.Join(mirrored, "/"), side, castling, ep}
	out = append(out, fields[4:]...)
	return strings.Join(out, " ")
}

func swapCase(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - 'a' + 'A')
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r - 'A' + 'a')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TestMirrorFEN(t *testing.T) {
	tests := []struct{ in, want string }{
		{
			"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1",
			"4k3/4p3/8/8/8/8/8/4K3 b - - 0 1",
		},
		{
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
		},
		{
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			"rnbqkbnr/pppp1ppp/8/4p3/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 2",
		},
		{
			"r3k3/8/8/8/8/8/8/4K2R w Kq - 4 20",
			"4k2r/8/8/8/8/8/8/R3K3 b Qk - 4 20",
		},
	}
	for _, tt := range tests {
		if got := mirrorFEN(tt.in); got != tt.want {
			t.Errorf("mirrorFEN(%q):\n got %q\nwant %q", tt.in, got, tt.want)
		}
	}
	// Mirror must be its own inverse
	for _, tt := range tests {
		if got := mirrorFEN(mirrorFEN(tt.in)); got != tt.in {
			t.Errorf("mirrorFEN not involutive for %q: got %q", tt.in, got)
		}
	}
}

// loadTestFENs reads the FEN prefix (first 4 fields) of every line in the
// STS EPD files. A local parser is used because importing the epd package
// here would create an import cycle (epd -> search -> evaluation).
func loadTestFENs(t *testing.T) []string {
	t.Helper()
	fens := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		// White castled short, Black castled long
		"2kr3r/pppq1ppp/2n2n2/4p3/4P3/2N2N2/PPPQ1PPP/R4RK1 w - - 8 12",
		// Passed pawns for both sides
		"8/2p5/8/1P6/8/5p2/6P1/4K2k w - - 0 40",
		// Fianchetto structures
		"rnbqk2r/ppppppbp/5np1/8/8/5NP1/PPPPPPBP/RNBQK2R w KQkq - 4 4",
		// En passant available
		"rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
	}
	for i := 1; i <= 6; i++ {
		path := filepath.Join("..", "..", "..", "testdata", fmt.Sprintf("STS%d.epd", i))
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		for _, line := range strings.Split(string(content), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			fens = append(fens, strings.Join(fields[:4], " ")+" 0 1")
		}
	}
	return fens
}

func TestEvaluationSymmetry(t *testing.T) {
	evaluator := NewEvaluator()
	failures := 0
	for _, fen := range loadTestFENs(t) {
		b, err := board.FromFEN(fen)
		if err != nil {
			t.Fatalf("bad FEN %q: %v", fen, err)
		}
		m, err := board.FromFEN(mirrorFEN(fen))
		if err != nil {
			t.Fatalf("bad mirrored FEN %q (from %q): %v", mirrorFEN(fen), fen, err)
		}

		orig := evaluator.Evaluate(b)
		mirror := evaluator.Evaluate(m)
		if orig != -mirror {
			failures++
			if failures <= 10 {
				t.Errorf("asymmetric eval: %d vs %d (mirror)\n  fen:    %s\n  mirror: %s",
					orig, mirror, fen, mirrorFEN(fen))
			}
		}
	}
	if failures > 10 {
		t.Errorf("... and %d more asymmetric positions", failures-10)
	}
}
