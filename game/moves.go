// Package game provides move parsing functionality for chess notation.
package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// MoveParser handles parsing of chess moves from various notations.
type MoveParser struct {
	isWhiteToMove bool
}

// NewMoveParser creates a new move parser for the specified player.
func NewMoveParser(isWhite bool) *MoveParser {
	return &MoveParser{isWhiteToMove: isWhite}
}

// SetCurrentPlayer updates the current player for move parsing.
func (mp *MoveParser) SetCurrentPlayer(isWhite bool) {
	mp.isWhiteToMove = isWhite
}

// ParseMove parses a move from string notation into a board.Move.
func (mp *MoveParser) ParseMove(notation string, gameBoard *board.Board) (board.Move, error) {
	notation = strings.TrimSpace(notation)
	notation = strings.ToLower(notation)

	if notation == "quit" || notation == "exit" {
		return board.Move{}, errors.New("QUIT")
	}
	if notation == "reset" {
		return board.Move{}, errors.New("RESET")
	}
	if notation == "fen" {
		return board.Move{}, errors.New("FEN")
	}
	if notation == "moves" {
		return board.Move{}, errors.New("MOVES")
	}

	if notation == "o-o" || notation == "0-0" {
		return mp.parseCastling(true, gameBoard)
	}
	if notation == "o-o-o" || notation == "0-0-0" {
		return mp.parseCastling(false, gameBoard)
	}

	if len(notation) >= 4 && len(notation) <= 5 {
		return mp.parseCoordinateMove(notation, gameBoard)
	}

	return mp.parseAlgebraicMove(notation, gameBoard)
}

func (mp *MoveParser) parseCastling(kingside bool, _ *board.Board) (board.Move, error) {
	var fromSquare, toSquare board.Square

	if mp.isWhiteToMove {
		fromSquare = board.Square{File: 4, Rank: 0}
		if kingside {
			toSquare = board.Square{File: 6, Rank: 0}
		} else {
			toSquare = board.Square{File: 2, Rank: 0}
		}
	} else {
		fromSquare = board.Square{File: 4, Rank: 7}
		if kingside {
			toSquare = board.Square{File: 6, Rank: 7}
		} else {
			toSquare = board.Square{File: 2, Rank: 7}
		}
	}

	return board.Move{
		From:       fromSquare,
		To:         toSquare,
		IsCastling: true,
		Promotion:  board.Empty,
	}, nil
}

func (mp *MoveParser) parseCoordinateMove(notation string, gameBoard *board.Board) (board.Move, error) {
	from, err := board.ParseSquare(notation[:2])
	if err != nil {
		return board.Move{}, fmt.Errorf("failed to parse from square: %w", err)
	}

	to, err := board.ParseSquare(notation[2:4])
	if err != nil {
		return board.Move{}, fmt.Errorf("failed to parse to square: %w", err)
	}

	piece := gameBoard.GetPiece(from.Rank, from.File)
	if piece == board.Empty {
		return board.Move{}, fmt.Errorf("no piece at square %s", notation[:2])
	}

	move := board.Move{From: from, To: to, Piece: piece, Promotion: board.Empty}

	if len(notation) == 5 {
		promotionChar := notation[4]
		promotion, err := mp.charToPiece(promotionChar)
		if err != nil {
			return board.Move{}, fmt.Errorf("failed to parse promotion: %w", err)
		}
		move.Promotion = promotion
	}

	return move, nil
}

func (mp *MoveParser) parseAlgebraicMove(notation string, gameBoard *board.Board) (board.Move, error) {
	notation = strings.TrimSuffix(notation, "+")
	notation = strings.TrimSuffix(notation, "#")

	if len(notation) >= 2 && notation[0] >= 'a' && notation[0] <= 'h' {
		return mp.parsePawnMove(notation, gameBoard)
	}

	if len(notation) >= 3 {
		return mp.parsePieceMove(notation, gameBoard)
	}

	return board.Move{}, fmt.Errorf("unsupported algebraic notation: %s", notation)
}

func (mp *MoveParser) parsePawnMove(_ string, _ *board.Board) (board.Move, error) {
	return board.Move{}, errors.New("algebraic notation not fully implemented - use coordinate notation (e.g., e2e4)")
}

func (mp *MoveParser) parsePieceMove(_ string, _ *board.Board) (board.Move, error) {
	return board.Move{}, errors.New("algebraic notation not fully implemented - use coordinate notation (e.g., e2e4)")
}

func (mp *MoveParser) charToPiece(char byte) (board.Piece, error) {
	switch char {
	case 'q':
		if mp.isWhiteToMove {
			return board.WhiteQueen, nil
		}
		return board.BlackQueen, nil
	case 'r':
		if mp.isWhiteToMove {
			return board.WhiteRook, nil
		}
		return board.BlackRook, nil
	case 'b':
		if mp.isWhiteToMove {
			return board.WhiteBishop, nil
		}
		return board.BlackBishop, nil
	case 'n':
		if mp.isWhiteToMove {
			return board.WhiteKnight, nil
		}
		return board.BlackKnight, nil
	default:
		return board.Empty, fmt.Errorf("invalid promotion piece: %c", char)
	}
}
