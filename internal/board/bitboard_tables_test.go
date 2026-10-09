package board

import "testing"

func TestKnightAttackPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		from int
		want []int
	}{
		{E4, []int{D2, F2, C3, G3, C5, G5, D6, F6}},
		{A1, []int{B3, C2}},         // corner
		{H8, []int{F7, G6}},         // corner
		{D1, []int{B2, F2, C3, E3}}, // edge
	}
	for _, tt := range tests {
		if got, want := GetKnightAttacks(tt.from), bitboardOf(tt.want...); got != want {
			t.Errorf("knight on %s: got %v, want %v", SquareToString(tt.from), got.BitList(), want.BitList())
		}
	}
}

func TestKingAttackPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		from int
		want []int
	}{
		{E4, []int{D3, E3, F3, D4, F4, D5, E5, F5}},
		{A1, []int{A2, B1, B2}},         // corner
		{H8, []int{G7, G8, H7}},         // corner
		{E1, []int{D1, F1, D2, E2, F2}}, // edge
	}
	for _, tt := range tests {
		if got, want := GetKingAttacks(tt.from), bitboardOf(tt.want...); got != want {
			t.Errorf("king on %s: got %v, want %v", SquareToString(tt.from), got.BitList(), want.BitList())
		}
	}
}

func TestPawnAttackPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		from  int
		color BitboardColor
		want  []int
	}{
		{"white centre", E4, BitboardWhite, []int{D5, F5}},
		{"black centre", E4, BitboardBlack, []int{D3, F3}},
		{"white edge file", A2, BitboardWhite, []int{B3}},
		{"black edge file", H7, BitboardBlack, []int{G6}},
		{"white on last rank", E8, BitboardWhite, nil},
		{"black on first rank", E1, BitboardBlack, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := GetPawnAttacks(tt.from, tt.color), bitboardOf(tt.want...); got != want {
				t.Errorf("got %v, want %v", got.BitList(), want.BitList())
			}
		})
	}
}

func TestBetweenSquares(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b int
		want []int
	}{
		{"same file", A1, A5, []int{A2, A3, A4}},
		{"same rank", A1, E1, []int{B1, C1, D1}},
		{"diagonal", A1, D4, []int{B2, C3}},
		{"anti-diagonal", H8, E5, []int{G7, F6}},
		{"not on a line", A1, B3, nil},
		{"same square", A1, A1, nil},
		{"adjacent", A1, A2, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := GetBetween(tt.a, tt.b), bitboardOf(tt.want...); got != want {
				t.Errorf("got %v, want %v", got.BitList(), want.BitList())
			}
		})
	}
}

func TestLineSquares(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b int
		want []int
	}{
		{"same file", A1, A3, []int{A1, A2, A3}},
		{"same rank", A1, C1, []int{A1, B1, C1}},
		{"diagonal", A1, C3, []int{A1, B2, C3}},
		{"not on a line", A1, B3, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := GetLine(tt.a, tt.b), bitboardOf(tt.want...); got != want {
				t.Errorf("got %v, want %v", got.BitList(), want.BitList())
			}
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	t.Parallel()
	for _, square := range []int{-1, 64, 100} {
		if GetKnightAttacks(square) != 0 {
			t.Errorf("GetKnightAttacks(%d) should return 0", square)
		}
		if GetKingAttacks(square) != 0 {
			t.Errorf("GetKingAttacks(%d) should return 0", square)
		}
		if GetPawnAttacks(square, BitboardWhite) != 0 {
			t.Errorf("GetPawnAttacks(%d, White) should return 0", square)
		}
		if GetBetween(square, E4) != 0 {
			t.Errorf("GetBetween(%d, E4) should return 0", square)
		}
	}
}
