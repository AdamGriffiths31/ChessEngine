package board

import (
	"fmt"
	"strings"
)

const (
	Cols       = 8
	Rows       = 8
	Horizontal = "─"
	Vertical   = "│"
	marginLeft = 3
	marginTop  = 1
	FirstCol   = 0
	FirstRow   = 0
	LastCol    = Cols - 1
	LastRow    = Rows - 1
)

func BuildTopBorder() string {
	return build("┌", "┬", "┐")
}

func BuildMiddleBorder() string {
	return build("├", "┼", "┤")
}

func BuildBottomBorder() string {
	return build("└", "┴", "┘")
}

func BuildBottomLabels() string {
	labels := ""
	for i := 0; i < Cols; i++ {
		c := i
		labels += fmt.Sprintf("%c", c+'A')
		if i != LastCol {
			labels += withMarginLeft("")
		}
	}
	return withMarginLeft(fmt.Sprintf("  %s\n", labels))
}

func build(left, middle, right string) string {
	border := left + Horizontal + strings.Repeat(Horizontal+Horizontal+middle+Horizontal, LastRow)
	border += Horizontal + Horizontal + right + "\n"
	return withMarginLeft(border)
}

func withMarginLeft(s string) string {
	return strings.Repeat(" ", marginLeft) + s
}
