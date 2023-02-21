package engine

func ParseFEN(fen string, pos *Board) {
	resetBoard(pos)

	fen = parsePiecePlacement(fen, pos)
	fen = parseActiveColor(fen, pos)
	fen = parseCastlingAvailability(fen, pos)
	fen = parseEnPassantTarget(fen, pos)

	//TODO Set moves from fen

	pos.PosistionKey = GeneratePositionKey(pos)
	UpdateListMaterial(pos)
}

func parsePiecePlacement(fen string, pos *Board) string {
	rank, file := Rank8, FileA
	for i, ch := range fen {
		switch {
		case ch == '/':
			rank--
			file = FileA
		case ch == ' ':
			i++
			return fen[i:]
		case '1' <= ch && ch <= '8':
			file += int(ch - '0')
		default:
			piece := getPieceType(ch)
			sq64 := rank*8 + file
			sq120 := Sqaure64ToSquare120[sq64]
			if piece != Empty {
				pos.Pieces[sq120] = piece
			}
			file++
		}
	}
	return fen
}

func parseActiveColor(fen string, pos *Board) string {
	if fen[0] == 'w' {
		pos.Side = White
	} else if fen[0] == 'b' {
		pos.Side = Black
	}
	return fen[2:]
}

func parseCastlingAvailability(fen string, pos *Board) string {
	if fen[0] == '-' {
		return fen[2:]
	}
	index := 0
	for fen[index] != ' ' {
		switch fen[index] {
		case 'K':
			pos.CastlePermission |= WhiteKingCastle
		case 'Q':
			pos.CastlePermission |= WhiteQueenCastle
		case 'k':
			pos.CastlePermission |= BlackKingCastle
		case 'q':
			pos.CastlePermission |= BlackQueenCastle
		}
		index++
	}
	index++
	return fen[index:]
}

func parseEnPassantTarget(fen string, pos *Board) string {
	if fen[0] == '-' {
		return fen[2:]
	}
	pos.EnPas = FileRankToSquare(int(fen[0])-'a', int(fen[1])-'1')
	return fen[3:]
}
