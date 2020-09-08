package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/diericx/iceetime/internal/app"
	"github.com/gin-gonic/gin"
)

func addStreamRoutes(rg *gin.RouterGroup, s app.IceetimeService) {
	stream := rg.Group("/streams")

	stream.GET("/torrent/:id", func(c *gin.Context) {
		idString := c.Param("id")
		torrentID, err := app.ParseTorrentIdFromString(idString)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error":   true,
				"message": err.Message,
			})
			return
		}

		err = nil
		torrent, err := s.GetTorrentByID(torrentID)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error":   true,
				"message": err.Message,
			})
			return
		}

		reader, err := s.GetFileReaderForFileInTorrent(torrent, torrent.MainFileIndex)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error":   true,
				"message": err.Message,
			})
			return
		}
		defer reader.Close()

		http.ServeContent(c.Writer, c.Request, torrent.Title, time.Time{}, reader)
	})

	stream.GET("/torrent/:id/transcode", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		id := c.Param("id")
		timeArg := c.Query("time")
		resolution := c.DefaultQuery("res", app.DefaultResolution)
		maxBitrate := c.DefaultQuery("max_bitrate", app.DefaultMaxBitrate)

		torrentID, err := app.ParseTorrentIdFromString(id)
		if err != nil {
			c.JSON(err.Code, gin.H{
				"error":   true,
				"message": err.Message,
			})
			return
		}

		streamURL := fmt.Sprintf("%s/%v", "http://127.0.0.1:8080/stream/torrent", torrentID)

		// Stream path
		torrent, err := s.GetTorrentByID(torrentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Error searching for your torrent",
			})
			return
		}
		if torrent == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Torrent not found on disk",
			})
			return
		}

		// Format time arg
		formattedTimeStr := app.FormatTimeString(timeArg)

		w.Header().Set("Transfer-Encoding", "chunked") // TODO: Is this necessary? not really sure what it does

		// TODO: Change these tracks from 0 default to query values
		cmdFF := s.Transcoder.NewTranscodeCommand(streamURL, formattedTimeStr, resolution, maxBitrate, 0, 0)
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
					"error":   true,
					"message": "Transcoding failed",
				})
				return

			}
		}
	})

	stream.GET("/stream/torrent/:id/transcode/metadata", func(c *gin.Context) {
		w := c.Writer

		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Invalid ID",
			})
			return
		}

		streamURL := fmt.Sprintf("%s/%v", "http://127.0.0.1:8080/stream/torrent", id)

		out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", streamURL).Output()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Error fetching metadata",
			})
			return
		}
		durString := string(out)
		durString = durString[:len(durString)-1]

		dur, err := strconv.ParseFloat(durString, 32)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Error parsing metadata",
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

}
