package engine

import (
	"math/rand"
	"time"
)

const MaxMoves = 2048

// const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
const StartFEN = "1QR2b1Q/P1K2bp1/1PPP1p2/1r2RB1q/2PqNBpp/p1p1ppnP/1nP4P/1k6 w - - 0 1"

const (
	False = iota
	True  = iota
)

var PceChar = []string{
	".",
	"P",
	"N",
	"B",
	"R",
	"Q",
	"K",
	"p",
	"n",
	"b",
	"r",
	"q",
	"k",
}

var SideChar = []string{
	"w",
	"b",
	"-",
}

var RankChars = []string{
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
}

var FileChars = []string{
	"a",
	"b",
	"c",
	"d",
	"e",
	"f",
	"g",
	"h",
}
var Pieces = map[int]string{
	0:  "Empty",
	1:  "WP",
	2:  "WN",
	3:  "WB",
	4:  "WR",
	5:  "WQ",
	6:  "WK",
	7:  "BP",
	8:  "BN",
	9:  "BB",
	10: "BR",
	11: "BQ",
	12: "BK",
	13: "OffBoard",
}

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
	OffBoard                       = 100
)

const (
	WhiteKingCastle  = 1
	WhiteQueenCastle = 2
	BlackKingCastle  = 4
	BlackQueenCastle = 8
)

var BitTable = [64]int{
	63, 30, 3, 32, 25, 41, 22, 33, 15, 50, 42, 13, 11, 53, 19, 34, 61, 29, 2,
	51, 21, 43, 45, 10, 18, 47, 1, 54, 9, 57, 0, 35, 62, 31, 40, 4, 49, 5, 52,
	26, 60, 6, 23, 44, 46, 27, 56, 16, 7, 39, 48, 24, 59, 14, 12, 55, 38, 28,
	58, 20, 37, 17, 36, 8,
}

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
	PieceList        [13][10]int
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

var SetMask [64]uint64
var ClearMask [64]uint64

var PieceKeys [13][120]uint64
var SideKey uint64
var CastleKeys [16]uint64

func init() {
	setSquares()
	setBitMasks()
	setPieceKeys()
}

// FileRankToSquare converts file & rank to the 120 sqaure
func FileRankToSquare(file, rank int) int {
	return (21 + file) + (rank * 10)
}

// GenerateRandomUint64 returns a random uint64
func generateRandomUint64() uint64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint64()
}

func setPieceKeys() {
	for i := 0; i < 13; i++ {
		for j := 0; j < 120; j++ {
			PieceKeys[i][j] = generateRandomUint64()
		}
	}

	SideKey = generateRandomUint64()
	for i := 0; i < 16; i++ {
		CastleKeys[i] = generateRandomUint64()
	}
}

// setSquares populates Sqaure120ToSquare64 & Sqaure64ToSquare120
func setSquares() {
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
			sq := FileRankToSquare(file, rank)
			Sqaure64ToSquare120[sqaure64] = sq
			Sqaure120ToSquare64[sq] = sqaure64
			sqaure64++
		}
	}
}

// setBitMasks populates ClearMask & SetMask
func setBitMasks() {
	for index := 0; index < 64; index++ {
		SetMask[index] = 0
		ClearMask[index] = 0
	}

	for index := 0; index < 64; index++ {
		SetMask[index] = 1 << index
		ClearMask[index] = ^SetMask[index]
	}
}
