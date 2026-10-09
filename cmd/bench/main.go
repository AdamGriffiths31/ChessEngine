// Package main provides the "bench" multi-tool binary, which bundles the
// chess engine's benchmarking utilities behind subcommands:
//
//	bench sts       - Strategic Test Suite benchmark (formerly cmd/sts)
//	bench profile   - CPU/memory profiling utility (formerly cmd/profile)
//	bench report    - regenerate the STS history table from tools/results/*.jsonl
//	bench sprt      - SPRT regression test between two UCI binaries
package main

import (
	"fmt"
	"log/slog"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: bench <subcommand> [flags]

Subcommands:
  sts        Strategic Test Suite (STS) benchmark
  profile    CPU/memory profiling utility
  report     Regenerate the STS history table from tools/results/*.jsonl
  sprt       SPRT regression test between two UCI binaries

Run "bench <subcommand> -h" for flags on a specific subcommand.`)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "sts":
		err = runSTS(os.Args[2:])
	case "profile":
		err = runProfile(os.Args[2:])
	case "report":
		err = runReport(os.Args[2:])
	case "sprt":
		err = runSPRT(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
