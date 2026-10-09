package book

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	startFEN       = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	performanceBin = "testdata/performance.bin"
)

func startBoard(t testing.TB) *board.Board {
	t.Helper()
	b, err := board.FromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func loadBook(t testing.TB, path string) *PolyglotBook {
	t.Helper()
	pb := NewPolyglotBook()
	if err := pb.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile(%s): %v", path, err)
	}
	return pb
}

// writeRawBook writes entries as 16-byte big-endian Polyglot records in the
// given order, followed by any trailing bytes, so tests can build both valid
// and malformed book files.
func writeRawBook(t *testing.T, name string, entries []PolyglotEntry, trailing []byte) string {
	t.Helper()
	var buf bytes.Buffer
	for _, e := range entries {
		if err := binary.Write(&buf, binary.BigEndian, e); err != nil {
			t.Fatal(err)
		}
	}
	buf.Write(trailing)

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// startMoves are the start-position book entries used by these tests: e2e4
// (weight 100), d2d4 (80) and g1f3 (60).
func startMoves(hash uint64) []PolyglotEntry {
	return []PolyglotEntry{
		{Hash: hash, Move: polyglotEncode(4, 1, 4, 3), Weight: 100},
		{Hash: hash, Move: polyglotEncode(3, 1, 3, 3), Weight: 80},
		{Hash: hash, Move: polyglotEncode(6, 0, 5, 2), Weight: 60},
	}
}

// startBook writes a valid book holding the three start-position moves plus
// one unrelated entry (hash 1, which sorts first).
func startBook(t *testing.T) string {
	t.Helper()
	hash := GetPolyglotHash().HashPosition(startBoard(t))
	entries := append([]PolyglotEntry{{Hash: 1, Move: polyglotEncode(4, 1, 4, 2)}}, startMoves(hash)...)
	return writeRawBook(t, "start.bin", entries, nil)
}

func TestPolyglotBook_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantEntries int
	}{
		{"empty", writeRawBook(t, "empty.bin", nil, nil), 0},
		{"single_entry", writeRawBook(t, "single.bin", []PolyglotEntry{{Hash: 1, Move: 1}}, nil), 1},
		{"start_position_book", startBook(t), 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := loadBook(t, tt.path)
			if !pb.IsLoaded() {
				t.Error("IsLoaded() = false after a successful load")
			}
			if got := pb.GetBookInfo().EntryCount; got != tt.wantEntries {
				t.Errorf("EntryCount = %d, want %d", got, tt.wantEntries)
			}
		})
	}

	t.Run("real_book_has_entries", func(t *testing.T) {
		if _, err := os.Stat(performanceBin); os.IsNotExist(err) {
			t.Skipf("%s not found", performanceBin)
		}
		if got := loadBook(t, performanceBin).GetBookInfo().EntryCount; got == 0 {
			t.Error("performance.bin loaded with no entries")
		}
	})
}

func TestPolyglotBook_LoadFromFileRejectsBadFiles(t *testing.T) {
	sorted := []PolyglotEntry{{Hash: 1, Move: 1}, {Hash: 2, Move: 1}}
	unsorted := []PolyglotEntry{{Hash: 2, Move: 1}, {Hash: 1, Move: 1}}

	tests := []struct {
		name    string
		path    string
		wantErr error // nil means any error is acceptable
	}{
		{"missing_file", "testdata/does-not-exist.bin", nil},
		{"size_not_multiple_of_16", writeRawBook(t, "ragged.bin", sorted, []byte{0xFF}), ErrInvalidBookFile},
		{"entries_not_sorted_by_hash", writeRawBook(t, "unsorted.bin", unsorted, nil), ErrInvalidBookFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPolyglotBook()
			err := pb.LoadFromFile(tt.path)
			if err == nil {
				t.Fatal("LoadFromFile succeeded, want an error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if pb.IsLoaded() {
				t.Error("IsLoaded() = true after a failed load")
			}
		})
	}
}

// Looks up the start position and checks the moves decoded from the raw entries
// (e2e4, d2d4, g1f3) come back in file order with their weights.
func TestPolyglotBook_LookupMoveFromFile(t *testing.T) {
	pb := loadBook(t, startBook(t))
	b := startBoard(t)
	hash := GetPolyglotHash().HashPosition(b)

	moves, err := pb.LookupMove(hash, b)
	if err != nil {
		t.Fatalf("LookupMove: %v", err)
	}

	want := []struct {
		from, to string
		weight   uint16
	}{
		{"e2", "e4", 100},
		{"d2", "d4", 80},
		{"g1", "f3", 60},
	}
	if len(moves) != len(want) {
		t.Fatalf("got %d moves, want %d", len(moves), len(want))
	}
	for i, w := range want {
		got := moves[i]
		if got.Move.From.String() != w.from || got.Move.To.String() != w.to || got.Weight != w.weight {
			t.Errorf("move %d = %s%s w%d, want %s%s w%d", i,
				got.Move.From.String(), got.Move.To.String(), got.Weight, w.from, w.to, w.weight)
		}
	}

	if _, err := pb.LookupMove(hash+1, b); !errors.Is(err, ErrPositionNotFound) {
		t.Errorf("unknown hash: error = %v, want ErrPositionNotFound", err)
	}
}

func TestPolyglotBook_LookupMoveBeforeLoad(t *testing.T) {
	_, err := NewPolyglotBook().LookupMove(1, startBoard(t))
	if !errors.Is(err, ErrBookNotLoaded) {
		t.Errorf("error = %v, want ErrBookNotLoaded", err)
	}
}

func BenchmarkRealBookLookup(b *testing.B) {
	if _, err := os.Stat(performanceBin); os.IsNotExist(err) {
		b.Skipf("%s not found", performanceBin)
	}
	pb := loadBook(b, performanceBin)
	bd := startBoard(b)
	hash := GetPolyglotHash().HashPosition(bd)

	b.ResetTimer()
	for range b.N {
		if _, err := pb.LookupMove(hash, bd); err != nil && !errors.Is(err, ErrPositionNotFound) {
			b.Fatalf("LookupMove: %v", err)
		}
	}
}
