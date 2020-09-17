package storm

import (
	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
)

type MovieTorrentLink struct {
	DB *storm.DB
}

func (r *MovieTorrentLink) Store(meta app.MovieTorrentLink) error {
	return r.DB.Save(&meta)
}

func (r *MovieTorrentLink) GetByImdbID(imdbID string) ([]app.MovieTorrentLink, error) {
	var movieTorrentLinks []app.MovieTorrentLink
	err := r.DB.Find("ImdbID", imdbID, &movieTorrentLinks)
	return movieTorrentLinks, err
}
