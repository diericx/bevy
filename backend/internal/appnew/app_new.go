package appnew

import "time"

// Torrent metadata
type Torrent struct {
	ID          int    // fkey
	InfoHash    string `storm:"unique" json:"infoHash"`
	Link        string
	LinkAuth    *BasicAuth // TODO: Encrypt or don't store this?
	Files       []string
	Size        int64
	Title       string
	Grabs       int
	Seeders     int
	Tracker     string
	MinRatio    float32
	MinSeedTime int
	CreatedAt   time.Time `json:"createdAt"`
}

// TorrentDAO handles storing Torrent objects
type TorrentDAO interface {
	Store(Torrent) error
	GetByID(int) (Torrent, error)
	Get() ([]Torrent, error)
	Remove(int) error
}

// TorrentService combines a torrent client with local data storage to
// create a highly functional torrent client.
type TorrentService interface {
	AddFromMagnet(magnet string) (hash string, err error)
	AddFromFile(filePath string) (hash string, err error)
	AddFromInfoHash(infoHash string) error

	GetByInfoHash(infoHash string) (Torrent, error)
	Get() ([]Torrent, error)

	Start(Torrent) error
	Stop(Torrent) error
	GetFileReader(Torrent, int) error

	Remove(Torrent) error
}

// Movie is a simple struct for holding metadata about a movie
type Movie struct {
	ID     int // fkey
	ImdbID string
	Title  string
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
