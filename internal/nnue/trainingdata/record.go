// Package trainingdata stores NNUE training positions in a packed binary
// format of RecordSize bytes each.
//
// A record holds the occupancy bitboard (bytes 0-7), one 4-bit piece index
// per occupied square in ascending square order (bytes 8-23), the search
// score in centipawns from White's side (bytes 24-25), the game result from
// White's side as 0, 1 or 2 for a loss, draw or win (byte 26) and the side to
// move (byte 27, 1 for Black). All integers are little-endian.
package trainingdata

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/nnue"
)

// RecordSize is the size of one packed position in bytes.
const RecordSize = 32

// Game results from White's side, as written in datagen text lines.
const (
	ResultBlackWins = "0.0"
	ResultDraw      = "0.5"
	ResultWhiteWins = "1.0"
)

const (
	maxPieces    = 32
	scoreOffset  = 24
	resultOffset = 26
	sideOffset   = 27
)

// Record is one packed training position.
type Record [RecordSize]byte

var pieces = func() (pieces [12]board.Piece) {
	for _, c := range "PNBRQKpnbrqk" {
		pieces[board.PieceToBitboardIndex(board.Piece(c))] = board.Piece(c)
	}
	return pieces
}()

// FormatLine returns the datagen text line for a position, newline included:
// the FEN, the search score from White's side and one of the Result values.
func FormatLine(fen string, whiteScore int, result string) string {
	return fmt.Sprintf("%s | %d | %s\n", fen, whiteScore, result)
}

func parseLine(line string) (Record, error) {
	var r Record
	parts := strings.Split(line, " | ")
	if len(parts) != 3 {
		return r, fmt.Errorf("want 3 fields separated by \" | \", got %d", len(parts))
	}
	score, err := strconv.Atoi(parts[1])
	if err != nil {
		return r, fmt.Errorf("score: %w", err)
	}
	var result byte
	switch parts[2] {
	case ResultBlackWins:
		result = 0
	case ResultDraw:
		result = 1
	case ResultWhiteWins:
		result = 2
	default:
		return r, fmt.Errorf("result %q is not %s, %s or %s", parts[2], ResultBlackWins, ResultDraw, ResultWhiteWins)
	}
	fields := strings.Fields(parts[0])
	if len(fields) < 2 {
		return r, fmt.Errorf("FEN %q has no side to move", parts[0])
	}
	if err := r.setPlacement(fields[0]); err != nil {
		return r, err
	}

	binary.LittleEndian.PutUint16(r[scoreOffset:], uint16(int16(max(math.MinInt16, min(math.MaxInt16, score)))))
	r[resultOffset] = result
	if fields[1] == "b" {
		r[sideOffset] = 1
	}
	return r, nil
}

func (r *Record) setPlacement(placement string) error {
	var occupancy board.Bitboard
	var squares [64]byte
	rank, file := 7, 0
	for i := 0; i < len(placement); i++ {
		c := placement[i]
		switch {
		case c == '/':
			rank, file = rank-1, 0
		case c >= '1' && c <= '8':
			file += int(c - '0')
		default:
			piece := board.PieceToBitboardIndex(board.Piece(c))
			if piece < 0 || rank < 0 || file > 7 {
				return fmt.Errorf("bad placement %q", placement)
			}
			occupancy = occupancy.SetBit(rank*8 + file)
			squares[rank*8+file] = byte(piece)
			file++
		}
	}
	if occupancy.PopCount() > maxPieces {
		return fmt.Errorf("placement %q has more than %d pieces", placement, maxPieces)
	}

	binary.LittleEndian.PutUint64(r[0:], uint64(occupancy))
	for k, rest := 0, occupancy; rest != 0; k++ {
		var square int
		square, rest = rest.PopLSB()
		r[8+k/2] |= squares[square] << (4 * (k % 2))
	}
	return nil
}

// BlackToMove reports whether Black is to move.
func (r *Record) BlackToMove() bool { return r[sideOffset] == 1 }

// Score returns the search score in centipawns from White's side.
func (r *Record) Score() int {
	return int(int16(binary.LittleEndian.Uint16(r[scoreOffset:])))
}

// Result returns the game result from White's side: 0, 0.5 or 1.
func (r *Record) Result() float64 { return float64(r[resultOffset]) / 2 }

// AppendFeatures appends the network inputs of every piece from White's
// perspective.
func (r *Record) AppendFeatures(dst []uint16) []uint16 {
	for k, rest := 0, r.occupancy(); rest != 0; k++ {
		var square int
		square, rest = rest.PopLSB()
		dst = append(dst, uint16(nnue.FeatureIndex(r.piece(k), square, 0)))
	}
	return dst
}

// Board returns the position with no castling rights or en passant square,
// which the network does not see.
func (r *Record) Board() *board.Board {
	b := board.NewBoard()
	for k, rest := 0, r.occupancy(); rest != 0; k++ {
		var square int
		square, rest = rest.PopLSB()
		b.SetPiece(square/8, square%8, pieces[r.piece(k)])
	}
	if r.BlackToMove() {
		b.SetSideToMove("b")
	}
	return b
}

func (r *Record) occupancy() board.Bitboard {
	return board.Bitboard(binary.LittleEndian.Uint64(r[0:]))
}

func (r *Record) piece(k int) int { return int(r[8+k/2]>>(4*(k%2))) & 0xF }
