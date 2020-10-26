package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/http"
	"github.com/diericx/iceetime/internal/app/repos/jackett"
	"github.com/diericx/iceetime/internal/app/repos/storm"
	"github.com/diericx/iceetime/internal/app/repos/tmdb"
	"github.com/diericx/iceetime/internal/app/services"

	"github.com/diericx/iceetime/internal/pkg/torrent"
)

func main() {
	var conf app.MainConfig
	if _, err := toml.DecodeFile(os.Getenv("CONFIG_FILE"), &conf); err != nil {
		panic(err)
	}
	if conf.Validate() != nil {
		fmt.Println("ERROR: Invalid config")
		fmt.Println(conf.Validate())
		os.Exit(1)
	}

	stormDBFilePath := filepath.Join(conf.TorrentClient.TorrentFilePath, ".iceetime.storm.db")
	stormDB, err := storm.OpenDB(stormDBFilePath)
	if err != nil {
		log.Fatalf("Couldn't open torrent file at %s. The file will be created if it doesn't exist, make sure the directory exists and user has proper permissions.", stormDBFilePath)
	}
	defer stormDB.Close()

	client, err := torrent.NewClient(conf.TorrentClient.TorrentFilePath, conf.TorrentClient.TorrentDataPath, 15, 30, 30)
	if err != nil {
		log.Fatalf("Couldn't start torrent client: %s", err.Error())
	}
	defer client.Close()

	//
	// Initialize repos
	//
	tmdbRepo := tmdb.TmdbRepo{
		APIKey: conf.Tmdb.APIKey,
	}

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
	tmdbService := services.Tmdb{
		TmdbRepo: tmdbRepo,
	}

	torrentService := services.NewTorrentService(
		client,
		&torrentMetaRepo,
		time.Second*15,
		conf.TorrentClient.MinSeeders,
		conf.TorrentClient.TorrentFilePath,
	)

	// Add and start (if running) torrents on disk
	torrentService.AddTorrentsOnDisk()
	err = torrentService.StartTorrentsAccordingToMetadata()
	if err != nil {
		log.Fatalf("Error starting existing torrents on disk: %s", err)
	}
	// Start maintinence thread which keeps meta up to date
	go torrentService.StartMetaRefreshForAllTorrentsLoop(conf.TorrentClient.MetaRefreshRate)

	releaseService := services.Release{
		ReleaseRepo: releaseRepo,
		Qualities:   conf.Qualities,
	}

	torrentLinkService := services.TorrentLink{
		MovieTorrentLinkRepo: movieTorrentLinkRepo,
	}

	// TODO: Input from config file
	transcoder := services.Transcoder{
		Config: conf.Transcoder,
	}

	httpHandler := http.HTTPHandler{
		TmdbService:        tmdbService,
		TorrentService:     torrentService,
		ReleaseService:     releaseService,
		TorrentLinkService: torrentLinkService,
		Transcoder:         transcoder,
		TorrentMetaRepo:    torrentMetaRepo,
		Qualities:          conf.Qualities,
		TorrentFilesPath:   conf.TorrentClient.TorrentFilePath,
	}

	httpHandler.Serve("secret-todo")
}
