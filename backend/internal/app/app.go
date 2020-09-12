package app

import (
	"io"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// TODO: input from config file
const DefaultResolution = "iw:ih"
const DefaultMaxBitrate = "50M"

// Torrent metadata
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

type TorrentService interface {
	AddFromMagnet(string) (*Torrent, error)
	AddFromFile(string) (*Torrent, error)
	LoadTorrentFilesFromCache() error
	GetByHashStr(string) (*Torrent, error)
	Get() ([]Torrent, error)
	GetReadSeekerForFileInTorrent(*Torrent, int) (io.ReadSeeker, error)

	// DownloadAll(*Torrent) error
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
	ID     int // fkey
	ImdbID string
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
