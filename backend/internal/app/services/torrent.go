package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

type Torrent struct {
	Client          app.TorrentClient
	TorrentMetaRepo app.TorrentMetaRepo
	// ReleaseRepo          ReleaseRepo
	// MovieTorrentLinkRepo MovieTorrentLinkRepo

	GetInfoTimeout   time.Duration
	TorrentFilesPath string
}

func NewTorrentService(client app.TorrentClient, torrentMetaRepo app.TorrentMetaRepo, getInfoTimeout time.Duration, torrentFilesPath string) (*Torrent, error) {
	s := Torrent{
		Client:           client,
		TorrentMetaRepo:  torrentMetaRepo,
		GetInfoTimeout:   getInfoTimeout,
		TorrentFilesPath: torrentFilesPath,
	}

	// Load torrents that exist in torrent files location
	s.addTorrentsViaFilesInPath()
	err := s.startTorrentsAccordingToMetadata()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// AddTorrentsViaFilesInDirectory adds all torrent files in a dir then async waits for info
func (s *Torrent) addTorrentsViaFilesInPath() error {
	var files []string
	err := filepath.Walk(s.TorrentFilesPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".torrent" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, file := range files {
		t, err := s.Client.AddFile(file)
		if err != nil {
			log.Printf("ERROR: Could not add file: %s\n", file)
		}
		wg.Add(1)
		go func() {
			<-t.GotInfo()
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func (s *Torrent) startTorrentsAccordingToMetadata() error {
	torrents, err := s.Get()
	if err != nil {
		return err
	}

	for _, t := range torrents {
		// Get meta
		meta, err := s.TorrentMetaRepo.GetByInfoHash(t.InfoHash().HexString())
		if err != nil {
			return err
		}
		if !meta.IsStopped {
			t.DownloadAll()
		}
	}

	return nil
}

func (s *Torrent) AddFromMagnet(magnet string, meta app.TorrentMeta) (torrent.Torrent, error) {
	t, err := s.Client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.GetInfoTimeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	// Save our custom meta
	meta.InfoHash = t.InfoHash().HexString()
	err = s.TorrentMetaRepo.Store(meta)
	if err != nil {
		t.Drop()
		return nil, err
	}

	// Save the entire meta to file for state recovery on restarts
	err = s.saveTorrentToFile(t)
	if err != nil {
		t.Drop()
		return nil, err
	}

	if !meta.IsStopped {
		t.DownloadAll()
	}

	return t, nil
}

func (s *Torrent) AddFromFile(file string, meta app.TorrentMeta) (torrent.Torrent, error) {
	t, err := s.Client.AddFile(file)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.GetInfoTimeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	// Save our custom meta
	meta.InfoHash = t.InfoHash().HexString()
	err = s.TorrentMetaRepo.Store(meta)
	if err != nil {
		t.Drop()
		return nil, err
	}

	if !meta.IsStopped {
		t.DownloadAll()
	}

	return t, nil
}

func (s *Torrent) Get() ([]torrent.Torrent, error) {
	torrents := s.Client.Torrents()
	return torrents, nil
}

func (s *Torrent) GetByInfoHashStr(infoHashStr string) (torrent.Torrent, error) {
	hash := metainfo.NewHashFromHex(infoHashStr)
	t, ok := s.Client.Torrent(hash)
	if !ok {
		return nil, errors.New("torrent not found")
	}
	return t, nil
}

func (s *Torrent) saveTorrentToFile(t torrent.Torrent) (err error) {
	file := filepath.Join(s.TorrentFilesPath, fmt.Sprintf("%s.torrent", t.InfoHash().HexString()))
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return
	}
	defer f.Close()
	return t.Metainfo().Write(f)
}
