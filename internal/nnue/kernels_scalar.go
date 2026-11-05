//go:build !goexperiment.simd

package nnue

func addColumn(acc, column *[Hidden]int16) {
	for i := range acc {
		acc[i] += column[i]
	}
}

func subColumn(acc, column *[Hidden]int16) {
	for i := range acc {
		acc[i] -= column[i]
	}
}

func screluDot(us, them, usWeights, themWeights *[Hidden]int16) int64 {
	var total int64
	for i := range us {
		v := int64(min(max(us[i], 0), QA))
		total += v * v * int64(usWeights[i])
		v = int64(min(max(them[i], 0), QA))
		total += v * v * int64(themWeights[i])
	}
	return total
}
