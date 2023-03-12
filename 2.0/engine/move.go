package engine

// MakeMoveInt Builds the move int
func MakeMoveInt(f, t, ca, pro, fl int) int {
	return f | t<<7 | ca<<14 | pro<<20 | fl
}
