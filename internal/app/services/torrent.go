package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/bevy/internal/app"
	"github.com/diericx/bevy/internal/pkg/torrent"
)

type Torrent struct {
	client           app.TorrentClient
	torrentMetaRepo  app.TorrentMetaRepo
	getInfoTimeout   time.Duration
	torrentFilesPath string

	torrentStatCache torrent.StatCache
}

func NewTorrentService(client app.TorrentClient, tmr app.TorrentMetaRepo, getInfoTimeout time.Duration, torrentFilesPath string) Torrent {
	return Torrent{
		client,
		tmr,
		getInfoTimeout,
		torrentFilesPath,
		make(torrent.StatCache),
	}
}

// StartMetaRefreshForAllTorrentsLoop goes through the current torrents in the client, takes a diff between what
// it and it's stored metadata, then updates the meta with said diff.
func (s *Torrent) StartMetaRefreshForAllTorrentsLoop(refreshRateInSeconds int) error {
	for {
		time.Sleep(time.Duration(refreshRateInSeconds) * time.Second)

		if s.torrentStatCache == nil {
			s.torrentStatCache = make(torrent.StatCache)
		}

		torrentMetas, err := s.torrentMetaRepo.Get()
		if err != nil {
			return err
		}
		for _, torrentMeta := range torrentMetas {
			t, ok := s.client.Torrent(torrentMeta.InfoHash)
			if !ok {
				log.Println("Error updating meta: torrent does not exist in client")
				continue
			}

			stats := t.Stats()
			cachedStats, ok := s.torrentStatCache[t.InfoHash()]

			// Add cache and move on if doesn't exist yet
			if !ok {
				s.torrentStatCache[t.InfoHash()] = t.Stats()
				continue
			}

			bytesWrittenDataDiff := stats.BytesWrittenData.Int64() - cachedStats.BytesWrittenData.Int64()
			bytesReadDataDiff := stats.BytesReadData.Int64() - cachedStats.BytesReadData.Int64()

			meta, err := s.torrentMetaRepo.GetByInfoHash(t.InfoHash())
			meta.SecondsAlive += refreshRateInSeconds

			// if it's completed downloading and seeding, start incrementing seed time
			if t.BytesMissing() == 0 && t.Seeding() {
				meta.SecondsSeedingWhileCompleted += refreshRateInSeconds
			}

			// If there is a difference, update meta with diff
			if bytesReadDataDiff > 0 || bytesWrittenDataDiff > 0 {
				// if meta doesn't exist, torrent might just be still connecting. Just move on
				if err != nil {
					log.Println("ERROR: ", err)
					continue
				}

				meta.BytesReadData += bytesReadDataDiff
				meta.BytesWrittenData += bytesWrittenDataDiff
				meta.DownloadSpeed = float32(bytesReadDataDiff) / float32(refreshRateInSeconds)

				s.torrentStatCache[t.InfoHash()] = t.Stats()
			} else {
				// Nothing downloaded so make sure download speed is set to 0
				if meta.DownloadSpeed != 0 {
					meta.DownloadSpeed = 0
					s.torrentStatCache[t.InfoHash()] = t.Stats()
				}
			}

			s.torrentMetaRepo.Store(meta)
		}
	}
}

