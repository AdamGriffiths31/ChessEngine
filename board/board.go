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

type Board struct {
	squares [8][8]Piece
}

func NewBoard() *Board {
	board := &Board{}
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