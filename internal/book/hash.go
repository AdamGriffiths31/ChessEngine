package book

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

var polyglotZobrist ZobristHash

func init() {
	for piece := range 12 {
		for square := range 64 {
			polyglotZobrist.pieceKeys[square][piece] = random64Poly[piece*64+square]
		}
	}
	copy(polyglotZobrist.castleKeys[:], random64Poly[768:])
	copy(polyglotZobrist.enPassantKeys[:], random64Poly[772:])
	polyglotZobrist.sideKey = random64Poly[780]
}

// ZobristHash generates a Zobrist hash for a chess position
// Compatible with Polyglot opening book format
type ZobristHash struct {
	pieceKeys     [64][12]uint64
	castleKeys    [4]uint64
	enPassantKeys [8]uint64
	sideKey       uint64
}

// GetPolyglotHash returns the Zobrist hash instance for Polyglot compatibility
func GetPolyglotHash() *ZobristHash {
	return &polyglotZobrist
}

// HashPosition generates a Zobrist hash for the given board position
// Compatible with standard Polyglot opening book format
func (zh *ZobristHash) HashPosition(b *board.Board) uint64 {
	var hash uint64

	for square, piece := range &b.Mailbox {
		if piece != board.Empty {
			hash ^= zh.pieceKeys[square][zh.GetPieceIndex(piece)]
		}
	}

	hash ^= zh.GetCastlingKey(b.GetCastlingRights())

	if b.HasEnPassantCapturer() {
		ep, _ := b.GetEnPassantTarget()
		hash ^= zh.enPassantKeys[ep.File]
	}

	if b.GetSideToMove() == "w" {
		hash ^= zh.sideKey
	}

	return hash
}

// GetPieceIndex returns the piece index for zobrist hashing.
// Official Polyglot order: BP(0), WP(1), BN(2), WN(3), BB(4), WB(5), BR(6), WR(7), BQ(8), WQ(9), BK(10), WK(11)
func (zh *ZobristHash) GetPieceIndex(piece board.Piece) int {
	switch piece {
	case board.BlackPawn:
		return 0
	case board.WhitePawn:
		return 1
	case board.BlackKnight:
		return 2
	case board.WhiteKnight:
		return 3
	case board.BlackBishop:
		return 4
	case board.WhiteBishop:
		return 5
	case board.BlackRook:
		return 6
	case board.WhiteRook:
		return 7
	case board.BlackQueen:
		return 8
	case board.WhiteQueen:
		return 9
	case board.BlackKing:
		return 10
	case board.WhiteKing:
		return 11
	default:
		return 0
	}
}

// GetSideKey returns the zobrist key for side to move
func (zh *ZobristHash) GetSideKey() uint64 {
	return zh.sideKey
}

// GetPieceKey returns the zobrist key for a piece at a given square
func (zh *ZobristHash) GetPieceKey(square int, pieceIndex int) uint64 {
	return zh.pieceKeys[square][pieceIndex]
}

// GetCastlingKey returns the combined zobrist key for a castling-rights
// string: the XOR of the individual per-right keys, exactly as HashPosition
// accumulates them (castleKeys[0]='K', [1]='Q', [2]='k', [3]='q'). XORing the
// values for an old and a new rights string therefore yields the incremental
// delta for the rights that changed. Returns 0 for no rights.
func (zh *ZobristHash) GetCastlingKey(castlingRights string) uint64 {
	var key uint64
	for _, right := range castlingRights {
		switch right {
		case 'K':
			key ^= zh.castleKeys[0]
		case 'Q':
			key ^= zh.castleKeys[1]
		case 'k':
			key ^= zh.castleKeys[2]
		case 'q':
			key ^= zh.castleKeys[3]
		}
	}
	return key
}

// GetEnPassantKey returns the zobrist key for en passant file
func (zh *ZobristHash) GetEnPassantKey(file int) uint64 {
	return zh.enPassantKeys[file]
}
