package trainingdata

import (
	"bytes"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/nnue"
)

var testLines = []string{
	"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1 | 31 | 0.5",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1 | -158 | 0.0",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1 | 2999 | 1.0",
	"8/8/8/8/8/8/8/K6k b - - 0 1 | -40000 | 0.0",
}

func TestParseLine_KeepsEveryField(t *testing.T) {
	for _, line := range testLines {
		r, err := parseLine(line)
		if err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		parts := strings.Split(line, " | ")
		fields := strings.Fields(parts[0])

		fen := r.Board().ToFEN()
		if got := strings.Fields(fen); got[0] != fields[0] || got[1] != fields[1] {
			t.Errorf("%q: board FEN %q, want placement %q and side %q", line, fen, fields[0], fields[1])
		}
		if r.BlackToMove() != (fields[1] == "b") {
			t.Errorf("%q: BlackToMove %v", line, r.BlackToMove())
		}
		wantScore := map[string]int{"31": 31, "-158": -158, "2999": 2999, "-40000": -32768}[parts[1]]
		if r.Score() != wantScore {
			t.Errorf("%q: score %d, want %d", line, r.Score(), wantScore)
		}
		wantResult := map[string]float64{"0.0": 0, "0.5": 0.5, "1.0": 1}[parts[2]]
		if r.Result() != wantResult {
			t.Errorf("%q: result %v, want %v", line, r.Result(), wantResult)
		}
	}
}

func TestAppendFeatures_MatchesTheBoardPieces(t *testing.T) {
	for _, line := range testLines {
		r, err := parseLine(line)
		if err != nil {
			t.Fatal(err)
		}
		b, err := board.FromFEN(strings.Split(line, " | ")[0])
		if err != nil {
			t.Fatal(err)
		}
		var want []uint16
		for piece, bb := range b.PieceBitboards {
			for rest := uint64(bb); rest != 0; rest &= rest - 1 {
				want = append(want, uint16(nnue.FeatureIndex(piece, bits.TrailingZeros64(rest), 0)))
			}
		}
		got := r.AppendFeatures(nil)
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%q: features %v, want %v", line, got, want)
		}
	}
}

func TestParseLine_RejectsMalformedLines(t *testing.T) {
	for _, line := range []string{
		"",
		"8/8/8/8/8/8/8/K6k w - - 0 1 | 12",
		"8/8/8/8/8/8/8/K6k w - - 0 1 | x | 0.5",
		"8/8/8/8/8/8/8/K6k w - - 0 1 | 12 | 0.7",
		"8/8/8/8/8/8/8/K6x w - - 0 1 | 12 | 0.5",
		"8/8/8/8/8/8/8/K6k | 12 | 0.5",
	} {
		if _, err := parseLine(line); err == nil {
			t.Errorf("%q parsed, want an error", line)
		}
	}
}

func TestLoad_PackedAndTextFilesGiveTheSameRecords(t *testing.T) {
	dir := t.TempDir()
	text := strings.Join(testLines, "\n") + "\nnot a position\n" + "8/8/8/8/8/8/8/K6k w - - 0 1 | 12"
	textPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(textPath, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	var packed bytes.Buffer
	w, err := NewWriter(&packed)
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := ScanText(strings.NewReader(text), w.Write)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if skipped != 2 {
		t.Errorf("ScanText skipped %d lines, want the bad line and the cut-off last line", skipped)
	}
	packedPath := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(packedPath, packed.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	fromText, textSkipped, err := Load(textPath)
	if err != nil {
		t.Fatal(err)
	}
	fromPacked, packedSkipped, err := Load(packedPath)
	if err != nil {
		t.Fatal(err)
	}
	if fromText.Len() != len(testLines) || textSkipped != 2 || packedSkipped != 0 {
		t.Errorf("text: %d records, %d skipped; packed: %d skipped", fromText.Len(), textSkipped, packedSkipped)
	}
	if !bytes.Equal(fromText, fromPacked) {
		t.Error("records loaded from the text and packed files differ")
	}
}

func TestLoad_RejectsATruncatedPackedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	content := append(magic[:], make([]byte, RecordSize+5)...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path); err == nil {
		t.Error("loaded a packed file with a partial record, want an error")
	}
}
