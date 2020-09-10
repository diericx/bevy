package torrent

import (
	"errors"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
)

type Client struct {
	timeout time.Duration
	c       *torrent.Client
}

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

func (client *Client) Close() {
	client.c.Close()
}

func (c *Client) AddFromMagnet(torrent app.Torrent) (app.Torrent, error) {
	t, err := c.c.AddMagnet(torrent.MagnetLink)
	if err != nil {
		return torrent, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(c.timeout):
		t.Drop()
		return torrent, errors.New("info grab timed out")
	}

	hashHexStr := t.InfoHash().HexString()
	info := t.Info()
	torrent.InfoHash = hashHexStr
	torrent.Size = info.Length
	torrent.Title = info.Name

	return torrent, nil
}

func (c *Client) AddFromFile(torrent app.Torrent) (app.Torrent, error) {
	t, err := c.c.AddTorrentFromFile(torrent.File)
	if err != nil {
		return torrent, err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(c.timeout):
		t.Drop()
		return torrent, errors.New("info grab timed out")
	}

	hashHexStr := t.InfoHash().HexString()
	info := t.Info()
	torrent.InfoHash = hashHexStr
	torrent.Size = info.Length
	torrent.Title = info.Name

	return torrent, nil
}

func (c *Client) Start(torrent app.Torrent) error {
	hash := metainfo.NewHashFromHex(torrent.InfoHash)
	t, ok := c.c.Torrent(hash)
	if !ok {
		return errors.New("torrent not found")
	}

	t.DownloadAll()
	return nil
}

func (c *Client) Stats(torrent app.Torrent) (app.TorrentStats, error) {
	hash := metainfo.NewHashFromHex(torrent.InfoHash)
	t, ok := c.c.Torrent(hash)
	if !ok {
		return app.TorrentStats{}, errors.New("torrent not found")
	}

	stats := t.Stats()

	return app.TorrentStats{
		TorrentStats:   stats,
		BytesCompleted: t.BytesCompleted(),
		IsSeeding:      t.Seeding(),
	}, nil

}
