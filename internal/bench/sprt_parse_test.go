package bench

import (
	"slices"
	"testing"
)

func TestParseSPRTStatus(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantLLR float64
	}{
		{
			name:    "standard status line",
			line:    "LLR: 2.94 (100.0%) (-2.94, 2.94) [0.00, 2.00]",
			wantOK:  true,
			wantLLR: 2.94,
		},
		{
			name:    "negative llr",
			line:    "LLR: -1.50 (-51.0%) (-2.94, 2.94) [0.00, 2.00]",
			wantOK:  true,
			wantLLR: -1.50,
		},
		{
			name:   "unrelated line",
			line:   "Finished game 12 (Dev vs Base): 1-0 {White wins by adjudication}",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, ok := parseSPRTStatus(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("parseSPRTStatus(%q) ok = %v, want %v", tc.line, ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if status.llr != tc.wantLLR {
				t.Errorf("parseSPRTStatus(%q) llr = %v, want %v", tc.line, status.llr, tc.wantLLR)
			}
		})
	}
}

func TestVerdictFromStatus(t *testing.T) {
	tests := []struct {
		name   string
		status sprtStatus
		want   SPRTVerdict
	}{
		{"below lbound accepts H0", sprtStatus{llr: -3.0, lbound: -2.94, ubound: 2.94}, SPRTAcceptH0},
		{"above ubound accepts H1", sprtStatus{llr: 3.0, lbound: -2.94, ubound: 2.94}, SPRTAcceptH1},
		{"between bounds continues", sprtStatus{llr: 0.1, lbound: -2.94, ubound: 2.94}, SPRTContinue},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := verdictFromStatus(tc.status); got != tc.want {
				t.Errorf("verdictFromStatus(%+v) = %v, want %v", tc.status, got, tc.want)
			}
		})
	}
}

// TestScoreLinePatterns checks the score and pentanomial patterns against
// lines captured from a real fastchess 1.8.2 run.
func TestScoreLinePatterns(t *testing.T) {
	games := gamesPattern.FindStringSubmatch("Games: 16, Wins: 11, Losses: 3, Draws: 2, Points: 12.0 (75.00 %)")
	if games == nil {
		t.Fatal("gamesPattern did not match a fastchess Games line")
	}
	if got, want := parseIntMatches(games), []int{16, 11, 3, 2}; !slices.Equal(got, want) {
		t.Errorf("games = %v, want %v", got, want)
	}

	ptnml := pentanomialPattern.FindStringSubmatch("Ptnml(0-2): [0, 0, 3, 2, 3], WL/DD Ratio: inf")
	if ptnml == nil {
		t.Fatal("pentanomialPattern did not match a fastchess Ptnml line")
	}
	if got, want := parseIntMatches(ptnml), []int{0, 0, 3, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("ptnml = %v, want %v", got, want)
	}
}
