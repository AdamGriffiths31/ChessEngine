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
// Uses the same calculation method as polyBook.go for compatibility
func (zh *ZobristHash) HashPosition(b *board.Board) uint64 {
	var hash uint64
	
	// Hash pieces on the board - use polyBook.go method: (64*polyPiece)+(8*rank)+file
	for square := 0; square < 64; square++ {
		file := square % 8
		rank := square / 8
		piece := b.GetPiece(rank, file)
		
		if piece != board.Empty {
			polyPiece := zh.getPieceIndex(piece)
			// Use the same indexing as polyBook.go: (64*polyPiece)+(8*rank)+file
			index := (64*polyPiece) + (8*rank) + file
			if index < 768 { // Ensure we don't exceed piece key range
				hash ^= random64Poly[index]
			}
		}
	}
	
	// Hash castling rights - use individual flags like polyBook.go (offset 768)
	castlingRights := b.GetCastlingRights()
	offset := 768
	for _, r := range castlingRights {
		switch r {
		case 'K': // White kingside
			hash ^= random64Poly[offset+0]
		case 'Q': // White queenside  
			hash ^= random64Poly[offset+1]
		case 'k': // Black kingside
			hash ^= random64Poly[offset+2]
		case 'q': // Black queenside
			hash ^= random64Poly[offset+3]
		}
	}
	
	// Hash en passant target - use offset 772 like polyBook.go
	if b.GetEnPassantTarget() != nil {
		file := b.GetEnPassantTarget().File
		offset = 772
		hash ^= random64Poly[offset+file]
	}
	
	// Hash side to move - only when side is White (offset 780)
	sideToMove := b.GetSideToMove()
	if sideToMove == "w" {
		offset = 780
		hash ^= random64Poly[offset]
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