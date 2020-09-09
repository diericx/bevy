package torrent

import "github.com/anacrolix/torrent"

func NewTorrentClient(torrentFilePath string, dataPath string, infoTimeout int, establishedConnsPerTorrent int, halfOpenConnsPerTorrent int) (*torrent.Client, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	config.EstablishedConnsPerTorrent = establishedConnsPerTorrent
	config.HalfOpenConnsPerTorrent = halfOpenConnsPerTorrent
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
