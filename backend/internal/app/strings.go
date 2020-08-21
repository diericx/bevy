package app

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func StringContainsAnyOf(s string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

func StringEndsInAny(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func ParseTorrentIdFromString(id string) (int, *Error) {
	idInt64, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return 0, NewError(err, http.StatusBadRequest, InvalidIDErr)
	}
	return int(idInt64), nil
}

func FormatTimeString(timeStr string) string {
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
