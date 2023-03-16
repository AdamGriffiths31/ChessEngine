package search

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/2.0/engine"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func InitPolyBook(h *EngineHolder) {
	h.UseBook = false
	file, err := os.Open("performance.bin")
	if err != nil {
		fmt.Println(os.Getwd())
		h.UseBook = false
		return
	}
	defer file.Close()

	fi, err := os.Stat("performance.bin")
	if err != nil {
		panic(fmt.Errorf("InitPolyBook: size  error %v", err))
	}
	size := fi.Size()

	NumEntries = uint64(size) / uint64(unsafe.Sizeof(PolyBookEntry{}))

	PolyEntry = make([]PolyBookEntry, NumEntries)

	reader := io.Reader(file)
	_ = binary.Read(reader, binary.LittleEndian, &PolyEntry)

	if NumEntries > 0 {
		h.UseBook = true
	}
}

func GetBookMove(p *engine.Position) int {
	var bookMoves [32]int
	polyKey := PolyKeyFromBoard(p)
	count := 0
	for i := 0; i < int(NumEntries); i++ {
		if polyKey == littleEndianToBigEndianUint64(PolyEntry[i].Key) {
			move := littleEndianToBigEndianUint16(PolyEntry[i].Move)
			tempMove := ConvertPolyMove(move, p)
			if tempMove != data.NoMove {
				bookMoves[count] = tempMove
				count++
				if count > 32 {
					break
				}
			}
		}
	}
	if count != 0 {
		randMove := rand.Intn(count)
		return bookMoves[randMove]
	}
	return data.NoMove
}

func ConvertPolyMove(polyMove uint16, p *engine.Position) int {

	ff := data.FileChars[(polyMove >> 6 & 7)]
	fr := data.RankChars[(polyMove >> 9 & 7)]
	tf := data.FileChars[(polyMove >> 0 & 7)]
	tr := data.RankChars[(polyMove >> 3 & 7)]
	pp := (polyMove >> 12 & 7)
	var move string
	promotedPiece := "q"
	if pp != 0 {
		switch pp {
		case 1:
			promotedPiece = "n"
		case 2:
			promotedPiece = "b"
		case 3:
			promotedPiece = "r"
		}
		move = fmt.Sprintf("%s%s%s%s%s", ff, fr, tf, tr, promotedPiece)
	} else {
		move = fmt.Sprintf("%s%s%s%s", ff, fr, tf, tr)
	}

	return p.ParseMove([]byte(move))
}

func PolyKeyFromBoard(p *engine.Position) uint64 {
	finalKey := uint64(0)
	for sq := 0; sq < 64; sq++ {
		piece := p.Board.PieceAt(sq)
		if piece != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			if piece > data.BK || piece < data.WP {
				panic(fmt.Errorf("PolyKeyFromBoard: piece error sq:%v value: %v", sq, piece))
			}
			polyPiece := PolyKindOfPiece[piece]
			rank := data.RanksBoard[data.Square64ToSquare120[sq]]
			file := data.FilesBoard[data.Square64ToSquare120[sq]]

			finalKey ^= Random64Poly[(64*polyPiece)+(8*rank)+file]
		}
	}
	//Castle
	offSet := 768
	if p.CastlePermission&data.WhiteKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+0]
	}
	if p.CastlePermission&data.WhiteQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+1]
	}
	if p.CastlePermission&data.BlackKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+2]
	}
	if p.CastlePermission&data.BlackQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+3]
	}
	//EnPas
	offSet = 772
	if hasPawnForCapture(p) {
		file := data.FilesBoard[p.EnPassant]
		finalKey ^= Random64Poly[offSet+file]
	}
	//Side
	if p.Side == data.White {
		offSet = 780
		finalKey ^= Random64Poly[offSet]
	}
	return finalKey
}

func hasPawnForCapture(p *engine.Position) bool {
	target := data.WP
	sqWithPawn := 0
	if p.Side == data.Black {
		target = data.BP
	}
	if p.EnPassant != data.NoSquare {
		if p.Side == data.White {
			sqWithPawn = p.EnPassant - 10
		} else {
			sqWithPawn = p.EnPassant + 10
		}
		if p.Board.PieceAt(sqWithPawn+1) == target {
			return true
		} else if p.Board.PieceAt(sqWithPawn-1) == target {
			return true
		}
	}
	return false
}

func clearEntries() {
	for i := range PolyEntry {
		PolyEntry[i].Key = 0
		PolyEntry[i].Move = 0
		PolyEntry[i].Weight = 0
		PolyEntry[i].Learn = 0
	}
}

func littleToBigEndian(l uint32) uint32 {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, l)
	return binary.BigEndian.Uint32(b)
}

func littleEndianToBigEndianUint64(littleEndian uint64) uint64 {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, littleEndian)
	return binary.BigEndian.Uint64(b)
}

func littleEndianToBigEndianUint16(n uint16) uint16 {
	return (n >> 8) | (n << 8)
}
