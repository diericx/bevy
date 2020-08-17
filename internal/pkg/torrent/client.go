package torrent

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"io"
	"net/http"
	"os"
)

type TorrentClient struct {
	torrentFilePath string
	dataPath        string
	client          *torrent.Client
}

func NewTorrentClient(torrentFilePath string, dataPath string) (*TorrentClient, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &TorrentClient{
		torrentFilePath: torrentFilePath,
		dataPath:        dataPath,
		client:          client,
	}, nil
}

func (c *TorrentClient) Close() {
	c.client.Close()
}

func (c *TorrentClient) AddFromMagnet(magnet string) (string, error) {
	t, err := c.client.AddMagnet(magnet)
	if err != nil {
		return "", err
	}
	<-t.GotInfo()
	// TODO: Start downloading?
	hash := t.InfoHash()
	return hash.AsString(), nil
}

func (c *TorrentClient) AddFromFileURL(fileURL string, name string) (string, error) {
	filePath := fmt.Sprintf("%s/%s.torrent", c.torrentFilePath, name)
	// Create file
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		panic(err)
	}
	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("Redirect")
	}

	resp, err := client.Do(req)
	if err != nil {
		if resp.StatusCode == http.StatusFound { //status code 302
			url, err := resp.Location()
			if err != nil {
				return "", err
			}
			// TODO: validate this assumption with url.Scheme
			return c.AddFromMagnet(url.String())
		}

		return "", err
	}

	// Get the data
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("couldn't reach file server with code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Create the file
	out, err = os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return "", err

	println("got here!")
	// Add torrent
	t, err := c.client.AddTorrentFromFile(filePath)
	if err != nil {
		return "", err
	}
	<-t.GotInfo()
	// TODO: remove file

	// TODO: Start downloading?
	hash := t.InfoHash()
	return hash.AsString(), nil
}

func (c *TorrentClient) RemoveByHash(hashString string) error {
	hash := metainfo.Hash{}
	hash.FromHexString(hashString)

	t, ok := c.client.Torrent(hash)
	if !ok {
		return errors.New("Torrent not found")
	}

	t.Drop()

	return nil
}
