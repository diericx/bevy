package app

import (
	"github.com/anacrolix/torrent"
	"gorm.io/gorm"
)

// Torrent metadata
type Torrent struct {
	gorm.Model
	InfoHash       string
	MagnetLink     string
	File           string
	Size           int64
	Title          string
	Grabs          int
	InitialSeeders int
	MinRatio       float32
	MinSeedTime    int
}

type TorrentStats struct {
	torrent.TorrentStats
	BytesCompleted int64
	BytesMissing   int64
	IsSeeding      bool
}

// TorrentDAO handles storing Torrent objects
type TorrentDAO interface {
	Store(Torrent) (Torrent, error)
	GetByID(uint) (Torrent, error)
	Get() ([]Torrent, error)
	Remove(uint) error
}

type TorrentClient interface {
	AddFromMagnet(Torrent) (Torrent, error)
	AddFromFile(Torrent) (Torrent, error)
	Stats(Torrent) (TorrentStats, error)
	Start(Torrent) error

	Close()
}

// Movie is a simple struct for holding metadata about a movie
type Movie struct {
	ID     int // fkey
	ImdbID string
	Title  string
	Year   int
}

// MovieDAO handles storing and retrieving Movies
type MovieDAO interface {
	Store(Movie) error
	GetByID(int) (Movie, error)
	Get() ([]Movie, error)
	Remove(int) error
}

// MovieTorrentLink handles linking a Movie to a specific file in a torrent
type MovieTorrentLink struct {
	MovieID   int
	TorrentID int
	FileIndex int
}

// MovieTorrentLinkDAO handles storing and retrieving MovieTorrentLinks
type MovieTorrentLinkDAO interface {
	Store(MovieTorrentLink) error

	GetByMovieID(int) (MovieTorrentLink, error)
	GetByTorrentID(int) (MovieTorrentLink, error)
	Get() ([]Movie, error)

	Remove(int) error
}
