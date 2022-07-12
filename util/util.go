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
		return "1"
	}
	return "0"
}

func Max(n1, n2 uint64) uint64 {
	if n1 > n2 {
		return n1
	}
	return n2
}

func RandRange(min, max uint64) uint64 {
	return min + (uint64(rand.Int63()) % (max - min))
}
