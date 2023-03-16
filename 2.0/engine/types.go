package engine

type Position struct {
	Board            Bitboard
	Play             int
	PositionKey      uint64
	Side             int
	CastlePermission int
	EnPassant        int
	FailHighFirst    float32
	FailHigh         float32
	MoveHistory      MoveHistory
	CurrentScore     int
	FiftyMove        int
}

type Bitboard struct {
	Pieces uint64
	//Black
	BlackPieces uint64
	BlackPawn   uint64
	BlackKnight uint64
	BlackBishop uint64
	BlackRook   uint64
	BlackQueen  uint64
	BlackKing   uint64
	//White
	WhitePieces uint64
	WhitePawn   uint64
	WhiteKnight uint64
	WhiteBishop uint64
	WhiteRook   uint64
	WhiteQueen  uint64
	WhiteKing   uint64
}

type MoveHistory struct {
	Killers [2][64]int
	History [13][120]int
}
