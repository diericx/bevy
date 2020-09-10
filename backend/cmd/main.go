package main

import (
	"log"
	"time"

	"github.com/diericx/iceetime/internal/pkg/http"
	"github.com/diericx/iceetime/internal/service"

	"github.com/diericx/iceetime/internal/pkg/sqlite"
	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	db, err := sqlite.InitSqliteDB("./downloads/torrents.db")
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}

	torrentDAO := sqlite.TorrentDAO{
		Db: db,
	}
	// TODO: Close this db connection??

	torrentClient, err := torrent.NewTorrentClient("./downloads", "./downloads", 15, 30, 30, time.Second*15)
	if err != nil {
		log.Panicf("Error starting torrent client: %s", err)
	}
	defer torrentClient.Close()

	torrentService := service.TorrentService{
		TorrentDAO:    &torrentDAO,
		TorrentClient: torrentClient,
	}

	httpHandler := http.HTTPHandler{
		TorrentService: torrentService,
	}

	httpHandler.Serve()
}
