package app

import (
	"io"
	"time"
)

type Error struct {
	OrigionalError error
	Code           int
	Message        string
}

func NewError(origionalError error, code int, message string) *Error {
	return &Error{
		OrigionalError: origionalError,
		Code:           code,
		Message:        message,
	}
}

func (e Error) Error() string {
	return e.Message
}

type Quality struct {
	Name  string `yaml:"name"`
	Regex string `yaml:"regex"`
}
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
type Indexer struct {
	Name       string     `yaml:"name"`
	URL        string     `yaml:"url"`
	BasicAuth  *BasicAuth `yaml:"basicAuth"`
	APIKey     string     `yaml:"apiKey"`
	Categories []string   `yaml:"categories"`
}

type Release struct {
	ID          int    `storm:"id,increment"` // primary key with auto increment
	TorrentHash string `storm:"unique"`       // Don't want to double down on torrents!
	TorrentName string // Store this here to easily infer quality
	ImdbID      string `storm:"index"`
	CreatedAt   time.Time
}

// Torrent metadata for a certain torrent
type Torrent struct {
	ID      string `storm:"id,increment"`
	Name    string // Note: Quality is inferred from this
	Size    int64
	Grabs   int
	Link    string `storm:"unique"`
	Tracker string
}

type Config struct {
	Indexers  []Indexer `yaml:"indexers"`
	Qualities []Quality `yaml:"qualities"`
}

type ReleaseDAO interface {
	Save(Release)
	GetByImdbIDAndMinQuality(imdbID string, minQuality Quality)
}

// ReleaseManager a list of torrents for a certain movie/show
type ReleaseManager interface {
	GetByImdbIDAndQuality(imdbID string, minQuality Quality) (*Release, *Error)
	AddFromTorznabQuery(imdbID string, minQuality Quality) (*Release, *Error)
}

// Torrents manages torrents
type Torrents interface {
	AddByURI(uri string) (ID string, Error *Error)
	Get(ID string) Torrent
	GetReadseeker(ID string) io.ReadSeeker
}

type Transcoder interface{}
