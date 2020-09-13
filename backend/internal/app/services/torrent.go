package services

import (
	"errors"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

type TorrentMetaRepo interface {
	Store(app.TorrentMeta) (*app.TorrentMeta, error)
	GetByInfoHash(string) (*app.TorrentMeta, error)
}

type ReleaseRepo interface {
	GetForMovie(imdbID string, title string, year string, minQuality int) ([]app.Release, error)
}

type MovieTorrentLinkRepo interface {
	Store(app.MovieTorrentLink) (*app.MovieTorrentLink, error)
	GetByImdbID(imdbID string) ([]app.MovieTorrentLink, error)
}

type TorrentClient interface {
	Close()
	AddMagnet(string)
}

type Torrent struct {
	Client               TorrentClient
	TorrentMetaRepo      TorrentMetaRepo
	ReleaseRepo          ReleaseRepo
	MovieTorrentLinkRepo MovieTorrentLinkRepo

	AddTimeout time.Duration
}

func (s *Torrent) AddFromMagnet(magnet string) (torrent.Torrent, error) {
	t, err := s.Client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.AddTimeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	return t, nil
}
