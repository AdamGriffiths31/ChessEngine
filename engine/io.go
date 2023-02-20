package engine

func SqaureString(square int) string {
	file := FilesBoard[square] + 'a'
	rank := RanksBoard[square] + '1'
	return string([]byte{byte(file), byte(rank)})
}

func PrintMove(move int) string {
	fromFile := FilesBoard[FromSquare(move)] + 'a'
	fromRank := RanksBoard[FromSquare(move)] + '1'
	toFile := FilesBoard[ToSqaure(move)] + 'a'
	toRank := RanksBoard[ToSqaure(move)] + '1'

	promotedStr := "q"
	promoted := Promoted(move)
	if promoted != 0 {
		if PieceKnight[promoted] == True {
			promotedStr = "n"
		} else if PieceRookQueen[promoted] == True && PieceBishopQueen[promoted] == False {
			promotedStr = "r"
		} else if PieceRookQueen[promoted] == False && PieceBishopQueen[promoted] == True {
			promotedStr = "b"
		}
		return string([]byte{byte(fromFile), byte(fromRank), byte(toFile), byte(toRank)}) + promotedStr
	}

	return string([]byte{byte(fromFile), byte(fromRank), byte(toFile), byte(toRank)})
}
