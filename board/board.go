package board

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func CheckBoard(pos *data.Board) {
	//TODO Need a config to turn this on off (perf is hit when its on and also not need when not deving)
	return
	var pieceNumber = [13]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var pawns = [3]uint64{0, 0}

	fakePos := &data.Board{}

	pawns[data.White] = pos.Pawns[data.White]
	pawns[data.Black] = pos.Pawns[data.Black]
	pawns[data.Both] = pos.Pawns[data.Both]

	for piece := data.WP; piece <= data.BK; piece++ {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq120 := pos.PieceList[piece][pieceNum]
			if pos.Pieces[sq120] != piece {
				panic(fmt.Errorf("CheckBoard: %v does not equal %v", pos.Pieces[sq120], piece))
			}
		}
	}

	for sq64 := 0; sq64 < 64; sq64++ {
		sq120 := data.Square64ToSquare120[sq64]
		piece := pos.Pieces[sq120]
		pieceNumber[piece]++

		SetPosPieceData(fakePos, sq120, piece)
	}

	for piece := data.WP; piece <= data.BK; piece++ {
		if pieceNumber[piece] != pos.PieceNumber[piece] {
			panic(fmt.Errorf("CheckBoard: piece %v [%v] does not match count %v - was %v", piece, data.Pieces[piece], pos.PieceNumber[piece], pieceNumber[piece]))
		}
	}

	pawnCountWhite := CountBits(pawns[data.White])
	if pawnCountWhite != pos.PieceNumber[data.WP] {
		panic(fmt.Errorf("CheckBoard: white pawn error wanted %v but got %v", pos.PieceNumber[data.WP], pawnCountWhite))
	}

	pawnCountBlack := CountBits(pawns[data.Black])
	if pawnCountBlack != pos.PieceNumber[data.BP] {
		panic(fmt.Errorf("CheckBoard: black pawn error wanted %v but got %v", pos.PieceNumber[data.BP], pawnCountBlack))
	}

	pawnCountBoth := CountBits(pawns[data.Both])
	if pawnCountBoth != pawnCountBlack+pawnCountWhite {
		panic(fmt.Errorf("CheckBoard: both pawn error wanted %v but got %v", pawnCountBlack+pawnCountWhite, pawnCountBoth))
	}

	for i := uint64(0); i < 64; i++ {
		if (pawns[data.White] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.White])
			if pos.Pieces[data.Square64ToSquare120[sq64]] != data.WP {
				panic(fmt.Errorf("CheckBoard: PopBit white  wanted %v but got %v", data.WP, pos.Pieces[data.Square64ToSquare120[sq64]]))
			}
		}
		if (pawns[data.Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.Black])
			if pos.Pieces[data.Square64ToSquare120[sq64]] != data.BP {
				panic(fmt.Errorf("CheckBoard: PopBit black  wanted %v but got %v", data.BP, pos.Pieces[data.Square64ToSquare120[sq64]]))
			}
		}
		if (pawns[data.Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.Both])
			if pos.Pieces[data.Square64ToSquare120[sq64]] != data.BP || pos.Pieces[data.Square120ToSquare64[sq64]] != data.WP {
				panic(fmt.Errorf("CheckBoard: PopBit both wanted %v but got %v", data.BP, pos.Pieces[data.Square64ToSquare120[sq64]]))
			}
		}
	}

	if CountBits(fakePos.BlackBishops) != CountBits(pos.BlackBishops) || CountBits(fakePos.BlackKnights) != CountBits(pos.BlackKnights) {
		panic(fmt.Errorf("CheckBoard: black min piece error\n%v-%v\n%v-%v", CountBits(fakePos.BlackBishops), CountBits(pos.BlackBishops), CountBits(fakePos.BlackKnights), CountBits(pos.BlackKnights)))
	}

	if CountBits(fakePos.WhiteBishops) != CountBits(pos.WhiteBishops) || CountBits(fakePos.WhiteKnights) != CountBits(pos.WhiteKnights) {
		panic(fmt.Errorf("CheckBoard: white min piece error"))
	}

	if CountBits(fakePos.WhiteRooks) != CountBits(pos.WhiteRooks) || CountBits(fakePos.BlackRooks) != CountBits(pos.BlackRooks) {
		panic(fmt.Errorf("CheckBoard: rooks piece error"))
	}

	if CountBits(fakePos.WhiteQueens) != CountBits(pos.WhiteQueens) || CountBits(fakePos.BlackQueens) != CountBits(pos.BlackQueens) {
		panic(fmt.Errorf("CheckBoard: queens piece error"))
	}

	if CountBits(fakePos.WhiteKing) != CountBits(pos.WhiteKing) || CountBits(fakePos.BlackKing) != CountBits(pos.BlackKing) {
		panic(fmt.Errorf("CheckBoard: kings piece error"))
	}

	if pos.Side != data.White && pos.Side != data.Black {
		panic(fmt.Errorf("CheckBoard: side was %v", pos.Side))
	}

	if pos.PositionKey != GeneratePositionKey(pos) {
		panic(fmt.Errorf("CheckBoard: PositionKey: %v did not match %v", pos.PositionKey, GeneratePositionKey(pos)))
	}

	if pos.EnPas != data.NoSquare {
		if pos.Side == data.White && data.RanksBoard[pos.EnPas] != data.Rank6 {
			panic(fmt.Errorf("CheckBoard: white EnPas error: %v was not on rank %v", pos.EnPas, data.Rank6))
		}

		if pos.Side == data.Black && data.RanksBoard[pos.EnPas] != data.Rank3 {
			panic(fmt.Errorf("CheckBoard: black EnPas error: %v was not on rank %v", pos.EnPas, data.Rank3))
		}
	}

	if pos.Pieces[pos.KingSquare[data.White]] != data.WK {
		panic(fmt.Errorf("CheckBoard: White king was not found at %v instead %v was found", pos.KingSquare[data.White], pos.Pieces[pos.KingSquare[data.White]]))
	}

	if pos.Pieces[pos.KingSquare[data.Black]] != data.BK {
		panic(fmt.Errorf("CheckBoard: Black king was not found at %v instead %v was found ", io.SquareString(pos.KingSquare[data.Black]), pos.Pieces[pos.KingSquare[data.Black]]))
	}
}

