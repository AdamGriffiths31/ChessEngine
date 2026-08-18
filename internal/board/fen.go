package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// FromFEN creates a new board from FEN (Forsyth-Edwards Notation) string
//
//nolint:gocyclo // refactored in Phase 4
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
				emptySquares, err := strconv.Atoi(string(char))
				if err != nil {
					return nil, fmt.Errorf("invalid FEN: failed to parse empty squares count: %w", err)
				}
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

	if len(parts) >= 2 {
		board.sideToMove = parts[1]
	}
	if len(parts) >= 3 {
		board.castlingRights = parts[2]
	}
	if len(parts) >= 4 {
		enPassantStr := parts[3]
		if enPassantStr != "-" {
			if len(enPassantStr) != 2 {
				return nil, errors.New("invalid FEN: invalid en passant square")
			}
			file := int(enPassantStr[0] - 'a')
			rank := int(enPassantStr[1] - '1')
			// A legal en passant target is always on rank 3 or rank 6
			// (index 2 or 5), the square a pawn lands on behind a double push.
			if file < 0 || file > 7 || (rank != 2 && rank != 5) {
				return nil, errors.New("invalid FEN: invalid en passant square")
			}
			board.enPassantSquare = Square{File: file, Rank: rank}
			board.hasEnPassant = true
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

	board.InitializeEvalScoresFromPosition()

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
