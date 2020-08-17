package main

import (
	"github.com/diericx/iceetime/internal/pkg/tmdb"
	"log"
	"net/http"
	"os"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/storm"
	"github.com/diericx/iceetime/internal/pkg/torznab"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func main() {
	config := app.Config{}
	// Open release manager config file
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Println("Config file not found: config.yaml")
		return
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		log.Println("Invalid config foudn in config.yaml: ", err)
		panic(err)
	}

	torrentDAO, err := storm.NewTorrentDAO("iceetime.db", config.Qualities)
	if err != nil {
		panic(err)
	}
	defer torrentDAO.Close()

	mediaMetaManager := omdb.NewMediaMetaManager(config.Tmdb.ApiKey)
	if err != nil {
		panic(err)
	}

	torznabIQH, _ := torznab.NewIndexerQueryHandler(mediaMetaManager, config.Indexers, config.Qualities) // TODO: handle this error

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

		// Attempt to get a torrent on disk
		torrentOnDisk, err := torrentDAO.GetByImdbIDAndMinQuality(imdbID, 0)
		if err != nil {
			println(err.Error())
			panic(err)
		}

		if torrentOnDisk != nil {
			c.JSON(200, torrentOnDisk)
			return
		}

		// Fetch torrent online
		torrent, err := torznabIQH.QueryMovie(imdbID, title, year, 1)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusNotFound, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
			return
		}
		if torrent == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
		}

		if torrent.MagnetLink != "" {

		}

		// Attempt to save first to catch any duplicate errors that will be caught by unique fields

		// TODO: save torrent
		// TODO: Add torrent to client

		// foundReleases, _ := releases.Get(imdbID, app.Quality{}) // TODO: manage this error
		c.JSON(200, torrent)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
