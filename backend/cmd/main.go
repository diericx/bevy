package main

import (
	"path/filepath"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/http"
	"github.com/diericx/iceetime/internal/app/repos/storm"
	"github.com/diericx/iceetime/internal/app/services"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	// TODO: input from config file
	torrentFilesPath := "./downloads"
	torrentDataPath := "./downloads"

	// TODO: input file location from config file
	stormDB, err := storm.OpenDB(filepath.Join(torrentFilesPath, ".iceetime.storm.db"))
	defer stormDB.Close()

	client, err := torrent.NewClient(torrentFilesPath, torrentDataPath, 15, 30, 30)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	torrentMetaRepo := storm.TorrentMeta{
		DB: stormDB,
	}

	torrentService, err := services.NewTorrentService(client, &torrentMetaRepo, time.Second*15, torrentFilesPath)
	if err != nil {
		panic(err)
	}

	// err = torrentService.LoadTorrentFilesFromCache()
	// if err != nil {
	// 	log.Panicf("Error loading torrent files from cache: %s", err)
	// }

	// TODO: Input from config file
	transcoderConfig := app.TranscoderConfig{}
	transcoderConfig.Video.Format = "ismv"
	transcoderConfig.Video.CompressionAlgo = "libx264"
	transcoderConfig.Audio.CompressionAlgo = "copy"
	transcoder := services.Transcoder{
		Config: transcoderConfig,
	}

	httpHandler := http.HTTPHandler{
		TorrentService:   *torrentService,
		Transcoder:       transcoder,
		TorrentFilesPath: torrentFilesPath,
	}

	httpHandler.Serve("secret-todo")
}
