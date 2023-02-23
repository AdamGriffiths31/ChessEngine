package engine

import (
	"math/rand"
	"time"
)

const MaxMoves = 2048
const MaxPositionMoves = 256
const MaxDepth = 64

// const StartFEN = "rnbqkbnr/pppppppp/8/8/8/7N/PPPPPPPP/RNBQKB1R b KQkq - 0 1"
const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

//const StartFEN = "2rr3k/pp3pp1/1nnqbN1p/3pN3/2pP4/2P3Q1/PPB4P/R4RK1 w - - 0 1"

//const StartFEN = "r1b1k2r/ppppnppp/2n2q2/2b5/3NP3/2P1B3/PP3PPP/RN1QKB1R w KQkq - 0 1"

//const StartFEN = "rnbqkbnr/ppp1pppp/8/3p4/4P3/2N5/PPPP1PPP/R1BQKBNR b KQkq - 0 2"

// const StartFEN = "rnbqkb1r/pp1p1pPp/8/2p1pP2/1P1P4/3P3P/P1P1P3/RNBQKBNR w WKkq e6 0 1"
// const StartFEN = "rnbqkbnr/p1p1p3/3p3p/1p1p4/2P1Pp2/8/PP1P1PpP/RNBQKB1R b KQkq e3 0 1"
// const StartFEN = "5k2/1n6/4n3/6N1/8/3N4/8/5K2 w - - 0 1"
// const StartFEN = "6k1/8/5r2/8/1nR5/5N2/86K1 b - - 0 1"
// const StartFEN = "r3k23/8/8/8/8/8/8/R3K2R b KQkq - 0 1"
// const StartFEN = "3rk2r/8/8/8/8/8/6p1/R3K2R b KQk - 0 1"
// const StartFEN = "r3k2r/p1ppqb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"

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
	"u",
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

type MoveList struct {
	Moves [MaxPositionMoves]Move
	Count int
}

type Move struct {
	Score int
	Move  int
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
	BigPiece         [2]int
	MajorPiece       [2]int
	MinPiece         [2]int
	Material         [2]int
	PieceList        [13][10]int
	CastlePermission int
	History          [MaxMoves]Undo

	PvTable *PVTable
	PvArray [MaxDepth]int

	SearchHistory [13][120]int
	SearchKillers [2][MaxDepth]int
}

type Undo struct {
	Move             int
	CastlePermission int
	EnPas            int
	FiftyMove        int
	PosistionKey     uint64
}

type PVEntry struct {
	PosistionKey uint64
	Move         int
}

type PVTable struct {
	PTable        []PVEntry
	NumberEntries int
}

type SearchInfo struct {
	StartTime int64
	StopTime  int64
	Depth     int
	DepthSet  int
	TimeSet   int
	MovesToGo int
	Infinite  int
	MoveTime  int
	Time      int
	Inc       int

	Node int64

	Quit    int
	Stopped int

	FailHigh      float32
	FailHighFirst float32
}

var Sqaure120ToSquare64 [120]int
var Sqaure64ToSquare120 [64]int

var SetMask [64]uint64
var ClearMask [64]uint64

var PieceKeys [13][120]uint64
var SideKey uint64
var CastleKeys [16]uint64

var PieceBig = [13]int{0, 0, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1}
var PieceMajor = [13]int{0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 1, 1}
var PieceMin = [13]int{0, 0, 1, 1, 0, 0, 0, 0, 1, 1, 0, 0, 0}
var PieceVal = [13]int{0, 100, 325, 325, 550, 1000, 5000, 100, 325, 325, 550, 1000, 5000}
var PieceCol = [13]int{Both, White, White, White, White, White, White, Black, Black, Black, Black, Black, Black}

var FilesBoard [120]int
var RanksBoard [120]int

var KnightDirection = [8]int{-8, -19, -21, -12, 8, 19, 21, 12}
var RookDirection = [4]int{-1, -10, 1, 10}
var BishopDirection = [4]int{-9, -11, 11, 9}
var KingDirection = [8]int{-1, -10, 1, 10, -9, -11, 11, 9}

var PiecePawn = [13]int{False, True, False, False, False, False, False, True, False, False, False, False, False}
var PieceKnight = [13]int{False, False, True, False, False, False, False, False, True, False, False, False, False}
var PieceKing = [13]int{False, False, False, False, False, False, True, False, False, False, False, False, True}
var PieceRookQueen = [13]int{False, False, False, False, True, True, False, False, False, False, True, True, False}
var PieceBishopQueen = [13]int{False, False, False, True, False, True, False, False, False, True, False, True, False}
var PieceSlides = [13]int{False, False, False, True, True, True, True, False, False, True, True, True, True}

