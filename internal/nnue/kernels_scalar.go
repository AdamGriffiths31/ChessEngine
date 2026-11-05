//go:build !goexperiment.simd

package nnue

func addColumn(acc, column *[Hidden]int16) {
	for i := range acc {
		acc[i] += column[i]
	}
}

func screluDot(acc, w *[Hidden]int16) int64 {
	var total int64
	for i := range acc {
		v := int64(min(max(acc[i], 0), QA))
		total += v * v * int64(w[i])
	}
	return total
}
