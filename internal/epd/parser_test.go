package epd

import (
	"reflect"
	"testing"
)

func TestParseEPD(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		line       string
		wantErr    bool
		wantBest   string
		wantAvoid  string
		wantID     string
		wantScores []MoveScore
	}{
		{name: "best move", line: "1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1 bm b8b7", wantBest: "b8b7"},
		{name: "castling best move", line: "r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 bm e1g1", wantBest: "e1g1"},
		{
			name:     "STS annotations",
			line:     `8/8/8/8/8/8/4P3/4K3 w - - bm e4; id "STS(v2.2) Open Files.001"; c0 "e4=10, Kf2=7, e3=3"; am Kd2;`,
			wantBest: "e4", wantAvoid: "Kd2", wantID: "STS(v2.2) Open Files.001",
			wantScores: []MoveScore{{"e4", 10}, {"Kf2", 7}, {"e3", 3}},
		},
		{name: "empty line", line: "", wantErr: true},
		{name: "comment line", line: "# This is a comment", wantErr: true},
		{name: "invalid FEN", line: "invalid_fen w - - 0 1 bm Nf3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := ParseEPD(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseEPD succeeded, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEPD: %v", err)
			}
			if pos.Board == nil {
				t.Fatal("Board is nil")
			}
			if pos.BestMove != tt.wantBest || pos.AvoidMove != tt.wantAvoid || pos.ID != tt.wantID {
				t.Errorf("got best=%q avoid=%q id=%q, want best=%q avoid=%q id=%q",
					pos.BestMove, pos.AvoidMove, pos.ID, tt.wantBest, tt.wantAvoid, tt.wantID)
			}
			if !reflect.DeepEqual(pos.MoveScores, tt.wantScores) {
				t.Errorf("MoveScores = %v, want %v", pos.MoveScores, tt.wantScores)
			}
		})
	}
}

// Comment and blank lines in a file are skipped, not reported as errors.
func TestParseEPDFile(t *testing.T) {
	t.Parallel()
	content := `1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1 bm b8b7
4r1k1/p1qr1p2/2pb1Bp1/1p5p/3P1n1R/1B3P2/PP3PK1/2Q4R w - - 0 1 bm c1f4
# This is a comment line
r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 bm e1g1
`

	positions, err := ParseEPDFile(content)
	if err != nil {
		t.Fatalf("ParseEPDFile: %v", err)
	}

	var got []string
	for _, p := range positions {
		got = append(got, p.BestMove)
	}
	if want := []string{"b8b7", "c1f4", "e1g1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("best moves = %v, want %v", got, want)
	}
}
