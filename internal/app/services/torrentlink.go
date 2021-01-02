package services

import (
	"github.com/diericx/bevy/internal/app"
	"github.com/diericx/bevy/internal/app/repos/storm"
	"github.com/diericx/bevy/internal/pkg/torrent"
)

type TorrentLink struct {
	MovieTorrentLinkRepo storm.MovieTorrentLink
}

func (s *TorrentLink) LinkTorrentToMovie(imdbID string, t torrent.Torrent, fileIndex int) (app.MovieTorrentLink, error) {
	link := app.MovieTorrentLink{
		ImdbID:          imdbID,
		TorrentInfoHash: t.InfoHash().HexString(),
		FileIndex:       fileIndex,
	}
	err := s.MovieTorrentLinkRepo.Store(link)
	return link, err
}

func (s *TorrentLink) GetLinksForMovie(imdbID string) ([]app.MovieTorrentLink, error) {
	return s.MovieTorrentLinkRepo.GetByImdbID(imdbID)
}
