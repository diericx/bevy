package services

import (
	"fmt"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/repos/jackett"
)

type Release struct {
	ReleaseRepo jackett.ReleaseRepo
	Qualities   []app.Quality
}

func (s *Release) QueryMovie(imdbID string, title string, year string, minQuality int) ([]app.Release, error) {
	return s.ReleaseRepo.QueryAllIndexers(
		imdbID,
		fmt.Sprintf("%s %s %s", title, year, s.Qualities[minQuality].Regex),
	)
}