// AddTorrentsOnDisk adds all torrent files in a dir then async waits for info
func (s *Torrent) AddTorrentsOnDisk() error {
	var files []string
	err := filepath.Walk(s.torrentFilesPath, func(path string, info os.FileInfo, err error) error {
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
		t, err := s.client.AddFile(file)
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

// StartTorrentsAccordingToMetadata goes through all the torrents, gets metadata and starts if they should be running
func (s *Torrent) StartTorrentsAccordingToMetadata() error {
	torrents := s.Get()

	for _, t := range torrents {
		// Get meta
		meta, err := s.torrentMetaRepo.GetByInfoHash(t.InfoHash())
		if err != nil {
			return err
		}
		if !meta.IsStopped {
			t.DownloadAll()
		}
	}

	return nil
}

// AddFromMagnet adds a magnet, saves the torrent info to a file and saves metadata with
// resulting torrent's info hash
func (s *Torrent) AddFromMagnet(magnet string, meta app.TorrentMeta) (torrent.Torrent, error) {
	t, err := s.client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.getInfoTimeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	// Save our custom meta
	var emptyHash metainfo.Hash
	if meta.InfoHash == emptyHash {
		meta.InfoHash = t.InfoHash()
	}
	if meta.Title == "" {
		meta.Title = t.Name()
	}
	err = s.torrentMetaRepo.Store(meta)
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

// AddFromFile adds a torrent from a file and creates it's own file in the files directory.
// A TorrentMeta is saved with the resulting torrent info hash.
func (s *Torrent) AddFromFile(file string, meta app.TorrentMeta) (torrent.Torrent, error) {
	t, err := s.client.AddFile(file)
	if err != nil {
		return nil, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(s.getInfoTimeout):
		t.Drop()
		return nil, errors.New("info grab timed out")
	}

	// Save our custom meta
	meta.InfoHash = t.InfoHash()
	err = s.torrentMetaRepo.Store(meta)
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

	return t, nil
}

func (s *Torrent) Get() []torrent.Torrent {
	return s.client.Torrents()
}

func (s *Torrent) GetByInfoHash(infoHash metainfo.Hash) (torrent.Torrent, error) {
	t, ok := s.client.Torrent(infoHash)
	if !ok {
		return nil, errors.New("torrent not found")
	}
	return t, nil
}

func (s *Torrent) GetByTitle(title string) (torrent.Torrent, error) {
	torrentMeta, err := s.torrentMetaRepo.GetByTitle(title)
	if err != nil {
		return nil, errors.New("torrent meta not found")
	}
	t, ok := s.client.Torrent(torrentMeta.InfoHash)
	if !ok {
		return nil, errors.New("torrent not found")
	}
	return t, nil
}

func (s *Torrent) AddBestTorrentFromReleases(releases []app.Release, q app.Quality) (torrent.Torrent, int, error) {
	for _, r := range releases {
		// Add to client to get hash
		t, err := s.addFromURLUknownScheme(
			r.Link,
			r.LinkAuth,
			app.TorrentMeta{
				Title:       r.Title,
				RatioToStop: r.MinRatio,
				HoursToStop: r.MinSeedTime,
			},
		)
		if err != nil {
			log.Printf("WARNING: could not add torrent magnet for %s\n Err: %s", r.Title, err)
			if t != nil {
				s.RemoveByHash(t.InfoHash())
			}
			continue
		}

		r.InfoHash = t.InfoHash().HexString()

		// Attempt to find a valid file
		index, terr := s.getValidFileInTorrent(t)
		if terr != nil {
			log.Printf("INFO: Passing on release %s because there was no valid file.", r.Title)
			s.RemoveByHash(t.InfoHash())
			continue
		}

		copyT := t // Why does this object need to be copied?? So weird..
		return copyT, index, nil
	}
	return nil, 0, nil
}

func (s *Torrent) RemoveByHash(hash metainfo.Hash) error {
	t, ok := s.client.Torrent(hash)
	if !ok {
		return errors.New("not found")
	}

	t.Drop()

	err := s.torrentMetaRepo.RemoveByInfoHash(hash)
	if err != nil {
		return err
	}

	return nil
}

// addFromURLUknownScheme will add the torrent if it is a magnet url, will download a file if it's a
// file or recursicely follow a redirect
func (s *Torrent) addFromURLUknownScheme(rawURL string, auth *app.BasicAuth, meta app.TorrentMeta) (torrent.Torrent, error) {
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
			return s.addFromURLUknownScheme(url.String(), auth, meta)
		}
		return nil, err
	}

	tempFilePath := fmt.Sprintf("%s/%s", s.torrentFilesPath, randomString(10))
	err = downloadFileFromResponse(resp, tempFilePath)
	defer os.Remove(tempFilePath)
	if err != nil {
		return nil, err
	}

	return s.AddFromFile(tempFilePath, meta)
}

func (s *Torrent) getValidFileInTorrent(t torrent.Torrent) (int, error) {
	// Get correct file
	files := t.Files()

	var bestMatchIndex int
	var bestMatchSize int64
	var foundMatch bool
	for i, file := range files {
		if app.StringEndsInAny(strings.ToLower(file.Path()), app.GetSupportedVideoFileFormats()) && !app.StringContainsAnyOf(strings.ToLower(file.Path()), app.GetBlacklistedFileNameContents()) {
			// Grab largest match
			if file.Length() > bestMatchSize {
				foundMatch = true
				bestMatchIndex = i
				bestMatchSize = file.Length()
			}
		}
	}
	if !foundMatch {
		return 0, errors.New("No valid video file found")
	}

	return bestMatchIndex, nil
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
	file := filepath.Join(s.torrentFilesPath, fmt.Sprintf("%s.torrent", t.InfoHash().HexString()))
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return
	}
	defer f.Close()
	return t.Metainfo().Write(f)
}
