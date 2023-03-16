package engine

import "github.com/AdamGriffiths31/ChessEngine/data"

// MakeMoveInt Builds the move int
func MakeMoveInt(f, t, ca, pro, fl int) int {
	return f | t<<7 | ca<<14 | pro<<20 | fl
}

func (p *Position) ParseMove(move []byte) int {
	if move[1] > '8' || move[1] < '1' {
		return data.NoMove
	}
	if move[3] > '8' || move[3] < '1' {
		return data.NoMove
	}
	if move[0] > 'h' || move[0] < 'a' {
		return data.NoMove
	}
	if move[2] > 'h' || move[2] < 'a' {
		return data.NoMove
	}

	from := data.FileRankToSquare(int(move[0]-'a'), int(move[1]-'1'))
	to := data.FileRankToSquare(int(move[2]-'a'), int(move[3]-'1'))

	ml := &MoveList{}
	p.GenerateAllMoves(ml)

	for MoveNum := 0; MoveNum < ml.Count; MoveNum++ {
		userMove := ml.Moves[MoveNum].Move
		if data.FromSquare(userMove) == from && data.ToSquare(userMove) == to {
			promPce := data.Promoted(userMove)
			if promPce != data.Empty {
				if data.PieceRookQueen[promPce] == data.True && data.PieceBishopQueen[promPce] == data.False && move[4] == 'r' {
					return userMove
				} else if data.PieceRookQueen[promPce] == data.False && data.PieceBishopQueen[promPce] == data.True && move[4] == 'b' {
					return userMove
				} else if data.PieceRookQueen[promPce] == data.True && data.PieceBishopQueen[promPce] == data.True && move[4] == 'q' {
					return userMove
				} else if data.PieceKnight[promPce] == data.True && move[4] == 'n' {
					return userMove
				}
				continue
			}
			return userMove
		}
	}

	return data.NoMove
}
