// Package openings provides Zobrist hashing functionality for chess positions compatible with Polyglot opening book format.
package openings

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Standard Polyglot Zobrist keys (initialized in init function)
var polyglotZobrist *ZobristHash

// ZobristHash generates a Zobrist hash for a chess position
// Compatible with Polyglot opening book format
type ZobristHash struct {
	pieceKeys     [64][12]uint64
	castleKeys    [16]uint64
	enPassantKeys [8]uint64
	sideKey       uint64
}

// GetPolyglotHash returns the Zobrist hash instance for Polyglot compatibility
func GetPolyglotHash() *ZobristHash {
	if polyglotZobrist == nil {
		polyglotZobrist = &ZobristHash{
			pieceKeys:     officialPolyglotPieceKeys,
			castleKeys:    officialPolyglotCastlingKeys,
			enPassantKeys: officialPolyglotEnPassantKeys,
			sideKey:       officialPolyglotSideKey,
		}
	}
	return polyglotZobrist
}

// HashPosition generates a Zobrist hash for the given board position
// Compatible with standard Polyglot opening book format
func (zh *ZobristHash) HashPosition(b *board.Board) uint64 {
	var hash uint64

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				polyPiece := zh.getPieceIndex(piece)
				square := rank*8 + file
				hash ^= zh.pieceKeys[square][polyPiece]
			}
		}
	}

	castlingRights := b.GetCastlingRights()
	for _, r := range castlingRights {
		switch r {
		case 'K':
			hash ^= zh.castleKeys[0]
		case 'Q':
			hash ^= zh.castleKeys[1]
		case 'k':
			hash ^= zh.castleKeys[2]
		case 'q':
			hash ^= zh.castleKeys[3]
		}
	}

	ep := b.GetEnPassantTarget()
	if ep != nil {
		epFile := ep.File
		side := b.GetSideToMove()
		var pawnRank int
		var pawnPiece board.Piece
		if side == "w" {
			pawnRank = 4
			pawnPiece = board.WhitePawn
		} else {
			pawnRank = 3
			pawnPiece = board.BlackPawn
		}
		for _, df := range []int{-1, 1} {
			adjFile := epFile + df
			if adjFile >= 0 && adjFile < 8 {
				if b.GetPiece(pawnRank, adjFile) == pawnPiece {
					hash ^= zh.enPassantKeys[epFile]
					break
				}
			}
		}
	}

	if b.GetSideToMove() == "w" {
		hash ^= zh.sideKey
	}

	return hash
}

// getPieceIndex converts a board piece to Zobrist array index
// Official Polyglot order: BP(0), WP(1), BN(2), WN(3), BB(4), WB(5), BR(6), WR(7), BQ(8), WQ(9), BK(10), WK(11)
func (zh *ZobristHash) getPieceIndex(piece board.Piece) int {
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

// HashMove generates a hash for a specific move (for move validation)
func (zh *ZobristHash) HashMove(move board.Move) uint64 {
	return uint64(move.From.File) |
		(uint64(move.From.Rank) << 3) |
		(uint64(move.To.File) << 6) |
		(uint64(move.To.Rank) << 9) |
		(uint64(move.Promotion) << 12)
}

// GetSideKey returns the zobrist key for side to move
func (zh *ZobristHash) GetSideKey() uint64 {
	return zh.sideKey
}

// GetPieceKey returns the zobrist key for a piece at a given square
func (zh *ZobristHash) GetPieceKey(square int, pieceIndex int) uint64 {
	return zh.pieceKeys[square][pieceIndex]
}

// GetPieceIndex returns the piece index for zobrist hashing
func (zh *ZobristHash) GetPieceIndex(piece board.Piece) int {
	return zh.getPieceIndex(piece)
}

// GetCastlingKey returns the zobrist key for castling rights
func (zh *ZobristHash) GetCastlingKey(castlingRights string) uint64 {
	castlingIndex := 0
	for _, right := range castlingRights {
		switch right {
		case 'K':
			castlingIndex |= 1
		case 'Q':
			castlingIndex |= 2
		case 'k':
			castlingIndex |= 4
		case 'q':
			castlingIndex |= 8
		}
	}
	return zh.castleKeys[castlingIndex]
}

// GetEnPassantKey returns the zobrist key for en passant file
func (zh *ZobristHash) GetEnPassantKey(file int) uint64 {
	return zh.enPassantKeys[file]
}
