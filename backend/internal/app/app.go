package app

import (
	"io"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// TODO: input from config file
const DefaultResolution = "iw:ih"
const DefaultMaxBitrate = "50M"

// These functions act as const arrays because go doesn't allow const arrays... I know pretty fucked up
// GetSupportedVideoFileFormats returns an array of strings that are the supported video formats
func GetSupportedVideoFileFormats() []string {
	return []string{".mkv", ".mp4"}
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
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Torrent is an actual torrent on our client, active or inactive. It is added to the client and we have info for it.
type Torrent struct {
	InfoHash       metainfo.Hash
	Stats          torrent.TorrentStats
	Length         int64
	BytesCompleted int64
	Name           string
	Seeding        bool

	RatioToStop  float32
	MinutesAlive int
	HoursToStop  int
	IsStopped    bool
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
	Grabs       string
	Seeders     int
	MinRatio    float32
	MinSeedTime int
}

// Indexer is info we need to hit an indexer for a list of torrents
type Indexer struct {
	Name                 string     `yaml:"name"`
	URL                  string     `yaml:"url"`
	BasicAuth            *BasicAuth `yaml:"basicAuth"`
	SupportsImdbIDSearch bool       `yaml:"supportsImdbIDSearch"`
	APIKey               string     `yaml:"apiKey"`
	Categories           string     `yaml:"categories"`
}

// Quality contains specifications for a specific quality of torrent and how to infer that quality from a name
type Quality struct {
	Name       string `yaml:"name"`
	Regex      string `yaml:"regex"`
	MinSize    int64  `yaml:"minSize"`
	MaxSize    int64  `yaml:"maxSize"`
	Resolution string `yaml:"resolution"`
}

// TorrentService handles CRUD actions for Torrents
type TorrentService interface {
	// Adding
	AddFromMagnet(string) (*Torrent, error)
	AddFromFile(string) (*Torrent, error)
	AddFromURLUknownScheme(rawURL string, auth *BasicAuth) (*Torrent, error)
	// Loading from previous state
	LoadTorrentFilesFromCache() error
	// Queries
	GetByHashStr(string) (*Torrent, error)
	Get() ([]Torrent, error)
	// Actions
	GetFiles(*Torrent) ([]TorrentFile, error)
	GetReadSeekerForFileInTorrent(*Torrent, int) (io.ReadSeeker, error)
	Start(*Torrent) error
}

// ReleaseService will find releases via attributes like imdbID, title, etc. but will not actually add them.
// It probably uses the torrent client to probe torrents for quality and files.
type ReleaseService interface {
	GetReleasesForMovie(imdbID string, title string, year string, minQuality int) ([]Release, error)
}

type TranscoderConfig struct {
	Video struct {
		Format          string `yaml:"format"`
		CompressionAlgo string `yaml:"compressionAlgo"`
	} `yaml:"video"`
	Audio struct {
		CompressionAlgo string `yaml:"compressionAlgo"`
	} `yaml:"audio"`
}

// Movie is a simple struct for holding metadata about a movie
type Movie struct {
	ID     int    // primary key
	ImdbID string `storm:"unique"`
	Title  string
	Year   int
}

// // MovieDAO handles storing and retrieving Movies
// type MovieDAO interface {
// 	Store(Movie) error
// 	GetByID(int) (Movie, error)
// 	Get() ([]Movie, error)
// 	Remove(int) error
// }

// // MovieTorrentLink handles linking a Movie to a specific file in a torrent
// type MovieTorrentLink struct {
// 	MovieID         int
// 	TorrentInfoHash string
// 	FileIndex       int
// }

// // MovieTorrentLinkDAO handles storing and retrieving MovieTorrentLinks
// type MovieTorrentLinkDAO interface {
// 	Store(MovieTorrentLink) error

// 	GetByMovieID(int) (MovieTorrentLink, error)
// 	GetByTorrentID(int) (MovieTorrentLink, error)
// 	Get() ([]Movie, error)

// 	Remove(int) error
// }
