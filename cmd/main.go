package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/storm"
	"github.com/diericx/iceetime/internal/pkg/torrent"
	"github.com/diericx/iceetime/internal/pkg/torznab"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
	Format Format `json:"format"`
}

func main() {
	config := app.Config{}

	// Open release manager config file
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Panicf("Config file not found: config.yaml: %s", err)
	}
	defer file.Close()

	// Decode config yaml
	d := yaml.NewDecoder(file)
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

	// Create main service
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

	r.GET("/stream/torrent/:id/transcode", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID",
			})
			return
		}

		streamURL := fmt.Sprintf("%s/%v", "http://127.0.0.1:8080/stream/torrent", id)

		// Stream path
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

		// Format time arg
		timeArg := c.Query("time")
		var formattedTimeString string
		timeFloat, err := strconv.ParseFloat(timeArg, 64)
		if err != nil {
			timeFloat = 0
		}
		timeInt := int(math.Round(timeFloat))
		timeDuration := time.Second * time.Duration(timeInt)
		formattedTimeString = fmtDuration(timeDuration)

		w.Header().Set("Transfer-Encoding", "chunked") // TODO: Is this necessary? not really sure what it does

		cmdFF := newFFMPEGTranscodeCommand(streamURL, formattedTimeString, config.Qualities[0].Resolution, config.FFMPEGConfig)
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
			signaled := status.Signaled()
			signal := status.Signal()
			log.Println(err, exitStatus, signaled, signal)
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

func newFFMPEGTranscodeCommand(input string, time string, resolution string, c app.FFMPEGConfig) *exec.Cmd {
	// Note: -ss flag needs to come before -i in order to skip encoding the entire first section
	ffmpegArgs := []string{
		"-ss", time,
		"-i", input,
		"-f", c.Video.Format,
		"-c:v", c.Video.CompressionAlgo,
		"-c:a", c.Audio.CompressionAlgo,
		"-b", "2000k",
		"-vf", fmt.Sprintf("scale=%s", resolution),
		"-threads", "0",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		// "-movflags", "frag_keyframe+empty_moov", // This was to allow mp4 encoding.. not sure what it implies
	}

	ffmpegArgs = append(ffmpegArgs, "-")

	cmdFF := exec.Command("ffmpeg", ffmpegArgs...)
	return cmdFF
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= h * time.Minute
	s := m / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
