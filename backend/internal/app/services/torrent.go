package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
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

// AddFromFile adds a torrent from a file and creates it's own file in the files directory. It does not handle
// the input file at all other than reading
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

// AddFromURLUknownScheme will add the torrent if it is a magnet url, will download a file if it's a
// file or recursicely follow a redirect
func (s *Torrent) AddFromURLUknownScheme(rawURL string, auth *app.BasicAuth, meta app.TorrentMeta) (torrent.Torrent, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "magnet" {
		return s.AddFromMagnet(rawURL, meta)
	}

	// Attempt to make http/s call
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		panic(err)
	}
	if auth != nil {
		req.SetBasicAuth(auth.Username, auth.Password)
	}
	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("Redirect")
	}

	resp, err := client.Do(req)
	if err != nil {
		if resp.StatusCode == http.StatusFound { //status code 302
			url, err := resp.Location()
			if err != nil {
				return nil, err
			}
			return s.AddFromURLUknownScheme(url.String(), auth, meta)
		}
		return nil, err
	}

	tempFilePath := fmt.Sprintf("%s/%s", s.TorrentFilesPath, randomString(10))
	err = downloadFileFromResponse(resp, tempFilePath)
	defer os.Remove(tempFilePath)
	if err != nil {
		return nil, err
	}

	return s.AddFromFile(tempFilePath, meta)
}

func (s *Torrent) GetReadSeekerForFileInTorrent(_t torrent.Torrent, fileIndex int) (io.ReadSeeker, error) {
	t, ok := s.Client.Torrent(_t.InfoHash())
	if !ok {
		return nil, errors.New("not found")
	}

	files := t.Files()
	return files[fileIndex].NewReader(), nil
}

func downloadFileFromResponse(resp *http.Response, filePath string) error {
	// Get the data
	if resp.StatusCode != 200 {
		return fmt.Errorf("couldn't reach file server with code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
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

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ123456788"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}
