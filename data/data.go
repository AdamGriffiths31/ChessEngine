package data

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	setSquares()
	setBitMasks()
	initSquareBB()
	setPieceKeys()
	setFilesAndRanks()
	setEvalMasks()
	initMvvLva()
	Init_sliders_attacks()
}

const MaxMoves = 2048
const MaxPositionMoves = 256
const MaxDepth = 64
const MaxWorkers = 32

//const StartFEN = "8/7p/5k2/5p2/p1p2P2/Pr1pPK2/1P1R3P/8 b - -"

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

const (
	UCIMode     = iota
	XboardMode  = iota
	ConsoleMode = iota
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
	NoSquare                       = 99
	OffBoard                       = 100
)

const (
	WhiteKingCastle  = 1
	WhiteQueenCastle = 2
	BlackKingCastle  = 4
	BlackQueenCastle = 8
)

const (
	PVNone  = iota
	PVAlpha = iota
	PVBeta  = iota
	PVExact = iota
)

var BitTable = [64]int{
	63, 30, 3, 32, 25, 41, 22, 33, 15, 50, 42, 13, 11, 53, 19, 34, 61, 29, 2,
	51, 21, 43, 45, 10, 18, 47, 1, 54, 9, 57, 0, 35, 62, 31, 40, 4, 49, 5, 52,
	26, 60, 6, 23, 44, 46, 27, 56, 16, 7, 39, 48, 24, 59, 14, 12, 55, 38, 28,
	58, 20, 37, 17, 36, 8,
}

type SearchWorker struct {
	Pos  *Board
	Info *SearchInfo
	Hash *PvHashTable

	Number    int
	Depth     int
	BestMove  int
	BestScore int
}
type MoveList struct {
	Moves [MaxPositionMoves]Move
	Count int
}

type Move struct {
	Score int
	Move  int
	Depth int
}

type PvHashTable struct {
	HashTable *PVTable
}

type Board struct {
	Pieces           [120]int
	KingSquare       [2]int
	Side             int
	EnPas            int
	FiftyMove        int
	Play             int
	HistoryPlay      int
	PositionKey      uint64
	PieceNumber      [13]int
	Pawns            [3]uint64
	BigPiece         [2]int
	MajorPiece       [2]int
	MinPiece         [2]int
	Material         [2]int
	PieceList        [13][10]int
	CastlePermission int
	History          [MaxMoves]Undo

	PvArray [MaxDepth]int

	SearchHistory [13][120]int
	SearchKillers [2][MaxDepth]int

	ColoredPiecesBB uint64
	WhitePiecesBB   uint64
	PiecesBB        uint64
}

type Undo struct {
	Move             int
	CastlePermission int
	EnPas            int
	FiftyMove        int
	PositionKey      uint64
}

type PVEntry struct {
	Age     int
	SMPData uint64
	SMPKey  uint64
}

type PVTable struct {
	PTable        []PVEntry
	NumberEntries int
	Hit           int
	Cut           int
	CurrentAge    int
}

type SearchInfo struct {
	StartTime    int64
	StopTime     int64
	Depth        int
	DepthSet     int
	TimeSet      int
	MovesToGo    int
	Infinite     int
	MoveTime     int
	Time         int
	Inc          int
	PostThinking bool

	Node int64

	Quit      int
	Stopped   bool
	ForceStop bool

	FailHigh      float32
	FailHighFirst float32

	GameMode int

	Cut     int
	NullCut int

	WorkerNumber int
}

type EngineOptions struct {
	UseBook bool
}

var EngineSettings EngineOptions

const Infinite = 32000
const ABInfinite = 30000
const Mate = ABInfinite - MaxDepth

var Square120ToSquare64 [120]int
var Square64ToSquare120 [64]int

var SetMask [64]uint64
var ClearMask [64]uint64
var FileBBMask [64]uint64
var RankBBMask [64]uint64
var SquareBB [64]uint64

var BlackPassedMask [64]uint64
var WhitePassedMask [64]uint64
var IsolatedMask [64]uint64

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

var Mirror64 = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

// FileRankToSquare converts file & rank to the 120 square
func FileRankToSquare(file, rank int) int {
	return (21 + file) + (rank * 10)
}

// GenerateRandomUint64 returns a random uint64
func generateRandomUint64() uint64 {
	return rand.Uint64()
}

// setPieceKeys sets the keys to a random uint64
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

