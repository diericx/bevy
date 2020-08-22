package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/ffmpeg"
	"github.com/diericx/iceetime/internal/pkg/storm"
	"github.com/diericx/iceetime/internal/pkg/torrent"
	"github.com/diericx/iceetime/internal/pkg/torznab"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

const defaultResolution = "iw:ih"
const defaultMaxBitrate = "50M"

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
	Format Format `json:"format"`
}

func main() {
	configLocation := os.Getenv("CONFIG_FILE")
	dbLocation := os.Getenv("TORRENT_DB_FILE")

	if configLocation == "" {
		log.Println("No config file")
		os.Exit(1)
	}
	if dbLocation == "" {
		log.Println("No db file location specified")
		os.Exit(1)
	}

	config := app.Config{}

	// Open release manager config file
	file, err := os.Open(configLocation)
	if err != nil {
		log.Println("Config file not found: ", configLocation)
		os.Exit(1)
	}
	defer file.Close()

	// Decode config yaml
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		log.Panicf("Invalid yaml in config: %s", err)
	}

	torrentDAO, err := storm.NewTorrentDAO(dbLocation, config.Qualities)
	if err != nil {
		panic(err)
	}
	defer torrentDAO.Close()

	torrentClient, err := torrent.NewTorrentClient(config.TorrentFilePath, config.TorrentDataPath, config.TorrentInfoTimeout, config.TorrentEstablishedConnsPerTorrent, config.TorrentHalfOpenConnsPerTorrent)
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

	// Create main service
	iceetimeService := app.IceetimeService{
		TorrentDAO:          torrentDAO,
		TorrentClient:       torrentClient,
		IndexerQueryHandler: torznabIQH,
		Qualities:           config.Qualities,
		MinSeeders:          config.MinSeeders,
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/find/movie", func(c *gin.Context) {
		imdbID := c.Query("imdbid")
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
		idString := c.Param("id")
		torrentID, err := app.ParseTorrentIdFromString(idString)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error": err.Message,
			})
			return
		}

		err = nil
		torrent, err := iceetimeService.GetTorrentByID(torrentID)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error": err.Message,
			})
			return
		}

		reader, err := iceetimeService.GetFileReaderForFileInTorrent(torrent, torrent.MainFileIndex)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error": err.Message,
			})
			return
		}
		defer reader.Close()

		http.ServeContent(c.Writer, c.Request, torrent.Title, time.Time{}, reader)
	})

	r.GET("/stream/torrent/:id/transcode", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		id := c.Param("id")
		timeArg := c.Query("time")
		resolution := c.DefaultQuery("res", defaultResolution)
		maxBitrate := c.DefaultQuery("max_bitrate", defaultMaxBitrate)

		torrentID, err := app.ParseTorrentIdFromString(id)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error": err.Message,
			})
			return
		}

		streamURL := fmt.Sprintf("%s/%v", "http://127.0.0.1:8080/stream/torrent", torrentID)

		// Stream path
		torrent, err := iceetimeService.GetTorrentByID(torrentID)
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

		// Format time arg
		formattedTimeStr := app.FormatTimeString(timeArg)

		w.Header().Set("Transfer-Encoding", "chunked") // TODO: Is this necessary? not really sure what it does

		// TODO: Change these tracks from 0 default to query values
		cmdFF := ffmpeg.NewTranscodeCommand(streamURL, formattedTimeStr, resolution, maxBitrate, config.FFMPEGConfig, 0, 0)
		cmdFF.Stdout = w
		cmdFF.Start()

		// Start a goroutine to listen for the request being dropped and end transcode if needed
		go func() {
			<-r.Context().Done()
			cmdFF.Process.Kill()
			println("Client Disconnected... Ending process.")
		}()

		// Async execute function
		if err := cmdFF.Wait(); err != nil {
			status := cmdFF.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			if exitStatus != 0 {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Transcoding failed",
				})
				return

			}
		}
	})

	r.GET("/stream/torrent/:id/transcode/metadata", func(c *gin.Context) {
		w := c.Writer

		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID",
			})
			return
		}

		streamURL := fmt.Sprintf("%s/%v", "http://127.0.0.1:8080/stream/torrent", id)

		out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", streamURL).Output()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error fetching metadata",
			})
			return
		}
		durString := string(out)
		durString = durString[:len(durString)-1]

		dur, err := strconv.ParseFloat(durString, 32)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error parsing metadata",
			})
			return
		}
		// TODO: get real metadata
		metadata := Metadata{
			Format: Format{
				Duration: int(dur),
			},
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			panic(err)
		}
		return
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
