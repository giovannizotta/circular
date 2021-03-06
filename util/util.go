package util

import "math/rand"

func All(v []bool) bool {
	for _, b := range v {
		if !b {
			return false
		}
	}
	return true
}

func GetDirection(from, to string) string {
	if from < to {
		return "0"
	}
	return "1"
}

func Min(n1, n2 uint64) uint64 {
	if n1 < n2 {
		return n1
	}
	return n2
}

func Max(n1, n2 uint64) uint64 {
	if n1 > n2 {
		return n1
	}
	return n2
}

func RandRange(min, max uint64) uint64 {
	if min == max {
		return min
	}
	if max < min {
		return 0
	}
	return min + (uint64(rand.Int63()) % (max - min))
}
