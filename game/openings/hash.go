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

	// 3. Hash en passant file (if any)
	if b.GetEnPassantTarget() != nil {
		file := b.GetEnPassantTarget().File
		hash ^= zh.enPassantKeys[file] // Use the array
	}

	// 4. Hash side to move - XOR only when BLACK to move
	if b.GetSideToMove() == "b" {
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

// getCastlingIndex converts castling rights to array index
// Polyglot format: bit 0=white O-O, bit 1=white O-O-O, bit 2=black O-O, bit 3=black O-O-O
func (zh *ZobristHash) getCastlingIndex(b *board.Board) int {
	var index int
	castlingRights := b.GetCastlingRights()

	// Convert "KQkq" format to bit indices
	for _, r := range castlingRights {
		switch r {
		case 'K': // White kingside
			index |= 1
		case 'Q': // White queenside
			index |= 2
		case 'k': // Black kingside
			index |= 4
		case 'q': // Black queenside
			index |= 8
		}
	}

	return index
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
