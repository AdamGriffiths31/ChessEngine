package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
)

// runReport regenerates the STS and Elo history tables from the JSONL
// result files under tools/results (sts_*.jsonl and elo_results.jsonl),
// replacing the hand-maintained sts_history.md/elo.md tables.
func runReport(args []string) error {
	fs := flag.NewFlagSet("bench report", flag.ExitOnError)
	dir := fs.String("dir", "tools/results", "directory containing sts_*.jsonl and elo_results.jsonl")
	output := fs.String("o", "", "write the report to this file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	report, err := bench.GenerateReport(*dir)
	if err != nil {
		return fmt.Errorf("failed to generate report: %v", err)
	}

	if *output == "" {
		fmt.Print(report)
		return nil
	}

	if err := os.WriteFile(*output, []byte(report), 0600); err != nil {
		return fmt.Errorf("failed to write report to %s: %v", *output, err)
	}
	fmt.Printf("Report written to %s\n", *output)
	return nil
}
