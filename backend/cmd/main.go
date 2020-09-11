package main

import (
	"log"
	"time"

	"github.com/diericx/iceetime/internal/pkg/http"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	torrentFilePath := "./downloads"
	torrentDataPath := "./downloads"

	torrentClient, err := torrent.NewTorrentClient(torrentFilePath, torrentDataPath, 15, 30, 30, time.Second*15)
	if err != nil {
		log.Panicf("Error starting torrent client: %s", err)
	}
	defer torrentClient.Close()

	torrentService := torrent.TorrentService{
		Client:           torrentClient,
		Timeout:          time.Second * 15,
		TorrentsLocation: torrentFilePath,
	}
	err = torrentService.LoadTorrentFilesFromCache()
	if err != nil {
		log.Panicf("Error loading torrent files from cache: %s", err)
	}

	httpHandler := http.HTTPHandler{
		TorrentService:  &torrentService,
		TorrentFilePath: torrentFilePath,
	}

	httpHandler.Serve("secret-todo")
}
