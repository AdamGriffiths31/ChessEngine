package bench

import "time"

// SPRTConfig configures an SPRT regression test between two UCI engine
// binaries, run via cutechess-cli.
type SPRTConfig struct {
	BasePath     string   // path to the baseline UCI binary
	DevPath      string   // path to the candidate UCI binary
	BaseOptions  []string // UCI options for Base, each "Name=Value" (e.g. "Skill Level=5")
	DevOptions   []string // UCI options for Dev, each "Name=Value"
	Elo0         float64  // H1 rejected if the true Elo gain is below this
	Elo1         float64  // H1 accepted if the true Elo gain is at or above this
	Alpha        float64  // max probability of accepting H1 when H0 is true
	Beta         float64  // max probability of accepting H0 when H1 is true
	TimeControl  string   // cutechess-cli tc= syntax, e.g. "10+0.1"
	OpeningsFile string   // path to an EPD opening book
	Concurrency  int      // concurrent games; keep engine Threads=1 so games share cores
	MaxGames     int      // safety cap if SPRT never reaches a bound
	CutechessCli string   // path to the cutechess-cli binary
}

// DefaultSPRTConfig returns the standard STC bounds. Concurrency defaults
// to 1 for deterministic, low-interference runs; raise it via the -concurrency
// flag to trade some determinism for wall-clock speed.
func DefaultSPRTConfig(rootPath string) SPRTConfig {
	return SPRTConfig{
		Elo0:         0,
		Elo1:         5,
		Alpha:        0.05,
		Beta:         0.05,
		TimeControl:  "10+0.1",
		OpeningsFile: rootPath + "/testdata/sprt_openings.epd",
		Concurrency:  1,
		MaxGames:     4000,
		CutechessCli: rootPath + "/tools/engines/cutechess-cli.exe",
	}
}

// SPRTVerdict mirrors cutechess-cli's sprt.h Result enum.
type SPRTVerdict int

const (
	// SPRTContinue means the games cap was reached with no decision.
	SPRTContinue SPRTVerdict = iota
	// SPRTAcceptH0 means the dev binary's Elo advantage is below Elo0 -
	// not a real improvement.
	SPRTAcceptH0
	// SPRTAcceptH1 means the dev binary's Elo advantage is at least Elo1.
	SPRTAcceptH1
)

// String returns a human-readable verdict label for logging.
func (v SPRTVerdict) String() string {
	switch v {
	case SPRTAcceptH0:
		return "H0 accepted (no improvement)"
	case SPRTAcceptH1:
		return "H1 accepted (improvement)"
	default:
		return "inconclusive (games cap reached)"
	}
}

// SPRTResult is the outcome of one SPRT run. Wins/Losses/Draws are all from
// the Dev binary's perspective (a Dev win is a Base loss), matching the
// "is Dev stronger than Base" framing SPRT answers.
type SPRTResult struct {
	Verdict    SPRTVerdict
	LLR        float64
	LBound     float64
	UBound     float64
	Wins       int
	Losses     int
	Draws      int
	GamesTotal int
	Config     SPRTConfig
	Timestamp  string
	Duration   time.Duration
	RawLog     string
}
