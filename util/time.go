package util

import (
	"fmt"
	"time"
)

func TimeTrack(start time.Time, name string) string {
	elapsed := time.Since(start)
	return fmt.Sprintf("%s took %.3fms", name, float64(elapsed.Microseconds())/1000)
}
