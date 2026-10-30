package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/internal/bench"
)

func parseEngineOptions(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func runSPRT(args []string) error {
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}
	cfg := bench.DefaultSPRTConfig(rootPath)

	fs := flag.NewFlagSet("bench sprt", flag.ExitOnError)
	base := fs.String("base", "", "path to the baseline UCI binary (required)")
	dev := fs.String("dev", "", "path to the candidate UCI binary (required)")
	elo0 := fs.Float64("elo0", cfg.Elo0, "SPRT elo0 bound (normalized Elo)")
	elo1 := fs.Float64("elo1", cfg.Elo1, "SPRT elo1 bound (normalized Elo)")
	alpha := fs.Float64("alpha", cfg.Alpha, "SPRT alpha (type I error)")
	beta := fs.Float64("beta", cfg.Beta, "SPRT beta (type II error)")
	tc := fs.String("tc", cfg.TimeControl, "fastchess time control, seconds+increment")
	openings := fs.String("openings", cfg.OpeningsFile, "EPD opening book path")
	concurrency := fs.Int("concurrency", cfg.Concurrency, "concurrent games")
	maxGames := fs.Int("games", cfg.MaxGames, "safety cap on total games")
	baseOptions := fs.String("base-options", "", `comma-separated UCI options for -base, e.g. "Skill Level=5,Threads=1"`)
	devOptions := fs.String("dev-options", "", `comma-separated UCI options for -dev, e.g. "Skill Level=5,Threads=1"`)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *base == "" || *dev == "" {
		fs.Usage()
		return fmt.Errorf("both -base and -dev are required")
	}

	cfg.BasePath = *base
	cfg.DevPath = *dev
	cfg.BaseOptions = parseEngineOptions(*baseOptions)
	cfg.DevOptions = parseEngineOptions(*devOptions)
	cfg.Elo0 = *elo0
	cfg.Elo1 = *elo1
	cfg.Alpha = *alpha
	cfg.Beta = *beta
	cfg.TimeControl = *tc
	cfg.OpeningsFile = *openings
	cfg.Concurrency = *concurrency
	cfg.MaxGames = *maxGames

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := bench.RunSPRT(ctx, cfg)
	if result == nil {
		return err
	}

	logger := bench.NewSPRTResultsLogger(rootPath)
	if logErr := logger.LogResult(result); logErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to log SPRT result: %v\n", logErr)
	}

	fmt.Printf("\nSPRT verdict: %s (llr=%.2f, bounds=[%.2f, %.2f], W-L-D=%d-%d-%d, games=%d)\n",
		result.Verdict, result.LLR, result.LBound, result.UBound,
		result.Wins, result.Losses, result.Draws, result.GamesTotal)

	return err
}
