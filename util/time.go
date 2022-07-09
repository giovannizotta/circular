package util

import (
	"log"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %.3fms", name, float64(elapsed.Microseconds())/1000)
}