// UpdateListMaterial sets the Material Lists
func UpdateListMaterial(pos *data.Board) {
	for i := 0; i < 120; i++ {
		sq := i
		piece := pos.Pieces[i]
		if piece == data.OffBoard || piece == data.Empty {
			continue
		}
		pos.PieceList[piece][pos.PieceNumber[piece]] = sq
		pos.PieceNumber[piece]++

		if piece == data.WK {
			pos.KingSquare[data.White] = sq
		}

		if piece == data.BK {
			pos.KingSquare[data.Black] = sq
		}

		SetPosPieceData(pos, sq, piece)

	}
}

func MirrorBoard(pos *data.Board) {
	var pieceArray [64]int
	swapPiece := [13]int{data.Empty, data.BP, data.BN, data.BB, data.BR, data.BQ, data.BK, data.WP, data.WN, data.WB, data.WR, data.WQ, data.WK}
	tempEnPas := data.NoSquare
	tempCastlePerm := 0
	tempSide := pos.Side ^ 1

	if pos.CastlePermission&data.WhiteKingCastle != 0 {
		tempCastlePerm |= data.BlackKingCastle
	}
	if pos.CastlePermission&data.WhiteQueenCastle != 0 {
		tempCastlePerm |= data.BlackQueenCastle
	}
	if pos.CastlePermission&data.BlackKingCastle != 0 {
		tempCastlePerm |= data.WhiteKingCastle
	}
	if pos.CastlePermission&data.BlackQueenCastle != 0 {
		tempCastlePerm |= data.WhiteQueenCastle
	}

	if pos.EnPas != data.NoSquare {
		tempEnPas = data.Square64ToSquare120[data.Mirror64[data.Square120ToSquare64[pos.EnPas]]]
	}

	for sq := 0; sq < 64; sq++ {
		pieceArray[sq] = pos.Pieces[data.Square64ToSquare120[data.Mirror64[sq]]]
	}

	resetBoard(pos)

	for sq := 0; sq < 64; sq++ {
		tp := swapPiece[pieceArray[sq]]
		pos.Pieces[data.Square64ToSquare120[sq]] = tp
	}

	pos.Side = tempSide
	pos.CastlePermission = tempCastlePerm
	pos.EnPas = tempEnPas

	pos.PositionKey = GeneratePositionKey(pos)
	UpdateListMaterial(pos)

	CheckBoard(pos)
}

// resetBoard restores Board to a default state
func resetBoard(pos *data.Board) {
	for i := 0; i < 120; i++ {
		pos.Pieces[i] = data.OffBoard
	}

	for i := 0; i < 64; i++ {
		pos.Pieces[data.Square64ToSquare120[i]] = data.Empty
	}

	for i := 0; i < 3; i++ {
		pos.Pawns[i] = 0
	}

	for i := 0; i < 13; i++ {
		pos.PieceNumber[i] = 0
	}

	pos.KingSquare[data.White], pos.KingSquare[data.Black] = data.NoSquare, data.NoSquare

	pos.Side = data.Both
	pos.EnPas = data.NoSquare
	pos.FiftyMove = 0

	pos.Play = 0
	pos.HistoryPlay = 0

	pos.CastlePermission = 0

	pos.PositionKey = 0

	pos.PiecesBB = 0
	pos.ColoredPiecesBB = 0
	pos.WhitePiecesBB = 0
}

