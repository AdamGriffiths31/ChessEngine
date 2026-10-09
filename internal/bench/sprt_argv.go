package bench

import "fmt"

// BuildSPRTArgs constructs the cutechess-cli argument list for an SPRT run
// between cfg.BasePath and cfg.DevPath. Pure function: no I/O, no process
// spawning - see sprt_runner.go for that. Dev registers before Base because
// cutechess-cli computes the SPRT LLR from the first-registered engine's
// perspective; reversed, "Dev wins" trended toward H0.
func BuildSPRTArgs(cfg SPRTConfig) []string {
	args := []string{
		"-engine", fmt.Sprintf("cmd=%s", cfg.DevPath), "name=Dev",
	}
	for _, opt := range cfg.DevOptions {
		args = append(args, fmt.Sprintf("option.%s", opt))
	}
	args = append(args, "-engine", fmt.Sprintf("cmd=%s", cfg.BasePath), "name=Base")
	for _, opt := range cfg.BaseOptions {
		args = append(args, fmt.Sprintf("option.%s", opt))
	}
	args = append(args,
		"-each", "proto=uci", fmt.Sprintf("tc=%s", cfg.TimeControl),
		"-games", fmt.Sprintf("%d", cfg.MaxGames),
		"-repeat",
		"-openings", fmt.Sprintf("file=%s", cfg.OpeningsFile), "format=epd", "order=random",
		"-concurrency", fmt.Sprintf("%d", cfg.Concurrency),
		"-sprt",
		fmt.Sprintf("elo0=%g", cfg.Elo0),
		fmt.Sprintf("elo1=%g", cfg.Elo1),
		fmt.Sprintf("alpha=%g", cfg.Alpha),
		fmt.Sprintf("beta=%g", cfg.Beta),
		"-maxmoves", "200",
		"-recover",
		"-draw", "movenumber=40", "movecount=8", "score=10",
		"-resign", "movecount=6", "score=800",
	)
	return args
}
