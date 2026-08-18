package board

import "testing"

// TestMoveIsValid documents the exact predicate mirrored from the inline
// TT-move sanity check formerly in internal/search/negamax.go: both From.File
// and To.File must be within [0,7], and From must differ from To. Note that
// only the File components are range-checked (Rank is not), matching the
// original inline code precisely.
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
