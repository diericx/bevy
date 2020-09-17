package services

import (
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/repos/storm"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

type TorrentLink struct {
	MovieTorrentLinkRepo storm.MovieTorrentLink
}

func (s *TorrentLink) LinkTorrentToMovie(imdbID string, t torrent.Torrent, fileIndex int) error {
	return s.MovieTorrentLinkRepo.Store(app.MovieTorrentLink{
		ImdbID:          imdbID,
		TorrentInfoHash: t.InfoHash().HexString(),
		FileIndex:       fileIndex,
	})
}

func (s *TorrentLink) GetLinksForMovie(imdbID string) ([]app.MovieTorrentLink, error) {
	return s.MovieTorrentLinkRepo.GetByImdbID(imdbID)
}
