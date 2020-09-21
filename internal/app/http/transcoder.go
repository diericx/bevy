package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/diericx/iceetime/internal/app"
	"github.com/gin-gonic/gin"
)

func (h *HTTPHandler) addTranscoderGroup(rg *gin.RouterGroup) {
	torrents := rg.Group("/transcoder")
	{
		torrents.GET("/from_url", func(c *gin.Context) {
			url := c.Query("url")
			timeArg := c.Query("time")
			resolution := c.DefaultQuery("res", app.DefaultResolution)
			maxBitrate := c.DefaultQuery("max_bitrate", app.DefaultMaxBitrate)

			// Format time arg
			formattedTimeStr := formatTimeString(timeArg)

			c.Writer.Header().Set("Transfer-Encoding", "chunked") // TODO: Is this necessary? not really sure what it does

			cmdFF := h.Transcoder.NewTranscodeCommand(url, formattedTimeStr, resolution, maxBitrate, 0, 0)
			cmdFF.Stdout = c.Writer
			cmdFF.Start()

			// Start a goroutine to listen for the request being dropped and end transcode if needed
			go func() {
				<-c.Request.Context().Done()
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
			} else {
				log.Println("Error encoding: ", err.Error())
			}
		})
		torrents.GET("/from_url/metadata", func(c *gin.Context) {
			url := c.Query("url")
			out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", url).Output()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   true,
					"message": fmt.Sprintf("Error fetching metadata", err.Error()),
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
					"message": fmt.Sprintf("Error fetching metadata", err.Error()),
				})
				return
			}

			// TODO: get full metadata
			metadata := Metadata{
				Format: Format{
					Duration: int(dur),
				},
			}

			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
			c.Writer.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(c.Writer).Encode(metadata); err != nil {
				panic(err)
			}
			return
		})
	}
}
