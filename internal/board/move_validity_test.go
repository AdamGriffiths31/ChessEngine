package board

import "testing"

// IsValid is the sanity check for moves read back from the transposition
// table: From and To files must be in [0,7] and From must differ from To. Ranks
// are deliberately not range-checked.
func TestMoveIsValid(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		move     Move
		expected bool
	}{
		{
			name:     "valid move",
			move:     Move{From: Square{File: 4, Rank: 1}, To: Square{File: 4, Rank: 3}},
			expected: true,
		},
		{
			name:     "from file below range",
			move:     Move{From: Square{File: -1, Rank: 1}, To: Square{File: 4, Rank: 3}},
			expected: false,
		},
		{
			name:     "from file above range",
			move:     Move{From: Square{File: 8, Rank: 1}, To: Square{File: 4, Rank: 3}},
			expected: false,
		},
		{
			name:     "to file below range",
			move:     Move{From: Square{File: 4, Rank: 1}, To: Square{File: -1, Rank: 3}},
			expected: false,
		},
		{
			name:     "to file above range",
			move:     Move{From: Square{File: 4, Rank: 1}, To: Square{File: 8, Rank: 3}},
			expected: false,
		},
		{
			name:     "from equals to",
			move:     Move{From: Square{File: 4, Rank: 3}, To: Square{File: 4, Rank: 3}},
			expected: false,
		},
		{
			name:     "zero-value move",
			move:     Move{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.move.IsValid(); got != tc.expected {
				t.Errorf("Move{%+v}.IsValid() = %v, want %v", tc.move, got, tc.expected)
			}
		})
	}
}
