package main

import (
	"log"
	"os"

	"github.com/diericx/iceetime/internal/pkg/http"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/ffmpeg"
	"github.com/diericx/iceetime/internal/pkg/storm"
	"github.com/diericx/iceetime/internal/pkg/torrent"
	"github.com/diericx/iceetime/internal/pkg/torznab"
	"gopkg.in/yaml.v2"
)

func main() {
	configLocation := os.Getenv("CONFIG_FILE")
	dbLocation := os.Getenv("TORRENT_DB_FILE")
	logFileLocation := os.Getenv("LOG_FILE")

	f, err := os.OpenFile(logFileLocation,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	logger := log.New(f, "prefix", log.LstdFlags)

	if configLocation == "" {
		logger.Println("No config file")
		os.Exit(1)
	}
	if dbLocation == "" {
		logger.Println("No db file location specified")
		os.Exit(1)
	}

	config := app.Config{}

	// Open release manager config file
	file, err := os.Open(configLocation)
	if err != nil {
		logger.Println("Config file not found: ", configLocation)
		os.Exit(1)
	}
	defer file.Close()

	// Decode config yaml
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		logger.Panicf("Invalid yaml in config: %s", err)
	}

	torrentDAO, err := storm.NewTorrentDAO(dbLocation, config.Qualities)
	if err != nil {
		logger.Panicf("Error starting torrent db access object: %s", err)
	}
	defer torrentDAO.Close()

	torrentClient, err := torrent.NewTorrentClient(config.TorrentFilePath, config.TorrentDataPath, config.TorrentInfoTimeout, config.TorrentEstablishedConnsPerTorrent, config.TorrentHalfOpenConnsPerTorrent)
	if err != nil {
		logger.Panicf("Error starting torrent client: %s", err)
	}
	defer torrentClient.Close()

	torznabIQH, _ := torznab.NewIndexerQueryHandler(config.Indexers, config.Qualities) // TODO: handle this error

	// Add all torrents from disk
	allTorrents, err := torrentDAO.All()
	if err != nil {
		logger.Panicf("Error starting indexer query handler: %s", err)
	}
	for _, t := range allTorrents {
		err := torrentClient.AddFromInfoHash(t.InfoHash)
		if err != nil {
			logger.Panicf("Error adding torrent from state on disk \nTitle: %s\nError: %s", t.Title, err)
		}
	}

	ffmpegTranscoder := ffmpeg.Transcoder{
		Config: config.TranscoderConfig,
	}

	// Create main service
	iceetimeService := app.IceetimeService{
		TorrentDAO:          torrentDAO,
		TorrentClient:       torrentClient,
		IndexerQueryHandler: torznabIQH,
		Qualities:           config.Qualities,
		MinSeeders:          config.MinSeeders,
		Transcoder:          ffmpegTranscoder,
	}

	httpHandler := http.HTTPHandler{
		Logger:          logger,
		IceetimeService: iceetimeService,
	}

	httpHandler.Serve()
}