// getPieceType returns returns the corresponding piece type integer
func getPieceType(ch rune) int {
	pieceMap := map[rune]int{
		'p': data.BP,
		'r': data.BR,
		'n': data.BN,
		'b': data.BB,
		'q': data.BQ,
		'k': data.BK,
		'P': data.WP,
		'R': data.WR,
		'N': data.WN,
		'B': data.WB,
		'Q': data.WQ,
		'K': data.WK,
	}

	piece, ok := pieceMap[ch]
	if !ok {
		panic(fmt.Errorf("getPieceType: could not find value for %v", ch))
	}

	return piece
}

func SetPosPieceData(pos *data.Board, sq int, piece int) {
	//King
	if piece == data.WK {
		SetBit(&pos.WhiteKing, data.Square120ToSquare64[sq])
	}

	if piece == data.BK {
		SetBit(&pos.BlackKing, data.Square120ToSquare64[sq])
	}

	//Queens
	if piece == data.WQ {
		SetBit(&pos.WhiteQueens, data.Square120ToSquare64[sq])
	}

	if piece == data.BQ {
		SetBit(&pos.BlackQueens, data.Square120ToSquare64[sq])
	}

	//Rooks
	if piece == data.WR {
		SetBit(&pos.WhiteRooks, data.Square120ToSquare64[sq])
	}

	if piece == data.BR {
		SetBit(&pos.BlackRooks, data.Square120ToSquare64[sq])
	}

	//Knights
	if piece == data.WN {
		SetBit(&pos.WhiteKnights, data.Square120ToSquare64[sq])
	}

	if piece == data.BN {
		SetBit(&pos.BlackKnights, data.Square120ToSquare64[sq])
	}

	//Bishops
	if piece == data.WB {
		SetBit(&pos.WhiteBishops, data.Square120ToSquare64[sq])
	}

	if piece == data.BB {
		SetBit(&pos.BlackBishops, data.Square120ToSquare64[sq])
	}

	//Pawns
	if piece == data.WP {
		SetBit(&pos.Pawns[data.White], data.Square120ToSquare64[sq])
		SetBit(&pos.Pawns[data.Both], data.Square120ToSquare64[sq])
	}

	if piece == data.BP {
		SetBit(&pos.Pawns[data.Black], data.Square120ToSquare64[sq])
		SetBit(&pos.Pawns[data.Both], data.Square120ToSquare64[sq])
	}

	if data.PieceCol[piece] == data.White {
		SetBit(&pos.WhitePiecesBB, data.Square120ToSquare64[sq])
	} else {
		SetBit(&pos.ColoredPiecesBB, data.Square120ToSquare64[sq])
	}
	SetBit(&pos.PiecesBB, data.Square120ToSquare64[sq])
}

func ClearPosPieceData(pos *data.Board, sq int, piece int) {
	//King
	if piece == data.WK {
		ClearBit(&pos.WhiteKing, data.Square120ToSquare64[sq])
	}

	if piece == data.BK {
		ClearBit(&pos.BlackKing, data.Square120ToSquare64[sq])
	}

	//Queens
	if piece == data.WQ {
		ClearBit(&pos.WhiteQueens, data.Square120ToSquare64[sq])
	}

	if piece == data.BQ {
		ClearBit(&pos.BlackQueens, data.Square120ToSquare64[sq])
	}

	//Rooks
	if piece == data.WR {
		ClearBit(&pos.WhiteRooks, data.Square120ToSquare64[sq])
	}

	if piece == data.BR {
		ClearBit(&pos.BlackRooks, data.Square120ToSquare64[sq])
	}

	//Knights
	if piece == data.WN {
		ClearBit(&pos.WhiteKnights, data.Square120ToSquare64[sq])
	}

	if piece == data.BN {
		ClearBit(&pos.BlackKnights, data.Square120ToSquare64[sq])
	}

	//Bishops
	if piece == data.WB {
		ClearBit(&pos.WhiteBishops, data.Square120ToSquare64[sq])
	}

	if piece == data.BB {
		ClearBit(&pos.BlackBishops, data.Square120ToSquare64[sq])
	}

	//Pawns
	if piece == data.WP {
		ClearBit(&pos.Pawns[data.White], data.Square120ToSquare64[sq])
		ClearBit(&pos.Pawns[data.Both], data.Square120ToSquare64[sq])
	}

	if piece == data.BP {
		ClearBit(&pos.Pawns[data.Black], data.Square120ToSquare64[sq])
		ClearBit(&pos.Pawns[data.Both], data.Square120ToSquare64[sq])
	}

	if data.PieceCol[piece] == data.White {
		ClearBit(&pos.WhitePiecesBB, data.Square120ToSquare64[sq])
	} else {
		ClearBit(&pos.ColoredPiecesBB, data.Square120ToSquare64[sq])
	}
	ClearBit(&pos.PiecesBB, data.Square120ToSquare64[sq])
}
