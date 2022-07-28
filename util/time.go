package util

import (
	"github.com/elementsproject/glightning/glightning"
	"time"
)

func TimeTrack(start time.Time, action string, loggingFunc func(level glightning.LogLevel, format string, v ...any)) {
	elapsed := time.Since(start)
	loggingFunc(glightning.Debug, "%s took %.3fms", action, float64(elapsed.Microseconds())/1000)
}
