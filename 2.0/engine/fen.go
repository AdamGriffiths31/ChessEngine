package engine

import (
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

// ParseFen updates the Position with data from the fen string
func (p *Position) ParseFen(fen string) {
	p.resetPosition()
	parts := strings.Fields(fen)
	p.Board = generateBitboardFromFen(parts[0])
	p.Side = determineSideToPlay(parts[1])
	p.CastlePermission = parseCastlingAvailability(parts[2])
	p.EnPassant = parseEnPassantTarget(parts[3])

	p.PositionKey = p.GeneratePositionKey()
}

// resetPosition clears the Position down to default
func (p *Position) resetPosition() {
	p.Board = Bitboard{}
	p.Play = 0
	p.CastlePermission = 0
	p.EnPassant = 0
	p.PositionKey = 0
}

// parseEnPassantTarget determines the En Passant square
func parseEnPassantTarget(fen string) int {
	if fen[0] == '-' || len(fen) == 1 {
		return data.Empty
	}
	if sq, ok := data.NameToSquareMap[fen]; ok {
		return sq
	}
	return data.Empty
}

// parseCastlingAvailability determines the castling rights for
// the given fen
func parseCastlingAvailability(fen string) int {
	result := 0
	for i, ch := range fen {
		if ch == 'K' {
			result |= data.WhiteKingCastle
		} else if ch == 'Q' {
			result |= data.WhiteQueenCastle
		} else if ch == 'k' {
			result |= data.BlackKingCastle
		} else if ch == 'q' {
			result |= data.BlackQueenCastle
		} else if ch == '-' && i == len(fen)-1 {
			break
		}
	}

	return result
}

// determineSideToPlay checks fen for either white or black to play
func determineSideToPlay(fen string) int {
	if fen == "w" {
		return data.White
	} else {
		return data.Black
	}
}

// generateBitboardFromFen populates the Bitboard bitboards for a
// given fen
func generateBitboardFromFen(fen string) Bitboard {
	board := Bitboard{}
	rank, file := data.Rank8, data.FileA
	for _, ch := range fen {
		if ch == '/' {
			rank--
			file = data.FileA
		} else if ch == ' ' {
			break
		} else if '1' <= ch && ch <= '8' {
			file += int(ch - '0')
		} else {
			piece := getPieceType(ch)
			sq64 := rank*8 + file
			board.SetPieceAtSquare(sq64, piece)
			file++
		}
	}

	return board
}

// getPieceType returns returns the corresponding piece type integer
func getPieceType(ch rune) int {
	pieceMap := map[rune]int{
		'p': BP,
		'r': BR,
		'n': BN,
		'b': BB,
		'q': BQ,
		'k': BK,
		'P': WP,
		'R': WR,
		'N': WN,
		'B': WB,
		'Q': WQ,
		'K': WK,
	}
	piece, ok := pieceMap[ch]
	if !ok {
		panic(fmt.Errorf("getPieceType: could not find value for %v", ch))
	}

	return piece
}
