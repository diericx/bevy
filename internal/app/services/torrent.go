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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

type Torrent struct {
	Client           app.TorrentClient
	TorrentMetaRepo  app.TorrentMetaRepo
	torrentStatCache torrent.StatCache
	// ReleaseRepo          ReleaseRepo
	// MovieTorrentLinkRepo MovieTorrentLinkRepo

	GetInfoTimeout   time.Duration
	MinSeeders       int
	TorrentFilesPath string
}

// UpdateMetaForAllTorrents goes through the current torrents in the client, takes a diff between what
// it and it's stored metadata, then updates the meta with said diff.
func (s *Torrent) UpdateMetaForAllTorrents() error {
	if s.torrentStatCache == nil {
		s.torrentStatCache = make(torrent.StatCache)
	}

	// torrents := s.Client.Torrents()
	torrentMetas, err := s.TorrentMetaRepo.Get()
	if err != nil {
		return err
	}
	for _, torrentMeta := range torrentMetas {
		t, err := s.GetByInfoHashStr(torrentMeta.InfoHash)
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		stats := t.Stats()
		cachedStats, ok := s.torrentStatCache[t.InfoHash().HexString()]

		// Add cache and move on if doesn't exist yet
		if !ok {
			s.torrentStatCache[t.InfoHash().HexString()] = t.Stats()
			continue
		}

		bytesWrittenDataDiff := stats.BytesWrittenData.Int64() - cachedStats.BytesWrittenData.Int64()
		bytesReadDataDiff := stats.BytesReadData.Int64() - cachedStats.BytesReadData.Int64()

		// If there is a difference, update meta with diff
		if bytesReadDataDiff > 0 || bytesWrittenDataDiff > 0 {
			meta, err := s.TorrentMetaRepo.GetByInfoHashStr(t.InfoHash().HexString())
			// if meta doesn't exist, torrent might just be still connecting. Just move on
			if err != nil {
				log.Println("ERROR: ", err)
				continue
			}

			meta.BytesReadData += bytesReadDataDiff
			meta.BytesWrittenData += bytesWrittenDataDiff
			s.TorrentMetaRepo.Store(meta)

			s.torrentStatCache[t.InfoHash().HexString()] = t.Stats()
		}
	}
	return nil
}

// AddTorrentsOnDisk adds all torrent files in a dir then async waits for info
func (s *Torrent) AddTorrentsOnDisk() error {
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

// StartTorrentsAccordingToMetadata goes through all the torrents, gets metadata and starts if they should be running
func (s *Torrent) StartTorrentsAccordingToMetadata() error {
	torrents, err := s.Get()
	if err != nil {
		return err
	}

	for _, t := range torrents {
		// Get meta
		meta, err := s.TorrentMetaRepo.GetByInfoHashStr(t.InfoHash().HexString())
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

func (s *Torrent) Get() ([]torrent.Torrent, error) {
	torrentMetas, err := s.TorrentMetaRepo.Get()
	if err != nil {
		return nil, err
	}
	torrents := []torrent.Torrent{}
	for _, meta := range torrentMetas {
		var hash metainfo.Hash
		err := hash.FromHexString(meta.InfoHash)
		if err != nil {
			return nil, err
		}
		torrent, ok := s.Client.Torrent(hash)
		if ok {
			torrents = append(torrents, torrent)
		}
	}
	return torrents, nil
}

func (s *Torrent) GetByInfoHashStr(infoHashStr string) (torrent.Torrent, error) {
	var hash metainfo.Hash
	err := hash.FromHexString(infoHashStr)
	if err != nil {
		return nil, err
	}
	t, ok := s.Client.Torrent(hash)
	if !ok {
		return nil, errors.New("torrent not found")
	}
	return t, nil
}

func (s *Torrent) GetReadSeekerForFileInTorrent(_t torrent.Torrent, fileIndex int) (io.ReadSeeker, error) {
	t, ok := s.Client.Torrent(_t.InfoHash())
	if !ok {
		return nil, errors.New("not found")
	}

	files := t.Files()
	return files[fileIndex].NewReader(), nil
}

func (s *Torrent) AddBestTorrentFromReleases(releases []app.Release, q app.Quality) (torrent.Torrent, int, error) {
	// sort torrents by seeders (to get most available torrents first)
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Seeders > releases[j].Seeders
	})

	for _, r := range releases {
		if float64(r.Size) < q.MinSize || float64(r.Size) > q.MaxSize {
			log.Printf("INFO: Passing on release %s because size %v is not correct.", r.Title, r.Size)
			continue
		}
		if r.Seeders < s.MinSeeders {
			log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", r.Title, r.Seeders, s.MinSeeders)
			continue
		}
		if stringContainsAnyOf(strings.ToLower(r.Title), app.GetBlacklistedTorrentNameContents()) {
			log.Printf("INFO: Passing on release %s because title contains one of these blacklisted words: %+v", r.Title, app.GetBlacklistedTorrentNameContents())
			continue
		}

		// Add to client to get hash
		t, err := s.AddFromURLUknownScheme(
			r.Link,
			r.LinkAuth,
			app.TorrentMeta{
				InfoHash:    r.InfoHash,
				RatioToStop: r.MinRatio,
				HoursToStop: r.MinSeedTime,
			},
		)
		if err != nil {
			log.Printf("WARNING: could not add torrent magnet for %s\n Err: %s", r.Title, err)
			s.RemoveByHash(t.InfoHash())
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
	t, ok := s.Client.Torrent(hash)
	if !ok {
		return errors.New("not found")
	}

	t.Drop()

	err := s.TorrentMetaRepo.RemoveByInfoHashStr(hash.HexString())
	if err != nil {
		return err
	}

	return nil
}

func (s *Torrent) getValidFileInTorrent(t torrent.Torrent) (int, error) {
	// Get correct file
	files := t.Files()

	for i, file := range files {
		if stringEndsInAny(strings.ToLower(file.Path()), app.GetSupportedVideoFileFormats()) && !stringContainsAnyOf(strings.ToLower(file.Path()), app.GetBlacklistedFileNameContents()) {
			return i, nil
		}
	}
	return 0, nil
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
