package torrent

import (
	"errors"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/diericx/iceetime/internal/app"
)

type TorrentService struct {
	TorrentDAO app.TorrentDAO
	Client     *torrent.Client
	Timeout    time.Duration
}

func (s *TorrentService) Add(torrentToAdd app.Torrent) (app.Torrent, error) {
	var t *torrent.Torrent
	if torrentToAdd.MagnetLink != "" {
		_t, err := s.Client.AddMagnet(torrentToAdd.MagnetLink)
		if err != nil {
			return app.Torrent{}, err
		}
		t = _t
	} else if torrentToAdd.File != "" {
		_t, err := s.Client.AddTorrentFromFile(torrentToAdd.File)
		if err != nil {
			return app.Torrent{}, err
		}
		t = _t
	} else {
		return app.Torrent{}, errors.New("must specify magnet or file")
	}

	select {
	case <-t.GotInfo():
	case <-time.After(s.Timeout):
		t.Drop()
		return app.Torrent{}, errors.New("info grab timed out")
	}

	// TODO: should we actually start downloading? Will this mess up readseekers?
	t.DownloadAll()

	// update some info
	hash := t.InfoHash()
	info := t.Info()
	torrentToAdd.InfoHash = hash.HexString()
	torrentToAdd.Size = info.Length
	torrentToAdd.Title = info.Name

	// Save to db
	newTorrent, err := s.TorrentDAO.Store(torrentToAdd)
	if err != nil {
		return app.Torrent{}, err
	}

	return newTorrent, nil
}

func (s *TorrentService) GetByID(id uint) (app.Torrent, error) {
	// TODO: fill in status from torrent client
	return s.TorrentDAO.GetByID(id)
}

func (s *TorrentService) Get() ([]app.Torrent, error) {
	// TODO: fill in status from torrent client
	return s.TorrentDAO.Get()
}
