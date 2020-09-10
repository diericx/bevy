package torrent

import (
	"time"

	"github.com/anacrolix/torrent"
)

func NewTorrentClient(torrentFilePath string, dataPath string, infoTimeout int, establishedConnsPerTorrent int, halfOpenConnsPerTorrent int, timeout time.Duration) (*Client, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	config.ListenPort = 42070
	config.EstablishedConnsPerTorrent = establishedConnsPerTorrent
	config.HalfOpenConnsPerTorrent = halfOpenConnsPerTorrent
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		timeout: timeout,
		c:       client,
	}, nil
}
