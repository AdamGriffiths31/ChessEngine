package board

import "testing"

// FuzzFromFEN feeds arbitrary strings to FromFEN. The only hard requirement
// is that malformed input never panics; well-formed input should parse to a
// *Board that can, in turn, be round-tripped through ToFEN without panicking.
func FuzzFromFEN(f *testing.F) {
	seeds := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"8/8/8/8/8/8/4P3/8 w - - 0 1",
		"8/8/8/8/4p3/8/4P3/8 w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1",
		"1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1",
		"",
		"invalid_fen w - - 0 1",
		"8/8/8/8/8/8/8/8 w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq d6 0 3",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, fen string) {
		b, err := FromFEN(fen)
		if err != nil {
			return
		}
		if b == nil {
			t.Fatalf("FromFEN(%q) returned nil board with nil error", fen)
		}
		// Cheap invariant: a successfully parsed board must be able to
		// round-trip through ToFEN without panicking.
		_ = b.ToFEN()
	})
}
