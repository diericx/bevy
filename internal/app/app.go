package app

import (
	"encoding/xml"
	"io"
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

type Quality struct {
	Index         int
	RegexPatterns []string
}

type Indexer struct {
	Name       string   `yaml:"name"`
	URL        string   `yaml:"url"`
	APIKey     string   `yaml:"apiKey"`
	Categories []string `yaml:"categories"`
}

type Release struct {
	Quality   Quality
	TorrentID string
	ImdbID    string
}

// Torrent metadata for a certain torrent
type Torrent struct {
	ID      string
	Tracker string
}

type ReleaseManagerConfig struct {
	Indexers []Indexer `yaml:"indexers"`
}

type Rss struct {
	XMLName  xml.Name  `xml:"rss"`
	Channels []Channel `xml:"channel"`
}
type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Items   []Item   `xml:"item"`
}
type Item struct {
	XMLName xml.Name `xml:"item"`
	Title   string   `xml:"title"`
	Size    int64    `xml:"size"`
	Grabs   int      `xml:"grabs"`
	Link    string   `xml:"link"`
	// TODO: Torznab attributes??
}

// ReleaseManager a list of torrents for a certain movie/show
type ReleaseManager interface {
	Get(imdbID string, minQuality Quality) (*Release, *Error)
	Add(imdbID string, minQuality Quality) (*Release, *Error)
}

// Torrents manages torrents
type Torrents interface {
	AddByURI(uri string) (ID string, Error *Error)
	Get(ID string) Torrent
	GetReadseeker(ID string) io.ReadSeeker
}

type Transcoder interface{}
