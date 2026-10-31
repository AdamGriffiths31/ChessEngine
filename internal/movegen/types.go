package movegen

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// Player represents a chess player (White or Black).
type Player int

const (
	// White represents the white player
	White Player = iota
	// Black represents the black player
	Black
)

func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

// MoveList represents a collection of chess moves with efficient storage.
// Should be obtained from GetMoveList() and released with ReleaseMoveList() for optimal performance.
type MoveList struct {
	Moves []board.Move
	Count int
}

// AddMove adds a move to the list
func (ml *MoveList) AddMove(move board.Move) {
	ml.Moves = append(ml.Moves, move)
	ml.Count++
}

// Clear empties the move list
func (ml *MoveList) Clear() {
	ml.Moves = ml.Moves[:0]
	ml.Count = 0
}

type castleWing struct {
	right  rune
	rookAt int
	kingTo int
	rookTo int
}

type sideInfo struct {
	color, enemy                            board.BitboardColor
	pawn, knight, bishop, rook, queen, king board.Piece
	promotions                              [4]board.Piece
	startRank, promotionRank, enPassantRank int
	push                                    int
	kingStart                               int
	kingside, queenside                     castleWing
}

var sides = [2]sideInfo{
	White: {
		color:         board.BitboardWhite,
		enemy:         board.BitboardBlack,
		pawn:          board.WhitePawn,
		knight:        board.WhiteKnight,
		bishop:        board.WhiteBishop,
		rook:          board.WhiteRook,
		queen:         board.WhiteQueen,
		king:          board.WhiteKing,
		promotions:    [4]board.Piece{board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight},
		startRank:     1,
		promotionRank: 7,
		enPassantRank: 4,
		push:          8,
		kingStart:     board.E1,
		kingside:      castleWing{right: 'K', rookAt: board.H1, kingTo: board.G1, rookTo: board.F1},
		queenside:     castleWing{right: 'Q', rookAt: board.A1, kingTo: board.C1, rookTo: board.D1},
	},
	Black: {
		color:         board.BitboardBlack,
		enemy:         board.BitboardWhite,
		pawn:          board.BlackPawn,
		knight:        board.BlackKnight,
		bishop:        board.BlackBishop,
		rook:          board.BlackRook,
		queen:         board.BlackQueen,
		king:          board.BlackKing,
		promotions:    [4]board.Piece{board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight},
		startRank:     6,
		promotionRank: 0,
		enPassantRank: 3,
		push:          -8,
		kingStart:     board.E8,
		kingside:      castleWing{right: 'k', rookAt: board.H8, kingTo: board.G8, rookTo: board.F8},
		queenside:     castleWing{right: 'q', rookAt: board.A8, kingTo: board.C8, rookTo: board.D8},
	},
}

func (s *sideInfo) pushPawns(pawns board.Bitboard) board.Bitboard {
	if s.color == board.BitboardWhite {
		return pawns.ShiftNorth()
	}
	return pawns.ShiftSouth()
}

func (s *sideInfo) pawnsAttacking(targets board.Bitboard) board.Bitboard {
	if s.color == board.BitboardWhite {
		return targets.ShiftSouthEast() | targets.ShiftSouthWest()
	}
	return targets.ShiftNorthEast() | targets.ShiftNorthWest()
}

func squareOf(square int) board.Square {
	return board.Square{File: square % 8, Rank: square / 8}
}
