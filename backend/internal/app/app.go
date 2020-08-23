package app

import (
	"log"
	"os/exec"
	"time"

	"github.com/anacrolix/torrent"
)

const InvalidTorrentErr string = "invalid torrent"
const InvalidIDErr string = "invalid id"
const IndexerQueryNoResultsErr string = "indexer query gave no results"
const NoValidTorrentsInQueryErr string = "no valid release found from query"
const TorrentByIDNotFoundErr string = "no torrent by that id was found"
const TorrentFileReaderErr string = "unable to get file reader for torrent"
const LocalDBQueryErr string = "unable to query local db"
const LocalDBSaveErr string = "unable to save to local db"

const DefaultResolution = "iw:ih"
const DefaultMaxBitrate = "50M"

func GetSupportedVideoFileFormats() []string {
	return []string{".mkv", ".mp4"}
}

func GetBlacklistedFileNameContents() []string {
	return []string{"sample"}
}

// TODO: These languages are only blacklisted because it's hard to support
func GetBlacklistedTorrentNameContents() []string {
	return []string{"fre", "french", "ita", "italian"}
}

type Error struct {
	OrigionalError error
	Code           int
	Message        string
}

func NewError(origionalError error, code int, message string) *Error {
	log.Printf("%s\n%s", message, origionalError)
	return &Error{
		OrigionalError: origionalError,
		Code:           code,
		Message:        message,
	}
}

func (e Error) Error() string {
	return e.Message
}

type MediaMeta struct {
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
}

type Quality struct {
	Name       string `yaml:"name"`
	Regex      string `yaml:"regex"`
	MinSize    int64  `yaml:"minSize"`
	MaxSize    int64  `yaml:"maxSize"`
	Resolution string `yaml:"resolution"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Indexer struct {
	Name                 string     `yaml:"name"`
	URL                  string     `yaml:"url"`
	BasicAuth            *BasicAuth `yaml:"basicAuth"`
	SupportsImdbIDSearch bool       `yaml:"supportsImdbIDSearch"`
	APIKey               string     `yaml:"apiKey"`
	Categories           string     `yaml:"categories"`
}

// Torrent metadata for a certain torrent
type Torrent struct {
	ID            int    `storm:"id,increment" json:"id"`
	Type          string // Movie, Episode, Season, Season Pack, etc.
	ImdbID        string `storm:"unique" json:"imdbID"`
	Title         string `json:"title"` // Note: Quality is inferred from this
	Size          int64  `json:"size"`
	InfoHash      string `storm:"unique" json:"infoHash"`
	Grabs         int    `json:"grabs"`
	Link          string
	LinkAuth      *BasicAuth // TODO: Encrypt or don't store this?
	Seeders       int        `json:"seeders"` // Note: subject to change
	MainFileIndex int
	Tracker       string
	MinRatio      float32
	MinSeedTime   int
	CreatedAt     time.Time `json:"createdAt"`
}

type Tmdb struct {
	ApiKey string `yaml:"apiKey"`
}

type FFMPEGConfig struct {
	Video struct {
		Format          string `yaml:"format"`
		CompressionAlgo string `yaml:"compressionAlgo"`
	} `yaml:"video"`
	Audio struct {
		CompressionAlgo string `yaml:"compressionAlgo"`
	} `yaml:"audio"`
}

type Config struct {
	Indexers                          []Indexer    `yaml:"indexers"`
	Qualities                         []Quality    `yaml:"qualities"`
	MinSeeders                        int          `yaml:"minSeeders"`
	TorrentInfoTimeout                int          `yaml:"torrentInfoTimeout"`
	TorrentFilePath                   string       `yaml:"torrentFilePath"`
	TorrentDataPath                   string       `yaml:"torrentDataPath"`
	TorrentHalfOpenConnsPerTorrent    int          `yaml:"torrentHalfOpenConnsPerTorrent"`
	TorrentEstablishedConnsPerTorrent int          `yaml:"torrentEstablishedConnsPerTorrent"`
	Tmdb                              Tmdb         `yaml:"tmdb"`
	FFMPEGConfig                      FFMPEGConfig `yaml:"ffmpeg"`
}

type TorrentDAO interface {
	Save(*Torrent) error
	GetByImdbIDAndMinQuality(imdbID string, minQuality int) (*Torrent, error)
	GetByID(id int) (*Torrent, error)
}

type TorrentClient interface {
	AddFromMagnet(magnet string) (hash string, err error)
	AddFromFile(filePath string) (hash string, err error)
	AddFromURLUknownScheme(rawURL string, auth *BasicAuth) (hash string, err error)
	AddFromInfoHash(infoHash string) error
	GetFiles(hash string) (files []string, err error)
	GetReaderForFileInTorrent(hashString string, fileIndex int) (torrent.Reader, error)
	RemoveByHash(hash string) error
}

// IndexerQueryHandler given inputs will handle querying indexers for torrents
type IndexerQueryHandler interface {
	QueryMovie(imdbID string, title string, year string, minQuality int) ([]Torrent, *Error)
}

type Transcoder interface {
	NewTranscodeCommand(input string, time string, resolution string, maxBitrate string, audioStream int, videoStream int) *exec.Cmd
}
