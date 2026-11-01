package board

import "strings"

// CastlingRights is a set of castling rights.
type CastlingRights uint8

// The individual castling rights.
const (
	WhiteKingside CastlingRights = 1 << iota
	WhiteQueenside
	BlackKingside
	BlackQueenside

	allCastlingRights = WhiteKingside | WhiteQueenside | BlackKingside | BlackQueenside
)

const castlingLetters = "KQkq"

var castlingStrings = buildCastlingStrings()

func buildCastlingStrings() [allCastlingRights + 1]string {
	var names [allCastlingRights + 1]string
	for rights := range names {
		var name strings.Builder
		for bit, letter := range castlingLetters {
			if rights&(1<<bit) != 0 {
				name.WriteRune(letter)
			}
		}
		names[rights] = name.String()
		if rights == 0 {
			names[rights] = "-"
		}
	}
	return names
}

// ParseCastlingRights reads rights from FEN notation, ignoring anything that
// is not one of the letters KQkq.
func ParseCastlingRights(s string) CastlingRights {
	var rights CastlingRights
	for bit, letter := range castlingLetters {
		if strings.ContainsRune(s, letter) {
			rights |= 1 << bit
		}
	}
	return rights
}

// String returns the rights in FEN notation, or "-" for none.
func (c CastlingRights) String() string {
	return castlingStrings[c&allCastlingRights]
}
