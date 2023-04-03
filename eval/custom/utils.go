package eval

import "github.com/AdamGriffiths31/ChessEngine/data"

const (
	darkSquares       = uint64(0xAA55AA55AA55AA55)
	scaleFactorNormal = 128
)

func OnlyOne(bb uint64) bool {
	return bb != 0 && !Several(bb)
}

func Several(bb uint64) bool {
	return bb&(bb-1) != 0
}

func flip(sq int) int {
	return data.Mirror64[sq]
}
