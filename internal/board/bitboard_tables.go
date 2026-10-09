package board

// Precomputed attack tables for chess pieces
var (
	// File and rank masks
	FileMasks [8]Bitboard
	RankMasks [8]Bitboard

	// Non-sliding piece attacks
	KnightAttacks [64]Bitboard
	KingAttacks   [64]Bitboard

	// Pawn attacks (separate for each color)
	WhitePawnAttacks [64]Bitboard
	BlackPawnAttacks [64]Bitboard

	BetweenTable [64][64]Bitboard // Squares between two squares (exclusive)
	LineTable    [64][64]Bitboard // All squares on the line between two squares (inclusive)

)

// init automatically initializes attack tables when the package is loaded
func init() {
	initializeFilesAndRanks()
	initializeKnightAttacks()
	initializeKingAttacks()
	initializePawnAttacks()
	initializeBetweenTable()
	initializeLineTable()
}

func initializeFilesAndRanks() {
	for file := range 8 {
		FileMasks[file] = FileMask(file)
	}

	for rank := range 8 {
		RankMasks[rank] = RankMask(rank)
	}
}

func initializeKnightAttacks() {
	knightMoves := [][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for square := range 64 {
		var attacks Bitboard
		file, rank := SquareToFileRank(square)

		for _, move := range knightMoves {
			newFile := file + move[0]
			newRank := rank + move[1]

			if newFile >= 0 && newFile <= 7 && newRank >= 0 && newRank <= 7 {
				targetSquare := FileRankToSquare(newFile, newRank)
				attacks = attacks.SetBit(targetSquare)
			}
		}

		KnightAttacks[square] = attacks
	}
}

func initializeKingAttacks() {
	kingMoves := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for square := range 64 {
		var attacks Bitboard
		file, rank := SquareToFileRank(square)

		for _, move := range kingMoves {
			newFile := file + move[0]
			newRank := rank + move[1]

			if newFile >= 0 && newFile <= 7 && newRank >= 0 && newRank <= 7 {
				targetSquare := FileRankToSquare(newFile, newRank)
				attacks = attacks.SetBit(targetSquare)
			}
		}

		KingAttacks[square] = attacks
	}
}

func initializePawnAttacks() {
	for square := range 64 {
		file, rank := SquareToFileRank(square)

		// White pawn attacks (moving up the board)
		var whiteAttacks Bitboard
		if rank < 7 { // Not on 8th rank
			// Attack to the left (southwest)
			if file > 0 {
				targetSquare := FileRankToSquare(file-1, rank+1)
				whiteAttacks = whiteAttacks.SetBit(targetSquare)
			}
			// Attack to the right (southeast)
			if file < 7 {
				targetSquare := FileRankToSquare(file+1, rank+1)
				whiteAttacks = whiteAttacks.SetBit(targetSquare)
			}
		}
		WhitePawnAttacks[square] = whiteAttacks

		// Black pawn attacks (moving down the board)
		var blackAttacks Bitboard
		if rank > 0 { // Not on 1st rank
			// Attack to the left (northwest)
			if file > 0 {
				targetSquare := FileRankToSquare(file-1, rank-1)
				blackAttacks = blackAttacks.SetBit(targetSquare)
			}
			// Attack to the right (northeast)
			if file < 7 {
				targetSquare := FileRankToSquare(file+1, rank-1)
				blackAttacks = blackAttacks.SetBit(targetSquare)
			}
		}
		BlackPawnAttacks[square] = blackAttacks
	}
}

func initializeBetweenTable() {
	for sq1 := range 64 {
		for sq2 := range 64 {
			var between Bitboard

			if sq1 != sq2 {
				file1, rank1 := SquareToFileRank(sq1)
				file2, rank2 := SquareToFileRank(sq2)

				fileDiff := file2 - file1
				rankDiff := rank2 - rank1

				// Check if squares are on the same line (rank, file, or diagonal)
				if fileDiff == 0 || rankDiff == 0 || abs(fileDiff) == abs(rankDiff) {
					// Normalize direction
					fileStep := 0
					rankStep := 0
					if fileDiff != 0 {
						fileStep = fileDiff / abs(fileDiff)
					}
					if rankDiff != 0 {
						rankStep = rankDiff / abs(rankDiff)
					}

					// Add squares between sq1 and sq2 (exclusive)
					currentFile := file1 + fileStep
					currentRank := rank1 + rankStep

					for currentFile != file2 || currentRank != rank2 {
						square := FileRankToSquare(currentFile, currentRank)
						between = between.SetBit(square)
						currentFile += fileStep
						currentRank += rankStep
					}
				}
			}

			BetweenTable[sq1][sq2] = between
		}
	}
}

func initializeLineTable() {
	for sq1 := range 64 {
		for sq2 := range 64 {
			var line Bitboard

			if sq1 != sq2 {
				file1, rank1 := SquareToFileRank(sq1)
				file2, rank2 := SquareToFileRank(sq2)

				fileDiff := file2 - file1
				rankDiff := rank2 - rank1

				// Check if squares are on the same line (rank, file, or diagonal)
				if fileDiff == 0 || rankDiff == 0 || abs(fileDiff) == abs(rankDiff) {
					line = line.SetBit(sq1).SetBit(sq2)

					line |= BetweenTable[sq1][sq2]
				}
			}

			LineTable[sq1][sq2] = line
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// GetKnightAttacks returns the knight attack pattern for a given square
func GetKnightAttacks(square int) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}
	return KnightAttacks[square]
}

// GetKingAttacks returns the king attack pattern for a given square
func GetKingAttacks(square int) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}
	return KingAttacks[square]
}

// GetPawnAttacks returns the pawn attack pattern for a given square and color
func GetPawnAttacks(square int, color BitboardColor) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}

	if color == BitboardWhite {
		return WhitePawnAttacks[square]
	}
	return BlackPawnAttacks[square]
}

// GetBetween returns the squares between two squares (exclusive)
func GetBetween(sq1, sq2 int) Bitboard {
	if sq1 < 0 || sq1 > 63 || sq2 < 0 || sq2 > 63 {
		return 0
	}
	return BetweenTable[sq1][sq2]
}

// GetLine returns all squares on the line between two squares (inclusive)
func GetLine(sq1, sq2 int) Bitboard {
	if sq1 < 0 || sq1 > 63 || sq2 < 0 || sq2 > 63 {
		return 0
	}
	return LineTable[sq1][sq2]
}
