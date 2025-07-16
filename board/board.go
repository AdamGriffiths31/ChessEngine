package board

import (
	"errors"
	"strconv"
	"strings"
)

type Piece rune

const (
	Empty       Piece = '.'
	WhitePawn   Piece = 'P'
	WhiteRook   Piece = 'R'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
	BlackPawn   Piece = 'p'
	BlackRook   Piece = 'r'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
)

type Square struct {
	File int // 0-7 (a-h)
	Rank int // 0-7 (1-8)
}

type Board struct {
	squares      [8][8]Piece
	castlingRights string // KQkq format
	enPassantTarget *Square // nil if no en passant
	halfMoveClock   int
	fullMoveNumber  int
	sideToMove      string // "w" or "b"
}

func NewBoard() *Board {
	board := &Board{
		castlingRights: "KQkq",
		enPassantTarget: nil,
		halfMoveClock: 0,
		fullMoveNumber: 1,
		sideToMove: "w",
	}
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			board.squares[rank][file] = Empty
		}
	}
	return board
}

func (b *Board) GetPiece(rank, file int) Piece {
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return Empty
	}
	return b.squares[rank][file]
}

func (b *Board) SetPiece(rank, file int, piece Piece) {
	if rank >= 0 && rank <= 7 && file >= 0 && file <= 7 {
		b.squares[rank][file] = piece
	}
}

func FromFEN(fen string) (*Board, error) {
	if fen == "" {
		return nil, errors.New("invalid FEN: missing board position")
	}
	
	parts := strings.Split(fen, " ")
	if len(parts) < 1 {
		return nil, errors.New("invalid FEN: missing board position")
	}

	boardPart := parts[0]
	ranks := strings.Split(boardPart, "/")
	
	if len(ranks) != 8 {
		return nil, errors.New("invalid FEN: must have exactly 8 ranks")
	}

	board := NewBoard()

	// Parse board position
	for rankIndex, rankStr := range ranks {
		// FEN ranks start from 8 (top) and go down to 1 (bottom)
		// Array index 0 should be rank 1, index 7 should be rank 8
		// So FEN rank 8 (rankIndex 0) goes to array index 7
		actualRank := 7 - rankIndex
		file := 0
		for _, char := range rankStr {
			if file >= 8 {
				return nil, errors.New("invalid FEN: too many files in rank")
			}

			if char >= '1' && char <= '8' {
				emptySquares, _ := strconv.Atoi(string(char))
				for i := 0; i < emptySquares; i++ {
					if file >= 8 {
						return nil, errors.New("invalid FEN: too many files in rank")
					}
					board.SetPiece(actualRank, file, Empty)
					file++
				}
			} else {
				piece := Piece(char)
				if !isValidPiece(piece) {
					return nil, errors.New("invalid FEN: invalid piece character")
				}
				board.SetPiece(actualRank, file, piece)
				file++
			}
		}
		
		if file != 8 {
			return nil, errors.New("invalid FEN: incorrect number of files in rank")
		}
	}

	// Parse additional FEN fields if available
	if len(parts) >= 2 {
		board.sideToMove = parts[1]
	}
	if len(parts) >= 3 {
		board.castlingRights = parts[2]
	}
	if len(parts) >= 4 {
		enPassantStr := parts[3]
		if enPassantStr != "-" {
			file := int(enPassantStr[0] - 'a')
			rank := int(enPassantStr[1] - '1')
			if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
				square := Square{File: file, Rank: rank}
				board.enPassantTarget = &square
			}
		}
	}
	if len(parts) >= 5 {
		if halfMove, err := strconv.Atoi(parts[4]); err == nil {
			board.halfMoveClock = halfMove
		}
	}
	if len(parts) >= 6 {
		if fullMove, err := strconv.Atoi(parts[5]); err == nil {
			board.fullMoveNumber = fullMove
		}
	}

	return board, nil
}

func isValidPiece(piece Piece) bool {
	validPieces := []Piece{
		WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
	}
	
	for _, validPiece := range validPieces {
		if piece == validPiece {
			return true
		}
	}
	return false
}

// Getter methods for board state
func (b *Board) GetCastlingRights() string {
	return b.castlingRights
}

func (b *Board) GetEnPassantTarget() *Square {
	return b.enPassantTarget
}

func (b *Board) GetHalfMoveClock() int {
	return b.halfMoveClock
}

func (b *Board) GetFullMoveNumber() int {
	return b.fullMoveNumber
}

func (b *Board) GetSideToMove() string {
	return b.sideToMove
}

// Setter methods for board state
func (b *Board) SetCastlingRights(rights string) {
	b.castlingRights = rights
}

func (b *Board) SetEnPassantTarget(target *Square) {
	b.enPassantTarget = target
}

func (b *Board) SetHalfMoveClock(clock int) {
	b.halfMoveClock = clock
}

func (b *Board) SetFullMoveNumber(num int) {
	b.fullMoveNumber = num
}

func (b *Board) SetSideToMove(side string) {
	b.sideToMove = side
}