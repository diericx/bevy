package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/storm"
	"github.com/diericx/iceetime/internal/pkg/torrent"
	"github.com/diericx/iceetime/internal/pkg/torznab"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func main() {
	config := app.Config{}
	// Open release manager config file
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Panicf("Config file not found: config.yaml: %s", err)
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		log.Panicf("Invalid yaml in config: %s", err)
	}

	torrentDAO, err := storm.NewTorrentDAO("iceetime-torrents.db", config.Qualities)
	if err != nil {
		panic(err)
	}
	defer torrentDAO.Close()

	torrentClient, err := torrent.NewTorrentClient(config.TorrentFilePath, config.TorrentDataPath, config.TorrentInfoTimeout)
	if err != nil {
		panic(err)
	}
	defer torrentClient.Close()

	torznabIQH, _ := torznab.NewIndexerQueryHandler(config.Indexers, config.Qualities) // TODO: handle this error

	// Add all torrents from disk
	allTorrents, err := torrentDAO.All()
	if err != nil {
		panic(err)
	}
	for _, t := range allTorrents {
		err := torrentClient.AddFromInfoHash(t.InfoHash)
		if err != nil {
			log.Panicf("Error adding torrent from state on disk \nTitle: %s\nError: %s", t.Title, err)
		}
	}

	iceetimeService := app.IceetimeService{
		TorrentDAO:          torrentDAO,
		TorrentClient:       torrentClient,
		IndexerQueryHandler: torznabIQH,
		Qualities:           config.Qualities,
		MinSeeders:          config.MinSeeders,
	}

	r := gin.Default()
	r.GET("/find/movie", func(c *gin.Context) {
		imdbID := c.Query("imdbID")
		title := c.Query("title")
		year := c.Query("year")
		if imdbID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing imdb id",
			})
			return
		}
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing title",
			})
			return
		}
		if year == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing year",
			})
			return
		}

		torrent, err := iceetimeService.FindLocallyOrFetchMovie(imdbID, title, year, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Message,
			})
			return
		}

		c.JSON(200, torrent)
	})

	r.GET("/stream/torrent/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID",
			})
		}

		torrent, err := torrentDAO.GetByID(int(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error searching for your torrent",
			})
			return
		}
		if torrent == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Torrent not found on disk",
			})
			return
		}

		reader, err := torrentClient.GetReaderForFileInTorrent(torrent.InfoHash, torrent.MainFileIndex)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Torrent not found in client",
			})
			return
		}
		defer reader.Close()

		http.ServeContent(c.Writer, c.Request, torrent.Title, time.Time{}, reader)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
