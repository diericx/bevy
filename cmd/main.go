package main

import (
	"github.com/diericx/iceetime/internal/pkg/tmdb"
	"log"
	"net/http"
	"os"

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

	torrentClient, err := torrent.NewTorrentClient("~/downloads", "~/downloads")
	if err != nil {
		panic(err)
	}
	defer torrentClient.Close()

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
			panic(err)
		}

		if torrentOnDisk != nil {
			c.JSON(200, torrentOnDisk)
			return
		}

		// Fetch torrent online
		torrent, terr := torznabIQH.QueryMovie(imdbID, title, year, 1)
		if terr != nil {
			log.Println(terr.Error())
			c.JSON(http.StatusNotFound, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
			return
		}
		if torrent == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
			return
		}

		// Add to client to get hash
		if torrent.MagnetLink != "" {
			hash, err := torrentClient.AddFromMagnet(torrent.MagnetLink)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusNotFound, gin.H{
					"error": "We ran into an issue adding the torrent magnet for that movie.",
				})
				return
			}
			torrent.InfoHash = hash
		} else if torrent.FileLink != "" {
			hash, err := torrentClient.AddFromFileURL(torrent.FileLink, torrent.Title)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusNotFound, gin.H{
					"error": "We ran into an issue adding the torrent file for that movie.",
				})
				return
			}
			torrent.InfoHash = hash
		}

		if err := torrentDAO.Save(torrent); err != nil {
			log.Println(err)
			err := torrentClient.RemoveByHash(torrent.InfoHash)
			if err != nil {
				log.Println("BRUTAL: Could not remove torrent after attempting an add. This is super bad!")
			}
			c.JSON(http.StatusNotFound, gin.H{
				"error": "We ran into an issue saving the torrent file for that movie.",
			})
			return
		}

		// foundReleases, _ := releases.Get(imdbID, app.Quality{}) // TODO: manage this error
		c.JSON(200, torrent)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
