package uci

import (
	"strings"
	"testing"
)

// FuzzParseGo feeds arbitrary "go" command argument strings (split the same
// way ParseCommand splits raw input) to ParseGo. ParseGo has no error return
// -- it silently ignores anything it can't parse -- so the only invariant
// worth asserting is that it never panics on malformed or adversarial input.
func FuzzParseGo(f *testing.F) {
	seeds := []string{
		"depth 6",
		"movetime 5000",
		"infinite",
		"wtime 300000 btime 300000 winc 2000 binc 2000",
		"movestogo 40 depth 10",
		"nodes 1000000",
		"depth",
		"movetime -1",
		"depth 99999999999999999999",
		"",
		"depth abc",
		"wtime btime winc binc",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	handler := NewProtocolHandler()

	f.Fuzz(func(_ *testing.T, argsStr string) {
		args := strings.Fields(argsStr)
		// No return value to validate beyond "did not panic" -- ParseGo
		// silently drops anything it can't parse.
		_ = handler.ParseGo(args)
	})
}
