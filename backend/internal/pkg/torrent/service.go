package torrent

import (
	"errors"
	"time"

	"fmt"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
)

type TorrentService struct {
	Timeout          time.Duration
	Client           *torrent.Client
	TorrentsLocation string
}

func (s *TorrentService) LoadTorrentFilesFromCache() error {
	var files []string
	err := filepath.Walk(s.TorrentsLocation, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".torrent" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, file := range files {
		s.AddFromFile(file)
	}

	return nil
}

func (s *TorrentService) AddFromMagnet(magnet string) (*app.Torrent, error) {
	t, err := s.Client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.Timeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	torrent := AnacrolixTorrentToApp(t)
	return &torrent, nil
}

func (s *TorrentService) AddFromFile(file string) (*app.Torrent, error) {
	t, err := s.Client.AddTorrentFromFile(file)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.Timeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	// Save for later recovery on restart
	s.saveTorrentFile(t)

	torrent := AnacrolixTorrentToApp(t)
	return &torrent, nil
}

func (s *TorrentService) Start(torrent app.Torrent) error {
	t, ok := s.Client.Torrent(torrent.InfoHash)
	if !ok {
		return errors.New("torrent not found")
	}

	t.DownloadAll()
	return nil
}

func (s *TorrentService) Get() ([]app.Torrent, error) {
	torrents := s.Client.Torrents()
	torrentsConverted := make([]app.Torrent, len(torrents))
	for i, torrent := range torrents {
		torrentsConverted[i] = AnacrolixTorrentToApp(torrent)
	}
	return torrentsConverted, nil
}

func (s *TorrentService) GetByHash(hash metainfo.Hash) (*app.Torrent, error) {
	t, ok := s.Client.Torrent(hash)
	if !ok {
		return nil, errors.New("torrent not found")
	}
	torrent := AnacrolixTorrentToApp(t)
	return &torrent, nil
}

func (s *TorrentService) cachedMetaInfo(infoHash metainfo.Hash) (ret *metainfo.MetaInfo) {
	file := filepath.Join(s.TorrentsLocation, fmt.Sprintf("%s.torrent", infoHash.HexString()))
	ret, err := metainfo.LoadFromFile(file)
	if err != nil {
		ret = nil
		return
	}
	if ret.HashInfoBytes() != infoHash {
		ret = nil
		return
	}
	return
}

func (s *TorrentService) saveTorrentFile(t *torrent.Torrent) (err error) {
	file := filepath.Join(s.TorrentsLocation, fmt.Sprintf("%s.torrent", t.InfoHash().HexString()))
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return
	}
	defer f.Close()
	return t.Metainfo().Write(f)
}
