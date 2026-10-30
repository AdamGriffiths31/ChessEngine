package bench

import (
	"regexp"
	"strconv"
)

// sprtStatus holds the numeric fields from one fastchess SPRT status line:
// the current log-likelihood ratio and the decision bounds it's compared
// against.
type sprtStatus struct {
	llr    float64
	lbound float64
	ubound float64
}

// sprtStatusPattern matches fastchess's periodic SPRT line, e.g.
// "LLR: 0.05 (1.6%) (-2.94, 2.94) [0.00, 2.00]". If a future fastchess
// version changes the layout, update this pattern rather than the call sites.
var sprtStatusPattern = regexp.MustCompile(`^LLR: (-?[\d.]+) \([^)]*\) \((-?[\d.]+), (-?[\d.]+)\)`)

// parseSPRTStatus extracts llr/lbound/ubound from a fastchess stdout line.
// ok is false for any line that isn't an SPRT status line.
func parseSPRTStatus(line string) (sprtStatus, bool) {
	m := sprtStatusPattern.FindStringSubmatch(line)
	if m == nil {
		return sprtStatus{}, false
	}

	llr, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return sprtStatus{}, false
	}
	lbound, err := strconv.ParseFloat(m[2], 64)
	if err != nil {
		return sprtStatus{}, false
	}
	ubound, err := strconv.ParseFloat(m[3], 64)
	if err != nil {
		return sprtStatus{}, false
	}

	return sprtStatus{llr: llr, lbound: lbound, ubound: ubound}, true
}

// gamesPattern matches fastchess's cumulative score line, e.g.
// "Games: 16, Wins: 11, Losses: 3, Draws: 2, Points: 12.0 (75.00 %)".
// The counts are for the first-registered engine (Dev).
var gamesPattern = regexp.MustCompile(`^Games: (\d+), Wins: (\d+), Losses: (\d+), Draws: (\d+)`)

// pentanomialPattern matches fastchess's game-pair distribution line, e.g.
// "Ptnml(0-2): [0, 0, 3, 2, 3], WL/DD Ratio: inf".
var pentanomialPattern = regexp.MustCompile(`^Ptnml\(0-2\): \[(\d+), (\d+), (\d+), (\d+), (\d+)\]`)

// parseIntMatches converts every submatch after the full match to an int.
// Callers' patterns only capture \d+, so conversion cannot fail.
func parseIntMatches(m []string) []int {
	out := make([]int, len(m)-1)
	for i, s := range m[1:] {
		out[i], _ = strconv.Atoi(s)
	}
	return out
}

// verdictFromStatus compares llr against the decision bounds the same way
// fastchess does (continue / accept H0 / accept H1), but from its
// already-printed numbers rather than by re-deriving the SPRT math.
func verdictFromStatus(status sprtStatus) SPRTVerdict {
	switch {
	case status.llr <= status.lbound:
		return SPRTAcceptH0
	case status.llr >= status.ubound:
		return SPRTAcceptH1
	default:
		return SPRTContinue
	}
}
