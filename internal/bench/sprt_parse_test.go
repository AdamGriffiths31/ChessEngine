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

func TestParseSPRTStop(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   SPRTVerdict
		wantOK bool
	}{
		{"H0 accepted", "SPRT ([0.00, 200.00]) completed - H0 was accepted", SPRTAcceptH0, true},
		{"H1 accepted", "SPRT ([0.00, 2.00]) completed - H1 was accepted", SPRTAcceptH1, true},
		{"status line", "LLR: 2.95 (100.2%) (-2.94, 2.94) [0.00, 2.00]", SPRTContinue, false},
		{"cutechess line", "SPRT: llr 2.95 (100.2%), lbound -2.94, ubound 2.94 - H1 was accepted", SPRTContinue, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseSPRTStop(tc.line)
			if got != tc.want || ok != tc.wantOK {
				t.Errorf("parseSPRTStop(%q) = %v, %v, want %v, %v", tc.line, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestScoreLinePatterns(t *testing.T) {
	games := gamesPattern.FindStringSubmatch("Games: 16, Wins: 11, Losses: 3, Draws: 2, Points: 12.0 (75.00 %)")
	if games == nil {
		t.Fatal("gamesPattern did not match a fastchess Games line")
	}
	got, err := parseIntMatches(games)
	if err != nil {
		t.Fatalf("parseIntMatches(games): %v", err)
	}
	if want := []int{16, 11, 3, 2}; !slices.Equal(got, want) {
		t.Errorf("games = %v, want %v", got, want)
	}

	ptnml := pentanomialPattern.FindStringSubmatch("Ptnml(0-2): [0, 0, 3, 2, 3], WL/DD Ratio: inf")
	if ptnml == nil {
		t.Fatal("pentanomialPattern did not match a fastchess Ptnml line")
	}
	got, err = parseIntMatches(ptnml)
	if err != nil {
		t.Fatalf("parseIntMatches(ptnml): %v", err)
	}
	if want := []int{0, 0, 3, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("ptnml = %v, want %v", got, want)
	}
}

func TestParseIntMatchesRejectsOverflow(t *testing.T) {
	if _, err := parseIntMatches([]string{"full", "99999999999999999999"}); err == nil {
		t.Error("parseIntMatches accepted a count that overflows int")
	}
}
