package openings

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// ZobristHash generates a Zobrist hash for a chess position
// Compatible with Polyglot opening book format
type ZobristHash struct {
	pieceKeys    [64][12]uint64 // 64 squares, 12 piece types
	castleKeys   [16]uint64     // 16 castling combinations
	enPassantKeys [8]uint64      // 8 files for en passant
	sideKey      uint64         // side to move
}

// Standard Polyglot Zobrist keys (hardcoded for compatibility)
var polyglotZobrist = &ZobristHash{
	pieceKeys: [64][12]uint64{
		// These would be the actual Polyglot Zobrist keys
		// For now using placeholder values - should be replaced with actual Polyglot keys
		{0x9D39247E33776D41, 0x2AF7398005AAA5C7, 0x44DB015024623547, 0x9C15F73E62A76AE2,
		 0x75834465489C0C89, 0x3290AC3A203001BF, 0x0FBBAD1F61042279, 0xE83A908FF2FB60CA,
		 0x0D7E765D58755C10, 0x1A083822CEAFE02D, 0x9C69A97284B578D7, 0x8D8BCA50E5F571DB},
		// ... (remaining 63 squares would have their 12 piece keys)
		// This is a simplified version - actual implementation would need all 768 keys
	},
	castleKeys: [16]uint64{
		0x0000000000000000, 0x0000000000000001, 0x0000000000000002, 0x0000000000000003,
		0x0000000000000004, 0x0000000000000005, 0x0000000000000006, 0x0000000000000007,
		0x0000000000000008, 0x0000000000000009, 0x000000000000000A, 0x000000000000000B,
		0x000000000000000C, 0x000000000000000D, 0x000000000000000E, 0x000000000000000F,
	},
	enPassantKeys: [8]uint64{
		0x0000000000000001, 0x0000000000000002, 0x0000000000000004, 0x0000000000000008,
		0x0000000000000010, 0x0000000000000020, 0x0000000000000040, 0x0000000000000080,
	},
	sideKey: 0xF8D626AAAF278509,
}

// GetPolyglotHash returns the Zobrist hash instance for Polyglot compatibility
func GetPolyglotHash() *ZobristHash {
	return polyglotZobrist
}

// HashPosition generates a Zobrist hash for the given board position
func (zh *ZobristHash) HashPosition(b *board.Board) uint64 {
	var hash uint64
	
	// Hash pieces on the board
	for square := 0; square < 64; square++ {
		file := square % 8
		rank := square / 8
		piece := b.GetPiece(rank, file)
		
		if piece != board.Empty {
			pieceIndex := zh.getPieceIndex(piece)
			hash ^= zh.pieceKeys[square][pieceIndex]
		}
	}
	
	// Hash castling rights
	castlingIndex := zh.getCastlingIndex(b)
	hash ^= zh.castleKeys[castlingIndex]
	
	// Hash en passant target
	if b.GetEnPassantTarget() != nil {
		file := b.GetEnPassantTarget().File
		hash ^= zh.enPassantKeys[file]
	}
	
	// Hash side to move
	if b.GetSideToMove() == "b" {
		hash ^= zh.sideKey
	}
	
	return hash
}

// getPieceIndex converts a board piece to Zobrist array index
// Polyglot order: WP, WN, WB, WR, WQ, WK, BP, BN, BB, BR, BQ, BK
func (zh *ZobristHash) getPieceIndex(piece board.Piece) int {
	switch piece {
	case board.WhitePawn:
		return 0
	case board.WhiteKnight:
		return 1
	case board.WhiteBishop:
		return 2
	case board.WhiteRook:
		return 3
	case board.WhiteQueen:
		return 4
	case board.WhiteKing:
		return 5
	case board.BlackPawn:
		return 6
	case board.BlackKnight:
		return 7
	case board.BlackBishop:
		return 8
	case board.BlackRook:
		return 9
	case board.BlackQueen:
		return 10
	case board.BlackKing:
		return 11
	default:
		return 0 // Should not happen
	}
}

// getCastlingIndex converts castling rights to array index
func (zh *ZobristHash) getCastlingIndex(b *board.Board) int {
	var index int
	
	// This would need to be implemented based on how the board stores castling rights
	// For now, returning 0 as placeholder
	// TODO: Implement proper castling rights indexing based on board representation
	
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