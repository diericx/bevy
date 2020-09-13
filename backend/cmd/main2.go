package main

import (
	"time"

	"github.com/diericx/iceetime/internal/app/services"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	client := torrent.NewClient("./downloads", "./downloads", 15, 30, 30, time.Second*15)
	defer client.Close()

	torrentService := services.Torrent{
		TorrentClient: client,
	}
}
