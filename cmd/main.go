package main

import (
	"github.com/diericx/iceetime/internal/pkg/tmdb"
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

	torrentDAO, err := storm.NewTorrentDAO("iceetime-torrents.db", config.Qualities)
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
			log.Println("[Torrent search cache hit]")
			c.JSON(200, torrentOnDisk)
			return
		}

		// Fetch torrent online
		torrent, terr := torznabIQH.QueryMovie(imdbID, title, year, 1)
		if terr != nil {
			log.Println(terr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
			return
		}
		if torrent == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "We ran into an issue looking for that movie.",
			})
			return
		}

		// Add to client to get hash
		hash, err := torrentClient.AddFromURLUknownScheme(torrent.Link)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "We ran into an issue adding the torrent magnet for that movie.",
			})
			return
		}
		torrent.InfoHash = hash

		// Save torrent to disk/cache
		if err := torrentDAO.Save(torrent); err != nil {
			log.Println(err)
			err := torrentClient.RemoveByHash(torrent.InfoHash)
			if err != nil {
				log.Println("BRUTAL: Could not remove torrent after attempting an add. This is super bad!")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "We ran into an issue saving the torrent file for that movie.",
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
			log.Println(err)
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

		reader, err := torrentClient.GetReader(torrent.InfoHash)
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
