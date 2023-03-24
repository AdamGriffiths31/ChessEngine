package search

import "github.com/AdamGriffiths31/ChessEngine/data"

func (e *Engine) MateIn(height int) int {
	return data.ABInfinite - height
}

func (e *Engine) MatedIn(height int) int {
	return -data.ABInfinite + height
}
