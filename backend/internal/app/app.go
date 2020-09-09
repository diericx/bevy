package app

import (
	"gorm.io/gorm"
)

// Torrent metadata
type Torrent struct {
	gorm.Model
	InfoHash    string
	MagnetLink  string
	File        string
	Size        int64
	Title       string
	Grabs       int
	Seeders     int
	MinRatio    float32
	MinSeedTime int
}

// TorrentDAO handles storing Torrent objects
type TorrentDAO interface {
	Store(Torrent) (Torrent, error)
	GetByID(uint) (Torrent, error)
	Get() ([]Torrent, error)
	Remove(uint) error
}

// TorrentService combines a torrent client with local data storage to
// create a highly functional torrent client.
type TorrentService interface {
	Add(torrent Torrent) (Torrent, error)

	// GetByInfoHash(infoHash string) (Torrent, error)
	GetByID(uint) (Torrent, error)
	Get() ([]Torrent, error)

	// Start(Torrent) error
	// Stop(Torrent) error
	// GetFileReader(Torrent, int) error

	// Remove(Torrent) error
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
