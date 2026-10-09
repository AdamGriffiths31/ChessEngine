package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
)

// runReport regenerates the STS history table from the JSONL
// result files under tools/results (sts_*.jsonl),
// replacing the hand-maintained sts_history.md table.
func runReport(args []string) error {
	fs := flag.NewFlagSet("bench report", flag.ExitOnError)
	dir := fs.String("dir", "tools/results", "directory containing sts_*.jsonl")
	output := fs.String("o", "", "write the report to this file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	report, err := bench.GenerateReport(*dir)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if *output == "" {
		fmt.Print(report)
		return nil
	}

	if err := os.WriteFile(*output, []byte(report), 0600); err != nil {
		return fmt.Errorf("failed to write report to %s: %w", *output, err)
	}
	fmt.Printf("Report written to %s\n", *output)
	return nil
}
