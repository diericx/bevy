package torrent

import (
	"errors"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
)

type TorrentService struct {
	timeout time.Duration
	c       *torrent.Client
}

func (client *Client) Close() {
	client.c.Close()
}

func (s *TorrentService) AddFromMagnet(torrent app.Torrent) (app.Torrent, error) {
	t, err := s.c.AddMagnet(torrent.MagnetLink)
	if err != nil {
		return torrent, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(c.timeout):
		t.Drop()
		return torrent, errors.New("info grab timed out")
	}

	info := t.Info()
	torrent.InfoHash = t.InfoHash()
	torrent.Length = info.Length
	torrent.Name = info.Name

	return torrent, nil
}

func (s *TorrentService) AddFromFile(torrent app.Torrent) (app.Torrent, error) {
	t, err := s.c.AddTorrentFromFile(torrent.File)
	if err != nil {
		return torrent, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(c.timeout):
		t.Drop()
		return torrent, errors.New("info grab timed out")
	}

	info := t.Info()
	torrent.InfoHash = t.InfoHash()
	torrent.Length = info.Length
	torrent.Name = info.Name

	return torrent, nil
}

func (s *TorrentService) Start(torrent app.Torrent) error {
	hash := metainfo.NewHashFromHex(torrent.InfoHash)
	t, ok := s.c.Torrent(hash)
	if !ok {
		return errors.New("torrent not found")
	}

	t.DownloadAll()
	return nil
}

func (s *TorrentService) Stats(torrent app.Torrent) (app.TorrentStats, error) {
	hash := metainfo.NewHashFromHex(torrent.InfoHash)
	t, ok := s.c.Torrent(hash)
	if !ok {
		return app.TorrentStats{}, errors.New("torrent not found")
	}

	stats := t.Stats()

	return app.TorrentStats{
		TorrentStats:   stats,
		BytesCompleted: t.BytesCompleted(),
		IsSeeding:      t.Seeding(),
	}, nil
}

func (s *TorrentService) Get() ([]Torrent, error) {
	torrents := s.c.Torrents()
	torrentsConverted := make([]Torrent, len())
}