// setSquares populates Square120ToSquare64 & Square64ToSquare120
func setSquares() {
	index := 0
	square64 := 0

	for index = 0; index < 120; index++ {
		Square120ToSquare64[index] = 65
	}

	for index = 0; index < 64; index++ {
		Square64ToSquare120[index] = 120
	}

	for rank := Rank1; rank < RankEmpty; rank++ {
		for file := FileA; file < FileEmpty; file++ {
			sq := FileRankToSquare(file, rank)
			Square64ToSquare120[square64] = sq
			Square120ToSquare64[sq] = square64
			square64++
		}
	}
}

func initSquareBB() {
	for sq := 0; sq < 64; sq++ {
		SquareBB[sq] = 1 << uint64(sq)
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

func setEvalMasks() {
	for sq := 0; sq < 8; sq++ {
		FileBBMask[sq] = 0
		RankBBMask[sq] = 0
	}

	for r := Rank8; r >= Rank1; r-- {
		for f := FileA; f <= FileH; f++ {
			sq := r*8 + f
			FileBBMask[f] |= 1 << sq
			RankBBMask[r] |= 1 << sq
		}
	}

	// Initialize IsolatedMask, WhitePassedMask, and BlackPassedMask
	for sq := 0; sq < 64; sq++ {
		IsolatedMask[sq] = 0
		WhitePassedMask[sq] = 0
		BlackPassedMask[sq] = 0

		tsq := sq + 8
		for tsq < 64 {
			WhitePassedMask[sq] |= 1 << tsq
			tsq += 8
		}

		tsq = sq - 8
		for tsq >= 0 {
			BlackPassedMask[sq] |= 1 << tsq
			tsq -= 8
		}

		if FilesBoard[Square64ToSquare120[sq]] > FileA {
			IsolatedMask[sq] |= FileBBMask[FilesBoard[Square64ToSquare120[sq]]-1]

			tsq = sq + 7
			for tsq < 64 {
				WhitePassedMask[sq] |= 1 << tsq
				tsq += 8
			}

			tsq = sq - 9
			for tsq >= 0 {
				BlackPassedMask[sq] |= 1 << tsq
				tsq -= 8
			}
		}

		if FilesBoard[Square64ToSquare120[sq]] < FileH {
			IsolatedMask[sq] |= FileBBMask[FilesBoard[Square64ToSquare120[sq]]+1]

			tsq = sq + 9
			for tsq < 64 {
				WhitePassedMask[sq] |= 1 << tsq
				tsq += 8
			}

			tsq = sq - 7
			for tsq >= 0 {
				BlackPassedMask[sq] |= 1 << tsq
				tsq -= 8
			}
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

func ToSquare(move int) int {
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

func NewBoardPos() *Board {
	pos := &Board{}
	return pos
}

func InitPvTable(table *PVTable) {
	if table == nil {
		table = &PVTable{}
	}
	var pvSize = 0x100000 * 64
	table.NumberEntries = pvSize / int(unsafe.Sizeof(PVEntry{}))
	table.NumberEntries -= 2
	fmt.Printf("PVTable: %v entries (%v)\n", table.NumberEntries, table.CurrentAge)
	table.PTable = make([]PVEntry, table.NumberEntries)
}

var BishopMask [64]uint64
var BishopAttacks [64][512]uint64
var BishopMagics = [64]uint64{
	0x89a1121896040240,
	0x2004844802002010,
	0x2068080051921000,
	0x62880a0220200808,
	0x4042004000000,
	0x100822020200011,
	0xc00444222012000a,
	0x28808801216001,
	0x400492088408100,
	0x201c401040c0084,
	0x840800910a0010,
	0x82080240060,
	0x2000840504006000,
	0x30010c4108405004,
	0x1008005410080802,
	0x8144042209100900,
	0x208081020014400,
	0x4800201208ca00,
	0xf18140408012008,
	0x1004002802102001,
	0x841000820080811,
	0x40200200a42008,
	0x800054042000,
	0x88010400410c9000,
	0x520040470104290,
	0x1004040051500081,
	0x2002081833080021,
	0x400c00c010142,
	0x941408200c002000,
	0x658810000806011,
	0x188071040440a00,
	0x4800404002011c00,
	0x104442040404200,
	0x511080202091021,
	0x4022401120400,
	0x80c0040400080120,
	0x8040010040820802,
	0x480810700020090,
	0x102008e00040242,
	0x809005202050100,
	0x8002024220104080,
	0x431008804142000,
	0x19001802081400,
	0x200014208040080,
	0x3308082008200100,
	0x41010500040c020,
	0x4012020c04210308,
	0x208220a202004080,
	0x111040120082000,
	0x6803040141280a00,
	0x2101004202410000,
	0x8200000041108022,
	0x21082088000,
	0x2410204010040,
	0x40100400809000,
	0x822088220820214,
	0x40808090012004,
	0x910224040218c9,
	0x402814422015008,
	0x90014004842410,
	0x1000042304105,
	0x10008830412a00,
	0x2520081090008908,
	0x40102000a0a60140,
}

var RookMask [64]uint64
var RookAttacks [64][4096]uint64
var RookMagics = [64]uint64{
	0xa8002c000108020,
	0x6c00049b0002001,
	0x100200010090040,
	0x2480041000800801,
	0x280028004000800,
	0x900410008040022,
	0x280020001001080,
	0x2880002041000080,
	0xa000800080400034,
	0x4808020004000,
	0x2290802004801000,
	0x411000d00100020,
	0x402800800040080,
	0xb000401004208,
	0x2409000100040200,
	0x1002100004082,
	0x22878001e24000,
	0x1090810021004010,
	0x801030040200012,
	0x500808008001000,
	0xa08018014000880,
	0x8000808004000200,
	0x201008080010200,
	0x801020000441091,
	0x800080204005,
	0x1040200040100048,
	0x120200402082,
	0xd14880480100080,
	0x12040280080080,
	0x100040080020080,
	0x9020010080800200,
	0x813241200148449,
	0x491604001800080,
	0x100401000402001,
	0x4820010021001040,
	0x400402202000812,
	0x209009005000802,
	0x810800601800400,
	0x4301083214000150,
	0x204026458e001401,
	0x40204000808000,
	0x8001008040010020,
	0x8410820820420010,
	0x1003001000090020,
	0x804040008008080,
	0x12000810020004,
	0x1000100200040208,
	0x430000a044020001,
	0x280009023410300,
	0xe0100040002240,
	0x200100401700,
	0x2244100408008080,
	0x8000400801980,
	0x2000810040200,
	0x8010100228810400,
	0x2000009044210200,
	0x4080008040102101,
	0x40002080411d01,
	0x2005524060000901,
	0x502001008400422,
	0x489a000810200402,
	0x1004400080a13,
	0x4000011008020084,
	0x26002114058042,
}

var rookRelevantBits = [64]int{
	12, 11, 11, 11, 11, 11, 11, 12,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	12, 11, 11, 11, 11, 11, 11, 12,
}

var bishopRelevantBits = [64]int{
	6, 5, 5, 5, 5, 5, 5, 6,
	5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5,
	6, 5, 5, 5, 5, 5, 5, 6,
}

func GetRookAttacks(occ uint64, sq int) uint64 {
	occ &= RookMask[sq]
	occ *= RookMagics[sq]
	occ >>= 64 - rookRelevantBits[sq]
	return RookAttacks[sq][occ]
}

func GetBishopAttacks(occ uint64, sq int) uint64 {
	occ &= BishopMask[sq]
	occ *= BishopMagics[sq]
	occ >>= 64 - bishopRelevantBits[sq]
	return BishopAttacks[sq][occ]
}

func maskBishopAttacks(square int) uint64 {
	var attacks uint64 = 0
	var f, r int

	tr := square / 8
	tf := square % 8

	for r, f = tr+1, tf+1; r <= 6 && f <= 6; r, f = r+1, f+1 {
		attacks |= (1 << (r*8 + f))
	}
	for r, f = tr+1, tf-1; r <= 6 && f >= 1; r, f = r+1, f-1 {
		attacks |= (1 << (r*8 + f))
	}
	for r, f = tr-1, tf+1; r >= 1 && f <= 6; r, f = r-1, f+1 {
		attacks |= (1 << (r*8 + f))
	}
	for r, f = tr-1, tf-1; r >= 1 && f >= 1; r, f = r-1, f-1 {
		attacks |= (1 << (r*8 + f))
	}

	return attacks
}
func maskRookAttacks(square int) uint64 {
	var attacks uint64 = 0

	tr := square / 8
	tf := square % 8

	for r := tr + 1; r <= 6; r++ {
		attacks |= (1 << (r*8 + tf))
	}
	for r := tr - 1; r >= 1; r-- {
		attacks |= (1 << (r*8 + tf))
	}
	for f := tf + 1; f <= 6; f++ {
		attacks |= (1 << (tr*8 + f))
	}
	for f := tf - 1; f >= 1; f-- {
		attacks |= (1 << (tr*8 + f))
	}

	return attacks
}

func Init_sliders_attacks() {

	for square := 0; square < 64; square++ {
		RookMask[square] = maskRookAttacks(square)
		BishopMask[square] = maskBishopAttacks(square)

		mask := maskRookAttacks(square)
		maskB := maskBishopAttacks(square)

		bitCount := countBits(mask)
		bitCountB := countBits(maskB)

		occupancyVariations := 1 << bitCount
		occupancyVariationsB := 1 << bitCountB

		for count := 0; count < occupancyVariations; count++ {
			occupancy := setOccupancy(count, bitCount, mask)
			magic_index := occupancy * RookMagics[square] >> (64 - uint64(rookRelevantBits[square]))
			RookAttacks[square][magic_index] = rookAttacksOnTheFly(square, occupancy)
		}

		for count := 0; count < occupancyVariationsB; count++ {
			occupancy := setOccupancy(count, bitCountB, maskB)
			magic_index := occupancy * BishopMagics[square] >> (64 - uint64(bishopRelevantBits[square]))
			BishopAttacks[square][magic_index] = bishopAttacksOnTheFly(square, occupancy)
		}

	}
}

func setOccupancy(index, bitsInMask int, attackMask uint64) uint64 {
	var occupancy uint64
	for count := 0; count < bitsInMask; count++ {
		square := getLs1bIndex(attackMask)
		attackMask &= attackMask - 1
		if index&(1<<count) != 0 {
			occupancy |= (1 << square)
		}
	}
	return occupancy
}

func countBits(b uint64) int {
	r := 0
	for ; b > 0; r++ {
		b &= b - 1
	}
	return r
}
func bishopAttacksOnTheFly(square int, block uint64) uint64 {
	var attacks uint64 = 0

	var f, r int

	tr := square / 8
	tf := square % 8

	for r, f = tr+1, tf+1; r <= 7 && f <= 7; r, f = r+1, f+1 {
		attacks |= (uint64(1) << (r*8 + f))
		if block&(uint64(1)<<(r*8+f)) != 0 {
			break
		}
	}

	for r, f = tr+1, tf-1; r <= 7 && f >= 0; r, f = r+1, f-1 {
		attacks |= (uint64(1) << (r*8 + f))
		if block&(uint64(1)<<(r*8+f)) != 0 {
			break
		}
	}

	for r, f = tr-1, tf+1; r >= 0 && f <= 7; r, f = r-1, f+1 {
		attacks |= (uint64(1) << (r*8 + f))
		if block&(uint64(1)<<(r*8+f)) != 0 {
			break
		}
	}

	for r, f = tr-1, tf-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		attacks |= (uint64(1) << (r*8 + f))
		if block&(uint64(1)<<(r*8+f)) != 0 {
			break
		}
	}

	return attacks
}

func rookAttacksOnTheFly(square int, block uint64) uint64 {
	var attacks uint64 = 0
	var f, r int

	tr := square / 8
	tf := square % 8

	for r = tr + 1; r <= 7; r++ {
		attacks |= (uint64(1) << (r*8 + tf))
		if (block & (uint64(1) << (r*8 + tf))) != 0 {
			break
		}
	}

	for r = tr - 1; r >= 0; r-- {
		attacks |= (uint64(1) << (r*8 + tf))
		if (block & (uint64(1) << (r*8 + tf))) != 0 {
			break
		}
	}

	for f = tf + 1; f <= 7; f++ {
		attacks |= (uint64(1) << (tr*8 + f))
		if (block & (uint64(1) << (tr*8 + f))) != 0 {
			break
		}
	}

	for f = tf - 1; f >= 0; f-- {
		attacks |= (uint64(1) << (tr*8 + f))
		if (block & (uint64(1) << (tr*8 + f))) != 0 {
			break
		}
	}

	return attacks
}

func getLs1bIndex(bitboard uint64) int {
	if bitboard != 0 {
		return countBits((bitboard & -bitboard) - 1)
	}
	return -1
}
