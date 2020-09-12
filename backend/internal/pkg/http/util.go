package http

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

func formatTimeString(timeStr string) string {
	timeFloat, err := strconv.ParseFloat(timeStr, 64)
	if err != nil {
		log.Println("WARNING: Could not format time string, falling back to 0")
		timeFloat = 0
	}
	timeInt := int(math.Round(timeFloat))
	timeDuration := time.Second * time.Duration(timeInt)
	return fmtDuration(timeDuration)
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
