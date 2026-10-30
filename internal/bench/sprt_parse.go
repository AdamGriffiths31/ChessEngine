package bench

import (
	"regexp"
	"strconv"
)

type sprtStatus struct {
	llr    float64
	lbound float64
	ubound float64
}

var sprtStatusPattern = regexp.MustCompile(`^LLR: (-?[\d.]+) \([^)]*\) \((-?[\d.]+), (-?[\d.]+)\)`)

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

var gamesPattern = regexp.MustCompile(`^Games: (\d+), Wins: (\d+), Losses: (\d+), Draws: (\d+)`)

var pentanomialPattern = regexp.MustCompile(`^Ptnml\(0-2\): \[(\d+), (\d+), (\d+), (\d+), (\d+)\]`)

func parseIntMatches(m []string) ([]int, error) {
	out := make([]int, len(m)-1)
	for i, s := range m[1:] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

var sprtStopPattern = regexp.MustCompile(`^SPRT \(.*\) completed - (H[01]) was accepted`)

func parseSPRTStop(line string) (SPRTVerdict, bool) {
	m := sprtStopPattern.FindStringSubmatch(line)
	if m == nil {
		return SPRTContinue, false
	}
	if m[1] == "H1" {
		return SPRTAcceptH1, true
	}
	return SPRTAcceptH0, true
}
