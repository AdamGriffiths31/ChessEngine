//go:build goexperiment.simd

package nnue

import "simd/archsimd"

func addColumn(acc, column *[Hidden]int16) {
	for i := 0; i < Hidden; i += 16 {
		archsimd.LoadInt16x16(acc[i:]).Add(archsimd.LoadInt16x16(column[i:])).Store(acc[i:])
	}
}

func subColumn(acc, column *[Hidden]int16) {
	for i := 0; i < Hidden; i += 16 {
		archsimd.LoadInt16x16(acc[i:]).Sub(archsimd.LoadInt16x16(column[i:])).Store(acc[i:])
	}
}

func screluDot(us, them, usWeights, themWeights *[Hidden]int16) int64 {
	zero := archsimd.BroadcastInt16x16(0)
	top := archsimd.BroadcastInt16x16(QA)
	var sum archsimd.Int32x8
	for i := 0; i < Hidden; i += 16 {
		v := archsimd.LoadInt16x16(us[i:]).Max(zero).Min(top)
		sum = sum.Add(v.DotProductPairs(v.Mul(archsimd.LoadInt16x16(usWeights[i:]))))
		v = archsimd.LoadInt16x16(them[i:]).Max(zero).Min(top)
		sum = sum.Add(v.DotProductPairs(v.Mul(archsimd.LoadInt16x16(themWeights[i:]))))
	}
	quad := sum.GetLo().Add(sum.GetHi())
	return int64(quad.GetElem(0)) + int64(quad.GetElem(1)) + int64(quad.GetElem(2)) + int64(quad.GetElem(3))
}
