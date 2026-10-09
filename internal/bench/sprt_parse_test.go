package bench

import "testing"

func TestParseSPRTStatus(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantLLR float64
	}{
		{
			name:    "standard status line",
			line:    "SPRT: llr 2.94 (100.0%), lbound -2.94, ubound 2.94",
			wantOK:  true,
			wantLLR: 2.94,
		},
		{
			name:    "negative llr",
			line:    "SPRT: llr -1.50 (-51.0%), lbound -2.94, ubound 2.94",
			wantOK:  true,
			wantLLR: -1.50,
		},
		{
			name:   "unrelated line",
			line:   "Finished game 12 (Dev vs Base): 1-0 {White wins}",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, ok := ParseSPRTStatus(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ParseSPRTStatus(%q) ok = %v, want %v", tc.line, ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if status.llr != tc.wantLLR {
				t.Errorf("ParseSPRTStatus(%q) llr = %v, want %v", tc.line, status.llr, tc.wantLLR)
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
