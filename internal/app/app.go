package app

import (
	"errors"
	"log"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/bevy/internal/pkg/torrent"
)

func GetDefaultTorrentMeta() TorrentMeta {
	return TorrentMeta{
		RatioToStop: 1,
		HoursToStop: 336,
		IsStopped:   false,
	}
}

// These functions act as const arrays because go doesn't allow const arrays... I know pretty fucked up

// GetSupportedVideoFileFormats returns an array of strings that are the supported video formats
func GetSupportedVideoFileFormats() []string {
	return []string{".mkv", ".mp4", ".mov", ".avi"}
}

// GetBlacklistedFileNameContents returns an array of strings that are blacklisted from torrent names
func GetBlacklistedFileNameContents() []string {
	return []string{"sample"}
}

// TODO: These languages are only blacklisted because it's hard to support
func GetBlacklistedTorrentNameContents() []string {
	return []string{"fre", "french", "ita", "italian"}
}

// BasicAuth info for basic auth http requests
type BasicAuth struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type TorrentMeta struct {
	ID int `storm:"id,increment"`
	// Would like storm to enforce this to be unique but it bugged out last time...
	InfoHash                     metainfo.Hash
	Title                        string  `json:"title"`
	RatioToStop                  float32 `json:"ratioToStop"`
	SecondsAlive                 int     `json:"secondsAlive"`
	SecondsSeedingWhileCompleted int     `json:"secondsSeedingWhileCompleted"`
	HoursToStop                  int     `json:"hourseToStop"`
	IsStopped                    bool    `json:"isStopped"`
	BytesWrittenData             int64   `json:"bytesWrittenData"`
	BytesReadData                int64   `json:"bytesReadData"`
	DownloadSpeed                float32 `json:"downloadSpeed"`
}

// TorrentFile represents a file in a torrent
type TorrentFile struct {
	Path string
	Size int64
}

// Release is a potential torrent for a specific piece of media. We use this info to decide whether or not we want to
// actually grab the torrent.
type Release struct {
	ImdbID      string
	Title       string
	Size        int64
	Link        string
	LinkAuth    *BasicAuth
	InfoHash    string
	Grabs       int
	Seeders     int
	MinRatio    float32
	MinSeedTime int
}

// Indexer is info we need to hit an indexer for a list of torrents
type Indexer struct {
	Name                 string     `toml:"name"`
	URL                  string     `toml:"url"`
	BasicAuth            *BasicAuth `toml:"basic_auth"`
	SupportsImdbIDSearch bool       `toml:"supports_imdb_id_search"`
	APIKey               string     `toml:"api_key"`
	Categories           string     `toml:"categories"`
}

// Quality contains specifications for a specific quality of torrent and how to infer that quality from a name
type Quality struct {
	Name            string  `toml:"name"`
	Regex           string  `toml:"regex"`
	MinSize         float64 `toml:"min_size"`
	MaxSize         float64 `toml:"max_size"`
	MinSeeders      int     `toml:"min_seeders"`
	Resolution      string  `toml:"resolution"`
	SeederScoreExpr string  `toml:"seeder_score_expr"`
	SizeScoreExpr   string  `toml:"size_score_expr"`
}

// MovieTorrentLink handles linking a Movie to a specific file in a torrent
type MovieTorrentLink struct {
	ID              int    `storm:"id,increment"`
	ImdbID          string `json:"imdbID"`
	TorrentInfoHash string `json:"torrentInfoHash"`
	FileIndex       int    `json:"fileIndex"`
}

type TorrentMetaRepo interface {
	Store(TorrentMeta) error
	GetByInfoHash(metainfo.Hash) (TorrentMeta, error)
	GetByTitle(string) (TorrentMeta, error)
	Get() ([]TorrentMeta, error)
	RemoveByInfoHash(metainfo.Hash) error
}

type ReleaseRepo interface {
	GetForMovie(imdbID string, title string, year string, minQuality int) ([]Release, error)
}

type MovieTorrentLinkRepo interface {
	Store(MovieTorrentLink) (*MovieTorrentLink, error)
	GetByImdbID(imdbID string) ([]MovieTorrentLink, error)
}

type TorrentClient interface {
	Close()
	AddMagnet(string) (torrent.Torrent, error)
	AddFile(string) (torrent.Torrent, error)
	Torrents() []torrent.Torrent
	Torrent(metainfo.Hash) (torrent.Torrent, bool)
}

func (i Indexer) Validate() error {
	if i.Name == "" {
		return errors.New("Name cannot be empty")
	}
	if i.URL == "" {
		return errors.New("URL cannot be empty")
	}
	if i.APIKey == "" {
		return errors.New("API key cannot be empty")
	}
	return nil
}

func (q Quality) Validate() error {
	if q.SeederScoreExpr == "" {
		return errors.New("Seeder score function cannot be empty")
	}
	if q.SizeScoreExpr == "" {
		return errors.New("Size score function cannot be empty")
	}
	if q.Name == "" {
		return errors.New("Name cannot be empty")
	}
	if q.Regex == "" {
		return errors.New("Regex cannot be empty")
	}
	if q.Resolution == "" {
		return errors.New("Resolution cannot be empty")
	}
	if q.MinSeeders <= 0 {
		return errors.New("MinSeeders cannot be less than or equal to 0")
	}
	return nil
}

// ValidateSizeSeedersAndName will trim the given release slice and return only those that are valid given
// size min maxes, seeder min maxes and name blacklists.
func (release Release) ValidateSizeSeedersAndName(quality Quality) error {
	if float64(release.Size) < quality.MinSize || float64(release.Size) > quality.MaxSize {
		log.Printf("INFO: Passing on release %s because size %v is not correct.", release.Title, release.Size)
		return errors.New("Invalid size")
	}
	if release.Seeders < int(quality.MinSeeders) {
		log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", release.Title, release.Seeders, quality.MinSeeders)
		return errors.New("Invalid seeders count")
	}
	if StringContainsAnyOf(strings.ToLower(release.Title), GetBlacklistedTorrentNameContents()) {
		log.Printf("INFO: Passing on release %s because title contains one of these blacklisted words: %+v", release.Title, GetBlacklistedTorrentNameContents())
		return errors.New("Invalid title")
	}
	return nil
}

func StringContainsAnyOf(s string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

func StringEndsInAny(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}
