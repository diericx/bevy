package storm

import (
	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
)

type MovieTorrentLink struct {
	DB *storm.DB
}

func (r *MovieTorrentLink) Store(meta app.TorrentMeta) error {
	return r.DB.Save(&meta)
}

func (r *MovieTorrentLink) GetByImdbID(imdbID string) (app.MovieTorrentLink, error) {
	var movieTorrentLink app.MovieTorrentLink
	err := r.DB.One("ImdbID", imdbID, &movieTorrentLink)
	return movieTorrentLink, err
}
