package epd

import "testing"

// FuzzParseEPD feeds arbitrary strings to ParseEPD. Malformed input must
// never panic; where parsing succeeds, the returned Position must at least
// carry a non-nil Board and support String() without panicking.
func FuzzParseEPD(f *testing.F) {
	seeds := []string{
		"1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1 bm b8b7",
		"r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 bm e1g1",
		"",
		"# This is a comment",
		"invalid_fen w - - 0 1 bm Nf3",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 bm e2e4; id \"test\";",
		"8/8/8/8/8/8/4P3/8 w - - 0 1 c0 \"Ba7=10, Qf6+=3, a5=3, h5=5\";",
		"r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 am Nf3",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, epdLine string) {
		pos, err := ParseEPD(epdLine)
		if err != nil {
			return
		}
		if pos == nil {
			t.Fatalf("ParseEPD(%q) returned nil position with nil error", epdLine)
		}
		if pos.Board == nil {
			t.Fatalf("ParseEPD(%q) returned position with nil Board", epdLine)
		}
		// Cheap invariant: String() must not panic on a successfully
		// parsed position.
		_ = pos.String()
	})
}
