package util

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
)

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

func RemoveBeforeCharacter(s string, sep string) string {
	if len(sep) == 0 {
		return s
	}
	splits := strings.Split(s, sep)
	return splits[len(splits)-1]
}

func GetCallInfo() string {
	pc := make([]uintptr, 15)
	offset := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:offset])
	frame, _ := frames.Next()
	filename := RemoveBeforeCharacter(frame.File, "/")
	function := RemoveBeforeCharacter(frame.Function, ".")
	return fmt.Sprintf("%s:%d %s: ", filename, frame.Line, function)
}

func GetMapValues[T any](m map[string]T) []T {
	values := make([]T, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
