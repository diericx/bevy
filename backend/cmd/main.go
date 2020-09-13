package main

import (
	"path/filepath"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/http"
	"github.com/diericx/iceetime/internal/app/repos/jackett"
	"github.com/diericx/iceetime/internal/app/repos/storm"
	"github.com/diericx/iceetime/internal/app/services"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	// TODO: input from config file
	torrentFilesPath := "./downloads"
	torrentDataPath := "./downloads"
	qualities := []app.Quality{
		app.Quality{
			Name:       "720p",
			Regex:      "720",
			MinSize:    5e8,
			MaxSize:    1e10,
			Resolution: "1280x720",
		},
		app.Quality{
			Name:       "1080p",
			Regex:      "1080",
			MinSize:    5e8,
			MaxSize:    1e10,
			Resolution: "1920x1080",
		},
	}
	indexers := []app.Indexer{
		app.Indexer{
			Name:       "1337x",
			URL:        "http://192.168.1.71:9117/api/v2.0/indexers/1337x/results/torznab",
			APIKey:     "0x7ym4k6c4nghc6nh6qi3s2pdyicxj19",
			Categories: "2000,100002,100004,100001,100054,100042,100070,100055,100003,100076,2010,2020,2030,2040,2045,2050,2060,2070,2080",
		},
	}

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

	releaseRepo := jackett.ReleaseRepo{
		Qualities: qualities,
		Indexers:  indexers,
	}

	torrentService, err := services.NewTorrentService(client, &torrentMetaRepo, time.Second*15, torrentFilesPath)
	if err != nil {
		panic(err)
	}

	releaseService := services.Release{
		ReleaseRepo: releaseRepo,
		Qualities:   qualities,
	}

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
		ReleaseService:   releaseService,
		Transcoder:       transcoder,
		TorrentFilesPath: torrentFilesPath,
	}

	httpHandler.Serve("secret-todo")
}
