package engine

type Game struct {
	position      *Position
	moves         []Move
	numberOfMoves uint16
}

func NewGame(
	position *Position,
	moves []Move,
	numberOfMoves uint16) Game {
	if position == nil {
		position = &Position{}
	}
	return Game{
		position,
		moves,
		numberOfMoves,
	}
}

func (g *Game) Position() *Position {
	return g.position
}

func ParseFen(fen string) Game {
	game := NewGame(nil, nil, 0)
	game.position.ParseFen(fen)
	return game
}

func (p *Position) Copy() *Position {
	newPos := &Position{
		Board:            p.Board.copy(),
		Play:             p.Play,
		PositionKey:      p.PositionKey,
		Side:             p.Side,
		CastlePermission: p.CastlePermission,
		EnPassant:        p.EnPassant,
		FailHighFirst:    p.FailHighFirst,
		FailHigh:         p.FailHigh,
		MoveHistory:      p.MoveHistory,
		CurrentScore:     p.CurrentScore,
		FiftyMove:        p.FiftyMove,
	}
	return newPos
}
