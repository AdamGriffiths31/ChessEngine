package util

import (
	"fmt"
	"time"
)

func TimeTrackNano(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %d ns (%fs)\n", name, elapsed.Nanoseconds(), elapsed.Seconds())
}

func TimeTrackMilliseconds(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %d ns (%fs)\n", name, elapsed.Milliseconds(), elapsed.Seconds())
}

func GetTimeMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func Min(l, r int) int {
	if l < r {
		return l
	}
	return r
}
