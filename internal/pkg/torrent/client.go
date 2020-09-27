package torrent

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type Torrent interface {
	BytesCompleted() int64
	BytesMissing() int64
	Seeding() bool
	GotInfo() <-chan struct{}
	Drop()
	DownloadAll()
	Length() int64
	Metainfo() metainfo.MetaInfo
	InfoHash() metainfo.Hash
	Files() []*torrent.File
	Name() string
	Stats() torrent.TorrentStats
}

type StatCache map[metainfo.Hash]torrent.TorrentStats

type Client struct {
	*torrent.Client
}

func NewClient(torrentFilePath string, dataPath string, infoTimeout int, establishedConnsPerTorrent int, halfOpenConnsPerTorrent int) (*Client, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	config.Seed = true
	// config.ListenPort = 42070
	config.EstablishedConnsPerTorrent = establishedConnsPerTorrent
	config.HalfOpenConnsPerTorrent = halfOpenConnsPerTorrent
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

func (c *Client) Close() {
	c.Client.Close()
}

func (c *Client) AddMagnet(uri string) (Torrent, error) {
	return c.Client.AddMagnet(uri)
}

func (c *Client) AddFile(file string) (Torrent, error) {
	return c.Client.AddTorrentFromFile(file)
}

func (c *Client) Torrents() []Torrent {
	// Simply converting type, must be a better way than this...
	aTs := c.Client.Torrents()
	torrents := make([]Torrent, len(aTs))
	for i, t := range aTs {
		torrents[i] = t
	}
	return torrents
}

func (c *Client) Torrent(h metainfo.Hash) (Torrent, bool) {
	return c.Client.Torrent(h)
}
