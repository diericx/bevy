package torrent

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type TorrentClient struct {
	torrentFilePath string
	dataPath        string
	infoTimeout     time.Duration
	client          *torrent.Client
}

func NewTorrentClient(torrentFilePath string, dataPath string, infoTimeout int) (*TorrentClient, error) {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = dataPath
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &TorrentClient{
		torrentFilePath: torrentFilePath,
		dataPath:        dataPath,
		infoTimeout:     time.Second * time.Duration(infoTimeout),
		client:          client,
	}, nil
}

func (c *TorrentClient) Close() {
	c.client.Close()
}

// AddFromURLUknownScheme will add the torrent if it is a magnet url, will download a file if it's a
// file or recursicely follow a redirect
func (c *TorrentClient) AddFromURLUknownScheme(rawURL string) (hash string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "magnet" {
		return c.AddFromMagnet(rawURL)
	}

	// Attempt to make http/s call
	req, err := http.NewRequest("GET", rawURL, nil)
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
			return c.AddFromURLUknownScheme(url.String())
		}
		return "", err
	}

	tempFilePath := fmt.Sprintf("%s/%s", c.torrentFilePath, RandomString(10))
	err = downloadFileFromResponse(resp, tempFilePath)
	defer os.Remove(tempFilePath)
	if err != nil {
		return "", err
	}

	return c.AddFromFile(tempFilePath)
}

func downloadFileFromResponse(resp *http.Response, filePath string) error {
	// Get the data
	if resp.StatusCode != 200 {
		return fmt.Errorf("couldn't reach file server with code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *TorrentClient) AddFromMagnet(magnet string) (string, error) {
	t, err := c.client.AddMagnet(magnet)
	if err != nil {
		return "", err
	}
	select {
	case <-t.GotInfo():
	case <-time.After(c.infoTimeout):
		return "", errors.New("info grab timed out")
	}
	// TODO: Start downloading?
	hash := t.InfoHash()
	return hash.HexString(), nil
}

func (c *TorrentClient) AddFromFile(filePath string) (string, error) {
	t, err := c.client.AddTorrentFromFile(filePath)
	if err != nil {
		return "", err
	}

	select {
	case <-t.GotInfo():
	case <-time.After(c.infoTimeout):
		return "", errors.New("info grab timed out")
	}
	// TODO: Start downloading?
	hash := t.InfoHash()
	return hash.HexString(), nil
}

func (c *TorrentClient) AddFromInfoHash(hashString string) error {
	hash := metainfo.Hash{}
	err := hash.FromHexString(hashString)
	if err != nil {
		return err
	}

	t, _ := c.client.AddTorrentInfoHash(hash)
	<-t.GotInfo()

	return nil
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

func (c *TorrentClient) GetReaderForFileInTorrent(hashString string, fileIndex int) (torrent.Reader, error) {
	hash := metainfo.Hash{}
	hash.FromHexString(hashString)

	t, ok := c.client.Torrent(hash)
	if !ok {
		return nil, errors.New("Torrent not found")
	}

	<-t.GotInfo()
	files := t.Files()

	return files[fileIndex].NewReader(), nil
}

func (c *TorrentClient) GetFiles(hashString string) ([]string, error) {
	hash := metainfo.Hash{}
	hash.FromHexString(hashString)

	t, ok := c.client.Torrent(hash)
	if !ok {
		return nil, errors.New("Torrent not found")
	}

	files := t.Files()
	var filePaths []string
	for _, file := range files {
		filePaths = append(filePaths, file.DisplayPath())
	}

	return filePaths, nil
}