var LoopSlidePiece = [8]int{WB, WR, WQ, 0, BB, BR, BQ, 0}
var LoopSlideIndex = [2]int{0, 4}
var LoopNonSlidePiece = [6]int{WN, WK, 0, BN, BK, 0}
var LoopNonSlideIndex = [2]int{0, 3}

var NumDir = [13]int{0, 0, 8, 4, 4, 8, 8, 0, 8, 4, 4, 8, 8}
var PieceDir = [13][8]int{
	{0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0},
	{-8, -19, -21, -12, 8, 19, 21, 12},
	{-9, -11, 11, 9, 0, 0, 0, 0},
	{-1, -10, 1, 10, 0, 0, 0, 0},
	{-1, -10, 1, 10, -9, -11, 11, 9},
	{-1, -10, 1, 10, -9, -11, 11, 9},
	{0, 0, 0, 0, 0, 0, 0},
	{-8, -19, -21, -12, 8, 19, 21, 12},
	{-9, -11, 11, 9, 0, 0, 0, 0},
	{-1, -10, 1, 10, 0, 0, 0, 0},
	{-1, -10, 1, 10, -9, -11, 11, 9},
	{-1, -10, 1, 10, -9, -11, 11, 9},
}

var CastlePerm = [120]int{
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 13, 15, 15, 15, 12, 15, 15, 14, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 7, 15, 15, 15, 3, 15, 15, 11, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
	15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
}

func init() {
	rand.Seed(time.Now().UnixNano())
	setSquares()
	setBitMasks()
	setPieceKeys()
	setFilesAndRanks()
	initMvvLva()
}

// FileRankToSquare converts file & rank to the 120 sqaure
func FileRankToSquare(file, rank int) int {
	return (21 + file) + (rank * 10)
}

// GenerateRandomUint64 returns a random uint64
func generateRandomUint64() uint64 {
	return rand.Uint64()
}

// setPieceKeys sets the keys to a random Uinit64
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
		SetMask[index] = uint64(0)
		ClearMask[index] = uint64(0)
	}

	for index := 0; index < 64; index++ {
		SetMask[index] |= uint64(1) << index
		ClearMask[index] = ^SetMask[index]
	}
}

// setFilesAndRanks populates FilesBoard & RanksBoard
func setFilesAndRanks() {
	for i := 0; i < 120; i++ {
		FilesBoard[i] = OffBoard
		RanksBoard[i] = OffBoard
	}

	for rank := Rank1; rank <= Rank8; rank++ {
		for file := FileA; file <= FileH; file++ {
			sq := FileRankToSquare(file, rank)
			FilesBoard[sq] = file
			RanksBoard[sq] = rank
		}
	}
}

/*
0000 0000 0000 0000 0000 0111 1111 -> From 0x7F
0000 0000 0000 0011 1111 1000 0000 -> To >> 7, 0x7F
0000 0000 0011 1100 0000 0000 0000 -> Captured >> 14, 0xF
0000 0000 0100 0000 0000 0000 0000 -> EP 0x40000
0000 0000 1000 0000 0000 0000 0000 -> Pawn Start 0x80000
0000 1111 0000 0000 0000 0000 0000 -> Promoted Piece >> 20, 0xF
0001 0000 0000 0000 0000 0000 0000 -> Castle 0x1000000
*/

func FromSquare(move int) int {
	return move & 0x7F
}

func ToSqaure(move int) int {
	return move >> 7 & 0x7F
}

func Captured(move int) int {
	return move >> 14 & 0xF
}

func Promoted(move int) int {
	return move >> 20 & 0xF
}

const MFLAGEP int = 0x40000
const MFLAGPS int = 0x80000
const MFLAGGCA int = 0x1000000

const MFLAGCAP int = 0x7C000
const MFLAGPRO int = 0xF00000

var NoMove = 0

var VictimScore = [13]int{0, 100, 200, 300, 400, 500, 600, 100, 200, 300, 400, 500, 600}
var MvvLvaScores [13][13]int

func initMvvLva() {
	for attacker := WP; attacker <= BK; attacker++ {
		for victim := WP; victim <= BK; victim++ {
			MvvLvaScores[victim][attacker] = VictimScore[victim] + 6 - (VictimScore[attacker] / 100)
		}
	}
}
