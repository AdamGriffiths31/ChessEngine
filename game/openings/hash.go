package openings

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// ZobristHash generates a Zobrist hash for a chess position
// Compatible with Polyglot opening book format
type ZobristHash struct {
	pieceKeys     [64][12]uint64 // 64 squares, 12 piece types
	castleKeys    [16]uint64     // 16 castling combinations
	enPassantKeys [8]uint64      // 8 files for en passant
	sideKey       uint64         // side to move
}

// Standard Polyglot Zobrist keys (initialized in init function)
var polyglotZobrist *ZobristHash

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

	// 1. Hash pieces on the board
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				polyPiece := zh.getPieceIndex(piece)
				square := rank*8 + file
				// Use the pre-initialized piece keys array
				hash ^= zh.pieceKeys[square][polyPiece]
			}
		}
	}

	// 2. Hash castling rights - XOR each individual right
	castlingRights := b.GetCastlingRights()
	for _, r := range castlingRights {
		switch r {
		case 'K': // White kingside
			hash ^= zh.castleKeys[0]
		case 'Q': // White queenside
			hash ^= zh.castleKeys[1]
		case 'k': // Black kingside
			hash ^= zh.castleKeys[2]
		case 'q': // Black queenside
			hash ^= zh.castleKeys[3]
		}
	}

	// 3. Hash en passant file (Polyglot: only if a pawn can actually capture)
	ep := b.GetEnPassantTarget()
	if ep != nil {
		epFile := ep.File
		side := b.GetSideToMove()
		var pawnRank int
		var pawnPiece board.Piece
		if side == "w" {
			pawnRank = 4 // White pawns on rank 5 (0-based)
			pawnPiece = board.WhitePawn
		} else {
			pawnRank = 3 // Black pawns on rank 4 (0-based)
			pawnPiece = board.BlackPawn
		}
		// Check for pawn on adjacent file
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

	// 4. Hash side to move - XOR only when White to move
	if b.GetSideToMove() == "w" {
		hash ^= zh.sideKey // Use the field
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
		return 0 // Should not happen
	}
}

// HashMove generates a hash for a specific move (for move validation)
func (zh *ZobristHash) HashMove(move board.Move) uint64 {
	// Simple hash for move validation - not part of position hash
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
