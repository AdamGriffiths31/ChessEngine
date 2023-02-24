package board

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func ParseFEN(fen string, pos *data.Board) {
	resetBoard(pos)

	fen = parsePiecePlacement(fen, pos)
	fen = parseActiveColor(fen, pos)
	fen = parseCastlingAvailability(fen, pos)
	fen = parseEnPassantTarget(fen, pos)

	//TODO Set moves from fen

	pos.PosistionKey = GeneratePositionKey(pos)
	UpdateListMaterial(pos)
}

func parsePiecePlacement(fen string, pos *data.Board) string {
	rank, file := data.Rank8, data.FileA
	for i, ch := range fen {
		switch {
		case ch == '/':
			rank--
			file = data.FileA
		case ch == ' ':
			i++
			return fen[i:]
		case '1' <= ch && ch <= '8':
			file += int(ch - '0')
		default:
			piece := getPieceType(ch)
			sq64 := rank*8 + file
			sq120 := data.Sqaure64ToSquare120[sq64]
			if piece != data.Empty {
				pos.Pieces[sq120] = piece
			}
			file++
		}
	}
	return fen
}

func parseActiveColor(fen string, pos *data.Board) string {
	if fen[0] == 'w' {
		pos.Side = data.White
	} else if fen[0] == 'b' {
		pos.Side = data.Black
	}
	return fen[2:]
}

func parseCastlingAvailability(fen string, pos *data.Board) string {
	if fen[0] == '-' {
		return fen[2:]
	}
	index := 0
	for fen[index] != ' ' {
		switch fen[index] {
		case 'K':
			pos.CastlePermission |= data.WhiteKingCastle
		case 'Q':
			pos.CastlePermission |= data.WhiteQueenCastle
		case 'k':
			pos.CastlePermission |= data.BlackKingCastle
		case 'q':
			pos.CastlePermission |= data.BlackQueenCastle
		}
		index++
	}
	index++
	return fen[index:]
}

func parseEnPassantTarget(fen string, pos *data.Board) string {
	if fen[0] == '-' {
		return fen[2:]
	}
	pos.EnPas = data.FileRankToSquare(int(fen[0])-'a', int(fen[1])-'1')
	return fen[3:]
}
