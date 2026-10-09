package bench

import (
	"regexp"
	"strconv"
)

// sprtStatus holds the numeric fields from one cutechess-cli SPRT status
// line: the current log-likelihood ratio and the decision bounds it's
// compared against.
type sprtStatus struct {
	llr    float64
	lbound float64
	ubound float64
}

// sprtStatusPattern matches cutechess-cli's periodic SPRT status line, e.g.
// "SPRT: llr 2.94 (100.0%), lbound -2.94, ubound 2.94". This is the
// commonly-documented shape; if a future cutechess-cli version changes the
// layout, update this pattern rather than the call sites.
var sprtStatusPattern = regexp.MustCompile(`llr (-?[\d.]+).*lbound (-?[\d.]+), ubound (-?[\d.]+)`)

// ParseSPRTStatus extracts llr/lbound/ubound from a cutechess-cli stdout
// line. ok is false for any line that isn't an SPRT status line.
func ParseSPRTStatus(line string) (sprtStatus, bool) {
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

// verdictFromStatus compares llr against the decision bounds the same way
// cutechess-cli's own sprt.cpp does (Continue/AcceptH0/AcceptH1), but from
// its already-printed numbers rather than by re-deriving the SPRT math.
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
