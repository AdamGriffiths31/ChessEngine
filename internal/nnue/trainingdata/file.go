package trainingdata

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// magic starts every packed file.
var magic = [8]byte{'N', 'N', 'U', 'E', 'P', 'O', 'S', '1'}

// Records is a run of packed positions held back to back.
type Records []byte

// Len returns the number of positions.
func (rs Records) Len() int { return len(rs) / RecordSize }

// At returns position i.
func (rs Records) At(i int) Record {
	return Record(rs[i*RecordSize : (i+1)*RecordSize])
}

// Writer appends packed positions to a file after its header.
type Writer struct {
	w *bufio.Writer
}

// NewWriter writes the header to w and returns a Writer for the records.
func NewWriter(w io.Writer) (*Writer, error) {
	bw := bufio.NewWriterSize(w, 1<<20)
	if _, err := bw.Write(magic[:]); err != nil {
		return nil, err
	}
	return &Writer{w: bw}, nil
}

// Write appends one record.
func (w *Writer) Write(r *Record) error {
	_, err := w.w.Write(r[:])
	return err
}

// Flush writes any buffered records.
func (w *Writer) Flush() error { return w.w.Flush() }

// Load reads a packed file, or a datagen text file when the header is
// missing. Text lines that do not parse are skipped and counted.
func Load(path string) (Records, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var header [len(magic)]byte
	n, err := io.ReadFull(f, header[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, 0, err
	}
	if n == len(magic) && header == magic {
		return loadPacked(f, path)
	}
	return loadText(io.MultiReader(bytes.NewReader(header[:n]), f))
}

func loadPacked(f *os.File, path string) (Records, int, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	size := info.Size() - int64(len(magic))
	if size%RecordSize != 0 {
		return nil, 0, fmt.Errorf("%s: %d bytes of records is not a multiple of %d", path, size, RecordSize)
	}
	records := make(Records, size)
	if _, err := io.ReadFull(f, records); err != nil {
		return nil, 0, err
	}
	return records, 0, nil
}

func loadText(r io.Reader) (Records, int, error) {
	var records Records
	skipped, err := ScanText(r, func(record *Record) error {
		records = append(records, record[:]...)
		return nil
	})
	return records, skipped, err
}

// ScanText parses datagen text lines and calls fn for each one that parses.
// It returns how many lines were skipped, including a cut-off last line.
func ScanText(r io.Reader, fn func(*Record) error) (skipped int, err error) {
	reader := bufio.NewReaderSize(r, 1<<20)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			if line != "" {
				skipped++
			}
			return skipped, nil
		}
		if err != nil {
			return skipped, err
		}
		record, err := parseLine(strings.TrimRight(line, "\r\n"))
		if err != nil {
			skipped++
			continue
		}
		if err := fn(&record); err != nil {
			return skipped, err
		}
	}
}
