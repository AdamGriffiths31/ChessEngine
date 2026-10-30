package board

func SquareToString(square int) string {
	if square < 0 || square > 63 {
		return "invalid"
	}
	file, rank := SquareToFileRank(square)
	return string(rune('a'+file)) + string(rune('1'+rank))
}

func StringToSquare(square string) int {
	if len(square) != 2 {
		return -1
	}
	file := int(square[0] - 'a')
	rank := int(square[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return -1
	}
	return FileRankToSquare(file, rank)
}

func bitboardOf(squares ...int) Bitboard {
	var bb Bitboard
	for _, sq := range squares {
		bb = bb.SetBit(sq)
	}
	return bb
}
