package app

import (
	"net/http"
	"strconv"
	"strings"
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
