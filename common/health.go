package common

import "math"

var Count int64

func IncrementCounter() int64 {
	if (math.MaxInt8 - Count) > 2 {
		Count++
	} else {
		Count = 0
	}
	return Count
}
