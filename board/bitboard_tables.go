package board

import "sync"

// Precomputed attack tables for chess pieces
var (
	// Basic file and rank masks (already defined in bitboard.go as constants)
	FileMasks [8]Bitboard
	RankMasks [8]Bitboard
	
	// Diagonal masks
	DiagonalMasks     [15]Bitboard // Main diagonals (a1-h8 direction)
	AntiDiagonalMasks [15]Bitboard // Anti-diagonals (a8-h1 direction)
	
	// Non-sliding piece attacks
	KnightAttacks [64]Bitboard
	KingAttacks   [64]Bitboard
	
	// Pawn attacks (separate for each color)
	WhitePawnAttacks [64]Bitboard
	BlackPawnAttacks [64]Bitboard
	
	// Pawn pushes (single and double)
	WhitePawnPushes       [64]Bitboard
	BlackPawnPushes       [64]Bitboard
	WhitePawnDoublePushes [64]Bitboard
	BlackPawnDoublePushes [64]Bitboard
	
	// Distance and connectivity tables
	DistanceTable [64][64]int      // Manhattan distance between squares
	BetweenTable  [64][64]Bitboard // Squares between two squares (exclusive)
	LineTable     [64][64]Bitboard // All squares on the line between two squares (inclusive)
	
	// Initialization flags
	tablesInitialized bool
	initMutex         sync.Once
)

// InitializeTables initializes all precomputed attack tables
// This should be called once before using any attack generation functions
func InitializeTables() {
	initMutex.Do(func() {
		initializeFilesAndRanks()
		initializeDiagonals()
		initializeKnightAttacks()
		initializeKingAttacks()
		initializePawnAttacks()
		initializePawnPushes()
		initializeDistanceTable()
		initializeBetweenTable()
		initializeLineTable()
		tablesInitialized = true
	})
}

// initializeFilesAndRanks initializes file and rank masks
func initializeFilesAndRanks() {
	for file := 0; file < 8; file++ {
		FileMasks[file] = FileMask(file)
	}
	
	for rank := 0; rank < 8; rank++ {
		RankMasks[rank] = RankMask(rank)
	}
}

// initializeDiagonals initializes diagonal and anti-diagonal masks
func initializeDiagonals() {
	// Main diagonals (a1-h8 direction)
	for diag := 0; diag < 15; diag++ {
		var mask Bitboard
		for square := 0; square < 64; square++ {
			file, rank := SquareToFileRank(square)
			// Main diagonal: rank - file is constant
			if rank-file == diag-7 {
				mask = mask.SetBit(square)
			}
		}
		DiagonalMasks[diag] = mask
	}
	
	// Anti-diagonals (a8-h1 direction)
	for diag := 0; diag < 15; diag++ {
		var mask Bitboard
		for square := 0; square < 64; square++ {
			file, rank := SquareToFileRank(square)
			// Anti-diagonal: rank + file is constant
			if rank+file == diag {
				mask = mask.SetBit(square)
			}
		}
		AntiDiagonalMasks[diag] = mask
	}
}

