//go:build goexperiment.simd

package nnue

import "simd/archsimd"

func addColumn(acc, column *[Hidden]int16) {
	for i := 0; i < Hidden; i += 16 {
		archsimd.LoadInt16x16(acc[i:]).Add(archsimd.LoadInt16x16(column[i:])).Store(acc[i:])
	}
}

func screluDot(acc, w *[Hidden]int16) int64 {
	zero := archsimd.BroadcastInt16x16(0)
	top := archsimd.BroadcastInt16x16(QA)
	var sum archsimd.Int32x8
	for i := 0; i < Hidden; i += 16 {
		v := archsimd.LoadInt16x16(acc[i:]).Max(zero).Min(top)
		sum = sum.Add(v.DotProductPairs(v.Mul(archsimd.LoadInt16x16(w[i:]))))
	}
	var lanes [8]int32
	sum.StoreArray(&lanes)
	var total int64
	for _, lane := range lanes {
		total += int64(lane)
	}
	return total
}
