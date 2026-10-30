package bench

import (
	"fmt"
	"path/filepath"
)

// BuildSPRTArgs constructs the fastchess argument list for an SPRT run
// between cfg.BasePath and cfg.DevPath. Dev registers before Base because
// fastchess reports results and the LLR from the first engine's perspective.
//
// Games are played in pairs (-games 2 -repeat): each opening once with each
// colour, which is what the pentanomial model expects.
func BuildSPRTArgs(cfg SPRTConfig) []string {
	args := []string{
		"-engine", fmt.Sprintf("cmd=%s", cfg.DevPath), "name=Dev", fmt.Sprintf("dir=%s", filepath.Dir(cfg.DevPath)),
	}
	for _, opt := range cfg.DevOptions {
		args = append(args, fmt.Sprintf("option.%s", opt))
	}
	args = append(args, "-engine", fmt.Sprintf("cmd=%s", cfg.BasePath), "name=Base", fmt.Sprintf("dir=%s", filepath.Dir(cfg.BasePath)))
	for _, opt := range cfg.BaseOptions {
		args = append(args, fmt.Sprintf("option.%s", opt))
	}
	args = append(args,
		"-each", "proto=uci", fmt.Sprintf("tc=%s", cfg.TimeControl),
		"-rounds", fmt.Sprintf("%d", (cfg.MaxGames+1)/2),
		"-games", "2",
		"-repeat",
		"-openings", fmt.Sprintf("file=%s", cfg.OpeningsFile), "format=epd", "order=random",
		"-concurrency", fmt.Sprintf("%d", cfg.Concurrency),
		"-sprt",
		fmt.Sprintf("elo0=%g", cfg.Elo0),
		fmt.Sprintf("elo1=%g", cfg.Elo1),
		fmt.Sprintf("alpha=%g", cfg.Alpha),
		fmt.Sprintf("beta=%g", cfg.Beta),
		"model=normalized",
		"-maxmoves", "200",
		"-recover",
		"-draw", "movenumber=40", "movecount=8", "score=10",
		"-resign", "movecount=6", "score=800",
		// fastchess otherwise writes a resumable config.json into the
		// working directory every 20 games.
		"-autosaveinterval", "0",
	)
	return args
}
