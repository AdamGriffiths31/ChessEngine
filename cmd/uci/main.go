// Package main implements the ChessEngine UCI protocol interface.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AdamGriffiths31/ChessEngine/internal/uci"
)

const defaultBookPath = "internal/book/testdata/performance.bin"

const envDebugLog = "CHESS_DEBUG_LOG"

func debugLogDefault() string {
	return os.Getenv(envDebugLog)
}

// stdout is never touched, since it is the UCI protocol channel.
func setupLogging(debugLogPath string) (closeFn func(), err error) {
	if debugLogPath == "" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))
		return func() {}, nil
	}

	f, err := os.OpenFile(filepath.Clean(debugLogPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open debug log file %s: %w", debugLogPath, err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return func() { _ = f.Close() }, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	debugLogPath := flag.String("debug-log", debugLogDefault(), "path to append JSON logs to, one record per move plus the UCI transcript (default: $"+envDebugLog+"; if empty, text warnings to stderr)")
	flag.Parse()

	closeLog, err := setupLogging(*debugLogPath)
	if err != nil {
		return err
	}
	defer closeLog()

	if flag.Arg(0) == "bench" {
		report, err := uci.BenchReport(flag.Args()[1:])
		if err != nil {
			return err
		}
		fmt.Println(report)
		return nil
	}

	engine := uci.NewUCIEngine()

	if _, err := os.Stat(defaultBookPath); err == nil {
		engine.SetBookFile(defaultBookPath)
	} else {
		engine.SetBookFile("")
	}

	if err := engine.Run(os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("UCI engine failed: %w", err)
	}
	return nil
}
