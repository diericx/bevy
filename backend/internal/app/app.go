package app

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

// // TODO: input from config file
// const DefaultResolution = "iw:ih"
// const DefaultMaxBitrate = "50M"

// // These functions act as const arrays because go doesn't allow const arrays... I know pretty fucked up
// // GetSupportedVideoFileFormats returns an array of strings that are the supported video formats
// func GetSupportedVideoFileFormats() []string {
// 	return []string{".mkv", ".mp4"}
// }

// // GetBlacklistedFileNameContents returns an array of strings that are blacklisted from torrent names
// func GetBlacklistedFileNameContents() []string {
// 	return []string{"sample"}
// }

// // TODO: These languages are only blacklisted because it's hard to support
// func GetBlacklistedTorrentNameContents() []string {
// 	return []string{"fre", "french", "ita", "italian"}
// }

// BasicAuth info for basic auth http requests
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TorrentMeta struct {
	InfoHash     string `storm:"unique"`
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

// MovieTorrentLink handles linking a Movie to a specific file in a torrent
type MovieTorrentLink struct {
	MovieID         int
	TorrentInfoHash string
	FileIndex       int
}

type TorrentMetaRepo interface {
	Store(TorrentMeta) error
	GetByInfoHash(string) (TorrentMeta, error)
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
