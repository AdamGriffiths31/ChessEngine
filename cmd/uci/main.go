// Package main implements the ChessEngine UCI protocol interface.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/internal/uci"
)

// defaultBookPath is the opening book used when present relative to the working
// directory at startup. Absent the file, the engine runs with no book.
const defaultBookPath = "internal/book/testdata/performance.bin"

// setupLogging wires the default slog logger. With no flag, warnings and
// above are written as text to stderr. When debugLogPath is non-empty, a JSON
// handler at Debug level writes to that file instead; stdout is never
// touched, since it is the UCI protocol channel.
func setupLogging(debugLogPath string) (closeFn func(), err error) {
	if debugLogPath == "" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))
		return func() {}, nil
	}

	f, err := os.OpenFile(debugLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) // #nosec G304 - path from CLI flag
	if err != nil {
		return nil, fmt.Errorf("failed to open debug log file %s: %w", debugLogPath, err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return func() { _ = f.Close() }, nil
}

func main() {
	debugLogPath := flag.String("debug-log", "", "path to write JSON debug logs at Debug level (default: text warnings to stderr)")
	flag.Parse()

	closeLog, err := setupLogging(*debugLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer closeLog()

	engine := uci.NewUCIEngine()

	if _, err := os.Stat(defaultBookPath); err == nil {
		engine.SetBookFile(defaultBookPath)
	} else {
		engine.SetBookFile("")
	}

	if err := engine.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "UCI engine failed: %v\n", err)
		os.Exit(1)
	}
}
