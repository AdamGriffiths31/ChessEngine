package engine

const MaxMoves = 2048

const (
	False = iota
	True  = iota
)

const (
	Empty = iota
	WP    = iota
	WN    = iota
	WB    = iota
	WR    = iota
	WQ    = iota
	WK    = iota
	BP    = iota
	BN    = iota
	BB    = iota
	BR    = iota
	BQ    = iota
	BK    = iota
	O     = iota
)

const (
	FileA     = iota
	FileB     = iota
	FileC     = iota
	FileD     = iota
	FileE     = iota
	FileF     = iota
	FileG     = iota
	FileH     = iota
	FileEmpty = iota
)

const (
	Rank1     = iota
	Rank2     = iota
	Rank3     = iota
	Rank4     = iota
	Rank5     = iota
	Rank6     = iota
	Rank7     = iota
	Rank8     = iota
	RankEmpty = iota
)

const (
	White = iota
	Black = iota
	Both  = iota
)

const (
	A1, B1, C1, D1, E1, F1, G1, H1 = 21, 22, 23, 24, 25, 26, 27, 28
	A2, B2, C2, D2, E2, F2, G2, H2 = 31, 32, 33, 34, 35, 36, 37, 38
	A3, B3, C3, D3, E3, F3, G3, H3 = 41, 42, 43, 44, 45, 46, 47, 48
	A4, B4, C4, D4, E4, F4, G4, H4 = 51, 52, 53, 54, 55, 56, 57, 58
	A5, B5, C5, D5, E5, F5, G5, H5 = 61, 62, 63, 64, 65, 66, 67, 68
	A6, B6, C6, D6, E6, F6, G6, H6 = 71, 72, 73, 74, 75, 76, 77, 78
	A7, B7, C7, D7, E7, F7, G7, H7 = 81, 82, 83, 84, 85, 86, 87, 88
	A8, B8, C8, D8, E8, F8, G8, H8 = 91, 92, 93, 94, 95, 96, 97, 98
	noSqaure                       = 99
)

const (
	WhiteKingCastle  = 1
	WhiteQueenCastle = 2
	BlackKingCastle  = 4
	BlackQueenCastle = 8
)

type Board struct {
	Pieces           [120]int
	KingSqaure       [2]int
	Side             int
	EnPas            int
	FiftyMove        int
	Play             int
	HistoryPlay      int
	PosistionKey     uint64
	PieceNumber      [13]int
	Pawns            [3]uint64
	BigPiece         [3]int
	MajorPiece       [3]int
	MinPiece         [3]int
	CastlePermission int
	History          [MaxMoves]Undo
}

type Undo struct {
	Move             int
	CastlePermission int
	EnPas            int
	FiftyMove        int
	PosistionKey     uint64
}

var Sqaure120ToSquare64 [120]int
var Sqaure64ToSquare120 [64]int

func SetSquares() {
	index := 0
	sqaure64 := 0

	for index = 0; index < 120; index++ {
		Sqaure120ToSquare64[index] = 65
	}

	for index = 0; index < 64; index++ {
		Sqaure64ToSquare120[index] = 120
	}

	for rank := Rank1; rank < RankEmpty; rank++ {
		for file := FileA; file < FileEmpty; file++ {
			sq := fileRankToSquare(file, rank)
			Sqaure64ToSquare120[sqaure64] = sq
			Sqaure120ToSquare64[sq] = sqaure64
			sqaure64++
		}
	}
}

func fileRankToSquare(file, rank int) int {
	return (21 + file) + (rank * 10)
}
