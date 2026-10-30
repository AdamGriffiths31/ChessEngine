package bench

import "time"

// SPRTConfig configures an SPRT regression test between two UCI engine
// binaries, run via fastchess.
type SPRTConfig struct {
	BasePath     string   // path to the baseline UCI binary
	DevPath      string   // path to the candidate UCI binary
	BaseOptions  []string // UCI options for Base, each "Name=Value" (e.g. "Skill Level=5")
	DevOptions   []string // UCI options for Dev, each "Name=Value"
	Elo0         float64  // H1 rejected if the true normalized Elo gain is below this
	Elo1         float64  // H1 accepted if the true normalized Elo gain is at or above this
	Alpha        float64  // max probability of accepting H1 when H0 is true
	Beta         float64  // max probability of accepting H0 when H1 is true
	TimeControl  string   // fastchess tc= syntax, e.g. "10+0.1"
	OpeningsFile string   // path to an EPD opening book
	Concurrency  int      // concurrent games
	MaxGames     int      // safety cap if SPRT never reaches a bound
	Fastchess    string   // path to the fastchess binary
}

// DefaultSPRTConfig returns the standard STC bounds: normalized Elo [0, 2].
// Concurrency defaults to 1.
func DefaultSPRTConfig(rootPath string) SPRTConfig {
	return SPRTConfig{
		Elo0:         0,
		Elo1:         2,
		Alpha:        0.05,
		Beta:         0.05,
		TimeControl:  "10+0.1",
		OpeningsFile: rootPath + "/testdata/8moves_v3.epd",
		Concurrency:  1,
		MaxGames:     4000,
		Fastchess:    rootPath + "/tools/engines/fastchess.exe",
	}
}

// SPRTVerdict is the outcome of an SPRT run.
type SPRTVerdict int

const (
	// SPRTContinue means fastchess accepted neither hypothesis: the games cap
	// was reached or the run ended early.
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
// the Dev binary's perspective (a Dev win is a Base loss).
type SPRTResult struct {
	Verdict    SPRTVerdict
	LLR        float64
	LBound     float64
	UBound     float64
	Wins       int
	Losses     int
	Draws      int
	GamesTotal int
	// Pentanomial counts game pairs by Dev's pair score: 0, 0.5, 1, 1.5, 2.
	Pentanomial [5]int
	Config      SPRTConfig
	Timestamp   string
	Duration    time.Duration
	RawLog      string
}
