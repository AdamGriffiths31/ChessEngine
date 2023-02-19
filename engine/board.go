package engine

import "fmt"

func ParseFEN(fen string, pos *Board) {
	resetBoard(pos)

	var piece int
	rank, file := Rank8, FileA
	counter := 0
	for _, ch := range fen {
		counter++
		switch {
		case ch == '/':
			rank--
			file = FileA
		case ch == ' ':
			goto end
		case '1' <= ch && ch <= '8':
			file += int(ch - '0')
		default:
			piece = GetPieceType(ch)
			sq64 := rank*8 + file
			sq120 := Sqaure64ToSquare120[sq64]
			if piece != Empty {
				pos.Pieces[sq120] = piece
			}
			file++
		}
	}
end:
	fen = fen[counter:]

	if fen[len(fen)-1] == 'w' {
		pos.Side = White
	} else {
		pos.Side = Black
	}
	fen = fen[2:]

	if fen[0] == '-' {
		fen = fen[2:]
	} else {
		for i := 0; i < 4; i++ {
			switch fen[i] {
			case 'K':
				pos.CastlePermission |= WhiteKingCastle
			case 'Q':
				pos.CastlePermission |= WhiteQueenCastle
			case 'k':
				pos.CastlePermission |= BlackKingCastle
			case 'q':
				pos.CastlePermission |= BlackQueenCastle
			}
		}
		fen = fen[5:]
	}

	if fen[0] == '-' {
		fen = fen[2:]
	} else {
		pos.EnPas = FileRankToSquare(int(fen[0])-'a', int(fen[1])-'1')
		fen = fen[3:]
	}

	pos.PosistionKey = GeneratePositionKey(pos)
}

func GetPieceType(ch rune) int {
	piece := 0
	switch ch {
	case 'p':
		piece = BP
	case 'r':
		piece = BR
	case 'n':
		piece = BN
	case 'b':
		piece = BB
	case 'q':
		piece = BQ
	case 'k':
		piece = BK
	case 'P':
		piece = WP
	case 'R':
		piece = WR
	case 'N':
		piece = WN
	case 'B':
		piece = WB
	case 'Q':
		piece = WQ
	case 'K':
		piece = WK
	}

	return piece
}

func PrintBoard(pos *Board) {
	println("Printing board...")
	for rank := Rank8; rank >= Rank1; rank-- {
		fmt.Printf("%v ", rank+1)
		for file := FileA; file <= FileH; file++ {
			sq := FileRankToSquare(file, rank)
			piece := pos.Pieces[sq]
			fmt.Printf("%3v", PceChar[piece])
		}
		fmt.Print("\n")

	}
	fmt.Print("  ")
	for file := FileA; file <= FileH; file++ {
		fmt.Printf("%3c", 'a'+file)
	}
	fmt.Print("\n")
	fmt.Printf("Side:%v\n", pos.Side)
	fmt.Printf("EnPas:%v\n", pos.EnPas)
	fmt.Printf("PosKey:%11X\n", pos.PosistionKey)

}

// resetBoard restores Board to a default state
func resetBoard(pos *Board) {
	for i := 0; i < 120; i++ {
		pos.Pieces[i] = OffBoard
	}

	for i := 0; i < 64; i++ {
		pos.Pieces[Sqaure64ToSquare120[i]] = Empty
	}

	for i := 0; i < 3; i++ {
		pos.BigPiece[i] = 0
		pos.MajorPiece[i] = 0
		pos.MinPiece[i] = 0
		pos.Pawns[i] = 0
	}

	for i := 0; i < 13; i++ {
		pos.PieceNumber[i] = 0
	}

	pos.KingSqaure[White], pos.KingSqaure[Black] = noSqaure, noSqaure

	pos.Side = Both
	pos.EnPas = noSqaure
	pos.FiftyMove = 0

	pos.Play = 0
	pos.HistoryPlay = 0

	pos.CastlePermission = 0

	pos.PosistionKey = 0
}
