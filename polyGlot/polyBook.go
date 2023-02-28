package polyglot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func InitPolyBook(entries *Entries) {
	file, err := os.Open("performance.bin")
	if err != nil {
		fmt.Println(os.Getwd())
		panic(fmt.Errorf("InitPolyBook: open performance.bin. %v", err))
	}
	defer file.Close()

	// position, err := file.Seek(0, io.SeekEnd)
	// if err != nil {
	// 	panic(fmt.Errorf("InitPolyBook: position error %v", err))
	// }

	fi, err := os.Stat("performance.bin")
	if err != nil {
		panic(fmt.Errorf("InitPolyBook: size  error %v", err))
	}
	size := fi.Size()

	NumEntries = uint64(size) / uint64(unsafe.Sizeof(PolyBookEntry{}))
	fmt.Printf("NumEntries: %d\n", NumEntries)

	entries.PolyEntry = make([]PolyBookEntry, NumEntries)

	for i := 0; i < int(NumEntries); i++ {
		data := readNextBytes(file, int(unsafe.Sizeof(PolyBookEntry{})))
		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.LittleEndian, &entries.PolyEntry[i])
		if err != nil {
			panic(fmt.Errorf("InitPolyBook: binary Read failed %v", err))
		}
	}

}

func PolyKeyFromBoard(pos *data.Board) uint64 {
	finalKey := uint64(0)
	for sq := 0; sq < 120; sq++ {
		piece := pos.Pieces[sq]
		if piece != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			if piece > data.BK || piece < data.WP {
				panic(fmt.Errorf("PolyKeyFromBoard: piece error sq:%v value: %v", sq, piece))
			}
			polyPiece := PolyKindOfPiece[piece]
			rank := data.RanksBoard[sq]
			file := data.FilesBoard[sq]

			finalKey ^= Random64Poly[(64*polyPiece)+(8*rank)+file]
		}
	}
	//Castle
	offSet := 768
	if pos.CastlePermission&data.WhiteKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+0]
	}
	if pos.CastlePermission&data.WhiteQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+1]
	}
	if pos.CastlePermission&data.BlackKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+2]
	}
	if pos.CastlePermission&data.BlackQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+3]
	}
	//EnPas
	offSet = 772
	if hasPawnForCapture(pos) {
		file := data.FilesBoard[pos.EnPas]
		finalKey ^= Random64Poly[offSet+file]
	}
	//Side
	if pos.Side == data.White {
		offSet = 780
		finalKey ^= Random64Poly[offSet]
	}
	return finalKey
}

func hasPawnForCapture(pos *data.Board) bool {
	target := data.WP
	sqWithPawn := 0
	if pos.Side == data.Black {
		target = data.BP
	}
	if pos.EnPas != data.NoSquare {
		if pos.Side == data.White {
			sqWithPawn = pos.EnPas - 10
		} else {
			sqWithPawn = pos.EnPas + 10
		}

		if pos.Pieces[sqWithPawn+1] == target {
			return true
		} else if pos.Pieces[sqWithPawn-1] == target {
			return true
		}
	}
	return false
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func clearEntries(entries *Entries) {
	for i := range entries.PolyEntry {
		entries.PolyEntry[i].Key = 0
		entries.PolyEntry[i].Move = 0
		entries.PolyEntry[i].Weight = 0
		entries.PolyEntry[i].Learn = 0
	}
}
