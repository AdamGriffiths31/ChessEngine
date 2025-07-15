package game

import (
	"errors"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

type MoveParser struct {
	currentPlayer Player
}

func NewMoveParser(player Player) *MoveParser {
	return &MoveParser{currentPlayer: player}
}

func (mp *MoveParser) SetCurrentPlayer(player Player) {
	mp.currentPlayer = player
}

func (mp *MoveParser) ParseMove(notation string, gameBoard *board.Board) (board.Move, error) {
	notation = strings.TrimSpace(notation)
	notation = strings.ToLower(notation)
	
	// Handle special commands
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
	
	// Handle castling
	if notation == "o-o" || notation == "0-0" {
		return mp.parseCastling(true, gameBoard)
	}
	if notation == "o-o-o" || notation == "0-0-0" {
		return mp.parseCastling(false, gameBoard)
	}
	
	// Handle simple coordinate moves (e.g., "e2e4")
	if len(notation) >= 4 && len(notation) <= 5 {
		return mp.parseCoordinateMove(notation)
	}
	
	// Handle algebraic notation (basic implementation)
	return mp.parseAlgebraicMove(notation, gameBoard)
}

func (mp *MoveParser) parseCastling(kingside bool, gameBoard *board.Board) (board.Move, error) {
	var fromSquare, toSquare board.Square
	
	if mp.currentPlayer == White {
		fromSquare = board.Square{File: 4, Rank: 0} // e1
		if kingside {
			toSquare = board.Square{File: 6, Rank: 0} // g1
		} else {
			toSquare = board.Square{File: 2, Rank: 0} // c1
		}
	} else {
		fromSquare = board.Square{File: 4, Rank: 7} // e8
		if kingside {
			toSquare = board.Square{File: 6, Rank: 7} // g8
		} else {
			toSquare = board.Square{File: 2, Rank: 7} // c8
		}
	}
	
	return board.Move{
		From:       fromSquare,
		To:         toSquare,
		IsCastling: true,
		Promotion:  board.Empty,
	}, nil
}

func (mp *MoveParser) parseCoordinateMove(notation string) (board.Move, error) {
	// Handle moves like "e2e4" or "e7e8q" (with promotion)
	from, err := board.ParseSquare(notation[:2])
	if err != nil {
		return board.Move{}, err
	}
	
	to, err := board.ParseSquare(notation[2:4])
	if err != nil {
		return board.Move{}, err
	}
	
	move := board.Move{From: from, To: to, Promotion: board.Empty}
	
	// Handle promotion
	if len(notation) == 5 {
		promotionChar := notation[4]
		promotion, err := mp.charToPiece(promotionChar)
		if err != nil {
			return board.Move{}, err
		}
		move.Promotion = promotion
	}
	
	return move, nil
}

func (mp *MoveParser) parseAlgebraicMove(notation string, gameBoard *board.Board) (board.Move, error) {
	// Very basic algebraic notation parsing
	// This is a simplified implementation that handles basic cases
	
	// Remove check/checkmate indicators
	notation = strings.TrimSuffix(notation, "+")
	notation = strings.TrimSuffix(notation, "#")
	
	// Handle pawn moves (e.g., "e4", "exd5")
	if len(notation) >= 2 && notation[0] >= 'a' && notation[0] <= 'h' {
		return mp.parsePawnMove(notation, gameBoard)
	}
	
	// Handle piece moves (e.g., "Nf3", "Bxf7")
	if len(notation) >= 3 {
		return mp.parsePieceMove(notation, gameBoard)
	}
	
	return board.Move{}, errors.New("unsupported algebraic notation")
}

func (mp *MoveParser) parsePawnMove(notation string, gameBoard *board.Board) (board.Move, error) {
	// For now, return an error since full algebraic parsing is complex
	return board.Move{}, errors.New("algebraic notation not fully implemented - use coordinate notation (e.g., e2e4)")
}

func (mp *MoveParser) parsePieceMove(notation string, gameBoard *board.Board) (board.Move, error) {
	// For now, return an error since full algebraic parsing is complex
	return board.Move{}, errors.New("algebraic notation not fully implemented - use coordinate notation (e.g., e2e4)")
}

func (mp *MoveParser) charToPiece(char byte) (board.Piece, error) {
	switch char {
	case 'q':
		if mp.currentPlayer == White {
			return board.WhiteQueen, nil
		}
		return board.BlackQueen, nil
	case 'r':
		if mp.currentPlayer == White {
			return board.WhiteRook, nil
		}
		return board.BlackRook, nil
	case 'b':
		if mp.currentPlayer == White {
			return board.WhiteBishop, nil
		}
		return board.BlackBishop, nil
	case 'n':
		if mp.currentPlayer == White {
			return board.WhiteKnight, nil
		}
		return board.BlackKnight, nil
	default:
		return board.Empty, errors.New("invalid promotion piece")
	}
}