// initializeKnightAttacks initializes knight attack patterns
func initializeKnightAttacks() {
	knightMoves := [][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	
	for square := 0; square < 64; square++ {
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

// initializeKingAttacks initializes king attack patterns
func initializeKingAttacks() {
	kingMoves := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1},           {0, 1},
		{1, -1},  {1, 0},  {1, 1},
	}
	
	for square := 0; square < 64; square++ {
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

// initializePawnAttacks initializes pawn attack patterns
func initializePawnAttacks() {
	for square := 0; square < 64; square++ {
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

// initializePawnPushes initializes pawn push patterns
func initializePawnPushes() {
	for square := 0; square < 64; square++ {
		file, rank := SquareToFileRank(square)
		
		// White pawn pushes
		var whitePush, whiteDoublePush Bitboard
		if rank < 7 { // Not on 8th rank
			// Single push
			targetSquare := FileRankToSquare(file, rank+1)
			whitePush = whitePush.SetBit(targetSquare)
			
			// Double push from 2nd rank
			if rank == 1 && rank < 6 {
				targetSquare = FileRankToSquare(file, rank+2)
				whiteDoublePush = whiteDoublePush.SetBit(targetSquare)
			}
		}
		WhitePawnPushes[square] = whitePush
		WhitePawnDoublePushes[square] = whiteDoublePush
		
		// Black pawn pushes
		var blackPush, blackDoublePush Bitboard
		if rank > 0 { // Not on 1st rank
			// Single push
			targetSquare := FileRankToSquare(file, rank-1)
			blackPush = blackPush.SetBit(targetSquare)
			
			// Double push from 7th rank
			if rank == 6 && rank > 1 {
				targetSquare = FileRankToSquare(file, rank-2)
				blackDoublePush = blackDoublePush.SetBit(targetSquare)
			}
		}
		BlackPawnPushes[square] = blackPush
		BlackPawnDoublePushes[square] = blackDoublePush
	}
}

// initializeDistanceTable initializes the distance table between all squares
func initializeDistanceTable() {
	for sq1 := 0; sq1 < 64; sq1++ {
		for sq2 := 0; sq2 < 64; sq2++ {
			file1, rank1 := SquareToFileRank(sq1)
			file2, rank2 := SquareToFileRank(sq2)
			
			fileDist := abs(file1 - file2)
			rankDist := abs(rank1 - rank2)
			
			// Manhattan distance (also known as taxicab distance)
			DistanceTable[sq1][sq2] = fileDist + rankDist
		}
	}
}

// initializeBetweenTable initializes the between table for all square pairs
func initializeBetweenTable() {
	for sq1 := 0; sq1 < 64; sq1++ {
		for sq2 := 0; sq2 < 64; sq2++ {
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

// initializeLineTable initializes the line table for all square pairs
func initializeLineTable() {
	for sq1 := 0; sq1 < 64; sq1++ {
		for sq2 := 0; sq2 < 64; sq2++ {
			var line Bitboard
			
			if sq1 != sq2 {
				file1, rank1 := SquareToFileRank(sq1)
				file2, rank2 := SquareToFileRank(sq2)
				
				fileDiff := file2 - file1
				rankDiff := rank2 - rank1
				
				// Check if squares are on the same line (rank, file, or diagonal)
				if fileDiff == 0 || rankDiff == 0 || abs(fileDiff) == abs(rankDiff) {
					// Add both endpoints
					line = line.SetBit(sq1).SetBit(sq2)
					
					// Add all squares between
					line = line | BetweenTable[sq1][sq2]
				}
			}
			
			LineTable[sq1][sq2] = line
		}
	}
}

// Helper function to get absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Getter functions for attack patterns

// GetKnightAttacks returns the knight attack pattern for a given square
func GetKnightAttacks(square int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	return KnightAttacks[square]
}

// GetKingAttacks returns the king attack pattern for a given square
func GetKingAttacks(square int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	return KingAttacks[square]
}

// GetPawnAttacks returns the pawn attack pattern for a given square and color
func GetPawnAttacks(square int, color BitboardColor) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	
	if color == BitboardWhite {
		return WhitePawnAttacks[square]
	}
	return BlackPawnAttacks[square]
}

// GetPawnPushes returns the pawn push pattern for a given square and color
func GetPawnPushes(square int, color BitboardColor) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	
	if color == BitboardWhite {
		return WhitePawnPushes[square]
	}
	return BlackPawnPushes[square]
}

// GetPawnDoublePushes returns the pawn double push pattern for a given square and color
func GetPawnDoublePushes(square int, color BitboardColor) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	
	if color == BitboardWhite {
		return WhitePawnDoublePushes[square]
	}
	return BlackPawnDoublePushes[square]
}

// GetDistance returns the Manhattan distance between two squares
func GetDistance(sq1, sq2 int) int {
	if !tablesInitialized {
		InitializeTables()
	}
	if sq1 < 0 || sq1 > 63 || sq2 < 0 || sq2 > 63 {
		return -1
	}
	return DistanceTable[sq1][sq2]
}

// GetBetween returns the squares between two squares (exclusive)
func GetBetween(sq1, sq2 int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if sq1 < 0 || sq1 > 63 || sq2 < 0 || sq2 > 63 {
		return 0
	}
	return BetweenTable[sq1][sq2]
}

// GetLine returns all squares on the line between two squares (inclusive)
func GetLine(sq1, sq2 int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if sq1 < 0 || sq1 > 63 || sq2 < 0 || sq2 > 63 {
		return 0
	}
	return LineTable[sq1][sq2]
}

// GetDiagonalMask returns the diagonal mask for a given square
func GetDiagonalMask(square int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	
	file, rank := SquareToFileRank(square)
	diagIndex := rank - file + 7
	return DiagonalMasks[diagIndex]
}

// GetAntiDiagonalMask returns the anti-diagonal mask for a given square
func GetAntiDiagonalMask(square int) Bitboard {
	if !tablesInitialized {
		InitializeTables()
	}
	if square < 0 || square > 63 {
		return 0
	}
	
	file, rank := SquareToFileRank(square)
	diagIndex := rank + file
	return AntiDiagonalMasks[diagIndex]
}

// IsTablesInitialized returns whether the attack tables have been initialized
func IsTablesInitialized() bool {
	return tablesInitialized
}

// ResetTablesForTesting resets the initialization state for testing purposes
// This should only be used in tests
func ResetTablesForTesting() {
	tablesInitialized = false
	initMutex = sync.Once{}
}