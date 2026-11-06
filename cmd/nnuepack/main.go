package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/nnue/trainingdata"
)

func main() {
	out := flag.String("out", "data.bin", "packed file to write; an existing file is replaced")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: nnuepack -out data.bin data.txt [more.txt ...]")
		os.Exit(2)
	}
	if err := run(*out, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "nnuepack:", err)
		os.Exit(1)
	}
}

func run(out string, inputs []string) error {
	start := time.Now()
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := trainingdata.NewWriter(f)
	if err != nil {
		return err
	}

	var written, skipped int
	for _, input := range inputs {
		n, s, err := pack(w, input)
		if err != nil {
			return fmt.Errorf("%s: %w", input, err)
		}
		fmt.Printf("%s: %d positions, %d lines skipped\n", input, n, s)
		written, skipped = written+n, skipped+s
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %d positions to %s in %s (%d lines skipped)\n",
		written, out, time.Since(start).Round(time.Second), skipped)
	return nil
}

func pack(w *trainingdata.Writer, path string) (written, skipped int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	skipped, err = trainingdata.ScanText(f, func(record *trainingdata.Record) error {
		written++
		return w.Write(record)
	})
	return written, skipped, err
}
