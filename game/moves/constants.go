package moves

// Board constants
const (
	MinRank = 0
	MaxRank = 7
	MinFile = 0
	MaxFile = 7
	BoardSize = 8
)

// Castling constants
const (
	KingsideFile = 6
	QueensideFile = 2
	KingsideRookFromFile = 7
	KingsideRookToFile = 5
	QueensideRookFromFile = 0
	QueensideRookToFile = 3
	KingStartFile = 4
)

// Move list capacity
const (
	InitialMoveListCapacity = 64
	MaxMoveListCapacity     = 512  // Pool capacity limit
	PoolPreAllocCapacity    = 256  // Pre-allocation size for pool
)

// Chess board dimensions for validation
const (
	ChessboardFiles = 8  // Number of files (a-h)
	ChessboardRanks = 8  // Number of ranks (1-8)
)

// Direction represents a movement direction with rank and file deltas
type Direction struct {
	RankDelta int
	FileDelta int
}

// Direction constants for pieces
var (
	// Rook directions (straight lines)
	RookDirections = []Direction{
		{1, 0},  // Up
		{-1, 0}, // Down
		{0, 1},  // Right
		{0, -1}, // Left
	}
	
	// Bishop directions (diagonals)
	BishopDirections = []Direction{
		{1, 1},   // Up-right
		{1, -1},  // Up-left
		{-1, 1},  // Down-right
		{-1, -1}, // Down-left
	}
	
	// Queen directions (combination of rook and bishop)
	QueenDirections = []Direction{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},     // Rook moves
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},   // Bishop moves
	}
	
	// Knight directions (L-shaped moves)
	KnightDirections = []Direction{
		{2, 1},   // Up 2, Right 1
		{2, -1},  // Up 2, Left 1
		{-2, 1},  // Down 2, Right 1
		{-2, -1}, // Down 2, Left 1
		{1, 2},   // Up 1, Right 2
		{1, -2},  // Up 1, Left 2
		{-1, 2},  // Down 1, Right 2
		{-1, -2}, // Down 1, Left 2
	}
)