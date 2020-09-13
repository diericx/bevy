package torrent

import (
	"time"

	"github.com/anacrolix/torrent"
)

type Torrent interface {
	BytesCompleted() int64
}

type Client struct {
	*torrent.Client
}

func NewClient(torrentFilePath string, dataPath string, infoTimeout int, establishedConnsPerTorrent int, halfOpenConnsPerTorrent int, timeout time.Duration) (*torrent.Client, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	// config.ListenPort = 42070
	config.EstablishedConnsPerTorrent = establishedConnsPerTorrent
	config.HalfOpenConnsPerTorrent = halfOpenConnsPerTorrent
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Close() {
	c.Close()
}

func (c *Client) AddMagnet(uri string) (Torrent, error) {
	return c.AddMagnet(uri)
}
