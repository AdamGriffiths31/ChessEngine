package engine

import "fmt"

func SqaureString(square int) string {
	if !SqaureOnBoard(square) {
		return "Sqaure not on board"
	}
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

func ParseMove(move []byte, pos *Board, info *SearchInfo) int {
	if move[0] == 't' {
		fmt.Printf("Take back\n")
		TakeMoveBack(pos)
		PrintBoard(pos)
		return NoMove
	}
	if move[0] == 'p' {
		PerftTest(4, StartFEN)
		return NoMove
	}
	if move[0] == 's' {
		SearchPosistion(pos, info)
		return NoMove
	}
	if move[0] == 'r' {
		max := GetPvLine(1, pos)
		fmt.Printf("PvLine of %d moves:", max)
		for i := 0; i < max; i++ {
			move := pos.PvArray[i]
			fmt.Printf("%s\n", PrintMove(move))
		}
		PrintBoard(pos)
	}
	if move[1] > '8' || move[1] < '1' {
		return NoMove
	}
	if move[3] > '8' || move[3] < '1' {
		return NoMove
	}
	if move[0] > 'h' || move[0] < 'a' {
		return NoMove
	}
	if move[2] > 'h' || move[2] < 'a' {
		return NoMove
	}

	from := FileRankToSquare(int(move[0]-'a'), int(move[1]-'1'))
	to := FileRankToSquare(int(move[2]-'a'), int(move[3]-'1'))

	ml := &MoveList{}
	GenerateAllMoves(pos, ml)

	for MoveNum := 0; MoveNum < ml.Count; MoveNum++ {
		userMove := ml.Moves[MoveNum].Move
		if FromSquare(userMove) == from && ToSqaure(userMove) == to {
			promPce := Promoted(userMove)
			if promPce != Empty {
				if PieceRookQueen[promPce] == True && PieceBishopQueen[promPce] == False && move[4] == 'r' {
					return userMove
				} else if PieceRookQueen[promPce] == False && PieceBishopQueen[promPce] == True && move[4] == 'b' {
					return userMove
				} else if PieceRookQueen[promPce] == True && PieceBishopQueen[promPce] == True && move[4] == 'q' {
					return userMove
				} else if PieceKnight[promPce] == True && move[4] == 'n' {
					return userMove
				}
				continue
			}
			return userMove
		}
	}

	return NoMove
}
