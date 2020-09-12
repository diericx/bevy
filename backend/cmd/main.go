package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/ffmpeg"
	"github.com/diericx/iceetime/internal/pkg/http"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	// TODO: input from config file
	torrentFilePath := "./downloads"
	torrentDataPath := "./downloads"

	// TODO: input file location from config file
	stormDB, err := storm.Open(filepath.Join(torrentFilePath, ".iceetime.storm.db"))
	defer stormDB.Close()

	torrentClient, err := torrent.NewTorrentClient(torrentFilePath, torrentDataPath, 15, 30, 30, time.Second*15)
	if err != nil {
		log.Panicf("Error starting torrent client: %s", err)
	}
	defer torrentClient.Close()

	torrentService := torrent.TorrentService{
		Client:           torrentClient,
		DB:               stormDB,
		Timeout:          time.Second * 15,
		TorrentsLocation: torrentFilePath,
	}
	err = torrentService.LoadTorrentFilesFromCache()
	if err != nil {
		log.Panicf("Error loading torrent files from cache: %s", err)
	}

	// TODO: Input from config file
	transcoderConfig := app.TranscoderConfig{}
	transcoderConfig.Video.Format = "ismv"
	transcoderConfig.Video.CompressionAlgo = "libx264"
	transcoderConfig.Audio.CompressionAlgo = "copy"
	transcoder := ffmpeg.Transcoder{
		Config: transcoderConfig,
	}

	httpHandler := http.HTTPHandler{
		TorrentService:  &torrentService,
		TorrentFilePath: torrentFilePath,
		Transcoder:      transcoder,
	}

	httpHandler.Serve("secret-todo")
}
