package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/http"
	"github.com/diericx/iceetime/internal/app/repos/jackett"
	"github.com/diericx/iceetime/internal/app/repos/storm"
	"github.com/diericx/iceetime/internal/app/services"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

type tomlConfig struct {
	Indexers      []app.Indexer
	Qualities     []app.Quality
	Transcoder    app.TranscoderConfig
	TmdbAPIKey    string
	TorrentClient app.TorrentClientConfig
}

func main() {
	// TODO: input from config file
	var conf tomlConfig
	if _, err := toml.DecodeFile(os.Getenv("CONFIG_FILE"), &conf); err != nil {
		panic(err)
	}

	// TODO: input file location from config file
	stormDB, err := storm.OpenDB(filepath.Join(conf.TorrentClient.TorrentFilePath, ".iceetime.storm.db"))
	defer stormDB.Close()

	client, err := torrent.NewClient(conf.TorrentClient.TorrentFilePath, conf.TorrentClient.TorrentDataPath, 15, 30, 30)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	//
	// Initialize repos
	//
	torrentMetaRepo := storm.TorrentMeta{
		DB: stormDB,
	}

	movieTorrentLinkRepo := storm.MovieTorrentLink{
		DB: stormDB,
	}

	releaseRepo := jackett.ReleaseRepo{
		Qualities: conf.Qualities,
		Indexers:  conf.Indexers,
	}

	//
	// Initialize services
	//
	torrentService := services.Torrent{
		Client:           client,
		TorrentMetaRepo:  &torrentMetaRepo,
		GetInfoTimeout:   time.Second * 15,
		MinSeeders:       conf.TorrentClient.MinSeeders,
		TorrentFilesPath: conf.TorrentClient.TorrentFilePath,
	}
	if err != nil {
		panic(err)
	}
	torrentService.AddTorrentsOnDisk()
	err = torrentService.StartTorrentsAccordingToMetadata()
	if err != nil {
		panic(err)
	}

	releaseService := services.Release{
		ReleaseRepo: releaseRepo,
		Qualities:   conf.Qualities,
	}

	torrentLinkService := services.TorrentLink{
		MovieTorrentLinkRepo: movieTorrentLinkRepo,
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
		TorrentService:     torrentService,
		ReleaseService:     releaseService,
		TorrentLinkService: torrentLinkService,
		Transcoder:         transcoder,
		Qualities:          conf.Qualities,
		TorrentFilesPath:   conf.TorrentClient.TorrentFilePath,
	}

	httpHandler.Serve("secret-todo")
}
